package testcontainers

import (
	"bufio"
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go/wait"
)

// customLoggerImplementation {
type inMemoryLogger struct {
	data []string
}

func (l *inMemoryLogger) Printf(format string, args ...interface{}) {
	l.data = append(l.data, fmt.Sprintf(format, args...))
}

// }

func TestCombineLifecycleHooks(t *testing.T) {
	prints := []string{}

	preCreateFunc := func(prefix string, hook string, lifecycleID int, hookID int) func(ctx context.Context, req *Request) error {
		return func(ctx context.Context, _ *Request) error {
			prints = append(prints, fmt.Sprintf("[%s] pre-%s hook %d.%d", prefix, hook, lifecycleID, hookID))
			return nil
		}
	}
	createdHookFunc := func(prefix string, hookType string, hook string, lifecycleID int, hookID int) func(ctx context.Context, c CreatedContainer) error {
		return func(ctx context.Context, _ CreatedContainer) error {
			prints = append(prints, fmt.Sprintf("[%s] %s-%s hook %d.%d", prefix, hookType, hook, lifecycleID, hookID))
			return nil
		}
	}
	startedHookFunc := func(prefix string, hookType string, hook string, lifecycleID int, hookID int) func(ctx context.Context, c StartedContainer) error {
		return func(ctx context.Context, _ StartedContainer) error {
			prints = append(prints, fmt.Sprintf("[%s] %s-%s hook %d.%d", prefix, hookType, hook, lifecycleID, hookID))
			return nil
		}
	}
	preCreatedFunc := func(prefix string, hook string, lifecycleID int, hookID int) func(ctx context.Context, c CreatedContainer) error {
		return createdHookFunc(prefix, "pre", hook, lifecycleID, hookID)
	}
	preStartedFunc := func(prefix string, hook string, lifecycleID int, hookID int) func(ctx context.Context, c StartedContainer) error {
		return startedHookFunc(prefix, "pre", hook, lifecycleID, hookID)
	}
	postCreatedFunc := func(prefix string, hook string, lifecycleID int, hookID int) func(ctx context.Context, c CreatedContainer) error {
		return createdHookFunc(prefix, "post", hook, lifecycleID, hookID)
	}
	postStartedFunc := func(prefix string, hook string, lifecycleID int, hookID int) func(ctx context.Context, c StartedContainer) error {
		return startedHookFunc(prefix, "post", hook, lifecycleID, hookID)
	}

	lifecycleHookFunc := func(prefix string, lifecycleID int) ContainerLifecycleHooks {
		return ContainerLifecycleHooks{
			PreCreates:     []ContainerRequestHook{preCreateFunc(prefix, "create", lifecycleID, 1), preCreateFunc(prefix, "create", lifecycleID, 2)},
			PostCreates:    []CreatedContainerHook{postCreatedFunc(prefix, "create", lifecycleID, 1), postCreatedFunc(prefix, "create", lifecycleID, 2)},
			PreStarts:      []CreatedContainerHook{preCreatedFunc(prefix, "start", lifecycleID, 1), preCreatedFunc(prefix, "start", lifecycleID, 2)},
			PostStarts:     []StartedContainerHook{postStartedFunc(prefix, "start", lifecycleID, 1), postStartedFunc(prefix, "start", lifecycleID, 2)},
			PostReadies:    []StartedContainerHook{postStartedFunc(prefix, "ready", lifecycleID, 1), postStartedFunc(prefix, "ready", lifecycleID, 2)},
			PreStops:       []StartedContainerHook{preStartedFunc(prefix, "stop", lifecycleID, 1), preStartedFunc(prefix, "stop", lifecycleID, 2)},
			PostStops:      []StartedContainerHook{postStartedFunc(prefix, "stop", lifecycleID, 1), postStartedFunc(prefix, "stop", lifecycleID, 2)},
			PreTerminates:  []StartedContainerHook{preStartedFunc(prefix, "terminate", lifecycleID, 1), preStartedFunc(prefix, "terminate", lifecycleID, 2)},
			PostTerminates: []StartedContainerHook{postStartedFunc(prefix, "terminate", lifecycleID, 1), postStartedFunc(prefix, "terminate", lifecycleID, 2)},
		}
	}

	defaultHooks := []ContainerLifecycleHooks{lifecycleHookFunc("default", 1), lifecycleHookFunc("default", 2)}
	userDefinedHooks := []ContainerLifecycleHooks{lifecycleHookFunc("user-defined", 1), lifecycleHookFunc("user-defined", 2), lifecycleHookFunc("user-defined", 3)}

	hooks := combineContainerHooks(defaultHooks, userDefinedHooks)

	// call all the hooks in the right order, honouring the lifecycle

	req := Request{}
	err := hooks.Creating(context.Background())(&req)
	require.NoError(t, err)

	c := &DockerContainer{}

	err = hooks.Created(context.Background())(c)
	require.NoError(t, err)
	err = hooks.Starting(context.Background())(c)
	require.NoError(t, err)
	err = hooks.Started(context.Background())(c)
	require.NoError(t, err)
	err = hooks.Readied(context.Background())(c)
	require.NoError(t, err)
	err = hooks.Stopping(context.Background())(c)
	require.NoError(t, err)
	err = hooks.Stopped(context.Background())(c)
	require.NoError(t, err)
	err = hooks.Terminating(context.Background())(c)
	require.NoError(t, err)
	err = hooks.Terminated(context.Background())(c)
	require.NoError(t, err)

	// assertions

	// There are 2 default container lifecycle hooks and 3 user-defined container lifecycle hooks.
	// Each lifecycle hook has 2 pre-create hooks and 2 post-create hooks.
	// That results in 16 hooks per lifecycle (8 defaults + 12 user-defined = 20)

	// There are 5 lifecycles (create, start, ready, stop, terminate),
	// but ready has only half of the hooks (it only has post), so we have 90 hooks in total.
	assert.Len(t, prints, 90)

	// The order of the hooks is:
	// - pre-X hooks: first default (2*2), then user-defined (3*2)
	// - post-X hooks: first user-defined (3*2), then default (2*2)

	for i := 0; i < 5; i++ {
		var hookType string
		// this is the particular order of execution for the hooks
		switch i {
		case 0:
			hookType = "create"
		case 1:
			hookType = "start"
		case 2:
			hookType = "ready"
		case 3:
			hookType = "stop"
		case 4:
			hookType = "terminate"
		}

		initialIndex := i * 20
		if i >= 2 {
			initialIndex -= 10
		}

		if hookType != "ready" {
			// default pre-hooks: 4 hooks
			assert.Equal(t, fmt.Sprintf("[default] pre-%s hook 1.1", hookType), prints[initialIndex])
			assert.Equal(t, fmt.Sprintf("[default] pre-%s hook 1.2", hookType), prints[initialIndex+1])
			assert.Equal(t, fmt.Sprintf("[default] pre-%s hook 2.1", hookType), prints[initialIndex+2])
			assert.Equal(t, fmt.Sprintf("[default] pre-%s hook 2.2", hookType), prints[initialIndex+3])

			// user-defined pre-hooks: 6 hooks
			assert.Equal(t, fmt.Sprintf("[user-defined] pre-%s hook 1.1", hookType), prints[initialIndex+4])
			assert.Equal(t, fmt.Sprintf("[user-defined] pre-%s hook 1.2", hookType), prints[initialIndex+5])
			assert.Equal(t, fmt.Sprintf("[user-defined] pre-%s hook 2.1", hookType), prints[initialIndex+6])
			assert.Equal(t, fmt.Sprintf("[user-defined] pre-%s hook 2.2", hookType), prints[initialIndex+7])
			assert.Equal(t, fmt.Sprintf("[user-defined] pre-%s hook 3.1", hookType), prints[initialIndex+8])
			assert.Equal(t, fmt.Sprintf("[user-defined] pre-%s hook 3.2", hookType), prints[initialIndex+9])
		}

		// user-defined post-hooks: 6 hooks
		assert.Equal(t, fmt.Sprintf("[user-defined] post-%s hook 1.1", hookType), prints[initialIndex+10])
		assert.Equal(t, fmt.Sprintf("[user-defined] post-%s hook 1.2", hookType), prints[initialIndex+11])
		assert.Equal(t, fmt.Sprintf("[user-defined] post-%s hook 2.1", hookType), prints[initialIndex+12])
		assert.Equal(t, fmt.Sprintf("[user-defined] post-%s hook 2.2", hookType), prints[initialIndex+13])
		assert.Equal(t, fmt.Sprintf("[user-defined] post-%s hook 3.1", hookType), prints[initialIndex+14])
		assert.Equal(t, fmt.Sprintf("[user-defined] post-%s hook 3.2", hookType), prints[initialIndex+15])

		// default post-hooks: 4 hooks
		assert.Equal(t, fmt.Sprintf("[default] post-%s hook 1.1", hookType), prints[initialIndex+16])
		assert.Equal(t, fmt.Sprintf("[default] post-%s hook 1.2", hookType), prints[initialIndex+17])
		assert.Equal(t, fmt.Sprintf("[default] post-%s hook 2.1", hookType), prints[initialIndex+18])
		assert.Equal(t, fmt.Sprintf("[default] post-%s hook 2.2", hookType), prints[initialIndex+19])
	}
}

func TestLifecycleHooks(t *testing.T) {
	prints := []string{}
	ctx := context.Background()
	// reqWithLifecycleHooks {
	lifecycleHooks := []ContainerLifecycleHooks{
		{
			PreCreates: []ContainerRequestHook{
				func(ctx context.Context, req *Request) error {
					prints = append(prints, fmt.Sprintf("pre-create hook 1: %#v", req))
					return nil
				},
				func(ctx context.Context, req *Request) error {
					prints = append(prints, fmt.Sprintf("pre-create hook 2: %#v", req))
					return nil
				},
			},
			PostCreates: []CreatedContainerHook{
				func(ctx context.Context, c CreatedContainer) error {
					prints = append(prints, fmt.Sprintf("post-create hook 1: %#v", c))
					return nil
				},
				func(ctx context.Context, c CreatedContainer) error {
					prints = append(prints, fmt.Sprintf("post-create hook 2: %#v", c))
					return nil
				},
			},
			PreStarts: []CreatedContainerHook{
				func(ctx context.Context, c CreatedContainer) error {
					prints = append(prints, fmt.Sprintf("pre-start hook 1: %#v", c))
					return nil
				},
				func(ctx context.Context, c CreatedContainer) error {
					prints = append(prints, fmt.Sprintf("pre-start hook 2: %#v", c))
					return nil
				},
			},
			PostStarts: []StartedContainerHook{
				func(ctx context.Context, c StartedContainer) error {
					prints = append(prints, fmt.Sprintf("post-start hook 1: %#v", c))
					return nil
				},
				func(ctx context.Context, c StartedContainer) error {
					prints = append(prints, fmt.Sprintf("post-start hook 2: %#v", c))
					return nil
				},
			},
			PostReadies: []StartedContainerHook{
				func(ctx context.Context, c StartedContainer) error {
					prints = append(prints, fmt.Sprintf("post-ready hook 1: %#v", c))
					return nil
				},
				func(ctx context.Context, c StartedContainer) error {
					prints = append(prints, fmt.Sprintf("post-ready hook 2: %#v", c))
					return nil
				},
			},
			PreStops: []StartedContainerHook{
				func(ctx context.Context, c StartedContainer) error {
					prints = append(prints, fmt.Sprintf("pre-stop hook 1: %#v", c))
					return nil
				},
				func(ctx context.Context, c StartedContainer) error {
					prints = append(prints, fmt.Sprintf("pre-stop hook 2: %#v", c))
					return nil
				},
			},
			PostStops: []StartedContainerHook{
				func(ctx context.Context, c StartedContainer) error {
					prints = append(prints, fmt.Sprintf("post-stop hook 1: %#v", c))
					return nil
				},
				func(ctx context.Context, c StartedContainer) error {
					prints = append(prints, fmt.Sprintf("post-stop hook 2: %#v", c))
					return nil
				},
			},
			PreTerminates: []StartedContainerHook{
				func(ctx context.Context, c StartedContainer) error {
					prints = append(prints, fmt.Sprintf("pre-terminate hook 1: %#v", c))
					return nil
				},
				func(ctx context.Context, c StartedContainer) error {
					prints = append(prints, fmt.Sprintf("pre-terminate hook 2: %#v", c))
					return nil
				},
			},
			PostTerminates: []StartedContainerHook{
				func(ctx context.Context, c StartedContainer) error {
					prints = append(prints, fmt.Sprintf("post-terminate hook 1: %#v", c))
					return nil
				},
				func(ctx context.Context, c StartedContainer) error {
					prints = append(prints, fmt.Sprintf("post-terminate hook 2: %#v", c))
					return nil
				},
			},
		},
	}
	// }

	req := Request{
		LifecycleHooks: lifecycleHooks,
	}
	startedContainer := &DockerContainer{}

	for _, hook := range lifecycleHooks {
		// TODO: instead of calling the hooks directly, we should create a Generic Container instead
		err := hook.Creating(ctx)(&req)
		require.NoError(t, err)

		err = hook.Created(ctx)(startedContainer)
		require.NoError(t, err)

		err = hook.Starting(ctx)(startedContainer)
		require.NoError(t, err)

		err = hook.Started(ctx)(startedContainer)
		require.NoError(t, err)

		err = hook.Readied(ctx)(startedContainer)
		require.NoError(t, err)

		err = hook.Stopping(ctx)(startedContainer)
		require.NoError(t, err)

		err = hook.Stopped(ctx)(startedContainer)
		require.NoError(t, err)

		// simulating container.Start again after the container has been stopped
		err = hook.Starting(ctx)(startedContainer)
		require.NoError(t, err)

		err = hook.Started(ctx)(startedContainer)
		require.NoError(t, err)

		err = hook.Readied(ctx)(startedContainer)
		require.NoError(t, err)
		// end of simulating container.Start again

		err = hook.Terminating(ctx)(startedContainer)
		require.NoError(t, err)

		err = hook.Terminated(ctx)(startedContainer)
		require.NoError(t, err)
	}

	lifecycleHooksIsHonouredFn(t, prints)
}

func TestLifecycleHooks_WithDefaultLogger(t *testing.T) {
	ctx := context.Background()

	// reqWithDefaultLogginHook {
	dl := inMemoryLogger{}

	req := Request{
		Image: nginxAlpineImage,
		LifecycleHooks: []ContainerLifecycleHooks{
			DefaultLoggingHook(&dl),
		},
		Started: true,
	}
	// }

	c, err := New(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, c)

	duration := 1 * time.Second
	err = c.Stop(ctx, &duration)
	require.NoError(t, err)

	err = c.Start(ctx)
	require.NoError(t, err)

	err = c.Terminate(ctx)
	require.NoError(t, err)

	require.Len(t, dl.data, 12)
}

func TestLifecycleHooks_WithMultipleHooks(t *testing.T) {
	ctx := context.Background()

	dl := inMemoryLogger{}

	req := Request{
		Image: nginxAlpineImage,
		LifecycleHooks: []ContainerLifecycleHooks{
			DefaultLoggingHook(&dl),
			DefaultLoggingHook(&dl),
		},
		Started: true,
	}

	c, err := New(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, c)

	duration := 1 * time.Second
	err = c.Stop(ctx, &duration)
	require.NoError(t, err)

	err = c.Start(ctx)
	require.NoError(t, err)

	err = c.Terminate(ctx)
	require.NoError(t, err)

	require.Len(t, dl.data, 24)
}

type linesTestLogger struct {
	data []string
}

func (l *linesTestLogger) Printf(format string, args ...interface{}) {
	l.data = append(l.data, fmt.Sprintf(format, args...))
}

func TestPrintContainerLogsOnError(t *testing.T) {
	ctx := context.Background()

	arrayOfLinesLogger := linesTestLogger{
		data: []string{},
	}

	req := Request{
		Image:      "docker.io/alpine",
		Cmd:        []string{"echo", "-n", "I am expecting this"},
		WaitingFor: wait.ForLog("I was expecting that").WithStartupTimeout(5 * time.Second),
		Started:    true,
		Logger:     &arrayOfLinesLogger,
	}

	container, err := New(ctx, req)
	// it should fail because the waiting for condition is not met
	if err == nil {
		t.Fatal(err)
	}
	TerminateContainerOnEnd(t, ctx, container)

	containerLogs, err := container.Logs(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer containerLogs.Close()

	// read container logs line by line, checking that each line is present in the stdout
	rd := bufio.NewReader(containerLogs)
	for {
		line, err := rd.ReadString('\n')
		if err != nil {
			if err.Error() == "EOF" {
				break
			}

			t.Fatal("Read Error:", err)
		}

		// the last line of the array should contain the line of interest,
		// but we are checking all the lines to make sure that is present
		found := false
		for _, l := range arrayOfLinesLogger.data {
			if strings.Contains(l, line) {
				found = true
				break
			}
		}
		assert.True(t, found, "container log line not found in the output of the logger: %s", line)
	}
}

func lifecycleHooksIsHonouredFn(t *testing.T, prints []string) {
	require.Len(t, prints, 24)

	assert.True(t, strings.HasPrefix(prints[0], "pre-create hook 1: "))
	assert.True(t, strings.HasPrefix(prints[1], "pre-create hook 2: "))

	assert.True(t, strings.HasPrefix(prints[2], "post-create hook 1: "))
	assert.True(t, strings.HasPrefix(prints[3], "post-create hook 2: "))

	assert.True(t, strings.HasPrefix(prints[4], "pre-start hook 1: "))
	assert.True(t, strings.HasPrefix(prints[5], "pre-start hook 2: "))

	assert.True(t, strings.HasPrefix(prints[6], "post-start hook 1: "))
	assert.True(t, strings.HasPrefix(prints[7], "post-start hook 2: "))

	assert.True(t, strings.HasPrefix(prints[8], "post-ready hook 1: "))
	assert.True(t, strings.HasPrefix(prints[9], "post-ready hook 2: "))

	assert.True(t, strings.HasPrefix(prints[10], "pre-stop hook 1: "))
	assert.True(t, strings.HasPrefix(prints[11], "pre-stop hook 2: "))

	assert.True(t, strings.HasPrefix(prints[12], "post-stop hook 1: "))
	assert.True(t, strings.HasPrefix(prints[13], "post-stop hook 2: "))

	assert.True(t, strings.HasPrefix(prints[14], "pre-start hook 1: "))
	assert.True(t, strings.HasPrefix(prints[15], "pre-start hook 2: "))

	assert.True(t, strings.HasPrefix(prints[16], "post-start hook 1: "))
	assert.True(t, strings.HasPrefix(prints[17], "post-start hook 2: "))

	assert.True(t, strings.HasPrefix(prints[18], "post-ready hook 1: "))
	assert.True(t, strings.HasPrefix(prints[19], "post-ready hook 2: "))

	assert.True(t, strings.HasPrefix(prints[20], "pre-terminate hook 1: "))
	assert.True(t, strings.HasPrefix(prints[21], "pre-terminate hook 2: "))

	assert.True(t, strings.HasPrefix(prints[22], "post-terminate hook 1: "))
	assert.True(t, strings.HasPrefix(prints[23], "post-terminate hook 2: "))
}
