# Garbage Collector

Typically, an integration test creates one or more containers. This can mean a
lot of containers running by the time everything is done. We need to have a way
to clean up after ourselves to keep our machines running smoothly.

Containers can be unused because:

1. Test is over and the container is not needed anymore.
2. Test failed, we do not need that container anymore because next build will
   create new containers.

## Terminate function

As we saw previously there are at least two ways to remove unused containers.
The primary method is to use the `Terminate(context.Conext)` function that is
available when a container is created. Use `defer` to ensure that it is called
on test completion.

!!!tip

    Remember to `defer` as soon as possible so you won't forget. The best time
    is as soon as you call `testcontainers.GenericContainer` but remember to
    check for the `err` first.

## Ryuk

[https://github.com/testcontainers/moby-ryuk](Ryuk) (also referred to as
`Reaper` in this package) removes containers/networks/volumes created by
Testcontainers-Go after a specified delay. It is a project developed by the
TestContainers organization and is used across the board for many of the
different language implementations.

When you run one test, you will see an additional container called `ryuk`
alongside all of the containers that were specified in your test. It relies on
container labels to determine which resources were created by the package
to determine the entities that are safe to remove. If a container is running
for more than 10 seconds, it will be killed.

!!!tip

    This feature can be disabled when creating a container

Even if you do not call Terminate, Ryuk ensures that the environment will be
kept clean and even cleans itself when there is nothing left to do.
