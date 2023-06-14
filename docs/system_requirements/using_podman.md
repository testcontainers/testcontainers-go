# Using Podman instead of Docker

_Testcontainers for Go_ supports the use of Podman (rootless or rootful) instead of Docker, but it requires some extra configuration.

1. We should start Podman's socket so it is available to Testcontainers:

```shell
systemctl enable --now --user podman podman.socket
```

2. Confirm socket exists and Podman is working:

```shell
ls -l $XDG_RUNTIME_DIR/podman/podman.sock
podman info
```

3. Set `DOCKER_HOST` environment variable for Testcontainers Configuration strategy:

```bash
export DOCKER_HOST=unix://$XDG_RUNTIME_DIR/podman/podman.sock
```

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
            Image: "docker.io/nginx:alpine"
        },
    }

    // ...
}
```

The `ProviderPodman` configures the `DockerProvider` with the correct default network for Podman to ensure complex network scenarios are working as with Docker.

## Fedora

The `DOCKER_HOST` environment variable must be set, as mentioned above. Additionally, SELinux may require a custom policy be applied to allow the reaper container to connect to and write to a socket. Once you experience the se-linux error, you can run the following commands to create and install a custom policy.

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

