package dockermodelrunner_test

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/testcontainers/testcontainers-go"
)

func TestMain(m *testing.M) {
	// Skip the tests if Docker Desktop is not running.
	// Because the Docker Model Runner is only available on Docker Desktop,
	// we can skip the tests if Docker Desktop is not running.
	// In the future, the Docker Model Runner will be available on Docker CE,
	// and we will not need to skip the tests.
	cli, err := testcontainers.NewDockerClientWithOpts(context.Background())
	if err != nil {
		log.Fatalf("failed to create docker client: %s", err)
	}

	info, err := cli.Info(context.Background())
	if err != nil {
		log.Fatalf("failed to get docker info: %s", err)
	}

	if info.OperatingSystem != "Docker Desktop" {
		log.Println("Skipping test that needs Docker Desktop")
		os.Exit(0)
	}

	os.Exit(m.Run())
}
