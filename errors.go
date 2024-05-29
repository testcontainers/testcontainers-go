package testcontainers

import "regexp"

var CreateFailDueToNameConflictRegex = regexp.MustCompile("Conflict. The container name .* is already in use by container .*")
