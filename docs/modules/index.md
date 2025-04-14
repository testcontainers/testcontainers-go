# Testcontainers for Go modules

In this section you'll discover how to create Go modules for _Testcontainers for Go_, including the steps to follow, the best practices, and the guidelines to consider before creating a new module.

## Before creating a new module

First, you have to consider the following things before creating a new module to the project:

1. Please check if the module you are interested in is already present in the [Modules Catalog](https://testcontainers.com/modules/?language=go).
If it's already there, you can use it as a dependency in your project, and you can also contribute to the module if you want to improve it!
2. If it is not present there, please consider where you are going to host the module: under your own GitHub account as a **Community module**, or under the `testcontainers-go` repository.
3. If you're a vendor interested in creating an **Official module** for your project/product, please contact us, see below.

In any case, we will be happy to assist you in the process of creating the module. Please feel free to ask for assistance in our [Slack](https://slack.testcontainers.org/) or in the [GitHub Discussions](https://github.com/testcontainers/testcontainers-go/discussions).

### Community modules

If you are going to host the module under your own GitHub account, please consider the following:

- you'll have to follow the same guidelines as the ones for the modules hosted under the `testcontainers-go` repository,
including the naming conventions, the documentation, the tests and the testable examples, and defining a proper CI workflow.
You'll find more information in the sections below.
- you'll be more autonomous in the development and release of the module, not having to wait for the maintainers to review and merge your PRs.
As a direct consequence, you'll be responsible for the maintenance, documentation and support of the module,
including updating the module to the latest version of _Testcontainers for Go_ if/when needed.
- once created, you'll need to add the module to the [Modules Catalog](https://testcontainers.com/modules/?language=go) in order to be listed there.
You can do this by submitting a PR to the [community repository](https://github.com/testcontainers/community-module-registry).
An example can be found [here](https://github.com/testcontainers/community-module-registry/pull/21/files).
- you'll need to add the module to the [Go documentation](https://pkg.go.dev) in order to be listed there. Please check our [Releasing docs](https://github.com/testcontainers/testcontainers-go/blob/main/RELEASING.md) for more information about triggering the Golang Proxy.

### Modules hosted under the `testcontainers-go` repository

If you still want to host the module under the `testcontainers-go` repository, please consider the following:

- we are not experts in all the technologies out there, so we are open to contributions from the community.
We are happy to review and merge your PRs, and we are also happy to help you with the development of the module.
But this is a shared responsibility, so we expect you to be involved in the maintenance, documentation and support of the module.
- the module will be part of the CI/CD pipeline of the `testcontainers-go` repository, so it will be tested and released with the rest of the modules.
Think of GitHub workflows, release notes, etc. Although it sounds great, which it is, it also means that it will increase the build time in our CI/CD pipeline on GitHub, including flaky tests, number of dependency updates, etc. So in the end it's more work for us.
- once created, we'll add the module to the [Modules Catalog](https://testcontainers.com/modules/?language=go) and to the [Go documentation](https://pkg.go.dev/github.com/testcontainers/testcontainers-go).

## Creating a new module

We have provided a command line tool to generate the scaffolding for the code of the example you are interested in. This tool will generate:

- a Go module for the example, including:
    - go.mod and go.sum files, including the current version of _Testcontainer for Go_.
    - a Go package named after the module, in lowercase
    - a Go file for the creation of the container.
    - a Go test file for running a simple test for your container, consuming the above struct and using the image flag as Docker image for the container.
    - a Go examples file for running the example in the docs site, also adding them to [https://pkg.go.dev](https://pkg.go.dev).
    - a Makefile to run the tests in a consistent manner
- a markdown file in the docs/modules directory including the snippets for both the creation of the container and a simple test. By default, this generated file will contain all the documentation for the module, including:
    - the version of _Testcontainers for Go_ in which the module was added.
    - a short introduction to the module.
    - a section for adding the module to the project dependencies.
    - a section for a usage example, including:
        - a snippet for creating the container, from the `examples_test.go` file in the Go module.
    - a section for the module reference, including:
        - the entrypoint function for creating the container.
        - the options for creating the container.
    - a section for the container methods.
- a new Nav entry for the module in the docs site, adding it to the `mkdocs.yml` file located at the root directory of the project.
- a GitHub workflow file in the .github/workflows directory to run the tests for the example.
- an entry in the VSCode workspace file, in order to include the new module in the project's workspace.

!!!info
    If you are hosting the module under your own GitHub account, please move the generated files to the new repository. Discard the following files and directories: `mkdocs.yml`, VSCode workspace, Sonarqube properties, and the `.github/workflows` directory, as they are specific to the `testcontainers-go` repository. You can use them as reference to create your own CI/CD pipeline.

### Command line flags

| Flag    | Short | Type   | Required | Description                                                                                                                                      |
|---------|-------|--------|----------|--------------------------------------------------------------------------------------------------------------------------------------------------|
| --name  | -n    | string | Yes      | Name of the module, use camel-case when needed. Only alphanumerical characters are allowed (leading character must be a letter).                 |
| --image | -i    | string | Yes      | Fully-qualified name of the Docker image to be used in the examples and tests (i.e. 'org/project:tag')                                             |
| --title | -t    | string | No       | A variant of the name supporting mixed casing (i.e. 'MongoDB'). Only alphanumerical characters are allowed (leading character must be a letter). |


### What is this tool not doing?

- If the module name or title does not contain alphanumerical characters, it will exit the generation.
- If the module already exists, it will exit without updating the existing files.

### How to run the tool

From the [`modulegen` directory]({{repo_url}}/tree/main/modulegen), please run:

```shell
go run . new module --name ${NAME_OF_YOUR_MODULE} --image "${REGISTRY}/${MODULE}:${TAG}" --title ${TITLE_OF_YOUR_MODULE}
```

!!!info
    In the case you just want to create [an example module](../examples/index.md), with no public API, please run:

    ```shell
    go run . new example --name ${NAME_OF_YOUR_MODULE} --image "${REGISTRY}/${MODULE}:${TAG}" --title ${TITLE_OF_YOUR_MODULE}
    ```

### Adding types and methods to the module

We are going to propose a set of steps to follow when adding types and methods to the module:

!!!warning
    The `StartContainer` function will be eventually deprecated and replaced with `Run`. We are keeping it in certain modules for backwards compatibility, but they will be removed in the future.

- Make sure a public `Container` type exists for the module. This type has to use composition to embed the `testcontainers.Container` type, promoting all the methods from it.
- Make sure a `Run` function exists and is public. This function is the entrypoint to the module and will define the initial values for a `testcontainers.GenericContainerRequest` struct, including the image in the function signature, the default exposed ports, wait strategies, etc. Therefore, the function must initialise the container request with the default values.
- Define container options for the module leveraging the `testcontainers.ContainerCustomizer` interface, that has one single method: `Customize(req *GenericContainerRequest) error`.

!!!warning
    The interface definition for `ContainerCustomizer` was changed to allow errors to be correctly processed.
    More specifically, the `Customize` method was changed from:

    ```go
    Customize(req *GenericContainerRequest)
    ```

    To:

    ```go
    Customize(req *GenericContainerRequest) error
    ```

- We consider that a best practice for the options is to define a function using the `With` prefix, that returns a function returning a modified `testcontainers.GenericContainerRequest` type. For that, the library already provides a `testcontainers.CustomizeRequestOption` type implementing the `ContainerCustomizer` interface, and we encourage you to use this type for creating your own customizer functions.
- At the same time, you could need to create your own container customizers for your module. Make sure they implement the `testcontainers.ContainerCustomizer` interface. Defining your own customizer functions is useful when you need to transfer a certain state that is not present at the `ContainerRequest` for the container, possibly using an intermediate Config struct.
- The options will be passed to the `Run` function as variadic arguments after the Go context, and they will be processed right after defining the initial `testcontainers.GenericContainerRequest` struct using a for loop.

```golang
// Config type represents an intermediate struct for transferring state from the options to the container
type Config struct {
    data string
}

// Run function is the entrypoint to the module
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
    cfg := Config{}

    req := testcontainers.ContainerRequest{
        Image: img,
        ...
    }
    genericContainerReq := testcontainers.GenericContainerRequest{
        ContainerRequest: req,
        Started:          true,
    }
    ...
    for _, opt := range opts {
        if err := opt.Customize(&genericContainerReq); err != nil {
            return nil, fmt.Errorf("customise: %w", err)
        }

        // If you need to transfer some state from the options to the container, you can do it here
        if myCustomizer, ok := opt.(MyCustomizer); ok {
            config.data = customizer.data
        }
    }
    ...
    container, err := testcontainers.GenericContainer(ctx, genericContainerReq)
    ...
    moduleContainer := &Container{Container: container}
    moduleContainer.initializeState(ctx, cfg)
    ...
    return moduleContainer, nil
}

// MyCustomizer type represents a container customizer for transferring state from the options to the container
type MyCustomizer struct {
    data string
}
// Customize method implementation
func (c MyCustomizer) Customize(req *testcontainers.GenericContainerRequest) testcontainers.ContainerRequest {
    req.ExposedPorts = append(req.ExposedPorts, "1234/tcp")
    return req.ContainerRequest
}
// WithMy function option to use the customizer
func WithMy(data string) testcontainers.ContainerCustomizer {
    return MyCustomizer{data: data}
}
```

- If needed, define public methods to extract information from the running container, using the `Container` type as receiver. E.g. a connection string to access a database:

```golang
func (c *Container) ConnectionString(ctx context.Context) (string, error) {...}
```

- Document the public API with Go comments.
- Extend the docs to describe the new API of the module. We usually define a parent `Module reference` section, including a `Container options` and a `Container methods` subsections; within each subsection, we define a nested subsection for each option and method, respectively.

### ContainerRequest options

In order to simplify the creation of the container for a given module, `Testcontainers for Go` provides a set of `testcontainers.CustomizeRequestOption` functions to customize the container request for the module. These options are:

- `testcontainers.WithImageSubstitutors`: a function that sets your own substitutions to the container images.
- `testcontainers.WithEnv`: a function that sets the environment variables for the container request.
- `testcontainers.WithExposedPorts`: a function that exposes additional ports from the container.
- `testcontainers.WithEntrypoint`: a function that completely replaces the container's entrypoint.
- `testcontainers.WithEntrypointArgs`: a function that appends commands to the container's entrypoint.
- `testcontainers.WithCmd`: a function that completely replaces the container's command.
- `testcontainers.WithCmdArgs`: a function that appends commands to the container's command.
- `testcontainers.WithLabels`: a function that adds Docker labels to the container.
- `testcontainers.WithFiles`: a function that copies files from the host into the container at creation time.
- `testcontainers.WithMounts`: a function that adds volume mounts to the container.
- `testcontainers.WithTmpfs`: a function that adds tmpfs mounts to the container.
- `testcontainers.WithHostPortAccess`: a function that enables the container to access a port that is already running in the host.
- `testcontainers.WithLogConsumers`: a function that sets the log consumers for the container request.
- `testcontainers.WithLogger`: a function that sets the logger for the container request.
- `testcontainers.WithWaitStrategy`: a function that sets the wait strategy for the container request.
- `testcontainers.WithWaitStrategyAndDeadline`: a function that sets the wait strategy for the container request with a deadline.
- `testcontainers.WithStartupCommand`: a function that sets the execution of a command when the container starts.
- `testcontainers.WithAfterReadyCommand`: a function that sets the execution of a command right after the container is ready (its wait strategy is satisfied).
- `testcontainers.WithDockerfile`: a function that sets the build from a Dockerfile for the container request.
- `testcontainers.WithNetwork`: a function that sets the network and the network aliases for the container request.
- `testcontainers.WithNewNetwork`: a function that sets the network aliases for a throw-away network for the container request.
- `testcontainers.WithConfigModifier`: a function that sets the config Docker type for the container request. Please see [Advanced Settings](../features/creating_container.md#advanced-settings) for more information.
- `testcontainers.WithHostConfigModifier`: a function that sets the host config Docker type for the container request. Please see [Advanced Settings](../features/creating_container.md#advanced-settings) for more information.
- `testcontainers.WithEndpointSettingsModifier`: a function that sets the endpoint settings Docker type for the container request. Please see [Advanced Settings](../features/creating_container.md#advanced-settings) for more information.
- `testcontainers.CustomizeRequest`: a function that merges the default options with the ones provided by the user. Recommended for completely customizing the container request.

### Update Go dependencies in the modules

To update the Go dependencies in the modules, please run:

```shell
$ cd modules
$ make tidy-examples
```

## Refreshing the modules

To refresh the modules, please run:

```shell
$ cd modulegen
$ go run . refresh
```

This command recreates all the project files for the modules and examples, including:

- the mkdocs.yml file, including all the modules and examples, excluding the `compose` module, as it has its own docs page.
- the dependabot config file, including all the modules, the examples and the modulegen module.
- the VSCode project file, including all the modules, the examples and the modulegen module.
- the Sonar properties file, including all the modules, the examples and the modulegen module.

Executing this command in a well-known state of the project, must not produce any changes in the project files.

## Interested in converting an example into a module?

The steps to convert an existing example, aka `${THE_EXAMPLE}`, into a module are the following:

1. Rename the module path at the `go.mod` file for your example.
1. Move the `examples/${THE_EXAMPLE}` directory to `modules/${THE_EXAMPLE}`.
1. In the `mkdocs.yml` file, move the entry for `${THE_EXAMPLE}` from examples to modules.
1. Move `docs/examples${THE_EXAMPLE}.md` file to `docs/modules/${THE_EXAMPLE}`, updating the references to the source code paths.
1. Update the GitHub workflow for `${THE_EXAMPLE}`, modifying names and paths.
