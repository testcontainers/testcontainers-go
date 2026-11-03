package testcontainers

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/moby/moby/client"
	"github.com/stretchr/testify/require"
)

func TestDockerContainerCopyToContainerUsesStrictSafeArchivePath(t *testing.T) {
	t.Setenv("GODEBUG", os.Getenv("GODEBUG")+",tarinsecurepath=0")

	const (
		containerID         = "copy-test-container"
		absoluteDestination = "/tmp/nested/copied.txt"
		relativeDestination = "tmp/nested/copied.txt"
		archiveName         = "tmp/nested/copied.txt"
		fileMode            = int64(0o640)
	)
	payload := []byte("copied through the Engine API")

	tests := []struct {
		name string
		copy func(t *testing.T, container *DockerContainer) error
	}{
		{
			name: "CopyToContainer absolute destination",
			copy: func(t *testing.T, container *DockerContainer) error {
				t.Helper()
				return container.CopyToContainer(t.Context(), payload, absoluteDestination, fileMode)
			},
		},
		{
			name: "CopyFileToContainer relative destination",
			copy: func(t *testing.T, container *DockerContainer) error {
				t.Helper()
				hostPath := filepath.Join(t.TempDir(), "copied.txt")
				require.NoError(t, os.WriteFile(hostPath, payload, 0o600))
				return container.CopyFileToContainer(t.Context(), hostPath, relativeDestination, fileMode)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			type copyRequest struct {
				method      string
				path        string
				destination string
				header      *tar.Header
				payload     []byte
				err         error
			}
			requests := make(chan copyRequest, 1)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				got := copyRequest{
					method:      r.Method,
					path:        r.URL.Path,
					destination: r.URL.Query().Get("path"),
				}
				gzr, err := gzip.NewReader(r.Body)
				if err == nil {
					defer gzr.Close()
					tr := tar.NewReader(gzr)
					got.header, got.err = tr.Next()
					if got.err == nil {
						got.payload, got.err = io.ReadAll(tr)
					}
					if got.err == nil {
						_, got.err = tr.Next()
						if errors.Is(got.err, io.EOF) {
							got.err = nil
						}
					}
				} else {
					got.err = err
				}
				requests <- got
				w.WriteHeader(http.StatusOK)
			}))
			t.Cleanup(server.Close)

			cli, err := client.New(client.WithHost(server.URL), client.WithAPIVersion("1.41"))
			require.NoError(t, err)
			container := &DockerContainer{
				ID: containerID,
				provider: &DockerProvider{
					client: cli,
				},
			}

			require.NoError(t, test.copy(t, container))
			got := <-requests
			require.NoError(t, got.err)
			require.Equal(t, http.MethodPut, got.method)
			require.Equal(t, "/v1.41/containers/"+containerID+"/archive", got.path)
			require.Equal(t, "/", got.destination)
			require.Equal(t, archiveName, got.header.Name)
			require.Equal(t, int64(len(payload)), got.header.Size)
			require.Equal(t, fileMode, got.header.Mode)
			require.Equal(t, byte(tar.TypeReg), got.header.Typeflag)
			require.Equal(t, payload, got.payload)
		})
	}
}
