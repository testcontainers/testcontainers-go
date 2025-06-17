package core

import (
	"errors"
	"fmt"
	"maps"
	"strings"
	"sync"

	"github.com/testcontainers/testcontainers-go/internal"
	"github.com/testcontainers/testcontainers-go/internal/config"
)

const (
	// LabelBase is the base label for all testcontainers labels.
	LabelBase = "org.testcontainers"

	// LabelLang specifies the language which created the test container.
	LabelLang = LabelBase + ".lang"

	// LabelReaper identifies the container as a reaper.
	LabelReaper = LabelBase + ".reaper"

	// LabelRyuk identifies the container as a ryuk.
	LabelRyuk = LabelBase + ".ryuk"

	// LabelSessionID specifies the session ID of the container.
	LabelSessionID = LabelBase + ".sessionId"

	// LabelVersion specifies the version of testcontainers which created the container.
	LabelVersion = LabelBase + ".version"

	// LabelReap specifies the container should be reaped by the reaper.
	LabelReap = LabelBase + ".reap"
)

// labelMerger provides thread-safe operations for merging labels
type labelMerger struct {
	mu     sync.RWMutex
	labels map[string]string
}

// newLabelMerger creates a new thread-safe label merger
func newLabelMerger(initial map[string]string) *labelMerger {
	if initial == nil {
		initial = make(map[string]string)
	}
	return &labelMerger{
		labels: initial,
	}
}

// MergeCustomLabelsSafeMerge merges labels from src to dst.
// If a key in src has [LabelBase] prefix returns an error.
func (m *labelMerger) Merge(src map[string]string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for key, value := range src {
		if strings.HasPrefix(key, LabelBase) {
			return fmt.Errorf("key %q has %q prefix", key, LabelBase)
		}
		m.labels[key] = value
	}
	return nil
}

// Labels returns a copy of the current labels
func (m *labelMerger) Labels() map[string]string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	labels := make(map[string]string, len(m.labels))
	for k, v := range m.labels {
		labels[k] = v
	}

	return labels
}

// DefaultLabels returns the standard set of labels which
// includes LabelSessionID if the reaper is enabled.
func DefaultLabels(sessionID string) map[string]string {
	labels := map[string]string{
		LabelBase:      "true",
		LabelLang:      "go",
		LabelVersion:   internal.Version,
		LabelSessionID: sessionID,
	}

	if !config.Read().RyukDisabled {
		labels[LabelReap] = "true"
	}

	return labels
}

// AddDefaultLabels adds the default labels for sessionID to target.
func AddDefaultLabels(sessionID string, target map[string]string) {
	maps.Copy(target, DefaultLabels(sessionID))
}

// MergeCustomLabels sets labels from src to dst.
// If a key in src has [LabelBase] prefix returns an error.
// If dst is nil returns an error.
func MergeCustomLabels(dst, src map[string]string) error {
	if dst == nil {
		return errors.New("destination map is nil")
	}

	merger := newLabelMerger(dst)
	return merger.Merge(src)
}
