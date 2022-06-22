package testcontainers

import (
	"context"
	"fmt"
	"github.com/docker/cli/cli-plugins/manager"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/connhelper"
	cliflags "github.com/docker/cli/cli/flags"
	cmd "github.com/docker/compose/v2/cmd/compose"
	"github.com/docker/compose/v2/pkg/compose"
	"github.com/docker/compose/v2/pkg/progress"
	"github.com/docker/docker/client"
	"os"
)

type NativeDockerComposeExecutor struct {
}

type memoryWriter struct {
	out []byte
}

type memoryProgressWriter struct {
	out memoryWriter
}

func (w *memoryProgressWriter) Start(context.Context) error {
	return nil
}

func (w *memoryProgressWriter) Stop() {
}

func (w *memoryProgressWriter) Event(e progress.Event) {
	fmt.Fprintln(&w.out, e.ID, e.Text, e.StatusText)
}

func (w *memoryProgressWriter) Events(events []progress.Event) {
	for _, e := range events {
		w.Event(e)
	}
}

func (w *memoryProgressWriter) TailMsgf(string, ...interface{}) {
}

func (w *memoryWriter) Write(p []byte) (n int, err error) {
	w.out = append(w.out, p...)
	n = len(p)
	return
}

func (e *NativeDockerComposeExecutor) Exec(options ComposeExecutorOptions) ExecError {
	cmds := []string{}

	if v, ok := options.Env[envProjectName]; ok {
		cmds = append(cmds, "--project-name", v)
	}

	if len(options.ComposeFiles) > 0 {
		for _, abs := range options.ComposeFiles {
			cmds = append(cmds, "-f", abs)
		}
	} else {
		cmds = append(cmds, "-f", "docker-compose.yml")
	}

	args := append(cmds, options.Args...)

	logger := memoryWriter{out: []byte{}}
	errLogger := memoryWriter{}
	progressLogger := memoryProgressWriter{out: logger}
	ctx := progress.WithContextWriter(options.Context, &progressLogger)

	dockerCli, err := command.NewDockerCli(command.WithOutputStream(&logger), command.WithErrorStream(&errLogger))

	if err != nil {
		return ExecError{Error: err}
	}

	err = dockerCli.Initialize(cliflags.NewClientOptions())
	if err != nil {
		return ExecError{Error: err}
	}

	command.WithInitializeClient(func(dockerCli *command.DockerCli) (client.APIClient, error) {
		cmd := "docker"
		if x := os.Getenv(manager.ReexecEnvvar); x != "" {
			cmd = x
		}
		var flags []string

		helper, err := connhelper.GetCommandConnectionHelper(cmd, flags...)
		if err != nil {
			return nil, err
		}

		return client.NewClientWithOpts(client.WithDialContext(helper.Dialer))
	})
	backend := compose.NewComposeService(dockerCli)

	if err != nil {
		return ExecError{}
	}
	root := cmd.RootCommand(dockerCli, backend)
	root.SetArgs(args)
	root.PersistentPreRunE = nil

	err = root.ExecuteContext(ctx)
	if err != nil {
		return ExecError{}
	}

	return ExecError{
		StdoutOutput: logger.out,
		StderrOutput: []byte{},
	}
}
