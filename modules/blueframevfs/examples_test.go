package blueframevfs_test

import (
	"context"
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/blueframevfs"
)

func ExampleRun() {
	// runBlueframeVFSContainer {
	ctx := context.Background()

	blueframevfsContainer, err := blueframevfs.Run(ctx, "edapt-docker-dev.artifactory.metro.ad.selinc.com/vfs:1.0.6-24289.081e1b3.develop")
	defer func() {
		if err := testcontainers.TerminateContainer(blueframevfsContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := blueframevfsContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}
