package client

import (
	"github.com/docker/docker/errdefs"
)

var permanentClientErrors = []func(error) bool{
	errdefs.IsNotFound,
	errdefs.IsInvalidParameter,
	errdefs.IsUnauthorized,
	errdefs.IsForbidden,
	errdefs.IsNotImplemented,
	errdefs.IsSystem,
}

func isPermanentClientError(err error) bool {
	for _, isErrFn := range permanentClientErrors {
		if isErrFn(err) {
			return true
		}
	}
	return false
}
