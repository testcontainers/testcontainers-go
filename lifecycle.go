package testcontainers

import (
	"context"
	"errors"
)

// ContainerRequestHook is a hook that will be called before a container is created.
// It can be used to modify container configuration before it is created,
// using the different lifecycle hooks that are available:
// - Creating
// For that, it will receive a ContainerRequestHook, modify it and return an error if needed.
type ContainerRequestHook func(ctx context.Context, req *Request) error

// CreatedContainerHook is a hook that will be called after a container is created
// It can be used to modify the state of the container after it is created,
// using the different lifecycle hooks that are available:
// - Created
// - Starting
// For that, it will receive a Container, modify it and return an error if needed.
type CreatedContainerHook func(ctx context.Context, container CreatedContainer) error

// StartedContainerHook is a hook that will be called after a container is started
// It can be used to modify the state of the container after it is started,
// using the different lifecycle hooks that are available:
// - Started
// - Readied
// - Stopping
// - Stopped
// - Terminating
// - Terminated
// For that, it will receive a Container, modify it and return an error if needed.
type StartedContainerHook func(ctx context.Context, container StartedContainer) error

// ContainerLifecycleHooks is a struct that contains all the hooks that can be used
// to modify the container lifecycle.
type ContainerLifecycleHooks struct {
	PreCreates     []ContainerRequestHook
	PostCreates    []CreatedContainerHook
	PreStarts      []CreatedContainerHook
	PostStarts     []StartedContainerHook
	PostReadies    []StartedContainerHook
	PreStops       []StartedContainerHook
	PostStops      []StartedContainerHook
	PreTerminates  []StartedContainerHook
	PostTerminates []StartedContainerHook
}

// Creating is a hook that will be called before a container is created.
func (c ContainerLifecycleHooks) Creating(ctx context.Context) func(req *Request) error {
	return func(req *Request) error {
		for _, hook := range c.PreCreates {
			if err := hook(ctx, req); err != nil {
				return err
			}
		}

		return nil
	}
}

// Created is a hook that will be called after a container is created
func (c ContainerLifecycleHooks) Created(ctx context.Context) func(container CreatedContainer) error {
	return createdContainerHookFn(ctx, c.PostCreates)
}

// Starting is a hook that will be called before a container is started
func (c ContainerLifecycleHooks) Starting(ctx context.Context) func(container CreatedContainer) error {
	return createdContainerHookFn(ctx, c.PreStarts)
}

// Started is a hook that will be called after a container is started
func (c ContainerLifecycleHooks) Started(ctx context.Context) func(container StartedContainer) error {
	return startedContainerHookFn(ctx, c.PostStarts)
}

// Readied is a hook that will be called after a container is ready
func (c ContainerLifecycleHooks) Readied(ctx context.Context) func(container StartedContainer) error {
	return startedContainerHookFn(ctx, c.PostReadies)
}

// Stopping is a hook that will be called before a container is stopped
func (c ContainerLifecycleHooks) Stopping(ctx context.Context) func(container StartedContainer) error {
	return startedContainerHookFn(ctx, c.PreStops)
}

// Stopped is a hook that will be called after a container is stopped
func (c ContainerLifecycleHooks) Stopped(ctx context.Context) func(container StartedContainer) error {
	return startedContainerHookFn(ctx, c.PostStops)
}

// Terminating is a hook that will be called before a container is terminated
func (c ContainerLifecycleHooks) Terminating(ctx context.Context) func(container StartedContainer) error {
	return startedContainerHookFn(ctx, c.PreTerminates)
}

// Terminated is a hook that will be called after a container is terminated
func (c ContainerLifecycleHooks) Terminated(ctx context.Context) func(container StartedContainer) error {
	return startedContainerHookFn(ctx, c.PostTerminates)
}

// combineContainerHooks it returns just one ContainerLifecycle hook, as the result of combining
// the default hooks with the user-defined hooks. The function will loop over all the default hooks,
// storing each of the hooks in a slice, and then it will loop over all the user-defined hooks,
// appending or prepending them to the slice of hooks. The order of hooks is the following:
// - for Pre-hooks, always run the default hooks first, then append the user-defined hooks
// - for Post-hooks, always run the user-defined hooks first, then the default hooks
func combineContainerHooks(defaultHooks, userDefinedHooks []ContainerLifecycleHooks) ContainerLifecycleHooks {
	preCreates := []ContainerRequestHook{}
	postCreates := []CreatedContainerHook{}
	preStarts := []CreatedContainerHook{}
	postStarts := []StartedContainerHook{}
	postReadies := []StartedContainerHook{}
	preStops := []StartedContainerHook{}
	postStops := []StartedContainerHook{}
	preTerminates := []StartedContainerHook{}
	postTerminates := []StartedContainerHook{}

	for _, defaultHook := range defaultHooks {
		preCreates = append(preCreates, defaultHook.PreCreates...)
		preStarts = append(preStarts, defaultHook.PreStarts...)
		preStops = append(preStops, defaultHook.PreStops...)
		preTerminates = append(preTerminates, defaultHook.PreTerminates...)
	}

	// append the user-defined hooks after the default pre-hooks
	// and because the post hooks are still empty, the user-defined post-hooks
	// will be the first ones to be executed
	for _, userDefinedHook := range userDefinedHooks {
		preCreates = append(preCreates, userDefinedHook.PreCreates...)
		postCreates = append(postCreates, userDefinedHook.PostCreates...)
		preStarts = append(preStarts, userDefinedHook.PreStarts...)
		postStarts = append(postStarts, userDefinedHook.PostStarts...)
		postReadies = append(postReadies, userDefinedHook.PostReadies...)
		preStops = append(preStops, userDefinedHook.PreStops...)
		postStops = append(postStops, userDefinedHook.PostStops...)
		preTerminates = append(preTerminates, userDefinedHook.PreTerminates...)
		postTerminates = append(postTerminates, userDefinedHook.PostTerminates...)
	}

	// finally, append the default post-hooks
	for _, defaultHook := range defaultHooks {
		postCreates = append(postCreates, defaultHook.PostCreates...)
		postStarts = append(postStarts, defaultHook.PostStarts...)
		postReadies = append(postReadies, defaultHook.PostReadies...)
		postStops = append(postStops, defaultHook.PostStops...)
		postTerminates = append(postTerminates, defaultHook.PostTerminates...)
	}

	return ContainerLifecycleHooks{
		PreCreates:     preCreates,
		PostCreates:    postCreates,
		PreStarts:      preStarts,
		PostStarts:     postStarts,
		PostReadies:    postReadies,
		PreStops:       preStops,
		PostStops:      postStops,
		PreTerminates:  preTerminates,
		PostTerminates: postTerminates,
	}
}

// applyCreatedLifecycleHooks applies created lifecycle hooks reporting the container logs on error if logError is true.
func (c *DockerContainer) applyCreatedLifecycleHooks(ctx context.Context, logError bool, hooks func(lifecycleHooks ContainerLifecycleHooks) []CreatedContainerHook) error {
	errs := make([]error, len(c.lifecycleHooks))
	for i, lifecycleHooks := range c.lifecycleHooks {
		errs[i] = createdContainerHookFn(ctx, hooks(lifecycleHooks))(c)
	}

	if err := errors.Join(errs...); err != nil {
		if logError {
			c.printLogs(ctx, err)
		}

		return err
	}

	return nil
}

// applyStartedLifecycleHooks applies started lifecycle hooks reporting the container logs on error if logError is true.
func (c *DockerContainer) applyStartedLifecycleHooks(ctx context.Context, logError bool, hooks func(lifecycleHooks ContainerLifecycleHooks) []StartedContainerHook) error {
	errs := make([]error, len(c.lifecycleHooks))
	for i, lifecycleHooks := range c.lifecycleHooks {
		errs[i] = startedContainerHookFn(ctx, hooks(lifecycleHooks))(c)
	}

	if err := errors.Join(errs...); err != nil {
		if logError {
			c.printLogs(ctx, err)
		}

		return err
	}

	return nil
}

// createdContainerHookFn is a helper function that will create a function to be returned by all the different
// container lifecycle hooks. The created function will iterate over all the hooks and call them one by one.
func createdContainerHookFn(ctx context.Context, containerHook []CreatedContainerHook) func(container CreatedContainer) error {
	return func(container CreatedContainer) error {
		errs := make([]error, len(containerHook))
		for i, hook := range containerHook {
			errs[i] = hook(ctx, container)
		}

		return errors.Join(errs...)
	}
}

// startedContainerHookFn is a helper function that will create a function to be returned by all the different
// container lifecycle hooks. The created function will iterate over all the hooks and call them one by one.
func startedContainerHookFn(ctx context.Context, containerHook []StartedContainerHook) func(container StartedContainer) error {
	return func(container StartedContainer) error {
		errs := make([]error, len(containerHook))
		for i, hook := range containerHook {
			errs[i] = hook(ctx, container)
		}

		return errors.Join(errs...)
	}
}
