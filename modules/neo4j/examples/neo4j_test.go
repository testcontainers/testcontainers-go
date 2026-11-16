package neo4j_test

import (
	"context"
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
}

func setupNeo4j(ctx context.Context) (*neo4j.Neo4jContainer, error) {
	return neo4j.Run(ctx,
		"neo4j:4.4",
		neo4j.WithAdminPassword(testPassword),
		neo4j.WithLabsPlugin(neo4j.Apoc),
		neo4j.WithNeo4jSetting("dbms.tx_log.rotation.size", "42M"),
	)
}

func createDriver(t *testing.T, ctx context.Context, container *neo4j.Neo4jContainer) neo.DriverWithContext {
	t.Helper()
	boltURL, err := container.BoltUrl(ctx)
	require.NoError(t, err)
	driver, err := neo.NewDriverWithContext(boltURL, neo.BasicAuth("neo4j", testPassword, ""))
	require.NoError(t, err)
	t.Cleanup(func() {
		err := driver.Close(ctx)
		require.NoErrorf(t, err, "failed to close neo: %s", err)
	})
	return driver
}
