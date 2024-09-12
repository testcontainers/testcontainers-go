package k6_test

import (
	"context"
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/k6"
)

func TestK6(t *testing.T) {
	testCases := []struct {
		title  string
		script string
		expect int
	}{
		{
			title:  "Passing test",
			script: "pass.js",
			expect: 0,
		},
		{
			title:  "Failing test",
			script: "fail.js",
			expect: 108,
		},
		{
			title:  "Passing remote test",
			script: "https://raw.githubusercontent.com/testcontainers/testcontainers-go/main/modules/k6/scripts/pass.js",
			expect: 0,
		},
		{
			title:  "Failing remote test",
			script: "https://raw.githubusercontent.com/testcontainers/testcontainers-go/main/modules/k6/scripts/fail.js",
			expect: 108,
		},
	}

	var cacheMount string
	t.Cleanup(func() {
		if cacheMount == "" {
			return
		}

		// Ensure the cache volume is removed as mounts that specify a volume
		// source as defined by the name are not removed automatically.
		provider, err := testcontainers.NewDockerProvider()
		require.NoError(t, err)
		defer provider.Close()

		require.NoError(t, provider.Client().VolumeRemove(context.Background(), cacheMount, true))
	})

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			ctx := context.Background()

			var options testcontainers.CustomizeRequestOption
			if !strings.HasPrefix(tc.script, "http") {
				absPath, err := filepath.Abs(filepath.Join("scripts", tc.script))
				if err != nil {
					t.Fatal(err)
				}
				options = k6.WithTestScript(absPath)
			} else {

				uri, err := url.Parse(tc.script)
				if err != nil {
					t.Fatal(err)
				}

				desc := k6.DownloadableFile{Uri: *uri, DownloadDir: t.TempDir()}
				options = k6.WithRemoteTestScript(desc)
			}

			ctr, err := k6.Run(ctx, "szkiba/k6x:v0.3.1", k6.WithCache(), options)
			if ctr != nil && cacheMount == "" {
				// First container, determine the cache mount.
				cacheMount, err = ctr.CacheMount(ctx)
				require.NoError(t, err)
			}
			testcontainers.CleanupContainer(t, ctr)
			require.NoError(t, err)

			// assert the result of the test
			state, err := ctr.State(ctx)
			require.NoError(t, err)
			require.Equal(t, tc.expect, state.ExitCode)
		})
	}
}
