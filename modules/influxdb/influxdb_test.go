package influxdb_test

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
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

	require.Truef(t, state.Running, "InfluxDB container is not running")
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

	require.Truef(t, state.Running, "InfluxDB container is not running")
}

func TestWithV2Env(t *testing.T) {
	username, password := "root", "password"
	token := "token"
	retention := "30d"
	usernameFile, passwordFile, tokenFile := "username.txt", "password.txt", "token.txt"

	tests := []struct {
		name           string
		influxConfig   influxdb.InfluxDBV2Config
		expectedEnvMap map[string]string
	}{
		{
			name: "With required only",
			influxConfig: influxdb.InfluxDBV2Config{
				Org:    "Org",
				Bucket: "Bucket",
			},
			expectedEnvMap: map[string]string{
				"DOCKER_INFLUXDB_INIT_MODE":         "setup",
				"DOCKER_INFLUXDB_INIT_ORG":          "Org",
				"DOCKER_INFLUXDB_INIT_BUCKET":       "Bucket",
				"DOCKER_INFLUXDB_INIT_AUTH_ENABLED": "false",
				"INFLUXDB_BIND_ADDRESS":             ":8088",
				"INFLUXDB_HTTP_BIND_ADDRESS":        ":8086",
				"INFLUXDB_REPORTING_DISABLED":       "true",
				"INFLUXDB_MONITOR_STORE_ENABLED":    "false",
				"INFLUXDB_HTTP_HTTPS_ENABLED":       "false",
				"INFLUXDB_HTTP_AUTH_ENABLED":        "false",
			},
		},
		{
			name: "With username and password",
			influxConfig: influxdb.InfluxDBV2Config{
				Org:      "Org",
				Bucket:   "Bucket",
				Username: &username,
				Password: &password,
			},
			expectedEnvMap: map[string]string{
				"DOCKER_INFLUXDB_INIT_MODE":         "setup",
				"DOCKER_INFLUXDB_INIT_ORG":          "Org",
				"DOCKER_INFLUXDB_INIT_BUCKET":       "Bucket",
				"DOCKER_INFLUXDB_INIT_AUTH_ENABLED": "false",
				"DOCKER_INFLUXDB_INIT_USERNAME":     username,
				"DOCKER_INFLUXDB_INIT_PASSWORD":     password,
				"INFLUXDB_BIND_ADDRESS":             ":8088",
				"INFLUXDB_HTTP_BIND_ADDRESS":        ":8086",
				"INFLUXDB_REPORTING_DISABLED":       "true",
				"INFLUXDB_MONITOR_STORE_ENABLED":    "false",
				"INFLUXDB_HTTP_HTTPS_ENABLED":       "false",
				"INFLUXDB_HTTP_AUTH_ENABLED":        "false",
			},
		},
		{
			name: "With token",
			influxConfig: influxdb.InfluxDBV2Config{
				Org:    "Org",
				Bucket: "Bucket",
				Token:  &token,
			},
			expectedEnvMap: map[string]string{
				"DOCKER_INFLUXDB_INIT_MODE":         "setup",
				"DOCKER_INFLUXDB_INIT_ORG":          "Org",
				"DOCKER_INFLUXDB_INIT_BUCKET":       "Bucket",
				"DOCKER_INFLUXDB_INIT_AUTH_ENABLED": "false",
				"DOCKER_INFLUXDB_INIT_ADMIN_TOKEN":  token,
				"INFLUXDB_BIND_ADDRESS":             ":8088",
				"INFLUXDB_HTTP_BIND_ADDRESS":        ":8086",
				"INFLUXDB_REPORTING_DISABLED":       "true",
				"INFLUXDB_MONITOR_STORE_ENABLED":    "false",
				"INFLUXDB_HTTP_HTTPS_ENABLED":       "false",
				"INFLUXDB_HTTP_AUTH_ENABLED":        "false",
			},
		},
		{
			name: "With retention",
			influxConfig: influxdb.InfluxDBV2Config{
				Org:       "Org",
				Bucket:    "Bucket",
				Retention: &retention,
			},
			expectedEnvMap: map[string]string{
				"DOCKER_INFLUXDB_INIT_MODE":         "setup",
				"DOCKER_INFLUXDB_INIT_ORG":          "Org",
				"DOCKER_INFLUXDB_INIT_BUCKET":       "Bucket",
				"DOCKER_INFLUXDB_INIT_AUTH_ENABLED": "false",
				"DOCKER_INFLUXDB_INIT_RETENTION":    retention,
				"INFLUXDB_BIND_ADDRESS":             ":8088",
				"INFLUXDB_HTTP_BIND_ADDRESS":        ":8086",
				"INFLUXDB_REPORTING_DISABLED":       "true",
				"INFLUXDB_MONITOR_STORE_ENABLED":    "false",
				"INFLUXDB_HTTP_HTTPS_ENABLED":       "false",
				"INFLUXDB_HTTP_AUTH_ENABLED":        "false",
			},
		},
		{
			name: "With files",
			influxConfig: influxdb.InfluxDBV2Config{
				Org:          "Org",
				Bucket:       "Bucket",
				UsernameFile: &usernameFile,
				PasswordFile: &passwordFile,
				TokenFile:    &tokenFile,
			},
			expectedEnvMap: map[string]string{
				"DOCKER_INFLUXDB_INIT_MODE":             "setup",
				"DOCKER_INFLUXDB_INIT_ORG":              "Org",
				"DOCKER_INFLUXDB_INIT_BUCKET":           "Bucket",
				"DOCKER_INFLUXDB_INIT_AUTH_ENABLED":     "false",
				"DOCKER_INFLUXDB_INIT_USERNAME_FILE":    usernameFile,
				"DOCKER_INFLUXDB_INIT_PASSWORD_FILE":    passwordFile,
				"DOCKER_INFLUXDB_INIT_ADMIN_TOKEN_FILE": tokenFile,
				"INFLUXDB_BIND_ADDRESS":                 ":8088",
				"INFLUXDB_HTTP_BIND_ADDRESS":            ":8086",
				"INFLUXDB_REPORTING_DISABLED":           "true",
				"INFLUXDB_MONITOR_STORE_ENABLED":        "false",
				"INFLUXDB_HTTP_HTTPS_ENABLED":           "false",
				"INFLUXDB_HTTP_AUTH_ENABLED":            "false",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			genericReq := &testcontainers.GenericContainerRequest{
				ContainerRequest: testcontainers.ContainerRequest{
					Image:        "influxdb:2.7.5-alpine",
					ExposedPorts: []string{"8086/tcp", "8088/tcp"},
					Env: map[string]string{
						"INFLUXDB_BIND_ADDRESS":          ":8088",
						"INFLUXDB_HTTP_BIND_ADDRESS":     ":8086",
						"INFLUXDB_REPORTING_DISABLED":    "true",
						"INFLUXDB_MONITOR_STORE_ENABLED": "false",
						"INFLUXDB_HTTP_HTTPS_ENABLED":    "false",
						"INFLUXDB_HTTP_AUTH_ENABLED":     "false",
					},
				},
			}

			opt := influxdb.WithV2Env(test.influxConfig)
			_ = opt(genericReq)

			assert.Equal(t, test.expectedEnvMap, genericReq.Env)
		})
	}
}

func TestV2Container_WithOptions(t *testing.T) {
	ctx := context.Background()

	username := "username"
	password := "password"
	org := "org"
	bucket := "bucket"
	authEnabled := true
	token := "influxdbv2token"

	influxdbContainer, err := influxdb.Run(ctx, "influxdb:2.7.11",
		influxdb.WithV2Env(influxdb.InfluxDBV2Config{
			Username:    &username,
			Password:    &password,
			Org:         org,
			Bucket:      bucket,
			Token:       &token,
			AuthEnabled: &authEnabled,
		}),
	)
	testcontainers.CleanupContainer(t, influxdbContainer)
	require.NoError(t, err)

	state, err := influxdbContainer.State(ctx)
	require.NoError(t, err)
	require.True(t, state.Running)

	// Query the InfluxDB API to verify the setup
	url, err := influxdbContainer.ConnectionUrl(ctx)
	require.NoError(t, err)

	// Initialize a new InfluxDB client
	client := influxdb2.NewClientWithOptions(url, token, influxdb2.DefaultOptions())
	defer client.Close()

	// Get the bucket
	influxBucket, err := client.BucketsAPI().FindBucketByName(ctx, bucket)
	require.NoError(t, err)
	require.Equal(t, bucket, influxBucket.Name)

	// Try to connect without authentication
	clientWithoutToken := influxdb2.NewClientWithOptions(url, "", influxdb2.DefaultOptions())
	defer clientWithoutToken.Close()

	_, err = clientWithoutToken.BucketsAPI().CreateBucketWithNameWithID(ctx, org, "example")
	require.Error(t, err, "Expected error when trying to create a bucket without authentication")
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

	expected0 := `[{"statement_id":0,"Series":[{"name":"h2o_feet","tags":{"location":"coyote_creek"},"columns":["time","location","max"],"values":[[1566977040,"coyote_creek",9.964]]},{"name":"h2o_feet","tags":{"location":"santa_monica"},"columns":["time","location","max"],"values":[[1566964440,"santa_monica",7.205]]}],"Messages":null}]`
	q := influxclient.NewQuery(`select "location", MAX("water_level") from "h2o_feet" group by "location"`, "NOAA_water_database", "s")
	response, err := cli.Query(q)
	require.NoError(t, err)

	require.NoError(t, response.Error())
	testJSON, err := json.Marshal(response.Results)
	require.NoError(t, err)

	assert.JSONEq(t, expected0, string(testJSON))
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
