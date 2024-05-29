package reaper

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/testcontainers/testcontainers-go/internal/config"
	"github.com/testcontainers/testcontainers-go/internal/core"
)

var mu sync.Mutex

type reaperMetaData struct {
	Endpoint  string
	SessionID string
}

var reaperInstance reaperMetaData // initialized with zero values

// ResetReaper resets the reaper instance.
// This is useful for tests that need to reset the reaper state,
// so do not use it in production code.
func ResetReaper() {
	mu.Lock()
	defer mu.Unlock()

	reaperInstance = reaperMetaData{}
}

// InitReaper initialises the reaper metadata instance needed
// to connect to the reaper, which is just the endpoint and the session ID.
func InitReaper(endpoint string, sessionID string) {
	mu.Lock()
	defer mu.Unlock()

	cfg := config.Read()
	if cfg.RyukDisabled {
		// calls to new when disabled are ignored
		return
	}

	reaperInstance = reaperMetaData{
		Endpoint:  endpoint,
		SessionID: sessionID,
	}
}

// Connect runs a goroutine which can be terminated by sending true into the returned channel
func Connect() (chan bool, error) {
	mu.Lock()
	defer mu.Unlock()

	cfg := config.Read()
	if cfg.RyukDisabled {
		// calls to connect when disabled are ignored
		return nil, nil
	}

	reaperEndpoint := reaperInstance.Endpoint
	sessionID := reaperInstance.SessionID

	conn, err := net.DialTimeout("tcp", reaperEndpoint, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("%w: Connecting to Ryuk on %s failed", err, reaperEndpoint)
	}

	terminationSignal := make(chan bool)
	go func(conn net.Conn) {
		sock := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
		defer conn.Close()

		labelFilters := []string{}
		for l, v := range core.DefaultLabels(sessionID) {
			labelFilters = append(labelFilters, fmt.Sprintf("label=%s=%s", l, v))
		}

		retryLimit := 3
		for retryLimit > 0 {
			retryLimit--

			if _, err := sock.WriteString(strings.Join(labelFilters, "&")); err != nil {
				continue
			}

			if _, err := sock.WriteString("\n"); err != nil {
				continue
			}

			if err := sock.Flush(); err != nil {
				continue
			}

			resp, err := sock.ReadString('\n')
			if err != nil {
				continue
			}

			if resp == "ACK\n" {
				break
			}
		}

		<-terminationSignal
	}(conn)
	return terminationSignal, nil
}
