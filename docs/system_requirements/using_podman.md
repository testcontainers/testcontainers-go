# Using Podman instead of Docker

_Testcontainers for Go_ supports the use of Podman (rootless or rootful) instead of Docker.

In most scenarios no special setup is required in _Testcontainers for Go_.
_Testcontainers for Go_ will automatically discover the socket based on the `DOCKER_HOST` environment variables.
Alternatively you can configure the host with a `.testcontainers.properties` file.
The discovered Docker host is taken into account when starting a reaper container.
The discovered socket is used to detect the use of Podman.

By default _Testcontainers for Go_ takes advantage of the default network settings both Docker and Podman are applying to newly created containers.
It only intervenes in scenarios where a `ContainerRequest` specifies networks and does not include the default network of the current container provider.
Unfortunately the default network for Docker is called _bridge_ where the default network in Podman is called _podman_.

In complex container network scenarios it may be required to explicitly make use of the `ProviderPodman` like so:

```go

package some_test

import (
    "testing"
    tc "github.com/testcontainers/testcontainers-go"
)

func TestSomething(t *testing.T) {
    req := tc.GenericContainerRequest{
        ProviderType: tc.ProviderPodman,
        ContainerRequest: tc.ContainerRequest{
            Image: "nginx:alpine"
        },
    }

    // ...
}
```

The `ProviderPodman` configures the `DockerProvider` with the correct default network for Podman to ensure complex network scenarios are working as with Docker.

## Podman socket activation

The reaper container needs to connect to the docker daemon to reap containers, so the podman socket service must be started:
```shell
> systemctl --user start podman.socket
```

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

