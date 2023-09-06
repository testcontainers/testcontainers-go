package wait

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
)

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
		wg := NewLogStrategy("docker").WithStartupTimeout(100 * time.Microsecond)
		err := wg.WaitUntilReady(context.Background(), target)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("no regexp", func(t *testing.T) {
		target := NopStrategyTarget{
			ReaderCloser: io.NopCloser(bytes.NewReader([]byte(loremIpsum))),
		}

		// get all words that start with "ip", end with "m" and has a whitespace before the "ip"
		wg := NewLogStrategy(`\sip[\w]+m`).WithStartupTimeout(100 * time.Microsecond).AsRegexp()
		err := wg.WaitUntilReady(context.Background(), target)
		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestWaitWithExactNumberOfOccurrences(t *testing.T) {
	t.Run("no regexp", func(t *testing.T) {
		target := NopStrategyTarget{
			ReaderCloser: io.NopCloser(bytes.NewReader([]byte("kubernetes\r\ndocker\n\rdocker"))),
		}
		wg := NewLogStrategy("docker").
			WithStartupTimeout(100 * time.Microsecond).
			WithOccurrence(2)
		err := wg.WaitUntilReady(context.Background(), target)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("as regexp", func(t *testing.T) {
		target := NopStrategyTarget{
			ReaderCloser: io.NopCloser(bytes.NewReader([]byte(loremIpsum))),
		}

		// get texts from "ip" to the next "m".
		// there are three occurrences of this pattern in the string:
		// one "ipsum mauris" and two "ipsum dolor sit am"
		wg := NewLogStrategy(`ip(.*)m`).WithStartupTimeout(100 * time.Microsecond).AsRegexp().WithOccurrence(3)
		err := wg.WaitUntilReady(context.Background(), target)
		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestWaitWithExactNumberOfOccurrencesButItWillNeverHappen(t *testing.T) {
	t.Run("no regexp", func(t *testing.T) {
		target := NopStrategyTarget{
			ReaderCloser: io.NopCloser(bytes.NewReader([]byte("kubernetes\r\ndocker"))),
		}
		wg := NewLogStrategy("containerd").
			WithStartupTimeout(100 * time.Microsecond).
			WithOccurrence(2)
		err := wg.WaitUntilReady(context.Background(), target)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("as regexp", func(t *testing.T) {
		target := NopStrategyTarget{
			ReaderCloser: io.NopCloser(bytes.NewReader([]byte(loremIpsum))),
		}

		// get texts from "ip" to the next "m".
		// there are only three occurrences matching
		wg := NewLogStrategy(`do(.*)ck.+`).WithStartupTimeout(100 * time.Microsecond).AsRegexp().WithOccurrence(4)
		err := wg.WaitUntilReady(context.Background(), target)
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestWaitShouldFailWithExactNumberOfOccurrences(t *testing.T) {
	t.Run("no regexp", func(t *testing.T) {
		target := NopStrategyTarget{
			ReaderCloser: io.NopCloser(bytes.NewReader([]byte("kubernetes\r\ndocker"))),
		}
		wg := NewLogStrategy("docker").
			WithStartupTimeout(100 * time.Microsecond).
			WithOccurrence(2)
		err := wg.WaitUntilReady(context.Background(), target)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("as regexp", func(t *testing.T) {
		target := NopStrategyTarget{
			ReaderCloser: io.NopCloser(bytes.NewReader([]byte(loremIpsum))),
		}

		// get "Maecenas".
		// there are only one occurrence matching
		wg := NewLogStrategy(`^Mae[\w]?enas\s`).WithStartupTimeout(100 * time.Microsecond).AsRegexp().WithOccurrence(2)
		err := wg.WaitUntilReady(context.Background(), target)
		if err == nil {
			t.Fatal("expected error")
		}
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
		wg := ForLog("docker").
			WithStartupTimeout(100 * time.Microsecond)

		{
			err := wg.WaitUntilReady(context.Background(), target)
			if err == nil {
				t.Fatal("no error")
			}

			expected := "container crashed with out-of-memory (OOMKilled)"
			if err.Error() != expected {
				t.Fatalf("expected %q, got %q", expected, err.Error())
			}
		}
	})

	t.Run("as regexp", func(t *testing.T) {
		wg := ForLog("docker").
			WithStartupTimeout(100 * time.Microsecond).AsRegexp()

		{
			err := wg.WaitUntilReady(context.Background(), target)
			if err == nil {
				t.Fatal("no error")
			}

			expected := "container crashed with out-of-memory (OOMKilled)"
			if err.Error() != expected {
				t.Fatalf("expected %q, got %q", expected, err.Error())
			}
		}
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
		wg := ForLog("docker").
			WithStartupTimeout(100 * time.Microsecond)

		{
			err := wg.WaitUntilReady(context.Background(), target)
			if err == nil {
				t.Fatal("no error")
			}

			expected := "container exited with code 1"
			if err.Error() != expected {
				t.Fatalf("expected %q, got %q", expected, err.Error())
			}
		}
	})

	t.Run("as regexp", func(t *testing.T) {
		wg := ForLog("docker").
			WithStartupTimeout(100 * time.Microsecond).AsRegexp()

		{
			err := wg.WaitUntilReady(context.Background(), target)
			if err == nil {
				t.Fatal("no error")
			}

			expected := "container exited with code 1"
			if err.Error() != expected {
				t.Fatalf("expected %q, got %q", expected, err.Error())
			}
		}
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
		wg := ForLog("docker").
			WithStartupTimeout(100 * time.Microsecond)

		{
			err := wg.WaitUntilReady(context.Background(), target)
			if err == nil {
				t.Fatal("no error")
			}

			expected := "unexpected container status \"dead\""
			if err.Error() != expected {
				t.Fatalf("expected %q, got %q", expected, err.Error())
			}
		}
	})

	t.Run("as regexp", func(t *testing.T) {
		wg := ForLog("docker").
			WithStartupTimeout(100 * time.Microsecond).AsRegexp()

		{
			err := wg.WaitUntilReady(context.Background(), target)
			if err == nil {
				t.Fatal("no error")
			}

			expected := "unexpected container status \"dead\""
			if err.Error() != expected {
				t.Fatalf("expected %q, got %q", expected, err.Error())
			}
		}
	})
}
