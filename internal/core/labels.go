package core

import (
	"github.com/testcontainers/testcontainers-go/internal"
	"github.com/testcontainers/testcontainers-go/internal/config"
)

const (
	LabelBase      = "org.testcontainers"
	LabelLang      = LabelBase + ".lang"
	LabelReaper    = LabelBase + ".reaper"
	LabelRyuk      = LabelBase + ".ryuk"
	LabelSessionID = LabelBase + ".sessionId"
	LabelVersion   = LabelBase + ".version"
)

// DefaultLabels returns the standard set of labels which
// includes LabelSessionID if the reaper is enabled.
func DefaultLabels(sessionID string) map[string]string {
	labels := map[string]string{
		LabelBase:    "true",
		LabelLang:    "go",
		LabelVersion: internal.Version,
	}

	if !config.Read().RyukDisabled {
		labels[LabelSessionID] = sessionID
	}

	return labels
}

// AddDefaultLabels adds the default labels for sessionID to target.
func AddDefaultLabels(sessionID string, target map[string]string) {
	for k, v := range DefaultLabels(sessionID) {
		target[k] = v
	}
}
