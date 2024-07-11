package core

import (
	"fmt"
	"strings"

	"github.com/testcontainers/testcontainers-go/internal"
)

const (
	LabelBase      = "org.testcontainers"
	LabelLang      = LabelBase + ".lang"
	LabelReaper    = LabelBase + ".reaper"
	LabelRyuk      = LabelBase + ".ryuk"
	LabelSessionID = LabelBase + ".sessionId"
	LabelVersion   = LabelBase + ".version"
)

// DefaultLabels returns the default map of Labels for a container
func DefaultLabels(sessionID string) map[string]string {
	return map[string]string{
		LabelBase:      "true",
		LabelLang:      "go",
		LabelSessionID: sessionID,
		LabelVersion:   internal.Version,
	}
}

// MergeCustomLabels merges default labels in-place to the custom labels.
//
// It is an error to use "org.testcontainers" as a prefix to custom labels.
func MergeCustomLabels(customLabels map[string]string, defaultLabels map[string]string) error {
	for customLabelKey := range customLabels {
		_, present := defaultLabels[customLabelKey]
		if present || strings.HasPrefix(customLabelKey, LabelBase) {
			return fmt.Errorf("custom labels cannot begin with %s or already be present", LabelBase)
		}
	}
	for defaultLabel, defaultValue := range defaultLabels {
		customLabels[defaultLabel] = defaultValue
	}

	return nil
}
