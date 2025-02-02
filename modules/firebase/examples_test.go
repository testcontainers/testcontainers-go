package firebase_test

import (
	"context"
	"fmt"
	"log"
	"path/filepath"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/firebase"
)

func ExampleRun() {
	ctx := context.Background()

	firebaseContainer, err := firebase.Run(ctx, "ghcr.io/u-health/docker-firebase-emulator:13.29.2",
		firebase.WithRoot(filepath.Join(".", "firebase")),
		firebase.WithCache(),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(firebaseContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	state, err := firebaseContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}
