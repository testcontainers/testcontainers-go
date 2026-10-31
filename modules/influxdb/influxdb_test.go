package influxdb_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/influxdb"
)

func TestV1Container(t *testing.T) {
	ctx := context.Background()
	influxDBContainer, err := influxdb.Run(ctx, "influxdb:1.8.10")
	testcontainers.CleanupContainer(t, influxDBContainer)
	require.NoError(t, err)

	state, err := influxDBContainer.State(ctx)
	require.NoError(t, err)

	require.Truef(t, state.Running, "InfluxDB container is not running")
}

func TestV2Container(t *testing.T) {
	ctx := context.Background()
	influxDBContainer, err := influxdb.Run(ctx,
		"influxdb:2.7.5-alpine",
		influxdb.WithDatabase("foo"),
		influxdb.WithUsername("root"),
		influxdb.WithPassword("password"),
	)
	testcontainers.CleanupContainer(t, influxDBContainer)
	require.NoError(t, err)

	state, err := influxDBContainer.State(ctx)
	require.NoError(t, err)

	require.Truef(t, state.Running, "InfluxDB container is not running")
}

func TestV2Options(t *testing.T) {
	// Ok base cases
	t.Run("with-username-password", func(t *testing.T) {
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

		opt := influxdb.WithV2Auth("org", "bucket", "username", "password")
		_ = opt(genericReq)

		require.Equal(t, map[string]string{
			"INFLUXDB_BIND_ADDRESS":          ":8088",
			"INFLUXDB_HTTP_BIND_ADDRESS":     ":8086",
			"INFLUXDB_REPORTING_DISABLED":    "true",
			"INFLUXDB_MONITOR_STORE_ENABLED": "false",
			"INFLUXDB_HTTP_HTTPS_ENABLED":    "false",
			"INFLUXDB_HTTP_AUTH_ENABLED":     "false",
			"DOCKER_INFLUXDB_INIT_MODE":      "setup",
			"DOCKER_INFLUXDB_INIT_USERNAME":  "username",
			"DOCKER_INFLUXDB_INIT_PASSWORD":  "password",
			"DOCKER_INFLUXDB_INIT_ORG":       "org",
			"DOCKER_INFLUXDB_INIT_BUCKET":    "bucket",
		}, genericReq.Env)
	})

	t.Run("with-org-bucket", func(t *testing.T) {
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

		opt := influxdb.WithV2("org", "bucket")
		_ = opt(genericReq)

		require.Equal(t, map[string]string{
			"INFLUXDB_BIND_ADDRESS":          ":8088",
			"INFLUXDB_HTTP_BIND_ADDRESS":     ":8086",
			"INFLUXDB_REPORTING_DISABLED":    "true",
			"INFLUXDB_MONITOR_STORE_ENABLED": "false",
			"INFLUXDB_HTTP_HTTPS_ENABLED":    "false",
			"INFLUXDB_HTTP_AUTH_ENABLED":     "false",
			"DOCKER_INFLUXDB_INIT_MODE":      "setup",
			"DOCKER_INFLUXDB_INIT_ORG":       "org",
			"DOCKER_INFLUXDB_INIT_BUCKET":    "bucket",
		}, genericReq.Env)
	})

	t.Run("with-auth-token", func(t *testing.T) {
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

		opt := influxdb.WithV2AdminToken("token")
		_ = opt(genericReq)

		require.Equal(t, map[string]string{
			"INFLUXDB_BIND_ADDRESS":            ":8088",
			"INFLUXDB_HTTP_BIND_ADDRESS":       ":8086",
			"INFLUXDB_REPORTING_DISABLED":      "true",
			"INFLUXDB_MONITOR_STORE_ENABLED":   "false",
			"INFLUXDB_HTTP_HTTPS_ENABLED":      "false",
			"INFLUXDB_HTTP_AUTH_ENABLED":       "false",
			"DOCKER_INFLUXDB_INIT_ADMIN_TOKEN": "token",
		}, genericReq.Env)
	})

	t.Run("with-retention", func(t *testing.T) {
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

		opt := influxdb.WithV2Retention(time.Hour * 24)
		_ = opt(genericReq)

		require.Equal(t, map[string]string{
			"INFLUXDB_BIND_ADDRESS":          ":8088",
			"INFLUXDB_HTTP_BIND_ADDRESS":     ":8086",
			"INFLUXDB_REPORTING_DISABLED":    "true",
			"INFLUXDB_MONITOR_STORE_ENABLED": "false",
			"INFLUXDB_HTTP_HTTPS_ENABLED":    "false",
			"INFLUXDB_HTTP_AUTH_ENABLED":     "false",
			"DOCKER_INFLUXDB_INIT_RETENTION": "24h0m0s",
		}, genericReq.Env)
	})

	t.Run("with-token-file", func(t *testing.T) {
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

		opt := influxdb.WithV2SecretsAdminToken("file")
		_ = opt(genericReq)

		require.Equal(t, map[string]string{
			"INFLUXDB_BIND_ADDRESS":                 ":8088",
			"INFLUXDB_HTTP_BIND_ADDRESS":            ":8086",
			"INFLUXDB_REPORTING_DISABLED":           "true",
			"INFLUXDB_MONITOR_STORE_ENABLED":        "false",
			"INFLUXDB_HTTP_HTTPS_ENABLED":           "false",
			"INFLUXDB_HTTP_AUTH_ENABLED":            "false",
			"DOCKER_INFLUXDB_INIT_ADMIN_TOKEN_FILE": "/run/secrets/file",
		}, genericReq.Env)
	})

	t.Run("with-username-and-password-file", func(t *testing.T) {
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

		opt := influxdb.WithV2SecretsAuth("org", "bucket", "file", "file2")
		_ = opt(genericReq)

		require.Equal(t, map[string]string{
			"INFLUXDB_BIND_ADDRESS":              ":8088",
			"INFLUXDB_HTTP_BIND_ADDRESS":         ":8086",
			"INFLUXDB_REPORTING_DISABLED":        "true",
			"INFLUXDB_MONITOR_STORE_ENABLED":     "false",
			"INFLUXDB_HTTP_HTTPS_ENABLED":        "false",
			"INFLUXDB_HTTP_AUTH_ENABLED":         "false",
			"DOCKER_INFLUXDB_INIT_MODE":          "setup",
			"DOCKER_INFLUXDB_INIT_USERNAME_FILE": "/run/secrets/file",
			"DOCKER_INFLUXDB_INIT_PASSWORD_FILE": "/run/secrets/file2",
			"DOCKER_INFLUXDB_INIT_ORG":           "org",
			"DOCKER_INFLUXDB_INIT_BUCKET":        "bucket",
		}, genericReq.Env)
	})

	// Empty fields provided as arguments
	t.Run("with-empty-username-password", func(t *testing.T) {
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

		opt := influxdb.WithV2Auth("org", "bucket", "", "")
		err := opt(genericReq)
		require.Error(t, err)
	})

	t.Run("with-empty-org-bucket", func(t *testing.T) {
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

		opt := influxdb.WithV2("", "")
		err := opt(genericReq)
		require.Error(t, err)
	})

	t.Run("with-empty-auth-token", func(t *testing.T) {
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

		opt := influxdb.WithV2AdminToken("")
		err := opt(genericReq)
		require.Error(t, err)
	})

	t.Run("with-empty-retention", func(t *testing.T) {
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

		opt := influxdb.WithV2Retention(0)
		err := opt(genericReq)
		require.Error(t, err)
	})

	t.Run("with-empty-token-file", func(t *testing.T) {
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

		opt := influxdb.WithV2SecretsAdminToken("")
		err := opt(genericReq)
		require.Error(t, err)
	})

	t.Run("with-empty-username-and-password-file", func(t *testing.T) {
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

		opt := influxdb.WithV2SecretsAuth("org", "bucket", "", "")
		err := opt(genericReq)
		require.Error(t, err)
	})

	// Conflicts
	t.Run("with-token-already-present-conflict", func(t *testing.T) {
		genericReq := &testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Image:        "influxdb:2.7.5-alpine",
				ExposedPorts: []string{"8086/tcp", "8088/tcp"},
				Env: map[string]string{
					"INFLUXDB_BIND_ADDRESS":            ":8088",
					"INFLUXDB_HTTP_BIND_ADDRESS":       ":8086",
					"INFLUXDB_REPORTING_DISABLED":      "true",
					"INFLUXDB_MONITOR_STORE_ENABLED":   "false",
					"INFLUXDB_HTTP_HTTPS_ENABLED":      "false",
					"INFLUXDB_HTTP_AUTH_ENABLED":       "false",
					"DOCKER_INFLUXDB_INIT_ADMIN_TOKEN": "im a token",
				},
			},
		}

		opt := influxdb.WithV2SecretsAdminToken("file")
		err := opt(genericReq)
		require.Error(t, err)
	})

	t.Run("with-username-password-already-present-conflict", func(t *testing.T) {
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
					"DOCKER_INFLUXDB_INIT_USERNAME":  "im a username",
					"DOCKER_INFLUXDB_INIT_PASSWORD":  "im a password",
				},
			},
		}

		opt := influxdb.WithV2SecretsAuth("org", "bucket", "file", "file2")
		err := opt(genericReq)
		require.Error(t, err)
	})
}

func TestRun_V2WithOptions(t *testing.T) {
	ctx := context.Background()

	username := "username"
	password := "password"
	org := "org"
	bucket := "bucket"
	token := "influxdbv2token"

	influxdbContainer, err := influxdb.Run(ctx, "influxdb:2.7.11",
		// influxdbV2EnvConfig {
		influxdb.WithV2Auth(org, bucket, username, password),
		influxdb.WithV2AdminToken(token),
		// }
	)
	testcontainers.CleanupContainer(t, influxdbContainer)
	require.NoError(t, err)

	state, err := influxdbContainer.State(ctx)
	require.NoError(t, err)
	require.True(t, state.Running)

	// Query the InfluxDB API to verify the setup
	endpoint, err := influxdbContainer.ConnectionUrl(ctx)
	require.NoError(t, err)

	response := influxRequest(t, ctx, http.MethodGet, endpoint+"/api/v2/buckets?name="+url.QueryEscape(bucket), token, nil)
	defer response.Body.Close()
	require.Equal(t, http.StatusOK, response.StatusCode)

	var buckets struct {
		Buckets []struct {
			Name string `json:"name"`
		} `json:"buckets"`
	}
	require.NoError(t, json.NewDecoder(response.Body).Decode(&buckets))
	require.Len(t, buckets.Buckets, 1)
	require.Equal(t, bucket, buckets.Buckets[0].Name)

	response = influxRequest(t, ctx, http.MethodPost, endpoint+"/api/v2/buckets", "", bytes.NewBufferString(`{"orgID":"`+org+`","name":"example","retentionRules":[]}`))
	defer response.Body.Close()
	require.GreaterOrEqual(t, response.StatusCode, http.StatusBadRequest, "Expected error when trying to create a bucket without authentication")
}

func TestWithInitDb(t *testing.T) {
	ctx := context.Background()
	influxDBContainer, err := influxdb.Run(ctx,
		"influxdb:1.8.10",
		influxdb.WithInitDb("testdata"),
	)
	testcontainers.CleanupContainer(t, influxDBContainer)
	require.NoError(t, err)

	if state, err := influxDBContainer.State(ctx); err != nil || !state.Running {
		require.NoError(t, err)
	}

	expected0 := `[{"statement_id":0,"Series":[{"name":"h2o_feet","tags":{"location":"coyote_creek"},"columns":["time","location","max"],"values":[[1566977040,"coyote_creek",9.964]]},{"name":"h2o_feet","tags":{"location":"santa_monica"},"columns":["time","location","max"],"values":[[1566964440,"santa_monica",7.205]]}],"Messages":null}]`
	query := url.Values{"db": {"NOAA_water_database"}, "epoch": {"s"}, "q": {`select "location", MAX("water_level") from "h2o_feet" group by "location"`}}
	response := influxRequest(t, ctx, http.MethodGet, influxDBContainer.MustConnectionUrl(ctx)+"/query?"+query.Encode(), "", nil)
	defer response.Body.Close()
	require.Equal(t, http.StatusOK, response.StatusCode)

	var result struct {
		Results []struct {
			StatementID int `json:"statement_id"`
			// Keep Series and Messages untagged to match the SDK-marshaled expectation.
			Series   json.RawMessage
			Messages json.RawMessage
			Err      string `json:"error,omitempty"`
		} `json:"results"`
		Err string `json:"error"`
	}
	require.NoError(t, json.NewDecoder(response.Body).Decode(&result))
	require.Empty(t, result.Err)
	for _, statement := range result.Results {
		require.Empty(t, statement.Err)
	}
	testJSON, err := json.Marshal(result.Results)
	require.NoError(t, err)

	assert.JSONEq(t, expected0, string(testJSON))
}

func TestWithConfigFile(t *testing.T) {
	influxVersion := "1.8.10"

	influxDBContainer, err := influxdb.Run(context.Background(),
		"influxdb:"+influxVersion,
		influxdb.WithConfigFile(filepath.Join("testdata", "influxdb.conf")),
	)
	testcontainers.CleanupContainer(t, influxDBContainer)
	require.NoError(t, err)

	if state, err := influxDBContainer.State(context.Background()); err != nil || !state.Running {
		require.NoError(t, err)
	}

	started := time.Now()
	/// influxConnectionUrl {
	ctx := context.Background()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, influxDBContainer.MustConnectionUrl(ctx)+"/ping", nil)
	require.NoError(t, err)
	response, err := (&http.Client{Timeout: 5 * time.Second}).Do(request)
	require.NoError(t, err)
	defer response.Body.Close()
	// }
	require.Equal(t, http.StatusNoContent, response.StatusCode)

	assert.Equal(t, "1.8.10", response.Header.Get("X-Influxdb-Version"))
	assert.Greater(t, time.Since(started), time.Duration(0))
}

func influxRequest(t *testing.T, ctx context.Context, method, endpoint, token string, body io.Reader) *http.Response {
	t.Helper()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	t.Cleanup(cancel)

	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	require.NoError(t, err)
	if token != "" {
		req.Header.Set("Authorization", "Token "+token)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	response, err := (&http.Client{Timeout: 5 * time.Second}).Do(req)
	require.NoError(t, err)
	return response
}
