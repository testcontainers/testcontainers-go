# Go version

From the [Go Release Policy](https://go.dev/doc/devel/release#policy):

> Each major Go release is supported until there are two newer major releases. For example, Go 1.5 was supported until the Go 1.7 release, and Go 1.6 was supported until the Go 1.8 release. We fix critical problems, including critical security problems, in supported releases as needed by issuing minor revisions (for example, Go 1.6.1, Go 1.6.2, and so on).

_Testcontainers for Go_ is tested against those two latest Go releases, therefore we recommend using any of them. You could check what versions are actually supported by looking at the [GitHub Action](../../.github/workflows/ci.yml) configuration, under the `test.strategy.matrix.go-version` key.
