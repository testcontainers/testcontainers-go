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

func DefaultLabels(sessionID string) map[string]string {
	return map[string]string{
		LabelBase:      "true",
		LabelLang:      "go",
		LabelSessionID: sessionID,
		LabelVersion:   internal.Version,
	}
}

// MergeCustomLabels sets labels from src to dst. Returns error if src label
// name has [LabelBase] prefix.
//
// NOTICE: dst labels must not be nil.
func MergeCustomLabels(dst, src map[string]string) error {
	for key := range src {
		if strings.HasPrefix(key, LabelBase) {
			format := "cannot use prefix %q for custom labels"
			return fmt.Errorf(format, LabelBase)
		}
	}
	for key, value := range src {
		dst[key] = value
	}
	return nil
}
