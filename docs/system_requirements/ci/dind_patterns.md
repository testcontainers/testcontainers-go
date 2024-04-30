# Patterns for running tests inside a Docker container

## 'Docker wormhole' pattern - Sibling docker containers

Testcontainers itself can be used from inside a container.
This is very useful for different CI scenarios like running everything in containers on Jenkins, or Docker-based CI tools such as Drone.

Testcontainers will automatically detect if it's inside a container and instead of "localhost" will use the default gateway's IP.

However, additional configuration is required if you use [volume mapping](../../features/files_and_mounts.md#volume-mapping). The following points need to be considered:

* The docker socket must be available via a volume mount
* The 'local' source code directory must be volume mounted *at the same path* inside the container that Testcontainers runs in, so that Testcontainers is able to set up the correct volume mounts for the containers it spawns.

### Docker-only example
If you run the tests with just `docker run ...` then make sure you add `-v $PWD:$PWD -w $PWD -v /var/run/docker.sock:/var/run/docker.sock` to the command, so it will look like this:
```bash
$ tree .
.
├── go.mod
└── internal
    └── platform
        └── integration_test.go
└── platform
    └── integration_test.go

$ docker run -it --rm -v $PWD:$PWD -w $PWD -v /var/run/docker.sock:/var/run/docker.sock golang:1.21 go test ./... -v
```

Where:

* `-v $PWD:$PWD` will add your current directory as a volume inside the container
* `-w $PWD` will set the current directory to this volume
* `-v /var/run/docker.sock:/var/run/docker.sock` will map the Docker socket


!!! warning
    If you are using Docker Desktop, you need to configure the `TC_HOST` environment variable to use the special DNS name
    `host.docker.internal` for accessing the host from within a container, which is provided by Docker Desktop:
    `-e TC_HOST=host.docker.internal`

### Docker Compose example

The same can be achieved with Docker Compose:

```yaml
tests:
  image: golang:1.21
  stop_signal: SIGKILL
  stdin_open: true
  tty: true
  working_dir: $PWD
  volumes:
    - $PWD:$PWD
    - /var/run/docker.sock:/var/run/docker.sock
  command: go test ./... -v
```

## Docker-in-Docker

While Docker-in-Docker (DinD) is generally considered an instrument of last resort, it is necessary for some CI environments.

[Drone CI](./drone.md) is one such example. Testcontainers has a Docker-in-Docker plugin (build image) for use with Drone,
which could be used as inspiration for setting up other similar testing using DinD.
