# Walk

Walk walks the strategies tree and calls the visit function for each node.

If visit function returns `wait.VisitStop`, the walk stops.
If visit function returns `wait.VisitRemove`, the current node is removed.

## Walk removing entries

<!--codeinclude-->
[Walk a strategy and remove the all FileStrategy entries found](../../../wait/walk_test.go) inside_block:walkRemoveFileStrategy
<!--/codeinclude-->
