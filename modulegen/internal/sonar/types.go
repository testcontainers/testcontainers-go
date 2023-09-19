package sonar

import (
	"sort"
	"strings"
)

type Config struct {
	Go             Go
	ProjectVersion string
}

type Go struct {
	Tests Tests
}

type Tests struct {
	ReportPaths string
}

func newConfig(tcVersion string, examples []string, modules []string) *Config {
	reportPaths := []string{"TEST-unit.xml", "modulegen/TEST-unit.xml"}
	for _, example := range examples {
		reportPaths = append(reportPaths, "examples/"+example+"/TEST-unit.xml")
	}
	for _, module := range modules {
		reportPaths = append(reportPaths, "modules/"+module+"/TEST-unit.xml")
	}
	sort.Strings(reportPaths)
	return &Config{
		Go: Go{
			Tests: Tests{
				ReportPaths: strings.Join(reportPaths, ","),
			},
		},
		ProjectVersion: tcVersion,
	}
}
