# Using Docker Sandboxes

_Testcontainers for Go_ can use Docker Sandboxes through the optional `sbx` runtime provider. The first operation that needs a Docker client starts one provider-owned sandbox per test process, running detached. Containers, including Ryuk, and native Compose stacks use that sandbox.

## Host prerequisites

Install the Docker `sbx` CLI following the [Docker Sandboxes documentation](https://docs.docker.com/ai/sandboxes/). Run your tests on a host able to start Docker Sandboxes, with local IPv4 TCP and UDP port publishing available. On Linux, the repository test harness requires readable and writable `/dev/kvm`, EROFS filesystem support, and `mkfs.erofs`.

Authenticate the CLI before running tests, on both Linux and macOS. The caller or CI runner must configure login; the provider uses existing authentication and does not log in for you. For the default CLI namespace:

```shell
sbx login
sbx version
```

`TESTCONTAINERS_SBX_APP_NAME` selects the provider's CLI namespace by adding `--app-name` to its CLI calls. If you set it, authenticate that same namespace first:

```shell
export TESTCONTAINERS_SBX_APP_NAME=tc-tests
sbx --app-name "$TESTCONTAINERS_SBX_APP_NAME" login
```

You can configure the namespace in Go with `sbx.Configure(sbx.WithAppName("tc-tests"))` before the first Testcontainers operation. The provider also accepts `WithBinary`, `WithWorkspace`, `WithBindRoots`, and `WithStartupTimeout`. Explicit options override the corresponding environment variables. See [configuration and behavior](../modules/sbx.md#configuration-and-behavior) for examples and defaults.

The provider automatically applies the published `sbx/docker-tcp:latest` kit when it creates the sandbox.

To install additional mixins at creation, call `sbx.Configure(sbx.WithKits(...))` before the first Testcontainers operation. `sbx.WithKitArgs(...)` supplies kit parameters. Configuration stays lazy and applies to the whole test process; it cannot change after provider initialization begins. See [Install additional kits](../modules/sbx.md#install-additional-kits) for a `TestMain` example and composition rules.

The kit exposes the sandbox's Docker API without authentication on a host loopback port. Other local users and processes can access it directly. Run only where all local users and processes are trusted.

## Register and select the provider

The sbx module is on `main` and cannot be fetched as a released module until it is published. See the [module reference](../modules/sbx.md#prerequisites-and-selection) for dependency instructions. Import the module in your test package to register it:

```go
package mypackage_test

import _ "github.com/testcontainers/testcontainers-go/modules/sbx"
```

Importing registers the provider without starting a sandbox or selecting it; adding the dependency alone does not register it.

`go test ./...` runs a separate test binary for each package. Every binary that uses Testcontainers with sbx must import the module, directly or through a shared test helper, or use a harness that registers it in that binary.

Select sbx before any Testcontainers operation:

```shell
TESTCONTAINERS_PROVIDER=sbx go test ./...
```

Alternatively, set `provider=sbx` in `~/.testcontainers.properties`. The environment variable takes precedence over the properties file. Runtime selection applies to the test process and cannot change after initialization.

## Host bind mounts

The workspace is allowed automatically. External bind sources, including temporary directories and Compose service bind mounts, require an allowed root configured with `sbx.WithBindRoots(...)` or `TESTCONTAINERS_SBX_BIND_ROOTS`. Existing tests that bind paths outside the workspace must configure those roots before the provider initializes. For example, to allow fixtures in two existing directories on macOS or Linux:

```shell
export TESTCONTAINERS_SBX_BIND_ROOTS="/opt/test-fixtures:/srv/shared-fixtures"
TESTCONTAINERS_PROVIDER=sbx go test ./...
```

Each root must be an existing absolute directory. See the [configuration reference](../modules/sbx.md#configuration-and-behavior) for validation rules. Only the requested bind source is mounted into the sandbox. A Compose file outside the workspace or an uploaded build context does not itself require a bind root; service bind sources do.

Bind source paths must not contain `..` components. Relative symlinks below a root are allowed only when resolution stays within one containing allowed root. Absolute symlinks below a root are rejected, as are relative links that escape to a separate allowed root. With nested roots, a broader containing root can authorize a relative link that stays within it. Keep roots and source paths stable and writable only by trusted users and processes: a host path can still change between provider validation and the external `sbx` mount.

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

## CI authentication

Authenticate sbx on the CI runner before running tests. With `DOCKERHUB_USERNAME` and `DOCKERHUB_TOKEN` supplied as CI secrets, use:

```shell
printf '%s\n' "$DOCKERHUB_TOKEN" | sbx login --username "$DOCKERHUB_USERNAME" --password-stdin
```

If `TESTCONTAINERS_SBX_APP_NAME` is configured, use the same namespace for login:

```shell
printf '%s\n' "$DOCKERHUB_TOKEN" | sbx --app-name "$TESTCONTAINERS_SBX_APP_NAME" login --username "$DOCKERHUB_USERNAME" --password-stdin
```

Repository CI authenticates each sbx job with these secrets before running tests and logs out afterward.

For relevant changes, live sbx runs only on pushes to `main` and same-repository pull requests. Fork pull requests retain normal unit and lint checks without live sbx or its secrets.
