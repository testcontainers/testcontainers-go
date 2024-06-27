package k6_test

import (
	"context"
	"net/url"
	"path/filepath"
	"strings"
	"testing"

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

	for _, tc := range testCases {
		tc := tc
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

			container, err := k6.Run(ctx, "szkiba/k6x:v0.3.1", k6.WithCache(), options)
			if err != nil {
				t.Fatal(err)
			}
			// Clean up the container after the test is complete
			t.Cleanup(func() {
				if err := container.Terminate(ctx); err != nil {
					t.Fatalf("failed to terminate container: %s", err)
				}
			})

			// assert the result of the test
			state, err := container.State(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if state.ExitCode != tc.expect {
				t.Fatalf("expected %d got %d", tc.expect, state.ExitCode)
			}
		})
	}
}
