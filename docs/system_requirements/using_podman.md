# Using Podman instead of Docker

_Testcontainers for Go_ supports the use of Podman (rootless or rootful) instead of Docker.

## Using the Podman module

Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

Use the [Podman module](../modules/podman.md) to register Podman as a runtime provider.
Podman must already be installed and running; the module does not install or start it.

Add the module to your project dependencies:

```shell
go get github.com/testcontainers/testcontainers-go/modules/podman
```

Import the module in each test package that uses Podman:

```go
package some_test

import (
    "testing"

    tc "github.com/testcontainers/testcontainers-go"
    _ "github.com/testcontainers/testcontainers-go/modules/podman"
)

func TestSomething(t *testing.T) {
    ctr, err := tc.Run(t.Context(), "docker.io/nginx:alpine")
    tc.CleanupContainer(t, ctr)
    if err != nil {
        t.Fatal(err)
    }

    // ...
}
```

The import registers the provider without connecting to Podman. Adding the dependency
to `go.mod` without importing it does not register it.

Select Podman before running your tests:

```shell
TESTCONTAINERS_PROVIDER=podman go test ./...
```

Alternatively, set `provider=podman` in `~/.testcontainers.properties`.
An explicit provider in code takes precedence over the environment variable,
which takes precedence over the properties file.

Select the runtime before any Testcontainers operation, including direct client
construction. Provider modules initialize before Docker host or socket discovery.

## Host and network configuration

Selecting the provider does not configure the Podman endpoint. The module retains
standard [Docker host detection](../features/configuration.md#docker-host-detection),
including `DOCKER_HOST` and the `docker.host` property in `~/.testcontainers.properties`.
The discovered host and socket are also used when starting Ryuk, the resource reaper.
See the platform-specific setup below for socket activation and Ryuk requirements.

By default _Testcontainers for Go_ takes advantage of the default network settings both Docker and Podman apply to newly created containers.
It only intervenes when a `ContainerRequest` specifies networks and does not include the default network of the current container provider.
The Podman module configures the default bridge network as `podman`, so per-container `ProviderPodman` options are unnecessary when Podman is selected through configuration.

## Migrating existing tests

Existing `ProviderPodman` calls, including `WithProvider(ProviderPodman)`, and automatic
Podman detection continue to work through the built-in compatibility path without
importing the module. `ProviderPodman` is deprecated, with removal deferred to a future
breaking release.

To migrate, import the Podman module and select it through configuration as shown
above, then remove per-container `ProviderPodman` options. Both paths use the same
Docker-compatible endpoint discovery and `podman` default bridge network.

## Podman socket activation

On Linux, the reaper container needs access to Podman's Docker-compatible API to reap containers, so the Podman socket service must be started:

```shell
systemctl --user start podman.socket
```

## macOS

On macOS, automatic Podman detection may fail and prevent Ryuk from starting.
Import the Podman module and explicitly select Podman as described above.
To use Ryuk:

1. Run Podman in rootful mode:

    ```shell
    podman machine stop
    podman machine set --rootful
    podman machine start
    ```

2. Add the following to `~/.testcontainers.properties`:

    ```properties
    ryuk.container.privileged=true
    ```

If you cannot use Ryuk, you can [disable it](../features/garbage_collector.md#ryuk),
but this is not recommended.

## Fedora

`DOCKER_HOST` environment variable must be set

```
> export DOCKER_HOST=unix://$XDG_RUNTIME_DIR/podman/podman.sock
```

SELinux may require a custom policy be applied to allow the reaper container to connect to and write to a socket. Once you experience the se-linux error, you can run the following commands to create and install a custom policy.

```
> sudo ausearch -c 'app' --raw | audit2allow -M my-podman
> sudo semodule -i my-podman.pp
```

The resulting my-podman.te file should look something like this:
```
module my-podman2 1.0;

require {
        type user_tmp_t;
        type container_runtime_t;
        type container_t;
        class sock_file write;
        class unix_stream_socket connectto;
}

#============= container_t ==============
allow container_t container_runtime_t:unix_stream_socket connectto;
allow container_t user_tmp_t:sock_file write;

```

**NOTE: It will take two rounds of installing a policy, then experiencing the next se-linux issue, install new policy, etc...**

