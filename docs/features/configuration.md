# Custom configuration

You can override some default properties if your environment requires that.

## Configuration locations
The configuration will be loaded from multiple locations. Properties are considered in the following order:

1. Environment variables
2. `.testcontainers.properties` in user's home folder. Example locations:  
**Linux:** `/home/myuser/.testcontainers.properties`  
**Windows:** `C:/Users/myuser/.testcontainers.properties`  
**macOS:** `/Users/myuser/.testcontainers.properties`

Note that when using environment variables, configuration property names should be set in upper 
case with underscore separators, preceded by `TESTCONTAINERS_` - e.g. `ryuk.disabled` becomes 
`TESTCONTAINERS_RYUK_DISABLED`.

### Supported properties

_Testcontainers for Go_ provides a struct type to represent the configuration:

<!--codeinclude-->
[Supported properties](../../config.go) inside_block:testcontainersConfig
<!--/codeinclude-->

You can read it with the `ReadConfig()` function:

```go
cfg := testcontainers.ReadConfig()
```

### Disabling Ryuk
Ryuk must be started as a privileged container.  
If your environment already implements automatic cleanup of containers after the execution,
but does not allow starting privileged containers, you can turn off the Ryuk container by setting
`TESTCONTAINERS_RYUK_DISABLED` **environment variable** to `true`.

!!!info
    For more information about Ryuk, see [Garbage Collector](garbage_collector.md).

## Customizing Docker host detection

Testcontainers will attempt to detect the Docker environment and configure everything to work automatically.

However, sometimes customization is required. Testcontainers will respect the following **environment variables**:

> **DOCKER_HOST** = unix:///var/run/docker.sock  
> See [Docker environment variables](https://docs.docker.com/engine/reference/commandline/cli/#environment-variables)
>
> **TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE**  
> Path to Docker's socket. Used by Ryuk, Docker Compose, and a few other containers that need to perform Docker actions.  
> Example: `/var/run/docker-alt.sock`

For advanced users, the Docker host connection can be configured **via configuration** in `~/.testcontainers.properties`.
The example below illustrates usage:

```properties
docker.host=tcp://my.docker.host:1234       # Equivalent to the DOCKER_HOST environment variable.
docker.tls.verify=1                         # Equivalent to the DOCKER_TLS_VERIFY environment variable
docker.cert.path=/some/path                 # Equivalent to the DOCKER_CERT_PATH environment variable
```
