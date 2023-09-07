# How to create a network

Apart from creating containers, `Testcontainers for Go` also allows you to create networks. This is useful when you need to connect multiple containers to the same network.

## Usage example

<!--codeinclude-->
[Creating a network](../../network_test.go) inside_block:createNetwork
<!--/codeinclude-->

<!--codeinclude-->
[Creating a network with IPAM](../../network_test.go) inside_block:withIPAM
<!--/codeinclude-->