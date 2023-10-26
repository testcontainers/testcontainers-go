package k6

import (
	"context"
	"path/filepath"
	"testing"
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
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.title, func(t *testing.T) {
			ctx := context.Background()

			absPath, err := filepath.Abs(filepath.Join("scripts", tc.script))
			if err != nil {
				t.Fatal(err)
			}

			container, err := RunContainer(ctx, WithCache(), WithTestScript(absPath))
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
