In more locked down / secured environments, it can be problematic to pull images from Docker Hub and run them without additional precautions.

An image name substitutor converts a Docker image name, as may be specified in code, to an alternative name. This is intended to provide a way to override image names, for example to enforce pulling of images from a private registry.

_Testcontainers for Go_ exposes an interface to perform this operations: `ImageSubstitutor`, and a No-operation implementation to be used as reference for custom implementations:

<!--codeinclude-->
[Image Substitutor Interface](../../generic.go) inside_block:imageSubstitutor
[Noop Image Substitutor](../../container_test.go) inside_block:noopImageSubstitutor
<!--/codeinclude-->
