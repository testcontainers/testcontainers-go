package testcontainersdocker

import "github.com/testcontainers/testcontainers-go/internal"

const (
	LabelBase      = "org.testcontainers"
	LabelLang      = LabelBase + ".lang"
	LabelReaper    = LabelBase + ".reaper"
	LabelSessionID = LabelBase + ".sessionId"
	LabelVersion   = LabelBase + ".version"
)

func DefaultLabels() map[string]string {
	return map[string]string{
		LabelBase:    "true",
		LabelLang:    "go",
		LabelVersion: internal.Version,
	}
}
