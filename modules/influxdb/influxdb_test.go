package influxdb_test

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	influxclient "github.com/influxdata/influxdb1-client/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/influxdb"
)

func TestV1Container(t *testing.T) {
	ctx := context.Background()
	influxDbContainer, err := influxdb.Run(ctx, "influxdb:1.8.10")
	testcontainers.CleanupContainer(t, influxDbContainer)
	require.NoError(t, err)

	state, err := influxDbContainer.State(ctx)
	require.NoError(t, err)

	if !state.Running {
		t.Fatal("InfluxDB container is not running")
	}
}

func TestV2Container(t *testing.T) {
	ctx := context.Background()
	influxDbContainer, err := influxdb.Run(ctx,
		"influxdb:2.7.5-alpine",
		influxdb.WithDatabase("foo"),
		influxdb.WithUsername("root"),
		influxdb.WithPassword("password"),
	)
	testcontainers.CleanupContainer(t, influxDbContainer)
	require.NoError(t, err)

	state, err := influxDbContainer.State(ctx)
	require.NoError(t, err)

	if !state.Running {
		t.Fatal("InfluxDB container is not running")
	}
}

func TestWithInitDb(t *testing.T) {
	ctx := context.Background()
	influxDbContainer, err := influxdb.Run(ctx,
		"influxdb:1.8.10",
		influxdb.WithInitDb("testdata"),
	)
	testcontainers.CleanupContainer(t, influxDbContainer)
	require.NoError(t, err)

	if state, err := influxDbContainer.State(ctx); err != nil || !state.Running {
		require.NoError(t, err)
	}

	cli, err := influxclient.NewHTTPClient(influxclient.HTTPConfig{
		Addr: influxDbContainer.MustConnectionUrl(ctx),
	})
	require.NoError(t, err)
	defer cli.Close()

	expected_0 := `[{"statement_id":0,"Series":[{"name":"h2o_feet","tags":{"location":"coyote_creek"},"columns":["time","location","max"],"values":[[1566977040,"coyote_creek",9.964]]},{"name":"h2o_feet","tags":{"location":"santa_monica"},"columns":["time","location","max"],"values":[[1566964440,"santa_monica",7.205]]}],"Messages":null}]`
	q := influxclient.NewQuery(`select "location", MAX("water_level") from "h2o_feet" group by "location"`, "NOAA_water_database", "s")
	response, err := cli.Query(q)
	require.NoError(t, err)

	if response.Error() != nil {
		t.Fatal(response.Error())
	}
	testJson, err := json.Marshal(response.Results)
	require.NoError(t, err)

	assert.JSONEq(t, expected_0, string(testJson))
}

func TestWithConfigFile(t *testing.T) {
	influxVersion := "1.8.10"

	influxDbContainer, err := influxdb.Run(context.Background(),
		"influxdb:"+influxVersion,
		influxdb.WithConfigFile(filepath.Join("testdata", "influxdb.conf")),
	)
	testcontainers.CleanupContainer(t, influxDbContainer)
	require.NoError(t, err)

	if state, err := influxDbContainer.State(context.Background()); err != nil || !state.Running {
		require.NoError(t, err)
	}

	/// influxConnectionUrl {
	cli, err := influxclient.NewHTTPClient(influxclient.HTTPConfig{
		Addr: influxDbContainer.MustConnectionUrl(context.Background()),
	})
	// }
	require.NoError(t, err)
	defer cli.Close()

	ping, version, err := cli.Ping(5 * time.Second)
	require.NoError(t, err)

	assert.Equal(t, "1.8.10", version)
	assert.Greater(t, ping, time.Duration(0))
}
