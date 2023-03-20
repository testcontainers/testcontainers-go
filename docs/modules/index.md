# Testcontainers for Go modules

In this section you'll discover how to create Go modules for _Testcontainers for Go_.

## Interested in adding a new module?

We have provided a command line tool to generate the scaffolding for the code of the example you are interested in. This tool will generate:

- a Go module for the example, including:
    - go.mod and go.sum files, including the current version of _Testcontainer for Go_.
    - a Go package named after the module, in lowercase
    - a Go file for the creation of the container, using a dedicated struct in which the image flag is set as Docker image.
    - a Go test file for running a simple test for your container, consuming the above struct.
    - a Makefile to run the tests in a consistent manner
    - a tools.go file including the build tools (i.e. `gotestsum`) used to build/run the example.
- a markdown file in the docs/modules directory including the snippets for both the creation of the container and a simple test.
- a new Nav entry for the module in the docs site, adding it to the `mkdocs.yml` file located at the root directory of the project.
- a GitHub workflow file in the .github/workflows directory to run the tests for the example.
- an entry in Dependabot's configuration file, in order to receive dependency updates.

### Command line flags

| Flag       | Type   | Required | Description                                                                                                                                                    |
|------------|--------|----------|----------------------------------------------------------------------------------------------------------------------------------------------------------------|
| -name      | string | Yes      | Name of the module, use camel-case when needed. Only alphanumerical characters are allowed (leading character must be a letter).                               |
| -image     | string | Yes      | Fully-qualified name of the Docker image to be used by the module (i.e. 'docker.io/org/project:tag')                                                           |
| -title     | string | No       | A variant of the name supporting mixed casing (i.e. 'MongoDB'). Only alphanumerical characters are allowed (leading character must be a letter).               |
| -as-module | bool   | No       | If set, the module will be generated as a Go module, under the modules directory. Otherwise, it will be generated as a subdirectory of the examples directory. |

### What is this tool not doing?

- If the module name or title does not contain alphanumerical characters, it will exit the generation.
- If the module already exists, it will exit without updating the existing files.

### How to run the tool

From the [`modulegen` directory]({{repo_url}}/tree/main/modulegen), please run:

```shell
go run . --name ${NAME_OF_YOUR_MODULE} --image "${REGISTRY}/${MODULE}:${TAG}" --title ${TITLE_OF_YOUR_MODULE}
```

or for creating a Go module:

```shell
go run . --name ${NAME_OF_YOUR_MODULE} --image "${REGISTRY}/${MODULE}:${TAG}" --title ${TITLE_OF_YOUR_MODULE} --as-module
```

## Update Go dependencies in the modules

To update the Go dependencies in the modules, please run:

```shell
$ cd modules
$ make tidy-examples
```
