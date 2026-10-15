package sbx

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const eventReconnectDelay = time.Second

type dockerEvent struct {
	Status string `json:"status"`
	Action string `json:"Action"`
	ID     string `json:"id"`
	Actor  struct {
		ID         string            `json:"ID"`
		Attributes map[string]string `json:"Attributes"`
	} `json:"Actor"`
}

func (m *portManager) startWatchers() {
	m.watchWG.Add(2)
	go m.periodicReconciler()
	go m.eventWatcher()
}

func (m *portManager) periodicReconciler() {
	defer m.watchWG.Done()
	ticker := time.NewTicker(portReconcilePeriod)
	defer ticker.Stop()
	for {
		select {
		case <-m.watchCtx.Done():
			return
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(m.watchCtx, portOperationTimeout)
			_ = m.reconcile(ctx)
			cancel()
		}
	}
}

func (m *portManager) eventWatcher() {
	defer m.watchWG.Done()
	for m.watchCtx.Err() == nil {
		_ = m.watchEventsOnce(m.watchCtx)
		timer := time.NewTimer(eventReconnectDelay)
		select {
		case <-m.watchCtx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			return
		case <-timer.C:
		}
		ctx, cancel := context.WithTimeout(m.watchCtx, portOperationTimeout)
		_ = m.reconcile(ctx)
		cancel()
	}
}

func (m *portManager) watchEventsOnce(ctx context.Context) error {
	filters, err := json.Marshal(map[string][]string{
		"type":  {"container"},
		"event": {"destroy", "die", "stop", "kill", "rename"},
	})
	if err != nil {
		return err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, m.engine.endpoint("/events", url.Values{"filters": {string(filters)}}), nil)
	if err != nil {
		return err
	}
	response, err := m.client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		body, readErr := io.ReadAll(io.LimitReader(response.Body, dockerResponseLimit))
		if readErr != nil {
			return readErr
		}
		return fmt.Errorf("docker events returned HTTP %d: %s", response.StatusCode, strings.TrimSpace(string(body)))
	}
	scanner := bufio.NewScanner(response.Body)
	scanner.Buffer(make([]byte, 4096), dockerResponseLimit)
	for scanner.Scan() {
		var event dockerEvent
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			continue
		}
		m.handleDockerEvent(ctx, event)
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return io.EOF
}

func (m *portManager) handleDockerEvent(ctx context.Context, event dockerEvent) {
	id := event.ID
	if id == "" {
		id = event.Actor.ID
	}
	if id == "" {
		return
	}
	token := event.Actor.Attributes[portLeaseLabel]
	if token != "" {
		m.bindObservedContainer(token, id, event.Actor.Attributes["name"])
	}
	m.mu.Lock()
	record := m.containers[id]
	m.mu.Unlock()
	if record == nil {
		return
	}
	opCtx, cancel := boundedContext(ctx, portOperationTimeout)
	_ = m.reconcileRecord(opCtx, record)
	cancel()
}

func (m *portManager) bindObservedContainer(token, id, name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.containers[id] != nil {
		return
	}
	record := m.creations[token]
	if record == nil {
		return
	}
	delete(m.creations, token)
	record.id, record.name = id, strings.TrimPrefix(name, "/")
	m.containers[id] = record
}
