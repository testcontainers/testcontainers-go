# Using Podman instead of Docker

_Testcontainers for Go_ supports the use of Podman (rootless or rootful) instead of Docker.
In most scenarios no special setup is required.
_Testcontainers for Go_ will automatically discover the socket based on the `DOCKER_HOST` or the `TC_HOST` environment variables.
Alternatively you can configure the host with a `.testcontainers.properties` file.
The discovered Docker host is also taken into account when starting a reaper container.

There's currently only one special case where additional configuration is necessary: complex container network scenarios.

By default _Testcontainers for Go_ takes advantage of the default network settings both Docker and Podman are applying to newly created containers.
It only intervenes in scenarios where a `ContainerRequest` specifies networks and does not include the default network of the current container provider.
Unfortunately the default network for Docker is called _bridge_ where the default network in Podman is called _podman_.
It is not even possible to create a network called _bridge_ with Podman as Podman does not allow creating a network with the same name as an already existing network mode.

In such scenarios it is possible to explicitly make use of the `ProviderPodman` like so:

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

The `ProviderPodman` configures the `DockerProvider` with the correct default network for Podman to ensure also complex network scenarios are working as with Docker.