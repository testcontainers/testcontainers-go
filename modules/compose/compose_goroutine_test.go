package compose

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// goroutineSnapshot captures the set of goroutine descriptions currently running,
// keyed by their stack trace (without goroutine IDs so we can compare across calls).
func goroutineSnapshot() map[string]int {
	buf := make([]byte, 1<<20) // 1 MB
	n := runtime.Stack(buf, true)
	stacks := string(buf[:n])

	counts := make(map[string]int)
	// Each goroutine block starts with "goroutine N [..."
	for _, block := range strings.Split(stacks, "\n\n") {
		block = strings.TrimSpace(block)
		if block == "" {
			continue
		}
		// Strip the first line ("goroutine N [state]:") to make it ID-independent.
		lines := strings.SplitN(block, "\n", 2)
		if len(lines) < 2 {
			continue
		}
		key := strings.TrimSpace(lines[1])
		counts[key]++
	}
	return counts
}

// goroutineLeaks returns descriptions of goroutines that are present in after but not before,
// excluding known background goroutines that are not related to the test.
func goroutineLeaks(before, after map[string]int) []string {
	var leaks []string
	for stack, afterCount := range after {
		beforeCount := before[stack]
		if afterCount > beforeCount {
			leaks = append(leaks, fmt.Sprintf("(+%d) %s", afterCount-beforeCount, stack))
		}
	}
	return leaks
}

// TestDockerComposeGoroutineLeak verifies that calling Up() followed by Down() does not
// leak net/http persistConn goroutines from the internal Docker CLI HTTP transport.
//
// Before the fix, two goroutines per Up+Down cycle were leaked permanently:
//
//	net/http.(*persistConn).readLoop
//	net/http.(*persistConn).writeLoop
func TestDockerComposeGoroutineLeak(t *testing.T) {
	path, _ := RenderComposeSimple(t)

	// Allow goroutines from previous tests to settle before taking baseline.
	time.Sleep(200 * time.Millisecond)
	before := goroutineSnapshot()

	compose, err := NewDockerCompose(path)
	require.NoError(t, err, "NewDockerCompose()")

	ctx := context.Background()

	err = compose.Up(ctx, Wait(true))
	require.NoError(t, err, "compose.Up()")

	err = compose.Down(ctx, RemoveOrphans(true), RemoveVolumes(true))
	require.NoError(t, err, "compose.Down()")

	// Allow goroutines to fully exit after Down().
	time.Sleep(500 * time.Millisecond)

	after := goroutineSnapshot()
	leaks := goroutineLeaks(before, after)

	// Filter to only HTTP transport goroutines that are the known leak.
	var httpLeaks []string
	for _, l := range leaks {
		if strings.Contains(l, "persistConn") {
			httpLeaks = append(httpLeaks, l)
		}
	}

	assert.Empty(t, httpLeaks,
		"net/http transport goroutines leaked after compose Down().\n"+
			"These indicate the Docker CLI HTTP client was not closed.\n"+
			"Leaked goroutines:\n%s", strings.Join(httpLeaks, "\n"))
}
