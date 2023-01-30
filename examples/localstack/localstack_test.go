package localstack

import (
	"context"
	"testing"
)

func TestRunInLegacyMode(t *testing.T) {
	tests := []struct {
		version string
		want    bool
	}{
		{"foo", true},
		{"latest", false},
		{"0.10.0", true},
		{"0.11", false},
		{"0.11.2", false},
		{"0.12", false},
		{"1.0", false},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			got := runInLegacyMode(tt.version)
			if got != tt.want {
				t.Errorf("runInLegacyMode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLocalStack(t *testing.T) {
	ctx := context.Background()

	container, err := setupLocalStack(ctx, defaultVersion, false)
	if err != nil {
		t.Fatal(err)
	}

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	// perform assertions
}
