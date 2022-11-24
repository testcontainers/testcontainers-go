# Code examples

In this section you'll discover how to create code examples for _Testcontainers for Go_.

## Interested in adding a new example?

We have provided a command line tool to generate the scaffolding for the code of the example you are interested in. This tool will generate:

- a Go module for the example, including:
    - go.mod and go.sum files, including the current version of _Testcontainer for Go_.
    - a Go package named after the example, in lowercase
    - a Go file for the creation of the container, using a dedicated struct in which the image flag is set as Docker image.
    - a Go test file for running a simple test for your container, consuming the above struct.
    - a Makefile to run the tests in a consistent manner
    - a tools.go file including the build tools (i.e. `gotestsum`) used to build/run the example.
- a markdown file in the docs/examples directory including the snippets for both the creation of the container and a simple test.
- a new Nav entry for the example in the docs site, adding it to the `mkdocs.yml` file located at the root directory of the project.
- a GitHub workflow file in the .github/workflows directory to run the tests for the example.

### Command line flags

| Flag | Type | Required | Description |
|------|------|----------|-------------|
| -name | string | Yes | Name of the example, use camel-case when needed. Only alphabetical characters are allowed. |
| -image | string | Yes | Fully-qualified name of the Docker image to be used by the example (i.e. 'docker.io/org/project:tag') |

### What is this tool not doing?

- If the example name does not contain alphabeticall characters, it will exit the generation.
- If the example already exists, it will exit without updating the existing files.

### How to run the tool

From the [`examples` directory]({{repo_url}}/tree/main/examples), please run:

```shell
go run . --name ${NAME_OF_YOUR_EXAMPLE} --image "${REGISTRY}/${EXAMPLE}:${TAG}"
```
