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

	lifecycleHookFunc := func(prefix string, lifecycleID int) LifecycleHooks {
		return LifecycleHooks{
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

	defaultHooks := []LifecycleHooks{lifecycleHookFunc("default", 1), lifecycleHookFunc("default", 2)}
	userDefinedHooks := []LifecycleHooks{lifecycleHookFunc("user-defined", 1), lifecycleHookFunc("user-defined", 2), lifecycleHookFunc("user-defined", 3)}

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
	lifecycleHooks := []LifecycleHooks{
		{
			PreCreates: []ContainerRequestHook{
				func(ctx context.Context, req *Request) error {
					prints = append(prints, "pre-create hook 1")
					return nil
				},
				func(ctx context.Context, req *Request) error {
					prints = append(prints, "pre-create hook 2")
					return nil
				},
			},
			PostCreates: []CreatedContainerHook{
				func(ctx context.Context, c CreatedContainer) error {
					prints = append(prints, "post-create hook 1")
					return nil
				},
				func(ctx context.Context, c CreatedContainer) error {
					prints = append(prints, "post-create hook 2")
					return nil
				},
			},
			PreStarts: []CreatedContainerHook{
				func(ctx context.Context, c CreatedContainer) error {
					prints = append(prints, "pre-start hook 1")
					return nil
				},
				func(ctx context.Context, c CreatedContainer) error {
					prints = append(prints, "pre-start hook 2")
					return nil
				},
			},
			PostStarts: []StartedContainerHook{
				func(ctx context.Context, c StartedContainer) error {
					prints = append(prints, "post-start hook 1")
					return nil
				},
				func(ctx context.Context, c StartedContainer) error {
					prints = append(prints, "post-start hook 2")
					return nil
				},
			},
			PostReadies: []StartedContainerHook{
				func(ctx context.Context, c StartedContainer) error {
					prints = append(prints, "post-ready hook 1")
					return nil
				},
				func(ctx context.Context, c StartedContainer) error {
					prints = append(prints, "post-ready hook 2")
					return nil
				},
			},
			PreStops: []StartedContainerHook{
				func(ctx context.Context, c StartedContainer) error {
					prints = append(prints, "pre-stop hook 1")
					return nil
				},
				func(ctx context.Context, c StartedContainer) error {
					prints = append(prints, "pre-stop hook 2")
					return nil
				},
			},
			PostStops: []StartedContainerHook{
				func(ctx context.Context, c StartedContainer) error {
					prints = append(prints, "post-stop hook 1")
					return nil
				},
				func(ctx context.Context, c StartedContainer) error {
					prints = append(prints, "post-stop hook 2")
					return nil
				},
			},
			PreTerminates: []StartedContainerHook{
				func(ctx context.Context, c StartedContainer) error {
					prints = append(prints, "pre-terminate hook 1")
					return nil
				},
				func(ctx context.Context, c StartedContainer) error {
					prints = append(prints, "pre-terminate hook 2")
					return nil
				},
			},
			PostTerminates: []StartedContainerHook{
				func(ctx context.Context, c StartedContainer) error {
					prints = append(prints, "post-terminate hook 1")
					return nil
				},
				func(ctx context.Context, c StartedContainer) error {
					prints = append(prints, "post-terminate hook 2")
					return nil
				},
			},
		},
	}
	// }

	c, err := Run(ctx, Request{
		Image:          nginxAlpineImage,
		LifecycleHooks: lifecycleHooks,
		Started:        true,
	})
	CleanupContainer(t, c)
	require.NoError(t, err)
	require.NotNil(t, c)

	duration := 1 * time.Second
	err = c.Stop(ctx, &duration)
	require.NoError(t, err)

	err = c.Start(ctx)
	require.NoError(t, err)

	err = c.Terminate(ctx)
	require.NoError(t, err)

	lifecycleHooksIsHonouredFn(t, prints)
}

func TestLifecycleHooks_WithDefaultLogger(t *testing.T) {
	ctx := context.Background()

	// reqWithDefaultLogginHook {
	dl := inMemoryLogger{}

	req := Request{
		Image: nginxAlpineImage,
		LifecycleHooks: []LifecycleHooks{
			DefaultLoggingHook(&dl),
		},
		Started: true,
	}
	// }

	c, err := Run(ctx, req)
	CleanupContainer(t, c)
	require.NoError(t, err)
	require.NotNil(t, c)

	duration := 1 * time.Second
	err = c.Stop(ctx, &duration)
	require.NoError(t, err)

	err = c.Start(ctx)
	require.NoError(t, err)

	err = c.Terminate(ctx)
	require.NoError(t, err)

	// Includes two additional entries for stop when terminate is called.
	require.Len(t, dl.data, 14)
}

func TestLifecycleHooks_WithMultipleHooks(t *testing.T) {
	ctx := context.Background()

	dl := inMemoryLogger{}

	req := Request{
		Image: nginxAlpineImage,
		LifecycleHooks: []LifecycleHooks{
			DefaultLoggingHook(&dl),
			DefaultLoggingHook(&dl),
		},
		Started: true,
	}

	c, err := Run(ctx, req)
	CleanupContainer(t, c)
	require.NoError(t, err)
	require.NotNil(t, c)

	duration := 1 * time.Second
	err = c.Stop(ctx, &duration)
	require.NoError(t, err)

	err = c.Start(ctx)
	require.NoError(t, err)

	err = c.Terminate(ctx)
	require.NoError(t, err)

	// Includes four additional entries for stop (twice) when terminate is called.
	require.Len(t, dl.data, 28)
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

	ctr, err := Run(ctx, req)
	CleanupContainer(t, ctr)
	// it should fail because the waiting for condition is not met
	if err == nil {
		t.Fatal(err)
	}

	containerLogs, err := ctr.Logs(ctx)
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
	t.Helper()

	expects := []string{
		"pre-create hook 1",
		"pre-create hook 2",
		"post-create hook 1",
		"post-create hook 2",
		"pre-start hook 1",
		"pre-start hook 2",
		"post-start hook 1",
		"post-start hook 2",
		"post-ready hook 1",
		"post-ready hook 2",
		"pre-stop hook 1",
		"pre-stop hook 2",
		"post-stop hook 1",
		"post-stop hook 2",
		"pre-start hook 1",
		"pre-start hook 2",
		"post-start hook 1",
		"post-start hook 2",
		"post-ready hook 1",
		"post-ready hook 2",
		// Terminate currently calls stop to ensure that child containers are stopped.
		"pre-stop hook 1",
		"pre-stop hook 2",
		"post-stop hook 1",
		"post-stop hook 2",
		"pre-terminate hook 1",
		"pre-terminate hook 2",
		"post-terminate hook 1",
		"post-terminate hook 2",
	}

	require.Equal(t, expects, prints)
}
