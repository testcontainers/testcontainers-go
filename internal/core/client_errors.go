package core

import "github.com/docker/docker/errdefs"

var permanentClientErrors = []func(error) bool{
	errdefs.IsNotFound,
	errdefs.IsInvalidParameter,
	errdefs.IsUnauthorized,
	errdefs.IsForbidden,
	errdefs.IsNotImplemented,
	errdefs.IsSystem,
}

// IsPermanentClientError returns true if the error is a permanent client error
// from the Docker client.
func IsPermanentClientError(err error) bool {
	for _, isErrFn := range permanentClientErrors {
		if isErrFn(err) {
			return true
		}
	}
	return false
}
