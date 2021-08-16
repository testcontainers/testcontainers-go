# Garbage Collector

Usually, one test creates at least one container. At the end it means a lot of
containers running. We need to have a way to keep the CI servers reliable
removing unused containers.

Containers can be unused because:

1. Test is over and the container is not needed anymore.
2. Test failed, we do not need that container anymore because next build will
   create new containers.

## Terminate function

As we saw previously there are at least two ways to remove unused containers.
The first one is to use the `Terminate(context.Conext)` function available when a
container is created. You can call it in your test or you use `defer` .

!!!tip

    Remember to `defer` as soon as possible so you won't forget. The best time
    is as soon as you call `testcontainers.GenericContainer` but remember to
    check for the `err` first.

## Ryuk

[https://github.com/testcontainers/moby-ryuk](ryuk) helps you to remove
containers/networks/volumes by given filter after specified delay.

It is a project developed by TestContainers, and it is used across the board for
Java, Go and any more.

When you run one test, you will see that there is not only the containers your
tests requires running, there is another one called `ryuk`. We refer to it as
`Reaper` as well in this library.

Based on container labels it removes resources created from testcontainers that
are running for more than 10 seconds.

!!!tip

    This feature can be disabled when creating a container

In this way even if you do not call Terminate, something will keep your
environment clean. It will also clean itself when there is nothing left to do.
