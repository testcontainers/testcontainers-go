# Using Docker Compose

Similar to generic containers support, it's also possible to run a bespoke set
of services specified in a docker-compose.yml file.

This is intended to be useful on projects where Docker Compose is already used
in dev or other environments to define services that an application may be
dependent upon.

## Using `docker compose` directly

!!!warning
	The minimal version of Go required to use this module is **1.21**.

```
go get github.com/testcontainers/testcontainers-go/modules/compose
```

Because `compose` v2 is implemented in Go it's possible for _Testcontainers for Go_ to
use [`github.com/docker/compose`](https://github.com/docker/compose) directly and skip any process execution/_docker-compose-in-a-container_ scenario.
The `ComposeStack` API exposes this variant of using `docker compose` in an easy way.

Before using the Compose module, there is some configuration that needs to be applied first.
It customizes the behaviour of the `Ryuk` container, which is used to clean up the resources created by the `docker compose` stack.
Please refer to [the Ryuk configuration](../configuration/#customizing-ryuk-the-resource-reaper) for more information.

### Usage

Use the convenience `NewDockerCompose(...)` constructor which creates a random identifier and takes a variable number
of stack files:

```go
package example_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	tc "github.com/testcontainers/testcontainers-go/modules/compose"
)

func TestSomething(t *testing.T) {
	compose, err := tc.NewDockerCompose("testdata/docker-compose.yml")
	require.NoError(t, err, "NewDockerComposeAPI()")

	t.Cleanup(func() {
		require.NoError(t, compose.Down(context.Background(), tc.RemoveOrphans(true), tc.RemoveImagesLocal), "compose.Down()")
	})

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	require.NoError(t, compose.Up(ctx, tc.Wait(true)), "compose.Up()")

	// do some testing here
}
```

Use the advanced `NewDockerComposeWith(...)` constructor allowing you to customise the compose execution with options:

- `StackIdentifier`: the identifier for the stack, which is used to name the network and containers. If not passed, a random identifier is generated.
- `WithStackFiles`: specify the Docker Compose stack files to use, as a variadic argument of string paths where the stack files are located.
- `WithStackReaders`: specify the Docker Compose stack files to use, as a variadic argument of `io.Reader` instances. It will create a temporary file in the temp dir of the given O.S., that will be removed after the `Down` method is called. You can use both `WithComposeStackFiles` and `WithComposeStackReaders` at the same time.

#### Compose Up options

- `Recreate`: recreate the containers. If any other value than `api.RecreateNever`, `api.RecreateForce` or `api.RecreateDiverged` is provided, the default value `api.RecreateForce` will be used.
- `RecreateDependencies`: recreate dependent containers. If any other value than `api.RecreateNever`, `api.RecreateForce` or `api.RecreateDiverged` is provided, the default value `api.RecreateForce` will be used.
- `RemoveOrphans`: remove orphaned containers when the stack is upped.
- `Wait`: will wait until the containers reached the running|healthy state.

#### Compose Down options

- `RemoveImages`: remove images after the stack is stopped. The `RemoveImagesAll` option will remove all images, while `RemoveImagesLocal` will remove only the images that don't have a tag.
- `RemoveOrphans`: remove orphaned containers after the stack is stopped.
- `RemoveVolumes`: remove volumes after the stack is stopped.

#### Example

```go
package example_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/docker/compose/v2/pkg/api"
	tc "github.com/testcontainers/testcontainers-go/modules/compose"
)

func TestSomethingElse(t *testing.T) {
	identifier := tc.StackIdentifier("some_ident")
	compose, err := tc.NewDockerComposeWith(tc.WithStackFiles("./testdata/docker-compose-simple.yml"), identifier)
	require.NoError(t, err, "NewDockerComposeAPIWith()")

	t.Cleanup(func() {
		require.NoError(t, compose.Down(context.Background(), tc.RemoveOrphans(true), tc.RemoveImagesLocal), "compose.Down()")
	})

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	require.NoError(t, compose.Up(ctx, tc.WithRecreate(api.RecreateNever), tc.Wait(true)), "compose.Up()")

	// do some testing here
}
```

### Interacting with compose services

To interact with service containers after a stack was started it is possible to get an `*tc.DockerContainer` instance via the `ServiceContainer(...)` function.
The function takes a **service name** (and a `context.Context`) and returns either a `*tc.DockerContainer` or an `error`.
This is different to the previous `LocalDockerCompose` API where service containers were accessed via their **container name** e.g. `mysql_1` or `mysql-1` (depending on the version of `docker compose`).

Furthermore, there's the convenience function `Services()` to get a list of all services **defined** by the current project.
Note that not all of them need necessarily be correctly started as the information is based on the given compose files.

### Wait strategies

Just like with regular test containers you can also apply wait strategies to `docker compose` services.
The `ComposeStack.WaitForService(...)` function allows you to apply a wait strategy to **a service by name**.
All wait strategies are executed in parallel to both improve startup performance by not blocking too long and to fail
early if something's wrong.

#### Example

```go
package example_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	tc "github.com/testcontainers/testcontainers-go/modules/compose"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestSomethingWithWaiting(t *testing.T) {
	identifier := tc.StackIdentifier("some_ident")
	compose, err := tc.NewDockerComposeWith(tc.WithStackFiles("./testdata/docker-compose-simple.yml"), identifier)
	require.NoError(t, err, "NewDockerComposeAPIWith()")

	t.Cleanup(func() {
		require.NoError(t, compose.Down(context.Background(), tc.RemoveOrphans(true), tc.RemoveImagesLocal), "compose.Down()")
	})

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	err = compose.
		WaitForService("nginx", wait.NewHTTPStrategy("/").WithPort("80/tcp").WithStartupTimeout(10*time.Second)).
		Up(ctx, tc.Wait(true))
	
	require.NoError(t, err, "compose.Up()")

	// do some testing here
}
```

### Compose environment

`docker compose` supports expansion based on environment variables.
The `ComposeStack` supports this as well in two different variants:

- `ComposeStack.WithEnv(m map[string]string) ComposeStack` to parameterize stacks from your test code
- `ComposeStack.WithOsEnv() ComposeStack` to parameterize tests from the OS environment e.g. in CI environments

### Docs

Also have a look at [ComposeStack](https://pkg.go.dev/github.com/testcontainers/testcontainers-go#ComposeStack) docs for
further information.

## Usage of the `docker compose` binary

_Note:_ this API is deprecated and superseded by `ComposeStack` which takes advantage of `compose` v2 being
implemented in Go as well by directly using the upstream project.

You can override Testcontainers' default behaviour and make it use a
docker compose binary installed on the local machine. This will generally yield
an experience that is closer to running docker compose locally, with the caveat
that Docker Compose needs to be present on dev and CI machines.

### Examples

```go
composeFilePaths := []string {"testdata/docker-compose.yml"}
identifier := strings.ToLower(uuid.New().String())

compose := tc.NewLocalDockerCompose(composeFilePaths, identifier)
execError := compose.
    WithCommand([]string{"up", "-d"}).
    WithEnv(map[string]string {
        "key1": "value1",
        "key2": "value2",
    }).
    Invoke()

err := execError.Error
if err != nil {
    return fmt.Errorf("Could not run compose file: %v - %v", composeFilePaths, err)
}
return nil
```

Note that the environment variables in the `env` map will be applied, if
possible, to the existing variables declared in the Docker Compose file.

In the following example, we demonstrate how to stop a Docker Compose created project using the
convenient `Down` method.

```go
composeFilePaths := []string{"testdata/docker-compose.yml"}

compose := tc.NewLocalDockerCompose(composeFilePaths, identifierFromExistingRunningCompose)
execError := compose.Down()
err := execError.Error
if err != nil {
    return fmt.Errorf("Could not run compose file: %v - %v", composeFilePaths, err)
}
return nil
```

