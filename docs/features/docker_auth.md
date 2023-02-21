# Authentication with Docker

Sometimes the Docker images you use live in a private Docker registry. For that reason, _Testcontainers for Go_ gives you the ability to read the Docker config file
in order to retrieve the authentication for a given registry.

!!!info
	If the `DOCKER_CONFIG` environment variable is set, _Testcontainers for Go_ will use the value of that variable as the path to the Docker config file. Otherwise, it will use the default path, which is `~/.docker/config.json`.

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
