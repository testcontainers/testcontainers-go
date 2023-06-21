# Files and volumes

## File mapping

It is possible to map a file or directory from your FileSystem into the container as a volume using the `Mounts` attribute at the container request struct:

<!--codeinclude-->
[Bind mounts](../../docker_auth_test.go) inside_block:bindMounts
<!--/codeinclude-->

## Volume mapping

It is also possible to map a volume from your Docker host into the container using the `Mounts` attribute at the container request struct:

<!--codeinclude-->
[Volume mounts](../../mounts_test.go) inside_block:volumeMounts
<!--/codeinclude-->

!!!tip
        This ability of creating volumes is also available for remote Docker hosts.
