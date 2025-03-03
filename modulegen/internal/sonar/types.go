package sonar

import (
	"sort"
	"strings"
)

// Config is a struct that contains the logic to generate the sonar-project.properties file.
type Config struct {
	Go             Go
	ProjectVersion string
}

// Go is a struct that contains the logic to generate the go files.
type Go struct {
	Tests Tests
}

// Tests is a struct that contains the logic to generate the test report paths.
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
