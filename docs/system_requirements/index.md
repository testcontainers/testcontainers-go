# System Requirements

## Go version

From the [Go Release Policy](https://go.dev/doc/devel/release#policy):

> Each major Go release is supported until there are two newer major releases. For example, Go 1.5 was supported until the Go 1.7 release, and Go 1.6 was supported until the Go 1.8 release. We fix critical problems, including critical security problems, in supported releases as needed by issuing minor revisions (for example, Go 1.6.1, Go 1.6.2, and so on).

_Testcontainers for Go_ is tested against those two latest Go releases, therefore we recommend using any of them. You could check what versions are actually supported by looking at the [GitHub Action](../../.github/workflows/ci.yml) configuration, under the `test.strategy.matrix.go-version` key.

## General Docker requirements

Testcontainers requires a Docker-API compatible container runtime. 
During development, Testcontainers is actively tested against recent versions of Docker on Linux, as well as against Docker Desktop on Mac and Windows. 
These Docker environments are automatically detected and used by Testcontainers without any additional configuration being necessary.

It is possible to configure Testcontainers to work for other Docker setups, such as a remote Docker host or Docker alternatives. 
However, these are not actively tested in the main development workflow, so not all Testcontainers features might be available and additional manual configuration might be necessary. 
If you have further questions about configuration details for your setup or whether it supports running Testcontainers-based tests, 
please contact the Testcontainers team and other users from the Testcontainers community on [Slack](https://slack.testcontainers.org/).
