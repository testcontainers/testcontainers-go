package wait_test

import (
	"testing"

	"github.com/testcontainers/testcontainers-go/wait"
)

// testPortScenarios tests the given strategy against a variety
// of scenarios for ports. The apply function is used to modify
// the waitBuilder before running the test.
func testPortScenarios(t *testing.T, strategy wait.Strategy, apply func(*testing.T, *waitBuilder) *waitBuilder) {
	t.Helper()

	for _, sendingRequest := range []bool{true, false} {
		testPrefix := "getting-port/"
		if sendingRequest {
			testPrefix = "sending-request/"
		}

		t.Run(testPrefix+"running", func(t *testing.T) {
			apply(t,
				newWaitBuilder().
					SendingRequest(sendingRequest),
			).Run(t, strategy)
		})

		t.Run(testPrefix+"oom", func(t *testing.T) {
			apply(t,
				newWaitBuilder().
					State(oom).
					SendingRequest(sendingRequest),
			).Run(t, strategy)
		})

		t.Run(testPrefix+"exited", func(t *testing.T) {
			apply(t,
				newWaitBuilder().
					State(exited).
					SendingRequest(sendingRequest),
			).Run(t, strategy)
		})

		t.Run(testPrefix+"dead", func(t *testing.T) {
			apply(t,
				newWaitBuilder().
					State(dead).
					SendingRequest(sendingRequest),
			).Run(t, strategy)
		})

		t.Run(testPrefix+"no-exposed-ports", func(t *testing.T) {
			var portErr wait.PortNotFoundErr
			apply(t,
				newWaitBuilder().
					MappedPorts().
					SendingRequest(sendingRequest).
					ErrorAs(&portErr),
			).Run(t, strategy)
		})
	}
}
