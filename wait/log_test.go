package wait_test

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go/wait"
)

const logTimeout = time.Millisecond * 200

const loremIpsum = `Lorem ipsum dolor sit amet,
consectetur adipiscing elit.
Donec a diam lectus.
Sed sit amet ipsum mauris.
Maecenas congue ligula ac quam viverra nec consectetur ante hendrerit.
Donec et mollis dolor.
Praesent et diam eget libero egestas mattis sit amet vitae augue.
Nam tincidunt congue enim, ut porta lorem lacinia consectetur.
Donec ut libero sed arcu vehicula ultricies a non tortor.
Lorem ipsum dolor sit amet, consectetur adipiscing elit.`

func TestWaitForLog(t *testing.T) {
	t.Run("no regexp", func(t *testing.T) {
		target := wait.NopStrategyTarget{
			ReaderCloser:   io.NopCloser(bytes.NewReader([]byte("docker"))),
			ContainerState: types.ContainerState{Running: true},
		}
		wg := wait.NewLogStrategy("docker").WithStartupTimeout(100 * time.Millisecond)
		err := wg.WaitUntilReady(context.Background(), target)
		require.NoError(t, err)
	})

	t.Run("no regexp", func(t *testing.T) {
		target := wait.NopStrategyTarget{
			ReaderCloser:   io.NopCloser(bytes.NewReader([]byte(loremIpsum))),
			ContainerState: types.ContainerState{Running: true},
		}

		// get all words that start with "ip", end with "m" and has a whitespace before the "ip"
		wg := wait.NewLogStrategy(`\sip[\w]+m`).WithStartupTimeout(100 * time.Millisecond).AsRegexp()
		err := wg.WaitUntilReady(context.Background(), target)
		require.NoError(t, err)
	})
}

func TestWaitWithExactNumberOfOccurrences(t *testing.T) {
	t.Run("no regexp", func(t *testing.T) {
		target := wait.NopStrategyTarget{
			ReaderCloser:   io.NopCloser(bytes.NewReader([]byte("kubernetes\r\ndocker\n\rdocker"))),
			ContainerState: types.ContainerState{Running: true},
		}
		wg := wait.NewLogStrategy("docker").
			WithStartupTimeout(100 * time.Millisecond).
			WithOccurrence(2)
		err := wg.WaitUntilReady(context.Background(), target)
		require.NoError(t, err)
	})

	t.Run("as regexp", func(t *testing.T) {
		target := wait.NopStrategyTarget{
			ReaderCloser:   io.NopCloser(bytes.NewReader([]byte(loremIpsum))),
			ContainerState: types.ContainerState{Running: true},
		}

		// get texts from "ip" to the next "m".
		// there are three occurrences of this pattern in the string:
		// one "ipsum mauris" and two "ipsum dolor sit am"
		wg := wait.NewLogStrategy(`ip(.*)m`).WithStartupTimeout(100 * time.Millisecond).AsRegexp().WithOccurrence(3)
		err := wg.WaitUntilReady(context.Background(), target)
		require.NoError(t, err)
	})
}

func TestWaitWithExactNumberOfOccurrencesButItWillNeverHappen(t *testing.T) {
	t.Run("no regexp", func(t *testing.T) {
		target := wait.NopStrategyTarget{
			ReaderCloser:   io.NopCloser(bytes.NewReader([]byte("kubernetes\r\ndocker"))),
			ContainerState: types.ContainerState{Running: true},
		}
		wg := wait.NewLogStrategy("containerd").
			WithStartupTimeout(logTimeout).
			WithOccurrence(2)
		err := wg.WaitUntilReady(context.Background(), target)
		require.Error(t, err)
	})

	t.Run("as regexp", func(t *testing.T) {
		target := wait.NopStrategyTarget{
			ReaderCloser:   io.NopCloser(bytes.NewReader([]byte(loremIpsum))),
			ContainerState: types.ContainerState{Running: true},
		}

		// get texts from "ip" to the next "m".
		// there are only three occurrences matching
		wg := wait.NewLogStrategy(`do(.*)ck.+`).WithStartupTimeout(100 * time.Millisecond).AsRegexp().WithOccurrence(4)
		err := wg.WaitUntilReady(context.Background(), target)
		require.Error(t, err)
	})
}

func TestWaitShouldFailWithExactNumberOfOccurrences(t *testing.T) {
	t.Run("no regexp", func(t *testing.T) {
		target := wait.NopStrategyTarget{
			ReaderCloser:   io.NopCloser(bytes.NewReader([]byte("kubernetes\r\ndocker"))),
			ContainerState: types.ContainerState{Running: true},
		}
		wg := wait.NewLogStrategy("docker").
			WithStartupTimeout(logTimeout).
			WithOccurrence(2)
		err := wg.WaitUntilReady(context.Background(), target)
		require.Error(t, err)
	})

	t.Run("as regexp", func(t *testing.T) {
		target := wait.NopStrategyTarget{
			ReaderCloser:   io.NopCloser(bytes.NewReader([]byte(loremIpsum))),
			ContainerState: types.ContainerState{Running: true},
		}

		// get "Maecenas".
		// there are only one occurrence matching
		wg := wait.NewLogStrategy(`^Mae[\w]?enas\s`).WithStartupTimeout(100 * time.Millisecond).AsRegexp().WithOccurrence(2)
		err := wg.WaitUntilReady(context.Background(), target)
		require.Error(t, err)
	})
}

// testForLog tests the given strategy with different container state scenarios.
func testForLog(t *testing.T, strategy wait.Strategy) {
	t.Helper()

	reader := strings.NewReader("")
	t.Run("running", func(t *testing.T) {
		reader.Reset(loremIpsum)
		newWaitBuilder().Logs(reader).Run(t, strategy)
	})

	t.Run("oom", func(t *testing.T) {
		reader.Reset(loremIpsum)
		newWaitBuilder().State(oom).Logs(reader).Run(t, strategy)
	})

	t.Run("exited", func(t *testing.T) {
		reader.Reset(loremIpsum)
		newWaitBuilder().State(exited).Logs(reader).Run(t, strategy)
	})

	t.Run("dead", func(t *testing.T) {
		reader.Reset(loremIpsum)
		newWaitBuilder().State(dead).Logs(reader).Run(t, strategy)
	})
}

func TestForLog(t *testing.T) {
	t.Run("no-regexp", func(t *testing.T) {
		strategy := wait.ForLog("ipsum").WithStartupTimeout(logTimeout)
		testForLog(t, strategy)
	})

	t.Run("as-regexp", func(t *testing.T) {
		strategy := wait.ForLog("i.sum").WithStartupTimeout(logTimeout).AsRegexp()
		testForLog(t, strategy)
	})
}
