package compose_test

import (
	"fmt"

	"github.com/testcontainers/testcontainers-go/modules/compose"
)

func ExampleNewLocalDockerCompose() {
	path := "/path/to/docker-compose.yml"

	_ = compose.NewLocalDockerCompose([]string{path}, "my_project")
}

func ExampleLocalDockerCompose() {
	_ = compose.LocalDockerCompose{
		Executable: "docker compose",
		ComposeFilePaths: []string{
			"/path/to/docker-compose.yml",
			"/path/to/docker-compose-1.yml",
			"/path/to/docker-compose-2.yml",
			"/path/to/docker-compose-3.yml",
		},
		Identifier: "my_project",
		Cmd: []string{
			"up", "-d",
		},
		Env: map[string]string{
			"FOO": "foo",
			"BAR": "bar",
		},
	}
}

func ExampleLocalDockerCompose_Down() {
	path := "/path/to/docker-compose.yml"

	stack := compose.NewLocalDockerCompose([]string{path}, "my_project")

	execError := stack.WithCommand([]string{"up", "-d"}).Invoke()
	if execError.Error != nil {
		_ = fmt.Errorf("Failed when running: %v", execError.Command)
	}

	execError = stack.Down()
	if execError.Error != nil {
		_ = fmt.Errorf("Failed when running: %v", execError.Command)
	}
}

func ExampleLocalDockerCompose_Invoke() {
	path := "/path/to/docker-compose.yml"

	stack := compose.NewLocalDockerCompose([]string{path}, "my_project")

	execError := stack.
		WithCommand([]string{"up", "-d"}).
		WithEnv(map[string]string{
			"bar": "BAR",
		}).
		Invoke()
	if execError.Error != nil {
		_ = fmt.Errorf("Failed when running: %v", execError.Command)
	}
}

func ExampleLocalDockerCompose_WithCommand() {
	path := "/path/to/docker-compose.yml"

	stack := compose.NewLocalDockerCompose([]string{path}, "my_project")

	stack.WithCommand([]string{"up", "-d"})
}

func ExampleLocalDockerCompose_WithEnv() {
	path := "/path/to/docker-compose.yml"

	stack := compose.NewLocalDockerCompose([]string{path}, "my_project")

	stack.WithEnv(map[string]string{
		"FOO": "foo",
		"BAR": "bar",
	})
}
