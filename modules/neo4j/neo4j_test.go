package neo4j_test

import (
	"context"
	"fmt"
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
		require.NoErrorf(t, err, "should have successfully connected to server but did not")
	})

	outer.Run("exercises APOC plugin", func(t *testing.T) {
		driver := createDriver(t, ctx, ctr)

		result, err := neo.ExecuteQuery(ctx, driver,
			"RETURN apoc.number.arabicToRoman(1986) AS output", nil,
			neo.EagerResultTransformer)
		require.NoErrorf(t, err, "expected APOC query to successfully run but did not")
		require.NotEmpty(t, result.Records)
		value, _ := result.Records[0].Get("output")
		require.Equalf(t, "MCMLXXXVI", value, "did not get expected roman number: %s", value)
	})

	outer.Run("is configured with custom Neo4j settings", func(t *testing.T) {
		env := getContainerEnv(t, ctx, ctr)

		require.Containsf(t, env, "NEO4J_dbms_tx__log_rotation_size=42M", "expected to custom setting to be exported but was not")
	})
}

func TestNeo4jWithEnterpriseLicense(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	images := map[string]string{
		"StandardEdition":   "neo4j:4.4",
		"EnterpriseEdition": "neo4j:4.4-enterprise",
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

			require.Containsf(t, env, "NEO4J_ACCEPT_LICENSE_AGREEMENT=yes", "expected to accept license agreement but did not")
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
		require.Errorf(t, err, "expected env to fail due to conflicting auth settings but did not")
		require.Nilf(t, ctr, "container must not be created with conflicting auth settings")
	})

	outer.Run("warns about overwrites of setting keys", func(t *testing.T) {
		// withSettings {
		logger := &inMemoryLogger{}
		ctr, err := neo4j.Run(ctx,
			"neo4j:4.4",
			testcontainers.WithLogger(logger), // needs to go before WithNeo4jSetting and WithNeo4jSettings
			neo4j.WithAdminPassword(testPassword),
			neo4j.WithNeo4jSetting("some.key", "value1"),
			neo4j.WithNeo4jSettings(map[string]string{"some.key": "value2"}),
			neo4j.WithNeo4jSetting("some.key", "value3"),
		)
		// }
		testcontainers.CleanupContainer(t, ctr)
		require.NoError(t, err)

		errorLogs := logger.Logs()
		require.Containsf(t, errorLogs, `setting "some.key" with value "value1" is now overwritten with value "value2"`+"\n", "expected setting overwrites to be logged")
		require.Containsf(t, errorLogs, `setting "some.key" with value "value2" is now overwritten with value "value3"`+"\n", "expected setting overwrites to be logged")
		require.Containsf(t, getContainerEnv(t, ctx, ctr), "NEO4J_some_key=value3", "expected custom setting to be set with last value")
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
	t.Helper()
	// boltURL {
	boltURL, err := container.BoltUrl(ctx)
	// }
	require.NoError(t, err)
	driver, err := neo.NewDriverWithContext(boltURL, neo.BasicAuth("neo4j", testPassword, ""))
	require.NoError(t, err)
	t.Cleanup(func() {
		err := driver.Close(ctx)
		require.NoErrorf(t, err, "failed to close neo: %s", err)
	})
	return driver
}

func getContainerEnv(t *testing.T, ctx context.Context, container *neo4j.Neo4jContainer) string {
	t.Helper()
	return testcontainers.RequireContainerExec(ctx, t, container, []string{"env"})
}

const logSeparator = "---$$$---"

type inMemoryLogger struct {
	buffer strings.Builder
}

func (iml *inMemoryLogger) Printf(msg string, args ...any) {
	iml.buffer.Write([]byte(fmt.Sprintf(msg, args...)))
	iml.buffer.Write([]byte(logSeparator))
}

func (iml *inMemoryLogger) Logs() []string {
	return strings.Split(iml.buffer.String(), logSeparator)
}
