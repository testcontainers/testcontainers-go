package testcontainers_test

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/exec"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestOverrideContainerRequest(t *testing.T) {
	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Env: map[string]string{
				"BAR": "BAR",
			},
			Image:        "foo",
			ExposedPorts: []string{"12345/tcp"},
			WaitingFor: wait.ForNop(
				func(_ context.Context, _ wait.StrategyTarget) error {
					return nil
				},
			),
			Networks: []string{"foo", "bar", "baaz"},
			NetworkAliases: map[string][]string{
				"foo": {"foo0", "foo1", "foo2", "foo3"},
			},
		},
	}

	toBeMergedRequest := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Env: map[string]string{
				"FOO": "FOO",
			},
			Image:        "bar",
			ExposedPorts: []string{"67890/tcp"},
			Networks:     []string{"foo1", "bar1"},
			NetworkAliases: map[string][]string{
				"foo1": {"bar"},
			},
			WaitingFor: wait.ForLog("foo"),
		},
	}

	// the toBeMergedRequest should be merged into the req
	err := testcontainers.CustomizeRequest(toBeMergedRequest)(&req)
	require.NoError(t, err)

	// toBeMergedRequest should not be changed
	require.Empty(t, toBeMergedRequest.Env["BAR"])
	require.Len(t, toBeMergedRequest.ExposedPorts, 1)
	require.Equal(t, "67890/tcp", toBeMergedRequest.ExposedPorts[0])

	// req should be merged with toBeMergedRequest
	require.Equal(t, "FOO", req.Env["FOO"])
	require.Equal(t, "BAR", req.Env["BAR"])
	require.Equal(t, "bar", req.Image)
	require.Equal(t, []string{"12345/tcp", "67890/tcp"}, req.ExposedPorts)
	require.Equal(t, []string{"foo", "bar", "baaz", "foo1", "bar1"}, req.Networks)
	require.Equal(t, []string{"foo0", "foo1", "foo2", "foo3"}, req.NetworkAliases["foo"])
	require.Equal(t, []string{"bar"}, req.NetworkAliases["foo1"])
	require.Equal(t, wait.ForLog("foo"), req.WaitingFor)
}

type msgsLogConsumer struct {
	msgs []string
}

// Accept prints the log to stdout
func (lc *msgsLogConsumer) Accept(l testcontainers.Log) {
	lc.msgs = append(lc.msgs, string(l.Content))
}

func TestWithLogConsumers(t *testing.T) {
	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:      "mysql:8.0.36",
			WaitingFor: wait.ForLog("port: 3306  MySQL Community Server - GPL"),
		},
		Started: true,
	}

	lc := &msgsLogConsumer{}

	err := testcontainers.WithLogConsumers(lc)(&req)
	require.NoError(t, err)

	ctx := context.Background()
	c, err := testcontainers.GenericContainer(ctx, req)
	testcontainers.CleanupContainer(t, c)
	// we expect an error because the MySQL environment variables are not set
	// but this is expected because we just want to test the log consumer
	require.ErrorContains(t, err, "container exited with code 1")
	require.NotEmpty(t, lc.msgs)
}

func TestWithLogConsumerConfig(t *testing.T) {
	lc := &msgsLogConsumer{}

	t.Run("add-to-nil", func(t *testing.T) {
		req := testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Image: "alpine",
			},
		}

		err := testcontainers.WithLogConsumerConfig(&testcontainers.LogConsumerConfig{
			Consumers: []testcontainers.LogConsumer{lc},
		})(&req)
		require.NoError(t, err)

		require.Equal(t, []testcontainers.LogConsumer{lc}, req.LogConsumerCfg.Consumers)
	})

	t.Run("replace-existing", func(t *testing.T) {
		req := testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Image: "alpine",
				LogConsumerCfg: &testcontainers.LogConsumerConfig{
					Consumers: []testcontainers.LogConsumer{testcontainers.NewFooLogConsumer(t)},
				},
			},
		}

		err := testcontainers.WithLogConsumerConfig(&testcontainers.LogConsumerConfig{
			Consumers: []testcontainers.LogConsumer{lc},
		})(&req)
		require.NoError(t, err)

		require.Equal(t, []testcontainers.LogConsumer{lc}, req.LogConsumerCfg.Consumers)
	})
}

func TestWithStartupCommand(t *testing.T) {
	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:      "alpine",
			Entrypoint: []string{"tail", "-f", "/dev/null"},
		},
		Started: true,
	}

	testExec := testcontainers.NewRawCommand([]string{"touch", ".testcontainers"}, exec.WithWorkingDir("/tmp"))

	err := testcontainers.WithStartupCommand(testExec)(&req)
	require.NoError(t, err)

	require.Len(t, req.LifecycleHooks, 1)
	require.Len(t, req.LifecycleHooks[0].PostStarts, 1)

	c, err := testcontainers.GenericContainer(context.Background(), req)
	testcontainers.CleanupContainer(t, c)
	require.NoError(t, err)

	_, reader, err := c.Exec(context.Background(), []string{"ls", "/tmp/.testcontainers"}, exec.Multiplexed())
	require.NoError(t, err)

	content, err := io.ReadAll(reader)
	require.NoError(t, err)
	require.Equal(t, "/tmp/.testcontainers\n", string(content))
}

func TestWithAfterReadyCommand(t *testing.T) {
	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:      "alpine",
			Entrypoint: []string{"tail", "-f", "/dev/null"},
		},
		Started: true,
	}

	testExec := testcontainers.NewRawCommand([]string{"touch", "/tmp/.testcontainers"})

	err := testcontainers.WithAfterReadyCommand(testExec)(&req)
	require.NoError(t, err)

	require.Len(t, req.LifecycleHooks, 1)
	require.Len(t, req.LifecycleHooks[0].PostReadies, 1)

	c, err := testcontainers.GenericContainer(context.Background(), req)
	testcontainers.CleanupContainer(t, c)
	require.NoError(t, err)

	_, reader, err := c.Exec(context.Background(), []string{"ls", "/tmp/.testcontainers"}, exec.Multiplexed())
	require.NoError(t, err)

	content, err := io.ReadAll(reader)
	require.NoError(t, err)
	require.Equal(t, "/tmp/.testcontainers\n", string(content))
}

func TestWithEnv(t *testing.T) {
	testEnv := func(t *testing.T, initial map[string]string, add map[string]string, expected map[string]string) {
		t.Helper()

		req := &testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Env: initial,
			},
		}
		opt := testcontainers.WithEnv(add)
		require.NoError(t, opt.Customize(req))
		require.Equal(t, expected, req.Env)
	}

	t.Run("add-to-existing", func(t *testing.T) {
		testEnv(t,
			map[string]string{"KEY1": "VAL1"},
			map[string]string{"KEY2": "VAL2"},
			map[string]string{
				"KEY1": "VAL1",
				"KEY2": "VAL2",
			},
		)
	})

	t.Run("add-to-nil", func(t *testing.T) {
		testEnv(t,
			nil,
			map[string]string{"KEY2": "VAL2"},
			map[string]string{"KEY2": "VAL2"},
		)
	})

	t.Run("override-existing", func(t *testing.T) {
		testEnv(t,
			map[string]string{
				"KEY1": "VAL1",
				"KEY2": "VAL2",
			},
			map[string]string{"KEY2": "VAL3"},
			map[string]string{
				"KEY1": "VAL1",
				"KEY2": "VAL3",
			},
		)
	})
}

func TestWithHostPortAccess(t *testing.T) {
	testHostPorts := func(t *testing.T, initial []int, add []int, expected []int) {
		t.Helper()

		req := &testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				HostAccessPorts: initial,
			},
		}
		opt := testcontainers.WithHostPortAccess(add...)
		require.NoError(t, opt.Customize(req))
		require.Equal(t, expected, req.HostAccessPorts)
	}

	t.Run("add-to-existing", func(t *testing.T) {
		testHostPorts(t,
			[]int{1, 2},
			[]int{3, 4},
			[]int{1, 2, 3, 4},
		)
	})

	t.Run("add-to-nil", func(t *testing.T) {
		testHostPorts(t,
			nil,
			[]int{3, 4},
			[]int{3, 4},
		)
	})
}

func TestWithEntrypoint(t *testing.T) {
	testEntrypoint := func(t *testing.T, initial []string, add []string, expected []string) {
		t.Helper()

		req := &testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Entrypoint: initial,
			},
		}
		opt := testcontainers.WithEntrypoint(add...)
		require.NoError(t, opt.Customize(req))
		require.Equal(t, expected, req.Entrypoint)
	}

	t.Run("replace-existing", func(t *testing.T) {
		testEntrypoint(t,
			[]string{"/bin/sh"},
			[]string{"pwd"},
			[]string{"pwd"},
		)
	})

	t.Run("replace-nil", func(t *testing.T) {
		testEntrypoint(t,
			nil,
			[]string{"/bin/sh", "-c"},
			[]string{"/bin/sh", "-c"},
		)
	})
}

func TestWithEntrypointArgs(t *testing.T) {
	testEntrypoint := func(t *testing.T, initial []string, add []string, expected []string) {
		t.Helper()

		req := &testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Entrypoint: initial,
			},
		}
		opt := testcontainers.WithEntrypointArgs(add...)
		require.NoError(t, opt.Customize(req))
		require.Equal(t, expected, req.Entrypoint)
	}

	t.Run("add-to-existing", func(t *testing.T) {
		testEntrypoint(t,
			[]string{"/bin/sh"},
			[]string{"-c", "echo hello"},
			[]string{"/bin/sh", "-c", "echo hello"},
		)
	})

	t.Run("add-to-nil", func(t *testing.T) {
		testEntrypoint(t,
			nil,
			[]string{"/bin/sh", "-c"},
			[]string{"/bin/sh", "-c"},
		)
	})
}

func TestWithExposedPorts(t *testing.T) {
	testPorts := func(t *testing.T, initial []string, add []string, expected []string) {
		t.Helper()

		req := &testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				ExposedPorts: initial,
			},
		}
		opt := testcontainers.WithExposedPorts(add...)
		require.NoError(t, opt.Customize(req))
		require.Equal(t, expected, req.ExposedPorts)
	}

	t.Run("add-to-existing", func(t *testing.T) {
		testPorts(t,
			[]string{"8080/tcp"},
			[]string{"9090/tcp"},
			[]string{"8080/tcp", "9090/tcp"},
		)
	})

	t.Run("add-to-nil", func(t *testing.T) {
		testPorts(t,
			nil,
			[]string{"8080/tcp"},
			[]string{"8080/tcp"},
		)
	})
}

func TestWithCmd(t *testing.T) {
	testCmd := func(t *testing.T, initial []string, add []string, expected []string) {
		t.Helper()

		req := &testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Cmd: initial,
			},
		}
		opt := testcontainers.WithCmd(add...)
		require.NoError(t, opt.Customize(req))
		require.Equal(t, expected, req.Cmd)
	}

	t.Run("replace-existing", func(t *testing.T) {
		testCmd(t,
			[]string{"echo"},
			[]string{"hello", "world"},
			[]string{"hello", "world"},
		)
	})

	t.Run("replace-nil", func(t *testing.T) {
		testCmd(t,
			nil,
			[]string{"echo", "hello"},
			[]string{"echo", "hello"},
		)
	})
}

func TestWithAlwaysPull(t *testing.T) {
	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: "alpine",
		},
	}

	opt := testcontainers.WithAlwaysPull()
	require.NoError(t, opt.Customize(&req))
	require.True(t, req.AlwaysPullImage)
}

func TestWithImagePlatform(t *testing.T) {
	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: "alpine",
		},
	}

	opt := testcontainers.WithImagePlatform("linux/amd64")
	require.NoError(t, opt.Customize(&req))
	require.Equal(t, "linux/amd64", req.ImagePlatform)
}

func TestWithCmdArgs(t *testing.T) {
	testCmd := func(t *testing.T, initial []string, add []string, expected []string) {
		t.Helper()

		req := &testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Cmd: initial,
			},
		}
		opt := testcontainers.WithCmdArgs(add...)
		require.NoError(t, opt.Customize(req))
		require.Equal(t, expected, req.Cmd)
	}

	t.Run("add-to-existing", func(t *testing.T) {
		testCmd(t,
			[]string{"echo"},
			[]string{"hello", "world"},
			[]string{"echo", "hello", "world"},
		)
	})

	t.Run("add-to-nil", func(t *testing.T) {
		testCmd(t,
			nil,
			[]string{"echo", "hello"},
			[]string{"echo", "hello"},
		)
	})
}

func TestWithLabels(t *testing.T) {
	testLabels := func(t *testing.T, initial map[string]string, add map[string]string, expected map[string]string) {
		t.Helper()

		req := &testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Labels: initial,
			},
		}
		opt := testcontainers.WithLabels(add)
		require.NoError(t, opt.Customize(req))
		require.Equal(t, expected, req.Labels)
	}

	t.Run("add-to-existing", func(t *testing.T) {
		testLabels(t,
			map[string]string{"key1": "value1"},
			map[string]string{"key2": "value2"},
			map[string]string{"key1": "value1", "key2": "value2"},
		)
	})

	t.Run("add-to-nil", func(t *testing.T) {
		testLabels(t,
			nil,
			map[string]string{"key1": "value1"},
			map[string]string{"key1": "value1"},
		)
	})
}

func TestWithLifecycleHooks(t *testing.T) {
	testHook := testcontainers.DefaultLoggingHook(nil)

	testLifecycleHooks := func(t *testing.T, replace bool, initial []testcontainers.ContainerLifecycleHooks, add []testcontainers.ContainerLifecycleHooks, expected []testcontainers.ContainerLifecycleHooks) {
		t.Helper()

		req := &testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				LifecycleHooks: initial,
			},
		}

		var opt testcontainers.CustomizeRequestOption
		if replace {
			opt = testcontainers.WithLifecycleHooks(add...)
		} else {
			opt = testcontainers.WithAdditionalLifecycleHooks(add...)
		}
		require.NoError(t, opt.Customize(req))
		require.Len(t, req.LifecycleHooks, len(expected))
		for i, hook := range expected {
			require.Equal(t, hook, req.LifecycleHooks[i])
		}
	}

	t.Run("replace-nil", func(t *testing.T) {
		testLifecycleHooks(t,
			true,
			nil,
			[]testcontainers.ContainerLifecycleHooks{testHook},
			[]testcontainers.ContainerLifecycleHooks{testHook},
		)
	})

	t.Run("replace-existing", func(t *testing.T) {
		testLifecycleHooks(t,
			true,
			[]testcontainers.ContainerLifecycleHooks{testHook},
			[]testcontainers.ContainerLifecycleHooks{testHook},
			[]testcontainers.ContainerLifecycleHooks{testHook},
		)
	})

	t.Run("add-to-nil", func(t *testing.T) {
		testLifecycleHooks(t,
			false,
			nil,
			[]testcontainers.ContainerLifecycleHooks{testHook},
			[]testcontainers.ContainerLifecycleHooks{testHook},
		)
	})

	t.Run("add-to-existing", func(t *testing.T) {
		testLifecycleHooks(t,
			false,
			[]testcontainers.ContainerLifecycleHooks{testHook},
			[]testcontainers.ContainerLifecycleHooks{testHook},
			[]testcontainers.ContainerLifecycleHooks{testHook, testHook},
		)
	})
}

func TestWithMounts(t *testing.T) {
	testMounts := func(t *testing.T, initial []testcontainers.ContainerMount, add []testcontainers.ContainerMount, expected testcontainers.ContainerMounts) {
		t.Helper()

		req := &testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Mounts: initial,
			},
		}
		opt := testcontainers.WithMounts(add...)
		require.NoError(t, opt.Customize(req))
		require.Equal(t, expected, req.Mounts)
	}

	t.Run("add-to-existing", func(t *testing.T) {
		testMounts(t,
			[]testcontainers.ContainerMount{
				{Source: testcontainers.GenericVolumeMountSource{Name: "source1"}, Target: "/tmp/target1"},
			},
			[]testcontainers.ContainerMount{
				{Source: testcontainers.GenericVolumeMountSource{Name: "source2"}, Target: "/tmp/target2"},
			},
			testcontainers.ContainerMounts{
				{Source: testcontainers.GenericVolumeMountSource{Name: "source1"}, Target: "/tmp/target1"},
				{Source: testcontainers.GenericVolumeMountSource{Name: "source2"}, Target: "/tmp/target2"},
			},
		)
	})

	t.Run("add-to-nil", func(t *testing.T) {
		testMounts(t,
			nil,
			[]testcontainers.ContainerMount{
				{Source: testcontainers.GenericVolumeMountSource{Name: "source1"}, Target: "/tmp/target1"},
			},
			testcontainers.ContainerMounts{
				{Source: testcontainers.GenericVolumeMountSource{Name: "source1"}, Target: "/tmp/target1"},
			},
		)
	})
}

func TestWithTmpfs(t *testing.T) {
	testTmpfs := func(t *testing.T, initial map[string]string, add map[string]string, expected map[string]string) {
		t.Helper()

		req := &testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Tmpfs: initial,
			},
		}
		opt := testcontainers.WithTmpfs(add)
		require.NoError(t, opt.Customize(req))
		require.Equal(t, expected, req.Tmpfs)
	}

	t.Run("add-to-existing", func(t *testing.T) {
		testTmpfs(t,
			map[string]string{"/tmp1": "size=100m"},
			map[string]string{"/tmp2": "size=200m"},
			map[string]string{"/tmp1": "size=100m", "/tmp2": "size=200m"},
		)
	})

	t.Run("add-to-nil", func(t *testing.T) {
		testTmpfs(t,
			nil,
			map[string]string{"/tmp1": "size=100m"},
			map[string]string{"/tmp1": "size=100m"},
		)
	})
}

func TestWithFiles(t *testing.T) {
	testFiles := func(t *testing.T, initial []testcontainers.ContainerFile, add []testcontainers.ContainerFile, expected []testcontainers.ContainerFile) {
		t.Helper()

		req := &testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Files: initial,
			},
		}
		opt := testcontainers.WithFiles(add...)
		require.NoError(t, opt.Customize(req))
		require.Equal(t, expected, req.Files)
	}

	t.Run("add-to-existing", func(t *testing.T) {
		testFiles(t,
			[]testcontainers.ContainerFile{{HostFilePath: "/tmp/file1", ContainerFilePath: "/container/file1"}},
			[]testcontainers.ContainerFile{{HostFilePath: "/tmp/file2", ContainerFilePath: "/container/file2"}},
			[]testcontainers.ContainerFile{
				{HostFilePath: "/tmp/file1", ContainerFilePath: "/container/file1"},
				{HostFilePath: "/tmp/file2", ContainerFilePath: "/container/file2"},
			},
		)
	})

	t.Run("add-to-nil", func(t *testing.T) {
		testFiles(t,
			nil,
			[]testcontainers.ContainerFile{{HostFilePath: "/tmp/file1", ContainerFilePath: "/container/file1"}},
			[]testcontainers.ContainerFile{{HostFilePath: "/tmp/file1", ContainerFilePath: "/container/file1"}},
		)
	})
}

func TestWithDockerfile(t *testing.T) {
	df := testcontainers.FromDockerfile{
		Context:    ".",
		Dockerfile: "Dockerfile",
		Repo:       "testcontainers",
		Tag:        "latest",
		BuildArgs:  map[string]*string{"ARG1": nil, "ARG2": nil},
	}

	req := &testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{},
	}

	opt := testcontainers.WithDockerfile(df)
	require.NoError(t, opt.Customize(req))
	require.Equal(t, df, req.FromDockerfile)
	require.Equal(t, ".", req.Context)
	require.Equal(t, "Dockerfile", req.Dockerfile)
	require.Equal(t, "testcontainers", req.Repo)
	require.Equal(t, "latest", req.Tag)
	require.Equal(t, map[string]*string{"ARG1": nil, "ARG2": nil}, req.BuildArgs)
}

func TestWithImageMount(t *testing.T) {
	cli, err := testcontainers.NewDockerClientWithOpts(context.Background())
	require.NoError(t, err)

	info, err := cli.Info(context.Background())
	require.NoError(t, err)

	// skip if the major version of the server is not v28 or greater
	if info.ServerVersion < "28.0.0" {
		t.Skipf("skipping test because the server version is not v28 or greater")
	}

	t.Run("valid", func(t *testing.T) {
		req := testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Image: "alpine",
			},
		}

		err := testcontainers.WithImageMount("alpine", "root/.ollama/models/", "/root/.ollama/models/")(&req)
		require.NoError(t, err)

		require.Len(t, req.Mounts, 1)

		src := req.Mounts[0].Source

		require.Equal(t, testcontainers.NewDockerImageMountSource("alpine", "root/.ollama/models/"), src)
		require.Equal(t, "alpine", src.Source())
		require.Equal(t, testcontainers.MountTypeImage, src.Type())

		dst := req.Mounts[0].Target
		require.Equal(t, testcontainers.ContainerMountTarget("/root/.ollama/models/"), dst)
	})

	t.Run("invalid", func(t *testing.T) {
		req := testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Image: "alpine",
			},
		}

		err := testcontainers.WithImageMount("alpine", "/root/.ollama/models/", "/root/.ollama/models/")(&req)
		require.Error(t, err)
	})

	t.Run("invalid-dots", func(t *testing.T) {
		req := testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Image: "alpine",
			},
		}

		err := testcontainers.WithImageMount("alpine", "../root/.ollama/models/", "/root/.ollama/models/")(&req)
		require.Error(t, err)
	})
}

func TestWithReuseByName_Succeeds(t *testing.T) {
	t.Parallel()
	req := &testcontainers.GenericContainerRequest{}
	containerName := "pg-test"

	opt := testcontainers.WithReuseByName(containerName)
	err := opt.Customize(req)

	require.NoError(t, err)
	require.True(t, req.Reuse)
	require.Equal(t, containerName, req.Name)
}

func TestWithReuseByName_ErrorsWithoutContainerNameProvided(t *testing.T) {
	t.Parallel()
	req := &testcontainers.GenericContainerRequest{}

	opt := testcontainers.WithReuseByName("")
	err := opt.Customize(req)

	require.ErrorContains(t, err, "container name must be provided")
	require.False(t, req.Reuse)
	require.Empty(t, req.Name)
}

func TestWithName(t *testing.T) {
	t.Parallel()
	req := &testcontainers.GenericContainerRequest{}

	opt := testcontainers.WithName("pg-test")
	err := opt.Customize(req)
	require.NoError(t, err)
	require.Equal(t, "pg-test", req.Name)

	t.Run("empty", func(t *testing.T) {
		req := &testcontainers.GenericContainerRequest{}

		opt := testcontainers.WithName("")
		err := opt.Customize(req)
		require.ErrorContains(t, err, "container name must be provided")
	})
}

func TestWithNoStart(t *testing.T) {
	t.Parallel()
	req := &testcontainers.GenericContainerRequest{}

	opt := testcontainers.WithNoStart()
	err := opt.Customize(req)
	require.NoError(t, err)
	require.False(t, req.Started)
}

func TestWithWaitStrategy(t *testing.T) {
	testDuration := 10 * time.Second
	defaultDuration := 60 * time.Second

	waitForFoo := wait.ForLog("foo")
	waitForBar := wait.ForLog("bar")

	testWaitFor := func(t *testing.T, replace bool, customDuration *time.Duration, initial wait.Strategy, add wait.Strategy, expected wait.Strategy) {
		t.Helper()

		req := &testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				WaitingFor: initial,
			},
		}

		var opt testcontainers.CustomizeRequestOption
		if replace {
			opt = testcontainers.WithWaitStrategy(add)
			if customDuration != nil {
				opt = testcontainers.WithWaitStrategyAndDeadline(*customDuration, add)
			}
		} else {
			opt = testcontainers.WithAdditionalWaitStrategy(add)
			if customDuration != nil {
				opt = testcontainers.WithAdditionalWaitStrategyAndDeadline(*customDuration, add)
			}
		}
		require.NoError(t, opt.Customize(req))
		require.Equal(t, expected, req.WaitingFor)
	}

	t.Run("replace-nil", func(t *testing.T) {
		t.Run("default-duration", func(t *testing.T) {
			testWaitFor(t,
				true,
				nil,
				nil,
				waitForFoo,
				wait.ForAll(waitForFoo).WithDeadline(defaultDuration),
			)
		})

		t.Run("custom-duration", func(t *testing.T) {
			testWaitFor(t,
				true,
				&testDuration,
				nil,
				waitForFoo,
				wait.ForAll(waitForFoo).WithDeadline(testDuration),
			)
		})
	})

	t.Run("replace-existing", func(t *testing.T) {
		t.Run("default-duration", func(t *testing.T) {
			testWaitFor(t,
				true,
				nil,
				waitForFoo,
				waitForBar,
				wait.ForAll(waitForBar).WithDeadline(defaultDuration),
			)
		})

		t.Run("custom-duration", func(t *testing.T) {
			testWaitFor(t,
				true,
				&testDuration,
				waitForFoo,
				waitForBar,
				wait.ForAll(waitForBar).WithDeadline(testDuration),
			)
		})
	})

	t.Run("add-to-nil", func(t *testing.T) {
		t.Run("default-duration", func(t *testing.T) {
			testWaitFor(t,
				false,
				nil,
				nil,
				waitForFoo,
				wait.ForAll(waitForFoo).WithDeadline(defaultDuration),
			)
		})

		t.Run("custom-duration", func(t *testing.T) {
			testWaitFor(t,
				false,
				&testDuration,
				nil,
				waitForFoo,
				wait.ForAll(waitForFoo).WithDeadline(testDuration),
			)
		})
	})

	t.Run("add-to-existing", func(t *testing.T) {
		t.Run("default-duration", func(t *testing.T) {
			testWaitFor(t,
				false,
				nil,
				waitForFoo,
				waitForBar,
				wait.ForAll(waitForFoo, waitForBar).WithDeadline(defaultDuration),
			)
		})

		t.Run("custom-duration", func(t *testing.T) {
			testWaitFor(t,
				false,
				&testDuration,
				waitForFoo,
				waitForBar,
				wait.ForAll(waitForFoo, waitForBar).WithDeadline(testDuration),
			)
		})
	})
}
