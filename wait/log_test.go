package wait

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/stretchr/testify/require"
)

const logTimeout = time.Second

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
		target := NopStrategyTarget{
			ReaderCloser: io.NopCloser(bytes.NewReader([]byte("docker"))),
		}
		wg := NewLogStrategy("docker").WithStartupTimeout(100 * time.Millisecond)
		err := wg.WaitUntilReady(context.Background(), target)
		require.NoError(t, err)
	})

	t.Run("no regexp", func(t *testing.T) {
		target := NopStrategyTarget{
			ReaderCloser: io.NopCloser(bytes.NewReader([]byte(loremIpsum))),
		}

		// get all words that start with "ip", end with "m" and has a whitespace before the "ip"
		wg := NewLogStrategy(`\sip[\w]+m`).WithStartupTimeout(100 * time.Millisecond).AsRegexp()
		err := wg.WaitUntilReady(context.Background(), target)
		require.NoError(t, err)
	})
}

func TestWaitWithExactNumberOfOccurrences(t *testing.T) {
	t.Run("no regexp", func(t *testing.T) {
		target := NopStrategyTarget{
			ReaderCloser: io.NopCloser(bytes.NewReader([]byte("kubernetes\r\ndocker\n\rdocker"))),
		}
		wg := NewLogStrategy("docker").
			WithStartupTimeout(100 * time.Millisecond).
			WithOccurrence(2)
		err := wg.WaitUntilReady(context.Background(), target)
		require.NoError(t, err)
	})

	t.Run("as regexp", func(t *testing.T) {
		target := NopStrategyTarget{
			ReaderCloser: io.NopCloser(bytes.NewReader([]byte(loremIpsum))),
		}

		// get texts from "ip" to the next "m".
		// there are three occurrences of this pattern in the string:
		// one "ipsum mauris" and two "ipsum dolor sit am"
		wg := NewLogStrategy(`ip(.*)m`).WithStartupTimeout(100 * time.Millisecond).AsRegexp().WithOccurrence(3)
		err := wg.WaitUntilReady(context.Background(), target)
		require.NoError(t, err)
	})
}

func TestWaitWithExactNumberOfOccurrencesButItWillNeverHappen(t *testing.T) {
	t.Run("no regexp", func(t *testing.T) {
		target := NopStrategyTarget{
			ReaderCloser: io.NopCloser(bytes.NewReader([]byte("kubernetes\r\ndocker"))),
		}
		wg := NewLogStrategy("containerd").
			WithStartupTimeout(logTimeout).
			WithOccurrence(2)
		err := wg.WaitUntilReady(context.Background(), target)
		require.Error(t, err)
	})

	t.Run("as regexp", func(t *testing.T) {
		target := NopStrategyTarget{
			ReaderCloser: io.NopCloser(bytes.NewReader([]byte(loremIpsum))),
		}

		// get texts from "ip" to the next "m".
		// there are only three occurrences matching
		wg := NewLogStrategy(`do(.*)ck.+`).WithStartupTimeout(100 * time.Millisecond).AsRegexp().WithOccurrence(4)
		err := wg.WaitUntilReady(context.Background(), target)
		require.Error(t, err)
	})
}

func TestWaitShouldFailWithExactNumberOfOccurrences(t *testing.T) {
	t.Run("no regexp", func(t *testing.T) {
		target := NopStrategyTarget{
			ReaderCloser: io.NopCloser(bytes.NewReader([]byte("kubernetes\r\ndocker"))),
		}
		wg := NewLogStrategy("docker").
			WithStartupTimeout(logTimeout).
			WithOccurrence(2)
		err := wg.WaitUntilReady(context.Background(), target)
		require.Error(t, err)
	})

	t.Run("as regexp", func(t *testing.T) {
		target := NopStrategyTarget{
			ReaderCloser: io.NopCloser(bytes.NewReader([]byte(loremIpsum))),
		}

		// get "Maecenas".
		// there are only one occurrence matching
		wg := NewLogStrategy(`^Mae[\w]?enas\s`).WithStartupTimeout(100 * time.Millisecond).AsRegexp().WithOccurrence(2)
		err := wg.WaitUntilReady(context.Background(), target)
		require.Error(t, err)
	})
}

func TestWaitForLogFailsDueToOOMKilledContainer(t *testing.T) {
	target := &MockStrategyTarget{
		LogsImpl: func(_ context.Context) (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader([]byte(""))), nil
		},
		StateImpl: func(_ context.Context) (*types.ContainerState, error) {
			return &types.ContainerState{
				OOMKilled: true,
			}, nil
		},
	}

	t.Run("no regexp", func(t *testing.T) {
		wg := ForLog("docker").WithStartupTimeout(logTimeout)

		err := wg.WaitUntilReady(context.Background(), target)
		expected := "container crashed with out-of-memory (OOMKilled)"
		require.EqualError(t, err, expected)
	})

	t.Run("as regexp", func(t *testing.T) {
		wg := ForLog("docker").WithStartupTimeout(logTimeout).AsRegexp()

		err := wg.WaitUntilReady(context.Background(), target)
		expected := "container crashed with out-of-memory (OOMKilled)"
		require.EqualError(t, err, expected)
	})
}

func TestWaitForLogFailsDueToExitedContainer(t *testing.T) {
	target := &MockStrategyTarget{
		LogsImpl: func(_ context.Context) (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader([]byte(""))), nil
		},
		StateImpl: func(_ context.Context) (*types.ContainerState, error) {
			return &types.ContainerState{
				Status:   "exited",
				ExitCode: 1,
			}, nil
		},
	}

	t.Run("no regexp", func(t *testing.T) {
		wg := ForLog("docker").WithStartupTimeout(logTimeout)

		err := wg.WaitUntilReady(context.Background(), target)
		expected := "container exited with code 1"
		require.EqualError(t, err, expected)
	})

	t.Run("as regexp", func(t *testing.T) {
		wg := ForLog("docker").WithStartupTimeout(logTimeout).AsRegexp()

		err := wg.WaitUntilReady(context.Background(), target)
		expected := "container exited with code 1"
		require.EqualError(t, err, expected)
	})
}

func TestWaitForLogFailsDueToUnexpectedContainerStatus(t *testing.T) {
	target := &MockStrategyTarget{
		LogsImpl: func(_ context.Context) (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader([]byte(""))), nil
		},
		StateImpl: func(_ context.Context) (*types.ContainerState, error) {
			return &types.ContainerState{
				Status: "dead",
			}, nil
		},
	}

	t.Run("no regexp", func(t *testing.T) {
		wg := ForLog("docker").WithStartupTimeout(logTimeout)

		err := wg.WaitUntilReady(context.Background(), target)
		expected := "unexpected container status \"dead\""
		require.EqualError(t, err, expected)
	})

	t.Run("as regexp", func(t *testing.T) {
		wg := ForLog("docker").WithStartupTimeout(logTimeout).AsRegexp()

		err := wg.WaitUntilReady(context.Background(), target)
		expected := "unexpected container status \"dead\""
		require.EqualError(t, err, expected)
	})
}
