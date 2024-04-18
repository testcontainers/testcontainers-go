# Registry

Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.30.0"><span class="tc-version">:material-tag: v0.30.0</span></a>

## Introduction

The Testcontainers module for Registry.

## Adding this module to your project dependencies

Please run the following command to add the Registry module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/registry
```

## Usage example

<!--codeinclude-->
[Creating a Registry container](../../modules/registry/examples_test.go) inside_block:runRegistryContainer
<!--/codeinclude-->

## Module reference

The Registry module exposes one entrypoint function to create the Registry container, and this function receives two parameters:

```golang
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*RegistryContainer, error)
```

- `context.Context`, the Go context.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the Registry container, you can pass options in a variadic way to configure it.

#### Image

If you need to set a different Registry Docker image, you can use `testcontainers.WithImage` with a valid Docker image
for Registry. E.g. `testcontainers.WithImage("registry:2.8.3")`.

{% include "../features/common_functional_options.md" %}

#### With Authentication

It's possible to enable authentication for the Registry container. By default, it is disabled, but you can enable it in two ways:

- You can use `WithHtpasswd` to enable authentication with a string representing the contents of a `htpasswd` file.
A temporary file will be created with the contents of the string and copied to the container.
- You can use `WithHtpasswdFile` to copy a `htpasswd` file from your local filesystem to the container.

In both cases, the `htpasswd` file will be copied into the `/auth` directory inside the container.

<!--codeinclude-->
[Htpasswd string](../../modules/registry/registry_test.go) inside_block:htpasswdString
[Htpasswd file](../../modules/registry/examples_test.go) inside_block:htpasswdFile
<!--/codeinclude-->

#### WithData

In the case you want to initialise the Registry with your own images, you can use `WithData` to copy a directory from your local filesystem to the container.
The directory will be copied into the `/data` directory inside the container.
The format of the directory should be the same as the one used by the Registry to store images.
Otherwise, the Registry will start but you won't be able to read any images from it.

<!--codeinclude-->
[Including data](../../modules/registry/examples_test.go) inside_block:htpasswdFile
<!--/codeinclude-->

### Container Methods

The Registry container exposes the following methods:

#### Address

This method returns the HTTP address string to connect to the Distribution Registry, so that you can use to connect to the Registry.
E.g. `http://localhost:32878/v2/_catalog`.

<!--codeinclude-->
[HTTP Address](../../modules/registry/registry_test.go) inside_block:httpAddress
<!--/codeinclude-->

#### ImageExists

The `ImageExists` method allows to check if an image exists in the Registry. It receives the Go context and the image reference as parameters.

!!! info
    The image reference should be in the format `my-registry:port/image:tag` in order to be pushed to the Registry.

#### PushImage

The `PushImage` method allows to push an image to the Registry. It receives the Go context and the image reference as parameters.

!!! info
    The image reference should be in the format `my-registry:port/image:tag` in order to be pushed to the Registry.

<!--codeinclude-->
[Pushing images to the registry](../../modules/registry/examples_test.go) inside_block:pushingImage
<!--/codeinclude-->

If the push operation is successful, the method will internally wait for the image to be available in the Registry, querying the Registry API, returning an error in case of any failure (e.g. pushing or waiting for the image).

#### DeleteImage

The `DeleteImage` method allows to delete an image from the Registry. It receives the Go context and the image reference as parameters.

!!! info
    The image reference should be in the format `image:tag` in order to be deleted from the Registry.

<!--codeinclude-->
[Deleting images from the registry](../../modules/registry/examples_test.go) inside_block:deletingImage
<!--/codeinclude-->
