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

## Docker Auth examples

### Creating a custom Docker config file

You can create a Docker config file directly when the environment running your tests
does not have the Docker CLI or a credential helper installed. Testcontainers reads
the credentials from this file without running `docker login`. The tests still need
access to a Docker-compatible container runtime.

For example, add the following permissions, environment variables and steps to a
GitHub Actions test job to pull private images from GitHub Container Registry (GHCR).
Keep your existing checkout and Go setup steps before these steps. The example uses
`sh`, `mktemp`, `printf`, `base64`, `tr`, `rm` and `rmdir`, and does not require the Docker CLI.

{% raw %}
```yaml
permissions:
  contents: read
  packages: read

env:
  # A non-empty value would take precedence over the config file.
  DOCKER_AUTH_CONFIG: ""
  # Set by the setup step; keep cleanup away from any inherited config.
  DOCKER_CONFIG: ""

steps:
  # Place your checkout and Go setup steps here.
  - name: Create Docker auth config
    shell: sh
    env:
      REGISTRY_USERNAME: ${{ github.actor }}
      REGISTRY_PASSWORD: ${{ secrets.GITHUB_TOKEN }}
    run: |
      umask 077
      config_dir=$(mktemp -d "$RUNNER_TEMP/testcontainers-docker-config.XXXXXX")
      printf 'DOCKER_CONFIG=%s\n' "$config_dir" >> "$GITHUB_ENV"
      auth=$(printf '%s:%s' "$REGISTRY_USERNAME" "$REGISTRY_PASSWORD" | base64 | tr -d '\r\n')
      printf '{"auths":{"ghcr.io":{"auth":"%s"}}}\n' "$auth" > "$config_dir/config.json"

  - name: Run tests
    run: go test ./...

  - name: Remove Docker auth config
    if: always()
    shell: sh
    run: |
      if [ -n "$DOCKER_CONFIG" ]; then
        rm -f "$DOCKER_CONFIG/config.json"
        rmdir "$DOCKER_CONFIG"
      fi
```
{% endraw %}

`DOCKER_CONFIG` points to the directory containing `config.json`. Writing it to
`GITHUB_ENV` makes it available to subsequent steps in the same job. The temporary
directory keeps this example separate from any existing Docker configuration.
The `auth` value is the base64 encoding of `username:password`, without a trailing
newline. Base64 is not encryption; do not print the generated config or upload it
as an artifact.

For GHCR, the workflow repository must have read access to the package in addition
to the `packages: read` permission. See GitHub's documentation on
[configuring a package's access control](https://docs.github.com/en/packages/learn-github-packages/configuring-a-packages-access-control-and-visibility).
Use an image such as `ghcr.io/your-organization/your-image:tag` in your tests.
For another registry, replace `ghcr.io` with its registry hostname and supply the
appropriate username and password through GitHub Actions secrets.
