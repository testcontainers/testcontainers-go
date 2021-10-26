package testcontainers

import (
	"fmt"
	"testing"
)

func Test_fetchComposeMajorVersion(t *testing.T) {
	version, err := fetchComposeMajorVersion("docker-compose")
	if err != nil {
		t.Error(err)
	}
	if version == 0 {
		t.Errorf("expected a valid version but got 0")
	}
}

func Test_parseComposeVersion(t *testing.T) {
	versions := [][]byte{
		[]byte("0"),
		[]byte("1"),
		[]byte("2"),
		[]byte("3"),
	}
	for expectedVersion, version := range versions {
		t.Run(fmt.Sprintf("version_%d", expectedVersion), func(t *testing.T) {
			got, _ := parseComposeVersion(version)
			if got != expectedVersion {
				t.Errorf("parseComposeVersion() got = %v, want %v", got, expectedVersion)
			}
		})
	}

	t.Run("error should return 0", func(t *testing.T) {
		got, err := parseComposeVersion([]byte("hello"))
		if err == nil {
			t.Error("expected error but got nil")
		}
		if got != 0 {
			t.Errorf("expected 0 but got %d", got)
		}
	})
}
