package testcontainers

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

type ComposeExecutorOptions struct {
	Pwd          string
	ComposeFiles []string
	Env          map[string]string
	Args         []string
	Context      context.Context
}

type ComposeExecutor func(options ComposeExecutorOptions) ExecError

func ContainerizedDockerComposeExecutor(options ComposeExecutorOptions) ExecError {
	req := ContainerRequest{
		Image: "docker/compose:1.29.2",
		Env:   options.Env,
		Mounts: Mounts(
			BindMount(
				coalesce(os.Getenv("TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE"), "/var/run/docker.sock"),
				"/var/run/docker.sock",
			),
			BindMount(options.Pwd, "/app"),
		),
		Cmd:        append([]string{"docker-compose"}, options.Args...),
		SkipReaper: true,
	}

	provider, err := ProviderDocker.GetProvider()
	if err != nil {
		return ExecError{
			Error: err,
		}
	}

	var composeFiles []string

	for _, p := range options.ComposeFiles {
		if !strings.HasPrefix(p, filepath.Clean(options.Pwd)+string(filepath.Separator)) {
			return ExecError{
				Error: fmt.Errorf("one of the compose files out of pwd directory"),
			}
		}

		composeFiles = append(composeFiles, "/app/"+filepath.ToSlash(p[len(options.Pwd)+1:]))
	}

	container, err := provider.RunContainer(options.Context, req)
	if err != nil {
		return ExecError{
			Command: req.Cmd,
			Error:   err,
		}
	}

	state, err := container.State(options.Context)

	defer container.StopLogProducer()
	_ = container.StartLogProducer(options.Context)
	//if dc.Logger != nil {
	//	container.FollowOutput(&loggerLogConsumer{Logger: dc.Logger})
	//}

	for err == nil && state.Running {
		time.Sleep(10 * time.Second)
		state, err = container.State(options.Context)
	}

	if state.Error != "" {
		err = fmt.Errorf(state.Error)
	}

	return ExecError{
		Command: req.Cmd,
		Error:   err,
	}
}

// LocalDockerComposeExecutor executes a program with arguments and environment variables inside a specific directory
func LocalDockerComposeExecutor(options ComposeExecutorOptions) ExecError {

	dirContext := options.Pwd
	environment := options.Env
	args := options.Args
	binary := "docker-compose"
	if runtime.GOOS == "windows" {
		binary = "docker-compose.exe"
	}

	if which(binary) != nil {
		return ExecError{
			Command: []string{binary},
			Error:   fmt.Errorf("Local Docker Compose not found. Is %s on the PATH?", binary),
		}
	}

	cmds := []string{}
	if len(options.ComposeFiles) > 0 {
		for _, abs := range options.ComposeFiles {
			cmds = append(cmds, "-f", abs)
		}
	} else {
		cmds = append(cmds, "-f", "docker-compose.yml")
	}

	args = append(cmds, args...)

	var errStdout, errStderr error

	cmd := exec.Command(binary, args...)
	cmd.Dir = dirContext
	cmd.Env = os.Environ()

	for key, value := range environment {
		cmd.Env = append(cmd.Env, key+"="+value)
	}

	stdoutIn, _ := cmd.StdoutPipe()
	stderrIn, _ := cmd.StderrPipe()

	stdout := newCapturingPassThroughWriter(os.Stdout)
	stderr := newCapturingPassThroughWriter(os.Stderr)

	err := cmd.Start()
	if err != nil {
		execCmd := []string{"Starting command", dirContext, binary}
		execCmd = append(execCmd, args...)

		return ExecError{
			// add information about the CMD and arguments used
			Command:      execCmd,
			StdoutOutput: stdout.Bytes(),
			StderrOutput: stderr.Bytes(),
			Error:        err,
			Stderr:       errStderr,
			Stdout:       errStdout,
		}
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		_, errStdout = io.Copy(stdout, stdoutIn)
		wg.Done()
	}()

	_, errStderr = io.Copy(stderr, stderrIn)
	wg.Wait()

	err = cmd.Wait()

	execCmd := []string{"Reading std", dirContext, binary}
	execCmd = append(execCmd, args...)

	return ExecError{
		Command:      execCmd,
		StdoutOutput: stdout.Bytes(),
		StderrOutput: stderr.Bytes(),
		Error:        err,
		Stderr:       errStderr,
		Stdout:       errStdout,
	}
}
