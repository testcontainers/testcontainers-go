# All Wait strategy

The All multi wait strategy holds a list of wait strategies. The execution of each strategy is first added, first executed.

Available Options:

- `WithDeadline` - the deadline for when all strategies must complete by, default is none.
- `WithStartupTimeoutDefault` - the startup timeout default to be used for each Strategy if not defined in seconds, default is 60 seconds.

<!--codeinclude-->
[ForAll Example](../../../wait/wait_examples_test.go) inside_block:ExampleForAll
<!--/codeinclude-->
