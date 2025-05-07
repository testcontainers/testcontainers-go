# Go version

From the [Go Release Policy](https://go.dev/doc/devel/release#policy):

> Each major Go release is supported until there are two newer major releases. For example, Go 1.5 was supported until the Go 1.7 release, and Go 1.6 was supported until the Go 1.8 release. We fix critical problems, including critical security problems, in supported releases as needed by issuing minor revisions (for example, Go 1.6.1, Go 1.6.2, and so on).

_Testcontainers for Go_ is tested against those two latest Go releases, therefore we recommend using any of them. You could check what versions are actually supported by looking at the [GitHub Action](https://github.com/testcontainers/testcontainers-go/blob/main/.github/workflows/ci.yml) configuration, under the `test.strategy.matrix.go-version` key.

## Updating the Go version

The project has a script to update the Go version used in the project. To update the Go version across all files, please run the following command:

```bash
DRY_RUN=false ./scripts/bump-go.sh 1.21
```

It will update:
- all `go.mod` files.
- all GitHub Actions workflows, and their matrices using the `go-version` strategy.
- the templates for the module generator.
- the Devcontainer configuration.
- the tag of the Go Docker image in the markdown documentation.
