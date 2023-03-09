package neo4j_test

import (
	"context"
	neo "github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/testcontainers/testcontainers-go/modules/neo4j"
	"testing"
)

const testPassword = "letmein!"

func TestNeo4j(outer *testing.T) {
	ctx := context.Background()

	container, err := setupNeo4j(ctx)
	if err != nil {
		outer.Fatal(err)
	}

	outer.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			outer.Fatalf("failed to terminate container: %s", err)
		}
	})

	outer.Run("connects via Bolt", func(t *testing.T) {
		driver := createDriver(t, ctx, container)

		err = driver.VerifyConnectivity(ctx)

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

}

func setupNeo4j(ctx context.Context) (*neo4j.Neo4jContainer, error) {
	return neo4j.StartContainer(ctx,
		neo4j.WithAdminPassword(testPassword),
		neo4j.WithLabsPlugin(neo4j.Apoc),
	)
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
