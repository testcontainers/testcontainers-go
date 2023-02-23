# Authentication with Docker

Sometimes the Docker images you use live in a private Docker registry. For that reason, _Testcontainers for Go_ gives you the ability to read the Docker configuration
and retrieve the authentication for a given registry. To achieve it, _Testcontainers for Go_ will internally check, in this particular order:
	
1. the `DOCKER_AUTH_CONFIG` environment variable, unmarshalling it from its JSON representation and using it as the Docker config.
2. the `DOCKER_CONFIG` environment variable, as an alternative path to the Docker config file.
3. else it will load the default Docker config file, which lives in the user's home, e.g. `~/.docker/config.json`
4. it will use the right Docker credential helper to retrieve the authentication (user, password and base64 representation) for the given registry.

To understand how the Docker credential helpers work, please refer to the [official documentation](https://docs.docker.com/engine/reference/commandline/login/#credential-helpers).

!!! info
	_Testcontainers for Go_ uses [https://github.com/cpuguy83/dockercfg](https://github.com/cpuguy83/dockercfg) to retrieve the authentication from the credential helpers.

To retrieve the credentials for a given registry, _Testcontainers for Go_ exposes the `AuthFromDockerConfig` function, which takes the registry name as a parameter and returns the authentication for that registry:

```go
auth, err := testcontainers.AuthFromDockerConfig("myregistry.com")
if err != nil {
	// do something with err
}
```

Once you have the authentication, you can use it to build a container request for an image living in a private registry, simply passing the
credentials in the `RegistryCred` field of the container request:

```go
req := ContainerRequest{
	FromDockerfile: testcontainers.FromDockerfile{
		Image: "myregistry.com/myimage:latest",
		RegistryCred: auth.Auth,
	},
}
```

In the case you are building an image from the Dockerfile, you can also pass the authentication in the `AuthConfigs` field of the `FromDockerfile` struct:

```go
req := ContainerRequest{
	FromDockerfile: testcontainers.FromDockerfile{
		Context: "/path/to/build/context",
		Dockerfile: "CustomDockerfile",
		AuthConfigs: map[string]types.AuthConfig{
			"myregistry.com": auth,
		},
		BuildArgs: map[string]*string {
			"FOO": "BAR",
		},
	},
}
```
