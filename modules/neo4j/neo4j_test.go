package neo4j_test

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"

	neo "github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/neo4j"
)

const testPassword = "letmein!"

func TestNeo4j(outer *testing.T) {
	outer.Parallel()

	ctx := context.Background()

	ctr, err := setupNeo4j(ctx)
	testcontainers.CleanupContainer(outer, ctr)
	require.NoError(outer, err)

	outer.Run("connects via Bolt", func(t *testing.T) {
		driver := createDriver(t, ctx, ctr)

		err := driver.VerifyConnectivity(ctx)
		if err != nil {
			t.Fatalf("should have successfully connected to server but did not: %s", err)
		}
	})

	outer.Run("exercises APOC plugin", func(t *testing.T) {
		driver := createDriver(t, ctx, ctr)

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
		env := getContainerEnv(t, ctx, ctr)

		if !strings.Contains(env, "NEO4J_dbms_tx__log_rotation_size=42M") {
			t.Fatal("expected to custom setting to be exported but was not")
		}
	})
}

func TestNeo4jWithEnterpriseLicense(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	images := map[string]string{
		"StandardEdition":   "docker.io/neo4j:4.4",
		"EnterpriseEdition": "docker.io/neo4j:4.4-enterprise",
	}

	for edition, img := range images {
		edition, img := edition, img
		t.Run(edition, func(t *testing.T) {
			t.Parallel()
			ctr, err := neo4j.Run(ctx,
				img,
				neo4j.WithAdminPassword(testPassword),
				neo4j.WithAcceptCommercialLicenseAgreement(),
			)
			testcontainers.CleanupContainer(t, ctr)
			require.NoError(t, err)

			env := getContainerEnv(t, ctx, ctr)

			if !strings.Contains(env, "NEO4J_ACCEPT_LICENSE_AGREEMENT=yes") {
				t.Fatal("expected to accept license agreement but did not")
			}
		})
	}
}

func TestNeo4jWithWrongSettings(outer *testing.T) {
	outer.Parallel()

	ctx := context.Background()

	outer.Run("without authentication", func(t *testing.T) {
		ctr, err := neo4j.Run(ctx, "neo4j:4.4")
		testcontainers.CleanupContainer(t, ctr)
		require.NoError(t, err)
	})

	outer.Run("auth setting outside WithAdminPassword raises error", func(t *testing.T) {
		ctr, err := neo4j.Run(ctx,
			"neo4j:4.4",
			neo4j.WithAdminPassword(testPassword),
			neo4j.WithNeo4jSetting("AUTH", "neo4j/thisisgonnafail"),
		)
		testcontainers.CleanupContainer(t, ctr)
		if err == nil {
			t.Fatalf("expected env to fail due to conflicting auth settings but did not")
		}
		if ctr != nil {
			t.Fatalf("container must not be created with conflicting auth settings")
		}
	})

	outer.Run("warns about overwrites of setting keys", func(t *testing.T) {
		// withSettings {
		logger := &inMemoryLogger{}
		ctr, err := neo4j.Run(ctx,
			"neo4j:4.4",
			neo4j.WithLogger(logger), // needs to go before WithNeo4jSetting and WithNeo4jSettings
			neo4j.WithAdminPassword(testPassword),
			neo4j.WithNeo4jSetting("some.key", "value1"),
			neo4j.WithNeo4jSettings(map[string]string{"some.key": "value2"}),
			neo4j.WithNeo4jSetting("some.key", "value3"),
		)
		// }
		testcontainers.CleanupContainer(t, ctr)
		require.NoError(t, err)

		errorLogs := logger.Logs()
		if !Contains(errorLogs, `setting "some.key" with value "value1" is now overwritten with value "value2"`+"\n") ||
			!Contains(errorLogs, `setting "some.key" with value "value2" is now overwritten with value "value3"`+"\n") {
			t.Fatalf("expected setting overwrites to be logged")
		}
		if !strings.Contains(getContainerEnv(t, ctx, ctr), "NEO4J_some_key=value3") {
			t.Fatalf("expected custom setting to be set with last value")
		}
	})

	outer.Run("rejects nil logger", func(t *testing.T) {
		ctr, err := neo4j.Run(ctx, "neo4j:4.4", neo4j.WithLogger(nil))
		testcontainers.CleanupContainer(t, ctr)
		if ctr != nil {
			t.Fatalf("container must not be created with nil logger")
		}
		if err == nil || err.Error() != "nil logger is not permitted" {
			t.Fatalf("expected config validation error but got no error")
		}
	})
}

func setupNeo4j(ctx context.Context) (*neo4j.Neo4jContainer, error) {
	return neo4j.Run(ctx,
		"neo4j:4.4",
		neo4j.WithAdminPassword(testPassword),
		// withLabsPlugin {
		neo4j.WithLabsPlugin(neo4j.Apoc),
		// }
		neo4j.WithNeo4jSetting("dbms.tx_log.rotation.size", "42M"),
	)
}

func createDriver(t *testing.T, ctx context.Context, container *neo4j.Neo4jContainer) neo.DriverWithContext {
	// boltURL {
	boltUrl, err := container.BoltUrl(ctx)
	// }
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
