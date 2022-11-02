# Testcontainers for Go code examples

Here you'll find code examples for Testcontainers for Go.

## Interested in adding a new example?

We have provided a command line tool to generate the scaffolding for the code of the example you are interested in. This tool will generate:

- a Go module for the example, including:
    - go.mod and go.sum files
    - a Go package named after the example, in lowercase
    - a Go file for the creation of the container, using a dedicated struct.
    - a Go test file for running a simple test for your container, consuming the above struct.
    - a Makefile to run the tests in a consistent manner
    - a tools.go file including the build tools (i.e. `gotestsum`) used to build/run the example.
- a markdown file in the docs/examples directory including the snippets for both the creation of the container and a simple test.

### What is this tool not doing?

- If the example already exists, it will exit without updating the existing files.
- You have to manually add the markdown entry in the docs to the [`mkdocs.yml`](../mkdocs.yml) file in the root directory of the project. It will generate the navigation menu for the docs website.

### How to run the tool

From the `examples` directory, please run:

```shell
go run main.go --name ${NAME_OF_YOUR_EXAMPLE}
```
