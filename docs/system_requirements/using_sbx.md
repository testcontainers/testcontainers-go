# Using Docker Sandboxes

_Testcontainers for Go_ can use Docker Sandboxes through the optional `sbx` runtime provider. The first operation that needs a Docker client starts one provider-owned sandbox per test process, running detached. Containers, including Ryuk, and native Compose stacks use that sandbox.

## Host prerequisites

Install the Docker `sbx` CLI following the [Docker Sandboxes documentation](https://docs.docker.com/ai/sandboxes/). Run your tests on a host able to start Docker Sandboxes, with local IPv4 TCP and UDP port publishing available. On Linux, the repository test harness requires readable and writable `/dev/kvm`, EROFS filesystem support, and `mkfs.erofs`.

Sign in using the ordinary CLI namespace and check the installation:

```shell
sbx login
sbx version
```

Local sandbox creation requires sign-in on Linux and macOS. The provider automatically applies the published `sbx/docker-tcp:latest` kit when it creates the sandbox.

## Register and select the provider

The sbx module is on `main` and cannot be fetched as a released module until it is published. See the [module reference](../modules/sbx.md#prerequisites-and-selection) for dependency instructions. Import the module in your test package to register it:

```go
package mypackage_test

import _ "github.com/testcontainers/testcontainers-go/modules/sbx"
```

Importing registers the provider without starting a sandbox or selecting it. Select sbx before any Testcontainers operation:

```shell
TESTCONTAINERS_PROVIDER=sbx go test ./...
```

Alternatively, set `provider=sbx` in `~/.testcontainers.properties`. The environment variable takes precedence over the properties file. Runtime selection applies to the test process and cannot change after initialization.

## Clean up the sandbox

Call `sbx.Cleanup` after `m.Run()` with a fresh, bounded context to remove the owned sandbox. Follow the full [cleanup in TestMain example](../modules/sbx.md#cleanup-in-testmain), including reporting cleanup errors before exiting. A named `sbx` import used for `Cleanup` also registers the provider, so it replaces the blank import above.

Ryuk cleans up container resources inside the sandbox; removing the sandbox requires the explicit cleanup call. Process termination before cleanup can leave a sandbox behind.

See [configuration and behavior](../modules/sbx.md#configuration-and-behavior) for runtime options and limitations.

## Run the repository tests

From the repository root on a host configured as above, run:

```shell
GOFLAGS=-v ./scripts/test-sbx.sh module
GOFLAGS=-v ./scripts/test-sbx.sh core
GOFLAGS=-v ./scripts/test-sbx.sh compose
```

The `core` mode runs the complete root Go-module `./...` suite. Optional nested Go modules are separate: `module` runs the supplemental sbx module suite, and `compose` runs runtime routing checks and a native Compose smoke test. Run `GOFLAGS=-v ./scripts/test-sbx.sh all` to run all three modes in sequence. The only sbx capability skips in core are host-network options, host networking, and `container:<id>` networking; sbx requires published ports for its managed forwarding and does not support those network modes.

The harness selects sbx, registers it in the test binaries, and arranges explicit sandbox cleanup; cleanup errors fail the test binary. The core Ryuk orphan test separately checks resource removal on the shared inner daemon after a child process exits. The harness uses ordinary `sbx login` credentials by default. See [the host test reference](../modules/sbx.md#run-the-provider-tests-on-a-host) for timeout settings and further details.

CI runs these tests on Ubuntu 24.04 with Go `1.26.x` and `1.27.x`, using the shared Go test workflow and its per-module dependency installation hook for `modules/sbx`. For relevant changes, live sbx runs only on pushes to `main` and same-repository pull requests, using `DOCKERHUB_USERNAME` and `DOCKERHUB_TOKEN`. Each run, attempt, and Go matrix leg gets a unique CLI namespace. Fork pull requests retain normal unit and lint checks without live sbx or its secrets.
