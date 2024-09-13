package gcloud

import (
	"context"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// RunBigTable creates an instance of the GCloud container type for BigTable.
func RunBigTable(ctx context.Context, img string, opts ...testcontainers.RequestCustomizer) (*Container, error) {
	req := testcontainers.Request{
		Image:        img,
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
		"gcloud beta emulators bigtable start --host-port 0.0.0.0:9000 --project=" + settings.ProjectID,
	}

	return newGCloudContainer(ctx, req, 9000, settings, "")
}
