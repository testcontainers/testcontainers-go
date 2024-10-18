# General Docker requirements

Testcontainers requires a Docker-API compatible container runtime. 
During development, Testcontainers is actively tested against recent versions of Docker on Linux, as well as against Docker Desktop on Mac and Windows. 
These Docker environments are automatically detected and used by Testcontainers without any additional configuration being necessary.

It is possible to configure Testcontainers to work for other Docker setups, such as a remote Docker host or Docker alternatives. 
However, these are not actively tested in the main development workflow, so not all Testcontainers features might be available and additional manual configuration might be necessary. Please see the [Docker host detection](../features/configuration.md#docker-host-detection) section for more information.

If you have further questions about configuration details for your setup or whether it supports running Testcontainers-based tests, 
please contact the Testcontainers team and other users from the Testcontainers community on [Slack](https://slack.testcontainers.org/).

## Using different container runtimes

_Testcontainers for Go_ automatically detects the selected Docker context and use it to run the tests on that container runtime. You can check the selected context by running:

```sh
docker context ls
NAME              DESCRIPTION                               DOCKER ENDPOINT                                                                                  ERROR
colima            colima                                    unix:///Users/mdelapenya/.colima/default/docker.sock                                             
default           Current DOCKER_HOST based configuration   unix:///var/run/docker.sock                                                                      
desktop-linux *   Docker Desktop                            unix:///Users/mdelapenya/.docker/run/docker.sock                                                 
orbstack          OrbStack                                  unix:///Users/mdelapenya/.orbstack/run/docker.sock                                               
podman            podman context                            unix:///var/folders/_j/nhbgdck523n3008dd3zlsm5m0000gn/T/podman/podman-machine-default-api.sock   
tcd               Testcontainers Desktop                    tcp://127.0.0.1:59908
```

It is possible to use any container runtime to satisfy the system requirements instead of Docker, as long as it is 100% Docker-API compatible, and a Docker context is created for it.

### Colima

Colima creates its own Docker context when it is installed. This context is called `colima`. You can set this as the active context by running:

```sh
docker context use colima
```

### Orbstack

Orbstack creates its own Docker context when it is installed. This context is called `orbstack`. You can set this as the active context by running:

```sh
docker context use orbstack
```

### Podman

Podman does not create its own Docker context when it is installed so, after starting a `podman-machine`, please create it with the following command:

```sh
docker context create podman --description "podman context" --docker "host=unix:///var/folders/_j/nhbgdck523n3008dd3zlsm5m0000gn/T/podman/podman-machine-default-api.sock"
```

!!! note
    The UNIX socket path could be different in your machine. You can find it by running `podman machine inspect`.

Then you can set this as the active context by running:

```sh
docker context use podman
```

### Rancher Desktop

Rancher Desktop creates its own Docker context when it is installed. This context is called `rancher-desktop`. You can set this as the active context by running:

```sh
docker context use rancher-desktop
```

### Testcontainers Desktop

Testcontainers Desktop creates its own Docker context when it is installed. This context is called `tcd`. You can set this as the active context by running:

```sh
docker context use tcd
```

Testcontainers Desktop allows you to switch between different container runtimes, such as Docker, Podman, and Colima, by just using its simple GUI. You can also run the containers in the cloud, using Docker's Testcontainers Cloud.
