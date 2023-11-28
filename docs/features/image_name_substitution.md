# Image name substitution

_Testcontainers for Go_ supports automatic substitution of Docker image names.

This allows the replacement of an image name specified in test code with an alternative name - for example, to replace the 
name of a Docker Hub image dependency with an alternative hosted on a private image registry.

This is advisable to avoid Docker Hub rate limiting, and some companies will prefer this for policy reasons.

!!!info
    As of November 2020 Docker Hub pulls are rate limited. As Testcontainers uses Docker Hub for standard images, some users may hit these rate limits and should mitigate accordingly. Suggested mitigations are noted in [this issue in Testcontainers for Java](https://github.com/testcontainers/testcontainers-java/issues/3099) at present.

This page describes two approaches for image name substitution:

* [Automatically modifying Docker Hub image names](#automatically-modifying-docker-hub-image-names), prefixing them with a private registry URL.
* [Using an Image Name Substitutor](#developing-a-custom-function-for-transforming-image-names-on-the-fly), developing a custom function for transforming image names on the fly.

!!!warning
    It is assumed that you have already set up a private registry hosting [all the Docker images your build requires](../supported_docker_environment/image_registry_rate_limiting.md#which-images-are-used-by-testcontainers).

## Automatically modifying Docker Hub image names

_Testcontainers for Go_ can be configured to modify Docker Hub image names on the fly to apply a prefix string.

Consider this if:

* Developers and CI machines need to use different image names. For example, developers are able to pull images from Docker Hub, but CI machines need to pull from a private registry.
* Your private registry has copies of images from Docker Hub where the names are predictable, and just adding a prefix is enough. 
  For example, `registry.mycompany.com/mirror/mysql:8.0.24` can be derived from the original Docker Hub image name (`mysql:8.0.24`) with a consistent prefix string: `registry.mycompany.com/mirror`

In this case, image name references in code are **unchanged**.
i.e. you would leave as-is:

<!--codeinclude-->
[Unchanged direct Docker Hub image name](../../container_test.go) inside_block:directDockerHubReference
<!--/codeinclude-->

You can then configure _Testcontainers for Go_ to apply a given prefix (e.g. `registry.mycompany.com/mirror`) to every image that it tries to pull from Docker Hub. Important to notice that **the prefix should not include a trailing slash**. This can be done in one of two ways:

* Setting the `TESTCONTAINERS_HUB_IMAGE_NAME_PREFIX=registry.mycompany.com/mirror` environment variable.
* Via config file, setting `hub.image.name.prefix` in the `~/.testcontainers.properties` file in your user home directory.

_Testcontainers for Go_ will automatically apply the prefix to every image that it pulls from Docker Hub - please verify that all [the required images](#images-used-by-testcontainers) exist in your registry.

_Testcontainers for Go_ will not apply the prefix to:

* non-Hub image names (e.g. where another registry is set)
* Docker Hub image names where the hub registry is explicitly part of the name (i.e. anything with a `docker.io` or `registry.hub.docker.com` host part)

## Developing a custom function for transforming image names on the fly

Consider this if:

* You have complex rules about which private registry images should be used as substitutes, e.g.:
    * non-deterministic mapping of names meaning that a [name prefix](#automatically-modifying-docker-hub-image-names) cannot be used, or
    * rules depending upon developer identity or location, or
* you wish to add audit logging of images used in the build, or
* you wish to prevent accidental usage of images that are not on an approved list.

In this case, image name references in code are **unchanged**. i.e. you would leave as-is:

<!--codeinclude-->
[Unchanged direct Docker Hub image name](../../container_test.go) inside_block:directDockerHubReference
<!--/codeinclude-->

You can implement a custom image name substitutor by:

* implementing the `ImageNameSubstitutor` interface, exposed by the `testcontainers` package.
* configuring _Testcontainers for Go_ to use your custom implementation, defined at the `ContainerRequest` level.

The following is an example image substitutor implementation prepending the `docker.io/` prefix, used in the tests:

<!--codeinclude-->
[Image Substitutor Interface](../../options.go) inside_block:imageSubstitutor
[Docker prefix Image Substitutor](../../container_test.go) inside_block:dockerImageSubstitutor
[Applying the substitutor](../../container_test.go) inside_block:applyImageSubstitutors
<!--/codeinclude-->

## Images used by Testcontainers

As of the current version of Testcontainers ({{latest_version}}):

* every image directly used by your tests
* images pulled by Testcontainers itself to support functionality:
    * [`testcontainers/ryuk`](https://hub.docker.com/r/testcontainers/ryuk) - performs fail-safe cleanup of containers, and always required (unless [Ryuk is disabled](./configuration.md#customizing-ryuk-the-resource-reaper)).
    * [`alpine`](https://hub.docker.com/r/_/alpine).
    * [`Docker in Docker`](https://hub.docker.com/_/docker).
    * [`nginx`](https://hub.docker.com/r/_/nginx).
    * [`delayed nginx`](https://hub.docker.com/r/menedev/delayed-nginx).
    * [`localstack`](https://hub.docker.com/r/localstack/localstack).
    * [`mysql`](https://hub.docker.com/r/_/mysql).
    * [`postgres`](https://hub.docker.com/r/_/postgres).
    * [`postgis`](https://hub.docker.com/r/postgis/postgis).
    * [`redis`](https://hub.docker.com/r/_/redis).
    * [`registry`](https://hub.docker.com/r/_/registry).
