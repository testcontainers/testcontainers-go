package wait

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
)

func TestWaitForLog(t *testing.T) {
	target := NopStrategyTarget{
		ReaderCloser: io.NopCloser(bytes.NewReader([]byte("docker"))),
	}
	wg := NewLogStrategy("docker").WithStartupTimeout(100 * time.Microsecond)
	err := wg.WaitUntilReady(context.Background(), target)
	if err != nil {
		t.Fatal(err)
	}
}

func TestWaitWithExactNumberOfOccurrences(t *testing.T) {
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
}

func TestWaitWithExactNumberOfOccurrencesButItWillNeverHappen(t *testing.T) {
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
}

func TestWaitShouldFailWithExactNumberOfOccurrences(t *testing.T) {
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
}
