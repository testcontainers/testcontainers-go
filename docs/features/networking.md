# Networking and communicating with containers

## Exposing container ports to the host

It is common to want to connect to a container from your test process, running on the test 'host' machine.
For example, you may be testing some code that needs to connect to a backend or data store container.

Generally, each required port needs to be explicitly exposed. For example, we can specify one or more ports as follows:

<!--codeinclude-->
[Exposing ports](../../docker_test.go) inside_block:exposePorts
<!--/codeinclude-->

Note that this exposed port number is from the *perspective of the container*. 

*From the host's perspective* Testcontainers actually exposes this on a random free port.
This is by design, to avoid port collisions that may arise with locally running software or in between parallel test runs.

Because there is this layer of indirection, it is necessary to ask Testcontainers for the actual mapped port at runtime.
This can be done using the `MappedPort` function, which takes the original (container) port as an argument:

<!--codeinclude-->
[Retrieving actual ports at runtime](../../container_test.go) inside_block:mappedPort
<!--/codeinclude-->

!!! warning
    Because the randomised port mapping happens during container startup, the container must be running at the time `MappedPort` is called. 
    You may need to ensure that the startup order of components in your tests caters for this.

## Getting the container host

When running with a local Docker daemon, exposed ports will usually be reachable on `localhost`.
However, in some CI environments they may instead be reachable on a different host.

As such, Testcontainers provides a convenience function to obtain an address on which the container should be reachable from the host machine.

<!--codeinclude-->
[Getting the container host](../../docker_test.go) inside_block:containerHost
<!--/codeinclude-->

It is normally advisable to use `Host` and `MappedPort` together when constructing addresses - for example:

<!--codeinclude-->
[Getting the container host and mapped port](../../docker_test.go) inside_block:buildingAddresses
<!--/codeinclude-->

!!! info
    Setting the `TC_HOST` environment variable overrides the host of the docker daemon where the container port is exposed. For example, `TC_HOST=172.17.0.1`.

## Docker's host networking mode

From [Docker documentation](https://docs.docker.com/network/drivers/host/):

> If you use the host network mode for a container, that container’s network stack is not isolated from the Docker host (the container shares the host’s networking namespace), and the container does not get its own IP-address allocated. For instance, if you run a container which binds to port 80 and you use host networking, the container’s application is available on port 80 on the host’s IP address.

But according to those docs, it's supported only for Linux hosts:

> The host networking driver only works on Linux hosts, and is not supported on Docker Desktop for Mac, Docker Desktop for Windows, or Docker EE for Windows Server.

In the case you need to skip a test on non-Linux hosts, you can use the `SkipIfDockerDesktop` function:

<!--codeinclude-->
[Skipping tests on non-Linux hosts](../../docker_test.go) inside_block:skipIfDockerDesktop
<!--/codeinclude-->

It will try to get a Docker client and obtain its Info. In the case the Operation System is "Docker Desktop", it will skip the test.

## Advanced networking

Docker provides the ability for you to create custom networks and place containers on one or more networks. Then, communication can occur between networked containers without the need of exposing ports through the host. With Testcontainers, you can do this as well. 

!!! tip
    Note that _Testcontainers for Go_ allows a container to be on multiple networks including network aliases.

For more information about how to create networks using _Testcontainers for Go_, please refer to the [How to create a network](./creating_networks.md) section.

<!--codeinclude-->
[Creating custom networks](../../network/network_test.go) inside_block:testNetworkAliases
<!--/codeinclude-->
