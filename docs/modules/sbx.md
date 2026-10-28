# Docker Sandboxes (sbx)

Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

The sbx module runs Testcontainers for Go inside a [Docker Sandbox](https://docs.docker.com/ai/sandboxes/). It registers an optional runtime provider. The first operation that needs a Docker client starts one sandbox for the test process, before client discovery and Ryuk. Containers and native Compose stacks in that process use the same sandbox.

## Prerequisites and selection

Install a supported Docker `sbx` CLI on a host that can run sandboxes with local IPv4 TCP and UDP port publishing. Authenticate the CLI with `sbx login` before running tests; the provider uses existing authentication and does not log in for you. See [Using Docker Sandboxes](../system_requirements/using_sbx.md) for host setup, CI login, and configured CLI namespaces.

The module is on `main` and cannot be fetched as a released module until it is published. After release, add the module dependency and import it in a test package:

```shell
go get github.com/testcontainers/testcontainers-go/modules/sbx
```

```go
import _ "github.com/testcontainers/testcontainers-go/modules/sbx"
```

The import registers `sbx`; adding the dependency alone does not. Registration does not start a sandbox.

`go test ./...` runs a separate test binary for each package. Every binary that uses Testcontainers with sbx must import the module, directly or through a shared test helper, or use a harness that registers it in that binary.

Select the provider before any Testcontainers operation:

```shell
TESTCONTAINERS_PROVIDER=sbx go test ./...
```

You can instead set `provider=sbx` in `~/.testcontainers.properties`. The environment variable takes precedence over the properties file. Select sbx for the test process before any Testcontainers operation.

The module creates the sandbox with `--kit sbx/docker-tcp:latest` and base image `docker.io/docker/sbx-kit-shell:1.0.0`. The `sbx` CLI pulls and loads the published Docker TCP kit. After creation, the module runs `sbx run --detach --name <name>` to keep its owned sandbox running until explicit cleanup, then discovers the Docker endpoint.

## Install additional kits

Use `sbx.Configure` with provider options before the first Testcontainers operation, typically at the beginning of `TestMain`, before `m.Run()` or provider health checks:

```go
if err := sbx.Configure(
    sbx.WithKits("./testdata/test-tools"),
    sbx.WithKitArgs(map[string]string{
        "test-tools.mode": "integration",
    }),
); err != nil {
    fmt.Fprintf(os.Stderr, "sbx configuration: %v\n", err)
    os.Exit(1)
}
```

Replace the example reference and arguments with your own mixin's settings. `WithKits` accepts multiple references, including OCI images, local directories, and Git sources. Use version tags or digests for reproducible published kits. Relative paths are resolved against the process working directory at initialization, independently of `TESTCONTAINERS_SBX_WORKSPACE`.

`Configure` configures the registered provider without selecting it or starting a sandbox. Continue selecting it through `TESTCONTAINERS_PROVIDER=sbx` or the `provider` property. If another provider is selected, the sbx configuration is unused. Containers still use the usual module `Run` functions and container customizers. Kit options apply to the sandbox environment; tools installed by a kit are not automatically installed inside its child containers.

`WithKits` appends references in the supplied order, alongside the required Docker TCP kit. `WithKitArgs` merges arguments, with later values overriding earlier values for the same key. An unqualified argument applies to every selected kit declaring it; `kit-handle.name` targets one kit. Empty values are allowed. Arguments are plain text stored by sbx, so use sbx credentials for secrets.

The provider passes repeated `--kit` and `--kit-arg` flags to `sbx create`. sbx validates sources, duplicate kit names, declared arguments, dependencies, and compatibility. Include any kits required by your mixins; sbx determines dependency order rather than treating flag order as installation order. Source policies and credentials must already permit unattended use. See [Use kits](https://docs.docker.com/ai/sandboxes/customize/use-kits/) for source formats, kit arguments, and composition rules.

Repeated `Configure` calls compose before initialization. Each call either applies all its options or returns an error without changing the previous configuration. Options copy their input slices and maps. Once initialization begins, configuration is frozen and further calls return `sbx.ErrProviderStarted`, including after initialization fails or the sandbox is cleaned up. Callers can check this with `errors.Is`. Kit creation runs within the startup timeout, configured with `WithStartupTimeout` or `TESTCONTAINERS_SBX_TIMEOUT`; creation errors retain the CLI diagnostic and trigger owned-sandbox cleanup.

Each test process shares one kit configuration. Configure every package's test binary, directly or through a shared test helper. Use a separate test process for a different kit combination. V3 sandboxes cannot install additional mixins with `sbx kit add` after creation.

## Cleanup in TestMain

Call `sbx.Cleanup` after `m.Run()` with a fresh, bounded context, then exit with the test result. A named import of `sbx` for `Cleanup` also registers the provider, so a blank import is unnecessary in the same test binary.

```go
package mypackage_test

import (
    "context"
    "fmt"
    "os"
    "testing"
    "time"

    sbx "github.com/testcontainers/testcontainers-go/modules/sbx"
)

func TestMain(m *testing.M) {
    code := m.Run()

    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
    if err := sbx.Cleanup(ctx); err != nil {
        fmt.Fprintf(os.Stderr, "sbx cleanup: %v\n", err)
        if code == 0 {
            code = 1
        }
    }
    cancel()
    os.Exit(code)
}
```

`Cleanup` removes the sandbox owned by this process and is safe to call when it was never started. If removal fails, it returns an error and can be retried; report that error so the run does not appear successful. Process termination before `TestMain` reaches cleanup can leave a sandbox behind.

## Configuration and behavior

Configure provider settings alongside kits, before the first Testcontainers operation:

```go
if err := sbx.Configure(
    sbx.WithBinary("sbx"),
    sbx.WithWorkspace("."),
    sbx.WithBindRoots("/opt/test-fixtures"),
    sbx.WithStartupTimeout(8*time.Minute),
    sbx.WithAppName("tc-tests"),
    sbx.WithKits("./testdata/test-tools"),
); err != nil {
    fmt.Fprintf(os.Stderr, "sbx configuration: %v\n", err)
    os.Exit(1)
}
```

Explicit options take precedence over environment variables; unset options use the environment and then the defaults below. The provider reads the environment during initialization and does not change it. Later settings override earlier ones, including across `Configure` calls. An overridden environment value is not validated or used, so a valid option can replace an invalid environment setting.

`WithBindRoots` replaces the entire additional-root list, including any environment setting. Call `WithBindRoots()` to allow only the workspace. `WithAppName("")` overrides an inherited namespace with the CLI default. Nonempty app names accept up to 20 letters, digits, hyphens, and underscores. Authenticate the chosen namespace before running tests, for example with `sbx --app-name tc-tests login`.

Binary and workspace options must be nonempty, startup timeouts must be positive, and bind roots must be absolute. Workspace and bind directories are checked at startup, so they may be created after `Configure`. Relative workspaces are resolved against the process working directory at startup. Relative kit sources remain relative to the process working directory, independently of the workspace setting.

| Option | Environment variable | Purpose | Default |
| --- | --- | --- | --- |
| `WithBinary(path)` | `TESTCONTAINERS_SBX_BINARY` | Path or name of the `sbx` executable | `sbx` on `PATH` |
| `WithWorkspace(path)` | `TESTCONTAINERS_SBX_WORKSPACE` | Workspace passed to the sandbox; paths inside it are visible to the sandbox | Current working directory |
| `WithBindRoots(paths...)` | `TESTCONTAINERS_SBX_BIND_ROOTS` | Additional allowed host bind directories; environment entries use the OS path-list separator (`:` on macOS/Linux) | No additional roots; workspace allowed automatically |
| `WithStartupTimeout(duration)` | `TESTCONTAINERS_SBX_TIMEOUT` | Sandbox initialization limit, including kit installation; environment values use Go durations such as `5m` | `5m` |
| `WithAppName(name)` | `TESTCONTAINERS_SBX_APP_NAME` | Optional `sbx --app-name` value for every CLI call | CLI default |

The provider exposes container TCP and UDP ports on `127.0.0.1`. `Host()` and `MappedPort()` return the local endpoint. A requested host port remains the same while its container runs. Stopping a container releases its forwarding; restarting tries to reclaim that same port and returns an error if another process took it. A fixed port already in use fails when the container is created.

Host bind sources must be existing absolute files or directories within the workspace or a configured bind root. Additional roots must be existing absolute directories. An empty setting adds no roots; empty list entries, relative paths, missing paths, and nondirectories fail configuration. The provider resolves root symlinks at initialization and retains their configured aliases, including macOS `/var` paths that resolve under `/private/var`.

For an authorized source outside the workspace, the provider asks `sbx` to mount only the requested source, preserving its configured path spelling and read-only setting. See [host bind mounts](../system_requirements/using_sbx.md#host-bind-mounts) for configuration and symlink restrictions. A bind of `/var/run/docker.sock` uses the sandbox's Docker socket, including for Ryuk. Other host sockets are unsupported. Named volumes live in the sandbox's Docker daemon.

For Compose, use the native `NewDockerCompose` or `NewDockerComposeWith` API. Its Docker client uses the selected sbx endpoint. The legacy shell based `NewLocalDockerCompose` API rejects sbx.

Current port support is local IPv4 TCP/UDP. IPv6 and arbitrary remote host bindings are unsupported, as are Docker `host` and `container` network modes. Published ports requested with `0.0.0.0` are exposed on local loopback only. Automatic Docker restart policies are unsupported because daemon-driven restarts bypass host port reservation; use explicit container start/restart instead. Host bind paths use POSIX syntax; Windows drive-letter binds are rejected.

## Run the provider tests on a host

From the repository root, with a supported `sbx` installation and local VM support:

```shell
./scripts/test-sbx.sh module
./scripts/test-sbx.sh core
./scripts/test-sbx.sh compose
./scripts/test-sbx.sh all
```

For verbose test output, run `GOFLAGS=-v ./scripts/test-sbx.sh module`. The harness preserves `GOFLAGS` for its Go commands.

`module` runs the sbx module suite, including a real nginx lifecycle, mapped port, exec, copy, logs, restart, occupied port test, and an idle-lifetime regression that verifies the sandbox stays running for more than 30 seconds. A separate test process creates a sandbox with the local v3 mixin in `modules/sbx/testdata/test-tools`, verifies its argument-generated file, and runs an ordinary nginx container. This check requires local kit builds to be available and allowed by the host's sbx configuration.

`core` registers sbx in every root-module test package and runs the complete root Go-module `./...` suite, with retries disabled. Optional nested Go modules are separate; the sbx module suite and native Compose checks supplement the core run. The only sbx capability skips are host-network options, host networking, and `container:<id>` networking: sbx manages forwarding for published ports and does not support those network modes.

`compose` runs runtime routing checks and a native Compose nginx smoke test. `all` runs those three modes in sequence, continuing with the remaining stages if one fails.

The harness selects sbx and calls `Cleanup` through `TestMain` for the core and Compose binaries; the module suite has its own `TestMain`. Owned-sandbox cleanup errors fail the test binary. Separately, the core Ryuk orphan test checks that Ryuk removes containers, networks, and volumes from the shared inner daemon after a child process exits. The harness uses a temporary Go workspace for registration and removes its temporary files when done.

The script accepts `TESTCONTAINERS_SBX_BINARY` for a custom CLI, `TESTCONTAINERS_SBX_TIMEOUT` for initialization, an optional `TESTCONTAINERS_SBX_APP_NAME` for CLI scoping, and `TESTCONTAINERS_SBX_GO_TEST_TIMEOUT` for the `go test` timeout (default `30m`). It sets `TESTCONTAINERS_SBX_WORKSPACE` to the repository root and enables the real integration tests. By default it uses the credentials from ordinary `sbx login`; unique sandbox names identify resources owned by each test process. If you explicitly set `TESTCONTAINERS_SBX_APP_NAME`, first sign in to that separate CLI namespace with `sbx --app-name <name> login`, then use the same name for the test command. The `--app-name` suffix accepts letters, digits, hyphens, and underscores, with a 20-character limit. For example:

```shell
TESTCONTAINERS_SBX_TIMEOUT=8m TESTCONTAINERS_SBX_GO_TEST_TIMEOUT=45m ./scripts/test-sbx.sh all
```

CI runs the same `all` mode on Ubuntu 24.04 with Go `1.26.x` and `1.27.x` when sbx, core runtime, Compose, or sbx CI files change. The shared Go test workflow installs sbx through the existing per-module dependency hook, selecting `modules/sbx` for this runtime. A machine without working sandbox virtualization cannot run these real tests.

CI uses the organization secrets `DOCKERHUB_USERNAME` and `DOCKERHUB_TOKEN`; the caller passes both explicitly to the shared Go test workflow. The account and token must be accepted by `sbx login --username <username> --password-stdin` and satisfy applicable sbx organization policy. Login, tests, and final logout use a unique CLI namespace derived from the workflow run, attempt, and Go matrix leg. Missing credentials fail trusted runs before installation; whitespace-only credentials fail login validation.

The authenticated integration job runs for relevant pushes to `main` and same-repository pull requests. Fork pull requests run normal unit and lint checks without live sbx or its secrets. The final CI gate uses the same trust condition. Login stores credentials for subsequent sbx operations, so limiting the token's environment to the login step does not make it safe to run untrusted code afterward. Host test commands continue to use ordinary `sbx login` unless an explicit namespace is supplied.
