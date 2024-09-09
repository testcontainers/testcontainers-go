package core

import (
	"github.com/testcontainers/testcontainers-go/internal"
)

const (
	LabelBase            = "org.testcontainers"
	LabelLang            = LabelBase + ".lang"
	LabelContainerHash   = LabelBase + ".hash"
	LabelCopiedFilesHash = LabelBase + ".copied_files.hash"
	LabelReaper          = LabelBase + ".reaper"
	LabelRyuk            = LabelBase + ".ryuk"
	LabelSessionID       = LabelBase + ".sessionId"
	LabelVersion         = LabelBase + ".version"
)

func DefaultLabels(sessionID string) map[string]string {
	return map[string]string{
		LabelBase:      "true",
		LabelLang:      "go",
		LabelSessionID: sessionID,
		LabelVersion:   internal.Version,
	}
}
