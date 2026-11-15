package bigquery_test

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	tcbigquery "github.com/testcontainers/testcontainers-go/modules/gcloud/bigquery"
	"github.com/testcontainers/testcontainers-go/modules/gcloud/internal/testrest"
)

//go:embed testdata/data.yaml
var dataYaml []byte

func TestBigQueryWithDataYAML(t *testing.T) {
	ctx := context.Background()

	t.Run("valid", func(t *testing.T) {
		bigQueryContainer, err := tcbigquery.Run(
			ctx,
			"ghcr.io/goccy/bigquery-emulator:0.6.1",
			tcbigquery.WithProjectID("test"),
			tcbigquery.WithDataYAML(bytes.NewReader(dataYaml)),
		)
		testcontainers.CleanupContainer(t, bigQueryContainer)
		require.NoError(t, err)

		projectID := bigQueryContainer.ProjectID()

		queryCtx, cancel := context.WithTimeout(ctx, time.Minute)
		defer cancel()
		client := &http.Client{Timeout: 30 * time.Second}
		defer client.CloseIdleConnections()

		// The emulator serves the BigQuery v2 REST API on its generated URI.
		endpoint := bigQueryContainer.URI() + "/bigquery/v2/projects/" + url.PathEscape(projectID) + "/queries"
		request := map[string]any{
			"query":         "SELECT * FROM dataset1.table_a where name = @name",
			"useLegacySql":  false,
			"parameterMode": "NAMED",
			"queryParameters": []any{map[string]any{
				"name":           "name",
				"parameterType":  map[string]string{"type": "STRING"},
				"parameterValue": map[string]string{"value": "bob"},
			}},
		}
		type queryResult struct {
			JobComplete  bool              `json:"jobComplete"`
			Errors       []json.RawMessage `json:"errors"`
			PageToken    string            `json:"pageToken"`
			JobReference struct {
				JobID    string `json:"jobId"`
				Location string `json:"location"`
			} `json:"jobReference"`
			Rows []struct {
				Fields []struct {
					Value string `json:"v"`
				} `json:"f"`
			} `json:"rows"`
		}
		var result queryResult
		require.NoError(t, testrest.Do(queryCtx, client, http.MethodPost, endpoint, request, &result))
		jobID, location := result.JobReference.JobID, result.JobReference.Location
		var val int64
		readRow := false
		for {
			require.Empty(t, result.Errors)
			if result.JobComplete {
				for _, row := range result.Rows {
					require.NotEmpty(t, row.Fields)
					val, err = strconv.ParseInt(row.Fields[0].Value, 10, 64)
					require.NoError(t, err)
					readRow = true
				}
				if result.PageToken == "" {
					break
				}
			}
			require.NotEmpty(t, jobID)
			require.NoError(t, queryCtx.Err())
			params := url.Values{"timeoutMs": {"10000"}}
			if result.PageToken != "" {
				params.Set("pageToken", result.PageToken)
			}
			if location != "" {
				params.Set("location", location)
			}
			result = queryResult{}
			require.NoError(t, testrest.Do(queryCtx, client, http.MethodGet, endpoint+"/"+url.PathEscape(jobID)+"?"+params.Encode(), nil, &result))
		}
		require.True(t, readRow)
		require.Equal(t, int64(30), val)
	})

	t.Run("multi-value-set", func(t *testing.T) {
		bigQueryContainer, err := tcbigquery.Run(
			ctx,
			"ghcr.io/goccy/bigquery-emulator:0.6.1",
			tcbigquery.WithProjectID("test"),
			tcbigquery.WithDataYAML(bytes.NewReader(dataYaml)),
			tcbigquery.WithDataYAML(bytes.NewReader(dataYaml)),
		)
		testcontainers.CleanupContainer(t, bigQueryContainer)
		require.ErrorContains(t, err, `data yaml already exists`)
	})

	t.Run("multi-value-not-set", func(t *testing.T) {
		noValueOption := func() testcontainers.ContainerCustomizer {
			return testcontainers.WithCmdArgs("--data-from-yaml")
		}

		bigQueryContainer, err := tcbigquery.Run(
			ctx,
			"ghcr.io/goccy/bigquery-emulator:0.6.1",
			noValueOption(), // because --project is always added last, this option will receive `--project` as value, which results in an error
			tcbigquery.WithProjectID("test"),
			tcbigquery.WithDataYAML(bytes.NewReader(dataYaml)),
		)
		testcontainers.CleanupContainer(t, bigQueryContainer)
		require.Error(t, err)
	})
}
