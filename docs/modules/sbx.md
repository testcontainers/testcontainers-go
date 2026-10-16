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

The import registers `sbx`; adding the dependency alone does not. Registration does not start a sandbox. Select the provider before any Testcontainers operation:

```shell
TESTCONTAINERS_PROVIDER=sbx go test ./...
```

You can instead set `provider=sbx` in `~/.testcontainers.properties`. The environment variable takes precedence over the properties file. Select sbx for the test process before any Testcontainers operation.

The module creates the sandbox with `--kit sbx/docker-tcp:latest` and base image `docker.io/docker/sbx-kit-shell:1.0.0`. The `sbx` CLI pulls and loads the published Docker TCP kit. After creation, the module runs `sbx run --detach --name <name>` to keep its owned sandbox running until explicit cleanup, then discovers the Docker endpoint.

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

| Variable | Purpose | Default |
| --- | --- | --- |
| `TESTCONTAINERS_SBX_BINARY` | Path or name of the `sbx` executable | `sbx` on `PATH` |
| `TESTCONTAINERS_SBX_WORKSPACE` | Workspace passed to the sandbox; paths inside it are visible to the sandbox | Current working directory |
| `TESTCONTAINERS_SBX_TIMEOUT` | Sandbox initialization limit, as a positive Go duration such as `5m` | `5m` |
| `TESTCONTAINERS_SBX_APP_NAME` | Optional `sbx --app-name` value for every CLI call | CLI default |

The provider exposes container TCP and UDP ports on `127.0.0.1`. `Host()` and `MappedPort()` return the local endpoint. A requested host port remains the same while its container runs. Stopping a container releases its forwarding; restarting tries to reclaim that same port and returns an error if another process took it. A fixed port already in use fails when the container is created.

Host bind sources must be existing absolute files or directories. The workspace is available through the sandbox; for a source outside it, the provider asks `sbx` to mount that path into the sandbox before container creation. A bind of `/var/run/docker.sock` uses the sandbox's Docker socket. Other host sockets are unsupported. Named volumes live in the sandbox.

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

`module` runs the sbx module suite, including a real nginx lifecycle, mapped port, exec, copy, logs, restart, occupied port test, and an idle-lifetime regression that verifies the sandbox stays running for more than 30 seconds.

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
