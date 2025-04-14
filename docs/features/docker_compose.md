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

Use the advanced `NewDockerComposeWith(...)` constructor allowing you to customise the compose execution with options:

- `StackIdentifier`: the identifier for the stack, which is used to name the network and containers. If not passed, a random identifier is generated.
- `WithStackFiles`: specify the Docker Compose stack files to use, as a variadic argument of string paths where the stack files are located.
- `WithStackReaders`: specify the Docker Compose stack files to use, as a variadic argument of `io.Reader` instances. It will create a temporary file in the temp dir of the given O.S., that will be removed after the `Down` method is called. You can use both `WithComposeStackFiles` and `WithComposeStackReaders` at the same time.

<!--codeinclude-->
[Define Compose File](../../modules/compose/compose_examples_test.go) inside_block:defineComposeFile
[Define Options](../../modules/compose/compose_examples_test.go) inside_block:defineStackWithOptions
[Start Compose Stack](../../modules/compose/compose_examples_test.go) inside_block:upComposeStack
[Get Service Names](../../modules/compose/compose_examples_test.go) inside_block:getServiceNames
[Get Service Container](../../modules/compose/compose_examples_test.go) inside_block:getServiceContainer
<!--/codeinclude-->

#### Compose Up options

- `Recreate`: recreate the containers. If any other value than `api.RecreateNever`, `api.RecreateForce` or `api.RecreateDiverged` is provided, the default value `api.RecreateForce` will be used.
- `RecreateDependencies`: recreate dependent containers. If any other value than `api.RecreateNever`, `api.RecreateForce` or `api.RecreateDiverged` is provided, the default value `api.RecreateForce` will be used.
- `RemoveOrphans`: remove orphaned containers when the stack is upped.
- `Wait`: will wait until the containers reached the running|healthy state.

#### Compose Down options

- `RemoveImages`: remove images after the stack is stopped. The `RemoveImagesAll` option will remove all images, while `RemoveImagesLocal` will remove only the images that don't have a tag.
- `RemoveOrphans`: remove orphaned containers after the stack is stopped.
- `RemoveVolumes`: remove volumes after the stack is stopped.

### Interacting with compose services

To interact with service containers after a stack was started it is possible to get a `*testcontainers.DockerContainer` instance via the `ServiceContainer(...)` function.
The function takes a **service name** (and a `context.Context`) and returns either a `*testcontainers.DockerContainer` or an `error`.

Furthermore, there's the convenience function `Services()` to get a list of all services **defined** by the current project.
Note that not all of them need necessarily be correctly started as the information is based on the given compose files.

### Wait strategies

Just like with the containers created by _Testcontainers for Go_, you can also apply wait strategies to `docker compose` services.
The `ComposeStack.WaitForService(...)` function allows you to apply a wait strategy to **a service by name**.
All wait strategies are executed in parallel to both improve startup performance by not blocking too long and to fail
early if something's wrong.

#### Example

<!--codeinclude-->
[Compose Example](../../modules/compose/compose_examples_test.go) inside_block:ExampleNewDockerComposeWith_waitForService
<!--/codeinclude-->

### Compose environment

`docker compose` supports expansion based on environment variables.
The `ComposeStack` supports this as well in two different variants:

- `ComposeStack.WithEnv(m map[string]string) ComposeStack` to parameterize stacks from your test code
- `ComposeStack.WithOsEnv() ComposeStack` to parameterize tests from the OS environment e.g. in CI environments

### Docs

Also have a look at [ComposeStack](https://pkg.go.dev/github.com/testcontainers/testcontainers-go#ComposeStack) docs for
further information.

## Usage of the deprecated Local `docker compose` binary

!!! warning
    This API is deprecated and superseded by `ComposeStack` which takes advantage of `compose` v2 being
    implemented in Go as well by directly using the upstream project.

You can override Testcontainers' default behaviour and make it use a
docker compose binary installed on the local machine. This will generally yield
an experience that is closer to running docker compose locally, with the caveat
that Docker Compose needs to be present on dev and CI machines.

### Examples

<!--codeinclude-->
[Invoke Example](../../modules/compose/compose_local_examples_test.go) inside_block:ExampleLocalDockerCompose_Invoke
<!--/codeinclude-->

Note that the environment variables in the `env` map will be applied, if
possible, to the existing variables declared in the Docker Compose file.

In the following example, we demonstrate how to stop a Docker Compose created project using the
convenient `Down` method.

<!--codeinclude-->
[Down Example](../../modules/compose/compose_local_examples_test.go) inside_block:ExampleLocalDockerCompose_Down
<!--/codeinclude-->
