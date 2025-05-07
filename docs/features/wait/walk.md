# Walk

Walk walks the strategies tree and calls the visit function for each node.

This allows modules to easily amend default wait strategies, updating or
removing specific strategies based on requirements of functional options.

For example removing a TLS strategy if a functional option enabled insecure mode
or changing the location of the certificate based on the configured user.

If visit function returns `wait.ErrVisitStop`, the walk stops.
If visit function returns `wait.ErrVisitRemove`, the current node is removed.

## Walk removing entries

The following example shows how to remove a strategy based on its type.
<!--codeinclude-->
[Remove FileStrategy entries](../../../wait/walk_test.go) inside_block:walkRemoveFileStrategy
<!--/codeinclude-->
