package wait_test

import (
	"context"
	"time"

	"github.com/testcontainers/testcontainers-go/wait"
)

func ExampleForAny() {
	// The following are run in parallel. Any that fail will fail the entire
	// ForAny. Any that succeed will succeed the ForAny. Either case cancels the
	// remaining waiting strategies.
	strategy := wait.ForAny(
		wait.ForLog("port: 3306  MySQL Community Server - GPL"),              // Timeout: 120s
		wait.ForExposedPort().WithStartupTimeout(180*time.Second),            // Timeout: 180s
		wait.ForListeningPort("3306/tcp").WithStartupTimeout(10*time.Second), // Timeout: 10s
	)

	// You can apply a default StartupTimeout for any inner strategy that didn't define it.
	strategy = strategy.WithStartupTimeoutDefault(120 * time.Second)

	// You can apply an overall deadline for all the strategies to complete within.
	strategy = strategy.WithDeadline(360 * time.Second)

	// Call WaitUntilReady, or pass the strategy to a function that uses it, to
	// start waiting.
	_ = strategy.WaitUntilReady(context.Background(), &wait.NopStrategyTarget{})
}

func ExampleForAll() {
	// The following are run in series. Any that fail will fail the entire
	// ForAll, and the remaining waiting strategies will be cancelled. All must
	// succeed for the ForAll to succeed.
	strategy := wait.ForAll(
		wait.ForLog("port: 3306  MySQL Community Server - GPL"),              // Timeout: 120s
		wait.ForExposedPort().WithStartupTimeout(180*time.Second),            // Timeout: 180s
		wait.ForListeningPort("3306/tcp").WithStartupTimeout(10*time.Second), // Timeout: 10s
	)

	// You can apply a default StartupTimeout for any inner strategy that didn't define it.
	strategy = strategy.WithStartupTimeoutDefault(120 * time.Second)

	// You can apply an overall deadline for all the strategies to complete within.
	strategy = strategy.WithDeadline(360 * time.Second)

	// Call WaitUntilReady, or pass the strategy to a function that uses it, to
	// start waiting.
	_ = strategy.WaitUntilReady(context.Background(), &wait.NopStrategyTarget{})
}
