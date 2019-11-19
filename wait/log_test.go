package wait

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"testing"
	"time"

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

func TestWaitForLog(t *testing.T) {
	target := noopStrategyTarget{
		ioReaderCloser: ioutil.NopCloser(bytes.NewReader([]byte("dude"))),
	}
	wg := NewLogStrategy("dude").WithStartupTimeout(100 * time.Microsecond)
	err := wg.WaitUntilReady(context.Background(), target)
	if err != nil {
		t.Fatal(err)
	}
}

func TestWaitWithMaxOccurrence(t *testing.T) {
	target := noopStrategyTarget{
		ioReaderCloser: ioutil.NopCloser(bytes.NewReader([]byte("hello\r\ndude\n\rdude"))),
	}
	wg := NewLogStrategy("dude").
		WithStartupTimeout(100 * time.Microsecond).
		WithOccurrence(2)
	err := wg.WaitUntilReady(context.Background(), target)
	if err != nil {
		t.Fatal(err)
	}
}

func TestWaitWithMaxOccurrenceButItWillNeverHappen(t *testing.T) {
	target := noopStrategyTarget{
		ioReaderCloser: ioutil.NopCloser(bytes.NewReader([]byte("hello\r\ndude"))),
	}
	wg := NewLogStrategy("blaaa").
		WithStartupTimeout(100 * time.Microsecond).
		WithOccurrence(2)
	err := wg.WaitUntilReady(context.Background(), target)
	if err == nil {
		t.Fatal("expected error")
	}
}
