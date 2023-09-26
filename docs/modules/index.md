# Testcontainers for Go modules

In this section you'll discover how to create Go modules for _Testcontainers for Go_.

## Interested in adding a new module?

We have provided a command line tool to generate the scaffolding for the code of the example you are interested in. This tool will generate:

- a Go module for the example, including:
    - go.mod and go.sum files, including the current version of _Testcontainer for Go_.
    - a Go package named after the module, in lowercase
    - a Go file for the creation of the container, using a dedicated struct in which the image flag is set as Docker image.
    - a Go test file for running a simple test for your container, consuming the above struct.
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
- an entry in Dependabot's configuration file, in order to receive dependency updates.
- an entry in the VSCode workspace file, in order to include the new module in the project's workspace.

### Command line flags

| Flag    | Short | Type   | Required | Description                                                                                                                                      |
|---------|-------|--------|----------|--------------------------------------------------------------------------------------------------------------------------------------------------|
| --name  | -n    | string | Yes      | Name of the module, use camel-case when needed. Only alphanumerical characters are allowed (leading character must be a letter).                 |
| --image | -i    | string | Yes      | Fully-qualified name of the Docker image to be used by the module (i.e. 'docker.io/org/project:tag')                                             |
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
    The `StartContainer` function will be eventually deprecated and replaced with `RunContainer`. We are keeping it in certain modules for backwards compatibility, but they will be removed in the future.

- Make sure a public `Container` type exists for the module. This type has to use composition to embed the `testcontainers.Container` type, promoting all the methods from it.
- Make sure a `RunContainer` function exists and is public. This function is the entrypoint to the module and will define the initial values for a `testcontainers.GenericContainerRequest` struct, including the image, the default exposed ports, wait strategies, etc. Therefore, the function must initialise the container request with the default values.
- Define container options for the module leveraging the `testcontainers.ContainerCustomizer` interface, that has one single method: `Customize(req *GenericContainerRequest)`.
- We consider that a best practice for the options is define a function using the `With` prefix, that returns a function returning a modified `testcontainers.GenericContainerRequest` type. For that, the library already provides a `testcontainers.CustomizeRequestOption` type implementing the `ContainerCustomizer` interface, and we encourage you to use this type for creating your own customizer functions.
- At the same time, you could need to create your own container customizers for your module. Make sure they implement the `testcontainers.ContainerCustomizer` interface. Defining your own customizer functions is useful when you need to transfer a certain state that is not present at the `ContainerRequest` for the container, possibly using an intermediate Config struct.
- The options will be passed to the `RunContainer` function as variadic arguments after the Go context, and they will be processed right after defining the initial `testcontainers.GenericContainerRequest` struct using a for loop.

```golang
// Config type represents an intermediate struct for transferring state from the options to the container
type Config struct {
    data string
}

// RunContainer function is the entrypoint to the module
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
    cfg := Config{}

    req := testcontainers.ContainerRequest{
        Image: "my-image",
        ...
    }
    genericContainerReq := testcontainers.GenericContainerRequest{
        ContainerRequest: req,
        Started:          true,
    }
    ...
    for _, opt := range opts {
        req = opt.Customize(&genericContainerReq)

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

- `testcontainers.CustomizeRequest`: a function that merges the default options with the ones provided by the user. Recommended for completely customizing the container request.
- `testcontainers.WithImage`: a function that sets the image for the container request.
- `testcontainers.WithConfigModifier`: a function that sets the config Docker type for the container request. Please see [Advanced Settings](../features/creating_container.md#advanced-settings) for more information.
- `testcontainers.WithEndpointSettingsModifier`: a function that sets the endpoint settings Docker type for the container request. Please see [Advanced Settings](../features/creating_container.md#advanced-settings) for more information.
- `testcontainers.WithHostConfigModifier`: a function that sets the host config Docker type for the container request. Please see [Advanced Settings](../features/creating_container.md#advanced-settings) for more information.
- `testcontainers.WithWaitStrategy`: a function that sets the wait strategy for the container request, adding all the passed wait strategies to the container request, using a `testcontainers.MultiStrategy` with 60 seconds of deadline. Please see [Wait strategies](../features/wait/multi.md) for more information.
- `testcontainers.WithWaitStrategyAndDeadline`: a function that sets the wait strategy for the container request, adding all the passed wait strategies to the container request, using a `testcontainers.MultiStrategy` with the passed deadline. Please see [Wait strategies](../features/wait/multi.md) for more information.

### Update Go dependencies in the modules

To update the Go dependencies in the modules, please run:

```shell
$ cd modules
$ make tidy-examples
```

## Interested in converting an example into a module?

The steps to convert an existing example, aka `${THE_EXAMPLE}`, into a module are the following:

1. Rename the module path at the `go.mod` file for your example.
1. Move the `examples/${THE_EXAMPLE}` directory to `modules/${THE_EXAMPLE}`.
1. Move the `${THE_EXAMPLE}` dependabot config from the examples section to the modules one, which is located at the bottom.
1. In the `mkdocs.yml` file, move the entry for `${THE_EXAMPLE}` from examples to modules.
1. Move `docs/examples${THE_EXAMPLE}.md` file to `docs/modules/${THE_EXAMPLE}`, updating the references to the source code paths.
1. Update the Github workflow for `${THE_EXAMPLE}`, modifying names and paths.
