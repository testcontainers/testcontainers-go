package testcontainers

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cenkalti/backoff/v4"
)

// A blackholed endpoint (SYNs dropped, no RST) must fail within the 5s dial
// bound and be classified retryable, so the spawner's 20s backoff can retry.
func TestReaperConnectDialBounded(t *testing.T) {
	r := &Reaper{Endpoint: "192.0.2.1:8080"} // TEST-NET-1, not routed
	start := time.Now()
	_, err := r.connect(context.Background())
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected dial error")
	}
	if elapsed > 10*time.Second {
		t.Fatalf("dial not bounded: took %v", elapsed)
	}
	s := &reaperSpawner{}
	rerr := s.retryError(err)
	var perm *backoff.PermanentError
	if errors.As(rerr, &perm) {
		t.Fatalf("timeout classified permanent, backoff would stop: %v", rerr)
	}
	t.Logf("dial failed after %v, retryable: %v", elapsed, err)
}
