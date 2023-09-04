package testcontainersdocker

import (
	"github.com/testcontainers/testcontainers-go/internal"
	"github.com/testcontainers/testcontainers-go/internal/testcontainerssession"
)

const (
	LabelBase      = "org.testcontainers"
	LabelLang      = LabelBase + ".lang"
	LabelReaper    = LabelBase + ".reaper"
	LabelRyuk      = LabelBase + ".ryuk"
	LabelSessionID = LabelBase + ".sessionId"
	LabelVersion   = LabelBase + ".version"
)

func DefaultLabels() map[string]string {
	return map[string]string{
		LabelBase:      "true",
		LabelLang:      "go",
		LabelSessionID: testcontainerssession.SessionID(),
		LabelVersion:   internal.Version,
	}
}
