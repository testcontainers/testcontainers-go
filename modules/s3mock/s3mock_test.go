package s3mock_test

import (
	"context"
	"encoding/xml"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/s3mock"
)

// TestS3Mock verifies that an S3Mock container starts correctly and that
// EndpointURL, HTTPSEndpointURL, and basic S3 operations all work as expected.
func TestS3Mock(t *testing.T) {
	ctx := context.Background()

	ctr, err := s3mock.Run(ctx, "adobe/s3mock:3.9.1")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	t.Run("EndpointURL", func(t *testing.T) {
		// endpointURL {
		endpointURL, err := ctr.EndpointURL(ctx)
		// }
		require.NoError(t, err)
		require.True(t, strings.HasPrefix(endpointURL, "http://"), "expected http scheme, got: %s", endpointURL)
	})

	t.Run("HTTPSEndpointURL", func(t *testing.T) {
		// httpsEndpointURL {
		httpsURL, err := ctr.HTTPSEndpointURL(ctx)
		// }
		require.NoError(t, err)
		// We only verify the scheme; making a full TLS request would require
		// trusting S3Mock's self-signed certificate, which is out of scope here.
		require.True(t, strings.HasPrefix(httpsURL, "https://"), "expected https scheme, got: %s", httpsURL)
	})

	t.Run("S3 operations", func(t *testing.T) {
		endpointURL, err := ctr.EndpointURL(ctx)
		require.NoError(t, err)
		bucketURL := endpointURL + "/testcontainers-s3mock"

		// S3Mock 3.9.1 documents unsigned PUT/GET requests; no signer is needed.
		s3Request(t, ctx, http.MethodPut, bucketURL, nil)

		const content = "hello from testcontainers"
		s3Request(t, ctx, http.MethodPut, bucketURL+"/hello.txt", strings.NewReader(content))

		got := s3Request(t, ctx, http.MethodGet, bucketURL+"/hello.txt", nil)
		require.Equal(t, content, string(got))
	})
}

// TestS3MockWithInitialBuckets verifies that WithInitialBuckets pre-creates
// the specified buckets so they are available immediately after startup.
func TestS3MockWithInitialBuckets(t *testing.T) {
	ctx := context.Background()

	// withInitialBuckets {
	ctr, err := s3mock.Run(ctx, "adobe/s3mock:3.9.1",
		s3mock.WithInitialBuckets("bucket1", "bucket2"),
	)
	// }
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	bucketNames := listBuckets(t, ctx, ctr)

	require.Contains(t, bucketNames, "bucket1")
	require.Contains(t, bucketNames, "bucket2")
}

// TestS3MockWithInitialBucketsEmpty verifies that calling WithInitialBuckets
// with no arguments is a no-op and does not create any buckets.
func TestS3MockWithInitialBucketsEmpty(t *testing.T) {
	ctx := context.Background()

	// WithInitialBuckets() with no arguments should be a no-op
	ctr, err := s3mock.Run(ctx, "adobe/s3mock:3.9.1",
		s3mock.WithInitialBuckets(),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	bucketNames := listBuckets(t, ctx, ctr)
	require.Empty(t, bucketNames, "expected no pre-created buckets when WithInitialBuckets is called with no arguments")
}

// s3Request bounds both the request duration and response body at the public endpoint.
func s3Request(t *testing.T, ctx context.Context, method, endpoint string, body io.Reader) []byte {
	t.Helper()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	require.NoError(t, err)
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	const maxResponseBytes = 1 << 20
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes+1))
	require.NoError(t, err)
	require.LessOrEqual(t, len(data), maxResponseBytes, "response exceeds limit")
	require.Equal(t, http.StatusOK, resp.StatusCode, "%s %s: %s", method, endpoint, data)
	return data
}

func listBuckets(t *testing.T, ctx context.Context, ctr *s3mock.Container) []string {
	t.Helper()
	endpointURL, err := ctr.EndpointURL(ctx)
	require.NoError(t, err)
	data := s3Request(t, ctx, http.MethodGet, endpointURL+"/", nil)
	var result struct {
		XMLName xml.Name `xml:"ListAllMyBucketsResult"`
		Buckets []struct {
			Name string `xml:"Name"`
		} `xml:"Buckets>Bucket"`
	}
	require.NoError(t, xml.Unmarshal(data, &result))
	bucketNames := make([]string, 0, len(result.Buckets))
	for _, bucket := range result.Buckets {
		bucketNames = append(bucketNames, bucket.Name)
	}
	return bucketNames
}
