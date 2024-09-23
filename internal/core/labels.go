package core

import (
	"errors"
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

func DefaultLabels(sessionID string) map[string]string {
	return map[string]string{
		LabelBase:      "true",
		LabelLang:      "go",
		LabelSessionID: sessionID,
		LabelVersion:   internal.Version,
	}
}

// MergeCustomLabels sets labels from src to dst.
// If a key in src has [LabelBase] prefix returns an error.
// If dst is nil returns an error.
func MergeCustomLabels(dst, src map[string]string) error {
	if dst == nil {
		return errors.New("destination map is nil")
	}
	for key, value := range src {
		if strings.HasPrefix(key, LabelBase) {
			return fmt.Errorf("key %q has %q prefix", key, LabelBase)
		}
		dst[key] = value
	}
	return nil
}
