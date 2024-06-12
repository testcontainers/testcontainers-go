# Using Colima with Docker

[Colima](https://github.com/abiosoft/colima) is a container runtime which
integrates with Docker's tooling and can be configured in various ways.

As of Colima v0.4.0 it's recommended to set the active Docker context to use
Colima. After the context is set _Testcontainers for Go_ will automatically be
configured to use Colima.

```bash
$ docker context ls
NAME        DESCRIPTION                               DOCKER ENDPOINT                                      KUBERNETES ENDPOINT      ORCHESTRATOR
colima      colima                                    unix:///Users/foobar/.colima/default/docker.sock
default *   Current DOCKER_HOST based configuration   unix:///Users/foobar/.colima/docker.sock

$ docker context use colima
colima
Current context is now "colima"

$ docker context ls
NAME       DESCRIPTION                               DOCKER ENDPOINT                                      KUBERNETES ENDPOINT       ORCHESTRATOR
colima *   colima                                    unix:///Users/foobar/.colima/default/docker.sock
default    Current DOCKER_HOST based configuration   unix:///var/run/docker.sock
```

If you're using an older version of Colima or have other applications that are
unaware of Docker context the following workaround is available:

- Locate your Docker Socket, see: [Colima's FAQ - Docker Socket Location](https://github.com/abiosoft/colima/blob/main/docs/FAQ.md#docker-socket-location)

- Create a symbolic link from the default Docker Socket to the expected location, and restart Colima with the `--network-address` flag.

```
    sudo ln -sf $HOME/.colima/default/docker.sock /var/run/docker.sock
    colima stop
    colima start --network-address
```

- Set the `DOCKER_HOST` environment variable to match the located Docker Socket

    * Example: `export DOCKER_HOST="unix://${HOME}/.colima/default/docker.sock"`

- As of testcontainers-go v0.14.0 set `TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE`
   to `/var/run/docker.sock` as the default value refers to your `DOCKER_HOST`
   environment variable.

    * Example: `export TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE="/var/run/docker.sock"`
