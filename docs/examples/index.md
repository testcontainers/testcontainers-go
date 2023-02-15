# Code examples

In this section you'll discover how to create code examples for _Testcontainers for Go_, which are almost the same as Go modules, but without exporting any public API.

Their main goal is to create shareable code snippets on how to use certain technology (e.g. a database, a web server), in order to explore its usage before converting the example module into a real Go module exposing a public API.

## Interested in adding a new example?

Please refer to the documentation of the code generation tool [here](../modules/index.md).

## Update Go dependencies in the examples

To update the Go dependencies in the examples, please run:

```shell
$ cd examples
$ make tidy-examples
```
