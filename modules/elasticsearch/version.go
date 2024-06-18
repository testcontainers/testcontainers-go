package elasticsearch

import (
	"fmt"
	"strings"

	"golang.org/x/mod/semver"
)

// isOSS returns true if the base image (without tag) is an OSS image
func isOSS(img string) bool {
	return strings.HasPrefix(img, DefaultBaseImageOSS)
}

// isAtLeastVersion returns true if the base image (without tag) is in a version or above
func isAtLeastVersion(img string, major int) bool {
	parts := strings.Split(img, ":")
	version := parts[len(parts)-1]

	if version == "latest" {
		return true
	}

	if !strings.HasPrefix(version, "v") {
		version = fmt.Sprintf("v%s", version)
	}

	if semver.IsValid(version) {
		return semver.Compare(version, fmt.Sprintf("v%d", major)) >= 0 // version >= v8.x
	}

	return false
}
