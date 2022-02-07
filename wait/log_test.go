package wait

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/go-connections/nat"
)

type noopStrategyTarget struct {
	ioReaderCloser io.ReadCloser
}

func (st noopStrategyTarget) Host(ctx context.Context) (string, error) {
	return "", nil
}

func (st noopStrategyTarget) MappedPort(ctx context.Context, n nat.Port) (nat.Port, error) {
	return n, nil
}

func (st noopStrategyTarget) Logs(ctx context.Context) (io.ReadCloser, error) {
	return st.ioReaderCloser, nil
}

func (st noopStrategyTarget) Exec(ctx context.Context, cmd []string) (int, error) {
	return 0, nil
}
func (st noopStrategyTarget) State(ctx context.Context) (*types.ContainerState, error) {
	return nil, nil
}

func TestWaitForLog(t *testing.T) {
	target := noopStrategyTarget{
		ioReaderCloser: ioutil.NopCloser(bytes.NewReader([]byte("docker"))),
	}
	wg := NewLogStrategy("docker").WithStartupTimeout(100 * time.Microsecond)
	err := wg.WaitUntilReady(context.Background(), target)
	if err != nil {
		t.Fatal(err)
	}
}

func TestWaitWithExactNumberOfOccurrences(t *testing.T) {
	target := noopStrategyTarget{
		ioReaderCloser: ioutil.NopCloser(bytes.NewReader([]byte("kubernetes\r\ndocker\n\rdocker"))),
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
	target := noopStrategyTarget{
		ioReaderCloser: ioutil.NopCloser(bytes.NewReader([]byte("kubernetes\r\ndocker"))),
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
	target := noopStrategyTarget{
		ioReaderCloser: ioutil.NopCloser(bytes.NewReader([]byte("kubernetes\r\ndocker"))),
	}
	wg := NewLogStrategy("docker").
		WithStartupTimeout(100 * time.Microsecond).
		WithOccurrence(2)
	err := wg.WaitUntilReady(context.Background(), target)
	if err == nil {
		t.Fatal("expected error")
	}
}
