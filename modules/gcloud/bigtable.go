package gcloud

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// RunBigTableContainer creates an instance of the GCloud container type for BigTable
func RunBigTableContainer(ctx context.Context, opts ...testcontainers.RequestCustomizer) (*Container, error) {
	req := testcontainers.Request{
		Image:        "gcr.io/google.com/cloudsdktool/cloud-sdk:367.0.0-emulators",
		ExposedPorts: []string{"9000/tcp"},
		WaitingFor:   wait.ForLog("running"),
		Started:      true,
	}

	settings, err := applyOptions(&req, opts)
	if err != nil {
		return nil, err
	}

	req.Cmd = []string{
		"/bin/sh",
		"-c",
		"gcloud beta emulators bigtable start --host-port 0.0.0.0:9000 " + fmt.Sprintf("--project=%s", settings.ProjectID),
	}

	ctr, err := testcontainers.Run(ctx, req)
	if err != nil {
		return nil, err
	}

	return newGCloudContainer(ctx, 9000, ctr, settings)
}
