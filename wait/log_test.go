package wait_test

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go/wait"
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
	t.Run("string", func(t *testing.T) {
		target := wait.NopStrategyTarget{
			ReaderCloser: readCloser("docker"),
		}
		wg := wait.NewLogStrategy("docker").WithStartupTimeout(100 * time.Millisecond)
		err := wg.WaitUntilReady(context.Background(), target)
		require.NoError(t, err)
	})

	t.Run("regexp", func(t *testing.T) {
		target := wait.NopStrategyTarget{
			ReaderCloser: readCloser(loremIpsum),
		}

		// get all words that start with "ip", end with "m" and has a whitespace before the "ip"
		wg := wait.NewLogStrategy(`\sip[\w]+m`).WithStartupTimeout(100 * time.Millisecond).AsRegexp()
		err := wg.WaitUntilReady(context.Background(), target)
		require.NoError(t, err)
	})

	t.Run("submatch/valid", func(t *testing.T) {
		target := wait.NopStrategyTarget{
			ReaderCloser: readCloser("three matches: ip1m, ip2m, ip3m"),
		}

		wg := wait.NewLogStrategy(`ip(\d)m`).WithStartupTimeout(100 * time.Millisecond).Submatch(func(pattern string, submatches [][][]byte) error {
			if len(submatches) != 3 {
				return wait.NewPermanentError(fmt.Errorf("%q matched %d times, expected %d", pattern, len(submatches), 3))
			}
			return nil
		})
		err := wg.WaitUntilReady(context.Background(), target)
		require.NoError(t, err)
	})

	t.Run("submatch/permanent-error", func(t *testing.T) {
		target := wait.NopStrategyTarget{
			ReaderCloser: readCloser("single matches: ip1m"),
		}

		wg := wait.NewLogStrategy(`ip(\d)m`).WithStartupTimeout(100 * time.Millisecond).Submatch(func(pattern string, submatches [][][]byte) error {
			if len(submatches) != 3 {
				return wait.NewPermanentError(fmt.Errorf("%q matched %d times, expected %d", pattern, len(submatches), 3))
			}
			return nil
		})
		err := wg.WaitUntilReady(context.Background(), target)
		require.Error(t, err)
		var permanentError *wait.PermanentError
		require.ErrorAs(t, err, &permanentError)
	})

	t.Run("submatch/temporary-error", func(t *testing.T) {
		target := newRunningTarget()
		expect := target.EXPECT()
		expect.Logs(anyContext).Return(readCloser(""), nil).Once()                 // No matches.
		expect.Logs(anyContext).Return(readCloser("ip1m, ip2m"), nil).Once()       // Two matches.
		expect.Logs(anyContext).Return(readCloser("ip1m, ip2m, ip3m"), nil).Once() // Three matches.
		expect.Logs(anyContext).Return(readCloser("ip1m, ip2m, ip3m, ip4m"), nil)  // Four matches.

		wg := wait.NewLogStrategy(`ip(\d)m`).WithStartupTimeout(400 * time.Second).Submatch(func(pattern string, submatches [][][]byte) error {
			switch len(submatches) {
			case 0, 2:
				// Too few matches.
				return fmt.Errorf("`%s` matched %d times, expected %d (temporary)", pattern, len(submatches), 3)
			case 3:
				// Expected number of matches should stop the wait.
				return nil
			default:
				// Should not be triggered.
				return wait.NewPermanentError(fmt.Errorf("`%s` matched %d times, expected %d (permanent)", pattern, len(submatches), 3))
			}
		})
		err := wg.WaitUntilReady(context.Background(), target)
		require.NoError(t, err)
	})
}

func TestWaitWithExactNumberOfOccurrences(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		target := wait.NopStrategyTarget{
			ReaderCloser: readCloser("kubernetes\r\ndocker\n\rdocker"),
		}
		wg := wait.NewLogStrategy("docker").
			WithStartupTimeout(100 * time.Millisecond).
			WithOccurrence(2)
		err := wg.WaitUntilReady(context.Background(), target)
		require.NoError(t, err)
	})

	t.Run("regexp", func(t *testing.T) {
		target := wait.NopStrategyTarget{
			ReaderCloser: readCloser(loremIpsum),
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
	t.Run("string", func(t *testing.T) {
		target := wait.NopStrategyTarget{
			ReaderCloser: readCloser("kubernetes\r\ndocker"),
		}
		wg := wait.NewLogStrategy("containerd").
			WithStartupTimeout(logTimeout).
			WithOccurrence(2)
		err := wg.WaitUntilReady(context.Background(), target)
		require.Error(t, err)
	})

	t.Run("regexp", func(t *testing.T) {
		target := wait.NopStrategyTarget{
			ReaderCloser: readCloser(loremIpsum),
		}

		// get texts from "ip" to the next "m".
		// there are only three occurrences matching
		wg := wait.NewLogStrategy(`do(.*)ck.+`).WithStartupTimeout(100 * time.Millisecond).AsRegexp().WithOccurrence(4)
		err := wg.WaitUntilReady(context.Background(), target)
		require.Error(t, err)
	})
}

func TestWaitShouldFailWithExactNumberOfOccurrences(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		target := wait.NopStrategyTarget{
			ReaderCloser: readCloser("kubernetes\r\ndocker"),
		}
		wg := wait.NewLogStrategy("docker").
			WithStartupTimeout(logTimeout).
			WithOccurrence(2)
		err := wg.WaitUntilReady(context.Background(), target)
		require.Error(t, err)
	})

	t.Run("regexp", func(t *testing.T) {
		target := wait.NopStrategyTarget{
			ReaderCloser: readCloser(loremIpsum),
		}

		// get "Maecenas".
		// there are only one occurrence matching
		wg := wait.NewLogStrategy(`^Mae[\w]?enas\s`).WithStartupTimeout(100 * time.Millisecond).AsRegexp().WithOccurrence(2)
		err := wg.WaitUntilReady(context.Background(), target)
		require.Error(t, err)
	})
}

func TestWaitForLogFailsDueToOOMKilledContainer(t *testing.T) {
	target := &wait.MockStrategyTarget{
		LogsImpl: func(_ context.Context) (io.ReadCloser, error) {
			return readCloser(""), nil
		},
		StateImpl: func(_ context.Context) (*container.State, error) {
			return &container.State{
				OOMKilled: true,
			}, nil
		},
	}

	t.Run("string", func(t *testing.T) {
		wg := wait.ForLog("docker").WithStartupTimeout(logTimeout)

		err := wg.WaitUntilReady(context.Background(), target)
		expected := "container crashed with out-of-memory (OOMKilled)"
		require.EqualError(t, err, expected)
	})

	t.Run("regexp", func(t *testing.T) {
		wg := wait.ForLog("docker").WithStartupTimeout(logTimeout).AsRegexp()

		err := wg.WaitUntilReady(context.Background(), target)
		expected := "container crashed with out-of-memory (OOMKilled)"
		require.EqualError(t, err, expected)
	})
}

func TestWaitForLogFailsDueToExitedContainer(t *testing.T) {
	target := &wait.MockStrategyTarget{
		LogsImpl: func(_ context.Context) (io.ReadCloser, error) {
			return readCloser(""), nil
		},
		StateImpl: func(_ context.Context) (*container.State, error) {
			return &container.State{
				Status:   "exited",
				ExitCode: 1,
			}, nil
		},
	}

	t.Run("string", func(t *testing.T) {
		wg := wait.ForLog("docker").WithStartupTimeout(logTimeout)

		err := wg.WaitUntilReady(context.Background(), target)
		expected := "container exited with code 1"
		require.EqualError(t, err, expected)
	})

	t.Run("regexp", func(t *testing.T) {
		wg := wait.ForLog("docker").WithStartupTimeout(logTimeout).AsRegexp()

		err := wg.WaitUntilReady(context.Background(), target)
		expected := "container exited with code 1"
		require.EqualError(t, err, expected)
	})
}

func TestWaitForLogFailsDueToUnexpectedContainerStatus(t *testing.T) {
	target := &wait.MockStrategyTarget{
		LogsImpl: func(_ context.Context) (io.ReadCloser, error) {
			return readCloser(""), nil
		},
		StateImpl: func(_ context.Context) (*container.State, error) {
			return &container.State{
				Status: "dead",
			}, nil
		},
	}

	t.Run("string", func(t *testing.T) {
		wg := wait.ForLog("docker").WithStartupTimeout(logTimeout)

		err := wg.WaitUntilReady(context.Background(), target)
		expected := "unexpected container status \"dead\""
		require.EqualError(t, err, expected)
	})

	t.Run("regexp", func(t *testing.T) {
		wg := wait.ForLog("docker").WithStartupTimeout(logTimeout).AsRegexp()

		err := wg.WaitUntilReady(context.Background(), target)
		expected := "unexpected container status \"dead\""
		require.EqualError(t, err, expected)
	})
}

// readCloser returns an io.ReadCloser that reads from s.
func readCloser(s string) io.ReadCloser {
	return io.NopCloser(strings.NewReader((s)))
}
