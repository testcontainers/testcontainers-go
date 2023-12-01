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
[Supported properties](../../internal/config/config.go) inside_block:testcontainersConfig
<!--/codeinclude-->

You can read it with the `ReadConfig()` function:

```go
cfg := testcontainers.ReadConfig()
```

For advanced users, the Docker host connection can be configured **via configuration** in `~/.testcontainers.properties`, but environment variables will take precedence.
Please see [Docker host detection](#docker-host-detection) for more information.

The example below illustrates how to configure the Docker host connection via properties file:

```properties
docker.host=tcp://my.docker.host:1234       # Equivalent to the DOCKER_HOST environment variable.
docker.tls.verify=1                         # Equivalent to the DOCKER_TLS_VERIFY environment variable
docker.cert.path=/some/path                 # Equivalent to the DOCKER_CERT_PATH environment variable
```

## Customizing images

Please read more about customizing images in the [Image name substitution](image_name_substitution.md) section.

## Customizing Ryuk, the resource reaper

1. Ryuk must be started as a privileged container. For that, you can set the `TESTCONTAINERS_RYUK_CONTAINER_PRIVILEGED` **environment variable**, or the  `ryuk.container.privileged` **property** to `true`.
1. If your environment already implements automatic cleanup of containers after the execution,
but does not allow starting privileged containers, you can turn off the Ryuk container by setting
`TESTCONTAINERS_RYUK_DISABLED` **environment variable** to `true`.
1. You can specify the connection timeout for Ryuk by setting the `ryuk.connection.timeout` **property**. The default value is 1 minute.
1. You can specify the reconnection timeout for Ryuk by setting the `ryuk.reconnection.timeout` **property**. The default value is 10 seconds.


!!!info
    For more information about Ryuk, see [Garbage Collector](garbage_collector.md).

## Docker host detection

_Testcontainers for Go_ will attempt to detect the Docker environment and configure everything to work automatically.

However, sometimes customization is required. _Testcontainers for Go_ will respect the following order:

1. Read the **tc.host** property in the `~/.testcontainers.properties` file. E.g. `tc.host=tcp://my.docker.host:1234`

2. Read the **DOCKER_HOST** environment variable. E.g. `DOCKER_HOST=unix:///var/run/docker.sock`
See [Docker environment variables](https://docs.docker.com/engine/reference/commandline/cli/#environment-variables) for more information.

3. Read the Go context for the **DOCKER_HOST** key. E.g. `ctx.Value("DOCKER_HOST")`. This is used internally for the library to pass the Docker host to the resource reaper.

4. Read the default Docker socket path, without the unix schema. E.g. `/var/run/docker.sock`

5. Read the **docker.host** property in the `~/.testcontainers.properties` file. E.g. `docker.host=tcp://my.docker.host:1234`

6. Read the rootless Docker socket path, checking in the following alternative locations:
    1. `${XDG_RUNTIME_DIR}/.docker/run/docker.sock`.
    2. `${HOME}/.docker/run/docker.sock`.
    3. `${HOME}/.docker/desktop/docker.sock`.
    4. `/run/user/${UID}/docker.sock`, where `${UID}` is the user ID of the current user.

7. The default Docker socket including schema will be returned if none of the above are set.

## Docker socket path detection

_Testcontainers for Go_ will attempt to detect the Docker socket path and configure everything to work automatically.

However, sometimes customization is required. _Testcontainers for Go_ will respect the following order:

1. Read the **tc.host** property in the `~/.testcontainers.properties` file. E.g. `tc.host=tcp://my.docker.host:1234`. If this property is set, the returned Docker socket path
will be the default Docker socket path: `/var/run/docker.sock`.

2. Read the **TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE** environment variable.
Path to Docker's socket. Used by Ryuk, Docker Compose, and a few other containers that need to perform Docker actions.  

    Example: `/var/run/docker-alt.sock`

3. If the Operative System retrieved by the Docker client is "Docker Desktop", and the host is running on Windows, it will return the `//var/run/docker.sock` UNC Path. Else return the default docker socket path for rootless docker.

4. Get the current Docker Host from the existing strategies: see [Docker host detection](#docker-host-detection).

5. If the socket contains the unix schema, the schema is removed (e.g. `unix:///var/run/docker.sock` -> `/var/run/docker.sock`)

6. Else, the default location of the docker socket is used: `/var/run/docker.sock`

In any case, if the docker socket schema is `tcp://`, the default docker socket path will be returned.
