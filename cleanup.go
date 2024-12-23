package testcontainers

import (
	"context"
	"fmt"
	"reflect"
	"time"
)

// DefaultTimeout for termination
var DefaultTimeout = 10 * time.Second

// TerminateOptions is a type that holds the options for terminating a container.
type TerminateOptions struct {
	Context context.Context
	Timeout *time.Duration
	Volumes []string
}

// TerminateOption is a type that represents an option for terminating a container.
type TerminateOption func(*TerminateOptions)

// StopContext returns a TerminateOption that sets the context.
// Default: context.Background().
func StopContext(ctx context.Context) TerminateOption {
	return func(c *TerminateOptions) {
		c.Context = ctx
	}
}

// StopTimeout returns a TerminateOption that sets the timeout.
// Default: See [Container.Stop].
func StopTimeout(timeout time.Duration) TerminateOption {
	return func(c *TerminateOptions) {
		c.Timeout = &timeout
	}
}

// RemoveVolumes returns a TerminateOption that sets additional volumes to remove.
// This is useful when the container creates named volumes that should be removed
// which are not removed by default.
// Default: nil.
func RemoveVolumes(volumes ...string) TerminateOption {
	return func(c *TerminateOptions) {
		c.Volumes = volumes
	}
}

// TerminateContainer calls [Container.Terminate] on the container if it is not nil.
//
// This should be called as a defer directly after [GenericContainer](...)
// or a modules Run(...) to ensure the container is terminated when the
// function ends.
func TerminateContainer(container Container, options ...TerminateOption) error {
	if isNil(container) {
		return nil
	}

	err := container.Terminate(context.Background(), options...)
	if !isCleanupSafe(err) {
		return fmt.Errorf("terminate: %w", err)
	}

	return nil
}

// isNil returns true if val is nil or an nil instance false otherwise.
func isNil(val any) bool {
	if val == nil {
		return true
	}

	valueOf := reflect.ValueOf(val)
	switch valueOf.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr, reflect.UnsafePointer, reflect.Interface, reflect.Slice:
		return valueOf.IsNil()
	default:
		return false
	}
}
