# Copying data into a container

Copying data of any type into a container is a very common practice when working with containers. This section will show you how to do it using _Testcontainers for Go_.

## Volume mapping

It is possible to map a Docker volume into the container using the `Mounts` attribute at the `ContainerRequest` struct. For that, please pass an instance of the `GenericVolumeMountSource` type, which allows you to specify the name of the volume to be mapped, and the path inside the container where it should be mounted:

<!--codeinclude-->
[Volume mounts](../../mounts_test.go) inside_block:volumeMounts
<!--/codeinclude-->

!!!tip
    This ability of creating volumes is also available for remote Docker hosts.

!!!warning
    Bind mounts are not supported, as it could not work with remote Docker hosts.

!!!tip
    It is recommended to copy data from your local host machine to a test container using the file copy API 
    described below, as it is much more portable.

## Mounting images

Since Docker v28, it is possible to mount the file system of an image into a container using the `Mounts` attribute at the `ContainerRequest` struct. For that, use the `DockerImageMountSource` type, which allows you to specify the name of the image to be mounted, and the subpath inside the container where it should be mounted, or simply call the `ImageMount` function, which does exactly that:

<!--codeinclude-->
[Image mounts](../../lifecycle_test.go) inside_block:imageMounts
<!--/codeinclude-->

!!!warning
    If the subpath is not a relative path, the creation of the container will fail.

!!!info
    Mounting images fails the creation of the container if the underlying container runtime does not support the `image mount` feature, which is available since Docker v28.

## Copying files to a container

If you would like to copy a file to a container, you can do it in two different manners:

1. Adding a list of files in the `ContainerRequest`, which will be copied before the container starts:

<!--codeinclude-->
[Copying a list of files](../../docker_files_test.go) inside_block:copyFileOnCreate
<!--/codeinclude-->

The `ContainerFile` struct will accept the following fields:

- `HostFilePath`: the path to the file in the host machine. Optional (see below).
- `Reader`: a `io.Reader` that will be used to copy the file to the container. Optional.
- `ContainerFilePath`: the path to the file in the container. Mandatory.
- `Mode`: the file mode, which is optional.

!!!info
    If the `Reader` field is set, the `HostFilePath` field will be ignored.

2. Using the `CopyFileToContainer` method on a `running` container:

<!--codeinclude-->
[Copying files to a running container](../../docker_files_test.go) inside_block:copyFileAfterCreate
[Wait for hello](../../testdata/waitForHello.sh)
<!--/codeinclude-->

## Copying directories to a container

It's also possible to copy an entire directory to a container, and that can happen before and/or after the container gets into the `Running` state. As an example, you could need to bulk-copy a set of files, such as a configuration directory that does not exist in the underlying Docker image.

It's important to notice that, when copying the directory to the container, the container path must exist in the Docker image. And this is a strong requirement for files to be copied _before_ the container is started, as we cannot create the full path at that time.

You can leverage the very same mechanism used for copying files to a container, but for directories.:

1. The first way is using the `Files` field in the `ContainerRequest` struct, as shown in the previous section, but using the path of a directory as `HostFilePath`. Like so:

<!--codeinclude-->
[Copying a directory using files](../../docker_files_test.go) inside_block:copyDirectoryToContainer
<!--/codeinclude-->

2. The second way uses the existing `CopyFileToContainer` method, which will internally check if the host path is a directory, calling the `CopyDirToContainer` method if needed:

<!--codeinclude-->
[Copying a directory to a running container](../../docker_files_test.go) inside_block:copyDirectoryToRunningContainerAsFile
<!--/codeinclude-->

3. The last third way uses the `CopyDirToContainer` method, directly, which, as you probably know, needs the existence of the parent directory in order to copy the directory:

<!--codeinclude-->
[Copying a directory to a running container](../../docker_files_test.go) inside_block:copyDirectoryToRunningContainerAsDir
<!--/codeinclude-->
