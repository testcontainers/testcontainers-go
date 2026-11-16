// Package mongotest supplies test-only official-shell access to generated MongoDB endpoints.
// It is imported only by tests, never by the module's production packages.
package mongotest

import (
	"context"
	"io"
	"net"
	"sync"
	"time"
)

// relay has exactly one immutable destination: the generated URI's authority,
// resolved from the test process just as it was by the Go client.
type relay struct {
	listener net.Listener
	cancel   context.CancelFunc
	done     chan struct{}
	mu       sync.Mutex
	sockets  map[net.Conn]struct{}
}

func newRelay(ctx context.Context, destination string) (*relay, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(ctx)
	r := &relay{listener: listener, cancel: cancel, done: make(chan struct{}), sockets: make(map[net.Conn]struct{})}
	stopped := make(chan struct{})
	go func() {
		defer close(stopped)
		<-ctx.Done()
		_ = listener.Close()
		r.mu.Lock()
		defer r.mu.Unlock()
		for c := range r.sockets {
			_ = c.Close()
		}
	}()
	go func() {
		defer close(r.done)
		var workers sync.WaitGroup
		for {
			c, err := listener.Accept()
			if err != nil {
				break
			}
			if !r.track(ctx, c) {
				continue
			}
			workers.Add(1)
			go func() {
				defer workers.Done()
				defer r.release(c)
				upstream, err := (&net.Dialer{Timeout: 3 * time.Second}).DialContext(ctx, "tcp", destination)
				if err != nil {
					return
				} // Closing the client makes mongosh report selection failure.
				if !r.track(ctx, upstream) {
					return
				}
				defer r.release(upstream)
				copied := make(chan struct{})
				go func() {
					defer close(copied)
					_, _ = io.Copy(upstream, c)
					_ = c.Close()
					_ = upstream.Close()
				}()
				_, _ = io.Copy(c, upstream)
				_ = c.Close()
				_ = upstream.Close()
				<-copied
			}()
		}
		cancel()
		workers.Wait()
		<-stopped
	}()
	return r, nil
}

func (r *relay) track(ctx context.Context, c net.Conn) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if ctx.Err() != nil {
		_ = c.Close()
		return false
	}
	if deadline, ok := ctx.Deadline(); ok {
		_ = c.SetDeadline(deadline)
	}
	r.sockets[c] = struct{}{}
	return true
}

func (r *relay) release(c net.Conn) {
	_ = c.Close()
	r.mu.Lock()
	delete(r.sockets, c)
	r.mu.Unlock()
}

func (r *relay) close() { r.cancel(); <-r.done }
