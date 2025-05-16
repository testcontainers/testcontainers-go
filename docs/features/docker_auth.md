# Authentication with Docker

Sometimes the Docker images you use live in a private Docker registry. For that reason, _Testcontainers for Go_ gives you the ability to read the Docker configuration
and retrieve the authentication for a given registry. To achieve it, _Testcontainers for Go_ will internally check, in this particular order:
	
1. the `DOCKER_AUTH_CONFIG` environment variable, unmarshalling the string value from its JSON representation and using it as the Docker config.
2. the `DOCKER_CONFIG` environment variable, as an alternative path to the Docker config file.
3. else it will load the default Docker config file, which lives in the user's home, e.g. `~/.docker/config.json`
4. it will use the right Docker credential helper to retrieve the authentication (user, password and base64 representation) for the given registry.

To understand how the Docker credential helpers work, please refer to the [official documentation](https://docs.docker.com/engine/reference/commandline/login/#credential-helpers).

!!! info
	_Testcontainers for Go_ uses [https://github.com/cpuguy83/dockercfg](https://github.com/cpuguy83/dockercfg) to retrieve the authentication from the credential helpers.

_Testcontainers for Go_ will automatically discover the credentials for a given Docker image from the Docker config, as described above. For that, it will extract the Docker registry from the image name, and for that registry will try to locate the authentication in the Docker config, returning an empty string if the registry is not found. As a consequence, all the fields to pass credentials to the container request will be deprecated.

```go
req := ContainerRequest{
	Image: "myregistry.com/myimage:latest",
}
```

In the case you are building an image from the Dockerfile, the authentication will be automatically retrieved from the Docker config, so you don't need to pass it explicitly:

<!--codeinclude-->
[Building From a Dockerfile does not need Auth credentials anymore](../../docker_test.go) inside_block:fromDockerfileWithBuildArgs
<!--/codeinclude-->

