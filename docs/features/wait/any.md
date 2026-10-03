# Any Wait strategy

The Any multi wait strategy holds a list of wait strategies. The execution of
each strategy is asynchronous: all run in their own goroutine. If any one
succeeds, the Wait will finish with success (no error) and the remaining
running wait strategies will be cancelled. If any one fails, the Wait will
finish with an error and the remaining running wait strategies will be
cancelled.

Available Options:

- `WithDeadline` - the deadline for when all strategies must complete by, default is none.
- `WithStartupTimeoutDefault` - the startup timeout default to be used for each Strategy if not defined in seconds, default is 60 seconds.

<!--codeinclude-->
[ForAny Example](../../../wait/wait_examples_test.go) inside_block:ExampleForAny
<!--/codeinclude-->
