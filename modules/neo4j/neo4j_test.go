package neo4j_test

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"

	neo "github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/testcontainers/testcontainers-go/modules/neo4j"
)

const testPassword = "letmein!"

func TestNeo4j(outer *testing.T) {
	outer.Parallel()

	ctx := context.Background()

	container := setupNeo4j(ctx, outer)

	outer.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			outer.Fatalf("failed to terminate container: %s", err)
		}
	})

	outer.Run("connects via Bolt", func(t *testing.T) {
		driver := createDriver(t, ctx, container)

		err := driver.VerifyConnectivity(ctx)

		if err != nil {
			t.Fatalf("should have successfully connected to server but did not: %s", err)
		}
	})

	outer.Run("exercises APOC plugin", func(t *testing.T) {
		driver := createDriver(t, ctx, container)

		result, err := neo.ExecuteQuery(ctx, driver,
			"RETURN apoc.number.arabicToRoman(1986) AS output", nil,
			neo.EagerResultTransformer)

		if err != nil {
			t.Fatalf("expected APOC query to successfully run but did not: %s", err)
		}
		if value, _ := result.Records[0].Get("output"); value != "MCMLXXXVI" {
			t.Fatalf("did not get expected roman number: %s", value)
		}
	})

	outer.Run("is configured with custom Neo4j settings", func(t *testing.T) {
		env := getContainerEnv(t, ctx, container)

		if !strings.Contains(env, "NEO4J_dbms_tx__log_rotation_size=42M") {
			t.Fatal("expected to custom setting to be exported but was not")
		}
	})
}

func TestNeo4jWithWrongSettings(outer *testing.T) {
	outer.Parallel()

	ctx := context.Background()

	outer.Run("without authentication", func(t *testing.T) {
		container, err := neo4j.RunContainer(ctx)
		if err != nil {
			t.Fatalf("expected env to successfully run but did not: %s", err)
		}
		t.Cleanup(func() {
			if err := container.Terminate(ctx); err != nil {
				outer.Fatalf("failed to terminate container: %s", err)
			}
		})
	})

	outer.Run("ignores auth setting outside WithAdminPassword", func(t *testing.T) {
		container, err := neo4j.RunContainer(ctx,
			neo4j.WithAdminPassword(testPassword),
			neo4j.WithNeo4jSetting("AUTH", "neo4j/thisisgonnabeignored"),
		)
		if err != nil {
			t.Fatalf("expected env to successfully run but did not: %s", err)
		}
		t.Cleanup(func() {
			if err := container.Terminate(ctx); err != nil {
				outer.Fatalf("failed to terminate container: %s", err)
			}
		})

		env := getContainerEnv(t, ctx, container)

		if !strings.Contains(env, "NEO4J_AUTH=neo4j/"+testPassword) {
			t.Fatalf("expected WithAdminPassword to have higher precedence than auth set with WithNeo4jSetting")
		}
	})

	outer.Run("warns about overwrites of setting keys", func(t *testing.T) {
		logger := &inMemoryLogger{}
		container, err := neo4j.RunContainer(ctx,
			neo4j.WithLogger(logger), // needs to go before WithNeo4jSetting and WithNeo4jSettings
			neo4j.WithAdminPassword(testPassword),
			neo4j.WithNeo4jSetting("some.key", "value1"),
			neo4j.WithNeo4jSettings(map[string]string{"some.key": "value2"}),
			neo4j.WithNeo4jSetting("some.key", "value3"),
		)
		if err != nil {
			t.Fatalf("expected env to successfully run but did not: %s", err)
		}
		t.Cleanup(func() {
			if err := container.Terminate(ctx); err != nil {
				outer.Fatalf("failed to terminate container: %s", err)
			}
		})

		errorLogs := logger.Logs()
		if !Contains(errorLogs, `setting "some.key" with value "value1" is now overwritten with value "value2"`+"\n") ||
			!Contains(errorLogs, `setting "some.key" with value "value2" is now overwritten with value "value3"`+"\n") {
			t.Fatalf("expected setting overwrites to be logged")
		}
		if !strings.Contains(getContainerEnv(t, ctx, container), "NEO4J_some_key=value3") {
			t.Fatalf("expected custom setting to be set with last value")
		}
	})

	outer.Run("rejects nil logger", func(t *testing.T) {
		container, err := neo4j.RunContainer(ctx, neo4j.WithLogger(nil))

		if container != nil {
			t.Fatalf("container must not be created with nil logger")
		}
		if err == nil || err.Error() != "nil logger is not permitted" {
			t.Fatalf("expected config validation error but got no error")
		}
	})
}

func setupNeo4j(ctx context.Context, t *testing.T) *neo4j.Neo4jContainer {
	// neo4jCreateContainer {
	container, err := neo4j.RunContainer(ctx,
		neo4j.WithAdminPassword(testPassword),
		neo4j.WithLabsPlugin(neo4j.Apoc),
		neo4j.WithNeo4jSetting("dbms.tx_log.rotation.size", "42M"),
	)
	// }
	if err != nil {
		t.Fatalf("expected container to successfully initialize but did not: %s", err)
	}
	return container
}

func createDriver(t *testing.T, ctx context.Context, container *neo4j.Neo4jContainer) neo.DriverWithContext {
	boltUrl, err := container.BoltUrl(ctx)
	if err != nil {
		t.Fatal(err)
	}
	driver, err := neo.NewDriverWithContext(boltUrl, neo.BasicAuth("neo4j", testPassword, ""))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := driver.Close(ctx); err != nil {
			t.Fatalf("failed to close neo: %s", err)
		}
	})
	return driver
}

func getContainerEnv(t *testing.T, ctx context.Context, container *neo4j.Neo4jContainer) string {
	exec, reader, err := container.Exec(ctx, []string{"env"})

	if err != nil {
		t.Fatalf("expected env to successfully run but did not: %s", err)
	}
	if exec != 0 {
		t.Fatalf("expected env to exit with status 0 but exited with: %d", exec)
	}
	envVars, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("expected to read all bytes from env output but did not: %s", err)
	}
	return string(envVars)
}

const logSeparator = "---$$$---"

type inMemoryLogger struct {
	buffer strings.Builder
}

func (iml *inMemoryLogger) Printf(msg string, args ...interface{}) {
	iml.buffer.Write([]byte(fmt.Sprintf(msg, args...)))
	iml.buffer.Write([]byte(logSeparator))
}

func (iml *inMemoryLogger) Logs() []string {
	return strings.Split(iml.buffer.String(), logSeparator)
}
