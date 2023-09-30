package influxdb_test

import (
	"context"
	"encoding/json"
	"fmt"
	db1 "github.com/influxdata/influxdb1-client/v2"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/influxdb"
	"testing"
)

func containerCleanup(t *testing.T, container testcontainers.Container) {
	if err := container.Terminate(context.Background()); err != nil {
		t.Fatalf("failed to terminate container: %s", err)
	}
}

func TestV1Container(t *testing.T) {
	ctx := context.Background()
	influxDbContainer, err := influxdb.RunContainer(ctx,
		testcontainers.WithImage("influxdb:1.8.10"),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer containerCleanup(t, influxDbContainer)

	state, err := influxDbContainer.State(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if state.Running {
		fmt.Println("InfluxDB container is running")
	} else {
		t.Fatal("InfluxDB container is not running")
	}
}

func TestV2Container(t *testing.T) {
	ctx := context.Background()
	influxDbContainer, err := influxdb.RunContainer(ctx,
		testcontainers.WithImage("influxdb:latest"),
		influxdb.WithDatabase("foo"),
		influxdb.WithUsername("root"),
		influxdb.WithPassword("password"),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer containerCleanup(t, influxDbContainer)

	state, err := influxDbContainer.State(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if state.Running {
		fmt.Println("InfluxDB container is running")
	} else {
		t.Fatal("InfluxDB container is not running")
	}
}

func TestWithInitDb(t *testing.T) {
	ctx := context.Background()
	influxDbContainer, err := influxdb.RunContainer(ctx,
		testcontainers.WithImage("influxdb:1.8.10"),
		influxdb.WithInitDb("."),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer containerCleanup(t, influxDbContainer)
	if state, err := influxDbContainer.State(ctx); err != nil || !state.Running {
		if err != nil {
			t.Fatal(err)
		}
		t.Fatal("InfluxDB container is not running")
	}
	fmt.Println("InfluxDB container is running")
	influxClient, err := db1.NewHTTPClient(db1.HTTPConfig{
		Addr: influxDbContainer.MustHaveConnectionUrl(ctx, false),
	})
	if err != nil {
		t.Fatal(err)
	}
	defer influxClient.Close()

	var (
		expected_0 = `[{"statement_id":0,"Series":[{"name":"h2o_feet","tags":{"location":"coyote_creek"},"columns":["time","location","max"],"values":[[1566977040,"coyote_creek",9.964]]},{"name":"h2o_feet","tags":{"location":"santa_monica"},"columns":["time","location","max"],"values":[[1566964440,"santa_monica",7.205]]}],"Messages":null}]`
	)
	q := db1.NewQuery(`select "location", MAX("water_level") from "h2o_feet" group by "location"`, "NOAA_water_database", "s")
	response, err := influxClient.Query(q)
	if err != nil {
		t.Fatal(err)
	}
	if response.Error() != nil {
		t.Fatal(response.Error())
	}
	testJson, err := json.Marshal(response.Results)
	assert.JSONEq(t, expected_0, string(testJson))
}
