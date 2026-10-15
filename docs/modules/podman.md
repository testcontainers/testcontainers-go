# Podman

Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

The Podman module registers Podman as a runtime provider for Testcontainers for Go.
It uses Podman's Docker-compatible API and its `podman` default bridge network.
Podman must already be running; the module does not start or install it.

## Adding this module to your project dependencies

```shell
go get github.com/testcontainers/testcontainers-go/modules/podman
```

Import the module in your test package to register the provider:

```go
import _ "github.com/testcontainers/testcontainers-go/modules/podman"
```

Adding the dependency to `go.mod` alone does not register it. Importing the package
registers the provider without connecting to Podman.

## Selecting Podman

Set the provider before running your tests:

```shell
TESTCONTAINERS_PROVIDER=podman go test ./...
```

Alternatively, set `provider=podman` in `~/.testcontainers.properties`.
An explicit provider in code takes precedence over the environment variable,
which takes precedence over the properties file.

The module retains standard Testcontainers endpoint and socket discovery.
See [Using Podman instead of Docker](../system_requirements/using_podman.md)
for socket activation, `DOCKER_HOST`, and platform-specific Ryuk requirements.

## Migrating existing tests

Existing `ProviderPodman` calls and automatic Podman detection continue to work
through the built-in compatibility path. You do not have to import the module
immediately. `ProviderPodman` is deprecated, with removal deferred to a future
breaking release.

To migrate, import this module and select Podman through configuration before any
Testcontainers operation. Per-container provider options are then unnecessary
for tests using that configured runtime.

## Testing the provider

The main CI pipeline runs the core Go module against native rootful Podman on
Ubuntu 24.04, with Ryuk enabled, for changes to core or `modules/podman`.
This job does not run the database modules or examples. It runs the core suite
with the Podman module imported through a temporary Go workspace, and separate
smoke tests without that import for automatic detection and `ProviderPodman`.
An orphan-cleanup test verifies that Ryuk deletes stopped containers, networks
and explicitly labelled volumes after the test process exits.

To run the same tests locally, configure a rootful Podman socket and Ryuk as
described in the system requirements, then run
`./.github/scripts/test-core-podman.sh` from the repository root.
The script temporarily adds provider imports to the core test binaries and
removes them when it exits. It does not modify core's module dependencies.
