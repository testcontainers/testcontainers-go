// Package testcbt supports the parent Bigtable behavior test without a Go SDK.
package testcbt

import (
	"context"
	"io"
	"net"
	"sync"
	"time"
)

// Relay forwards the host-access bridge to one unchanged generated endpoint.
// It owns its listener, connections and copying goroutines until Close returns.
type Relay struct {
	listener net.Listener
	target   string
	ctx      context.Context
	cancel   context.CancelFunc
	mu       sync.Mutex
	conns    map[net.Conn]struct{}
	closed   bool
	wg       sync.WaitGroup
	once     sync.Once
}

func NewRelay(target string) (*Relay, error) {
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	r := &Relay{listener: listener, target: target, ctx: ctx, cancel: cancel, conns: make(map[net.Conn]struct{})}
	r.wg.Add(1)
	go r.accept()
	return r, nil
}

func (r *Relay) Port() int {
	return r.listener.Addr().(*net.TCPAddr).Port
}

func (r *Relay) track(conn net.Conn) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		conn.Close()
		return false
	}
	r.conns[conn] = struct{}{}
	return true
}

func (r *Relay) release(conn net.Conn) {
	conn.Close()
	r.mu.Lock()
	delete(r.conns, conn)
	r.mu.Unlock()
}

func (r *Relay) accept() {
	defer r.wg.Done()
	for {
		conn, err := r.listener.Accept()
		if err != nil {
			return
		}
		if !r.track(conn) {
			return
		}
		r.wg.Add(1)
		go r.forward(conn)
	}
}

func (r *Relay) forward(client net.Conn) {
	defer r.wg.Done()
	defer r.release(client)
	upstream, err := (&net.Dialer{Timeout: 2 * time.Second}).DialContext(r.ctx, "tcp", r.target)
	if err != nil {
		return
	}
	if !r.track(upstream) {
		return
	}
	defer r.release(upstream)
	done := make(chan struct{})
	go func() {
		defer close(done)
		// EOF and copy errors both terminate the stream; cbt reports failed RPCs.
		_, _ = io.Copy(upstream, client)
		// End the opposite copy even when the peer keeps its socket open.
		upstream.Close()
		client.Close()
	}()
	_, _ = io.Copy(client, upstream)
	client.Close()
	upstream.Close()
	<-done
}

func (r *Relay) Close() {
	r.once.Do(func() {
		r.cancel()
		r.listener.Close()
		r.mu.Lock()
		r.closed = true
		for conn := range r.conns {
			conn.Close()
		}
		r.mu.Unlock()
		r.wg.Wait()
	})
}
