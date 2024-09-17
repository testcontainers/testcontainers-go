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

// MergeCustomLabels sets labels from src to dst. Returns error if src label
// name has [LabelBase] prefix.
//
// NOTICE: The dst labels must not be nil so we can set the labels.
func MergeCustomLabels(dst, src map[string]string) error {
	if dst == nil && len(src) > 0 {
		return errors.New("cannot merge custom labels because destination map is nil")
	}
	for key, value := range src {
		if strings.HasPrefix(key, LabelBase) {
			format := "cannot use prefix %q for custom labels"
			return fmt.Errorf(format, LabelBase)
		}
		dst[key] = value
	}
	return nil
}
