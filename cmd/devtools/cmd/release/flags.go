package release

import "fmt"

const (
	bumpTypeFlag    string = "bump-type"
	preReleaseFlag  string = "pre-release"
	bumpTypeInfoMsg string = "Must be 'major', 'minor', 'patch' or 'prerel'"
	dryRunFlag      string = "dry-run"
)

func parseBumpType(bumpType string) error {
	if bumpType != "major" && bumpType != "minor" && bumpType != "patch" && bumpType != "prerel" {
		return fmt.Errorf("invalid bump type: %s. %s", bumpType, bumpTypeInfoMsg)
	}

	return nil
}
