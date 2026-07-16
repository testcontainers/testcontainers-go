package sftp_test

import (
	"context"
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/sftp"
)

func ExampleRun() {
	// runSFTPContainer {
	ctx := context.Background()

	sftpContainer, err := sftp.Run(ctx, "atmoz/sftp:latest",
		sftp.WithUser("alice", "secret"),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(sftpContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := sftpContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}
