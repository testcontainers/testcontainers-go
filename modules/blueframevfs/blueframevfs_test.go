package blueframevfs_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/blueframevfs"
)

func TestBlueframeVFS(t *testing.T) {
	ctx := context.Background()

	ctr, err := blueframevfs.Run(ctx, "edapt-docker-dev.artifactory.metro.ad.selinc.com/vfs:1.0.6-24289.081e1b3.develop")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// perform assertions

	// make a request to vfs to get the running status
	// assert that the response is 200
	// assert that the response body contains "running"
	assert.True(t, ctr.IsRunning())

	// sleep for 55 seconds
	time.Sleep(55 * time.Second)
	ctr.Terminate(ctx)

}
