package papercutsmtp_test

import (
	"context"
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/papercutsmtp"
)

func ExampleRun() {
	// runPapercutSMTPContainer {
	ctx := context.Background()

	papercutsmtpContainer, err := papercutsmtp.Run(ctx, "changemakerstudiosus/papercut-smtp:latest")
	defer func() {
		if err := testcontainers.TerminateContainer(papercutsmtpContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := papercutsmtpContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}
