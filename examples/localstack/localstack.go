package localstack

import (
	"context"
	"fmt"
	"strings"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/internal/testcontainersdocker"
	"github.com/testcontainers/testcontainers-go/wait"
	"golang.org/x/mod/semver"
)

const defaultVersion = "0.11.2"

// localStackContainer represents the LocalStack container type used in the module
type localStackContainer struct {
	testcontainers.Container
	legacyMode bool
}

func runInLegacyMode(version string) bool {
	if version == "latest" {
		return false
	}

	if !strings.HasPrefix(version, "v") {
		version = fmt.Sprintf("v%s", version)
	}

	if semver.IsValid(version) {
		return semver.Compare(version, "v0.11") < 0
	}

	fmt.Printf("Version %s is not a semantic version, LocalStack will run in legacy mode.\n", version)
	fmt.Printf("Consider using \"setupLocalStack(context context.Context, version string, legacyMode bool)\" constructor if you want to disable legacy mode.")
	return true
}

// setupLocalStack creates an instance of the LocalStack container type
func setupLocalStack(ctx context.Context, version string, legacyMode bool) (*localStackContainer, error) {
	if version == "" {
		version = defaultVersion
	}

	/*
		Do not run in legacy mode when the version is a valid semver version greater than the v0.11 and legacyMode is false
			| runInLegacyMode | legacyMode | result |
			|-----------------|------------|--------|
			| false           | false      | false  |
			| false           | true       | true   |
			| true            | false      | true   |
			| true            | true       | true   |
	*/
	legacyMode = !runInLegacyMode(version) && !legacyMode

	req := testcontainers.ContainerRequest{
		Image:      "localstack/localstack:0.11.2",
		Binds:      []string{fmt.Sprintf("%s:/var/run/docker.sock", testcontainersdocker.ExtractDockerHost(ctx))},
		WaitingFor: wait.ForLog(".*Ready\\.\n").WithOccurrence(1),
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	return &localStackContainer{Container: container, legacyMode: legacyMode}, nil
}
