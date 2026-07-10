# Nginx

Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

## Introduction

The Testcontainers module for [Nginx](https://nginx.org/), a high-performance web server commonly used as a reverse proxy, load balancer, and HTTP cache.

## Adding this module to your project dependencies

Please run the following command to add the Nginx module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/nginx
```

## Usage example

<!--codeinclude-->
[Creating a Nginx container](../../modules/nginx/examples_test.go) inside_block:runNginxContainer
<!--/codeinclude-->

## Module Reference

### Run function

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

The Nginx module exposes one entrypoint function to create the Nginx container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "nginx:1.25")`.

### Container Options

When starting the Nginx container, you can pass options in a variadic way to configure it.

#### WithConfigFile

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

`WithConfigFile` mounts a custom `nginx.conf` file at `/etc/nginx/nginx.conf` in the container, replacing the default Nginx main configuration.
The `hostPath` argument must be an absolute path to the configuration file on the host.

```golang
nginx.Run(ctx, "nginx:1.25", nginx.WithConfigFile("/path/to/nginx.conf"))
```

#### WithCustomConfig

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

`WithCustomConfig` mounts a configuration snippet at `/etc/nginx/conf.d/default.conf` in the container.
Use this option to customise the default virtual-host behaviour (locations, upstreams, proxy settings, etc.) without replacing the entire main configuration.
The `hostPath` argument must be an absolute path to the configuration file on the host.

```golang
nginx.Run(ctx, "nginx:1.25", nginx.WithCustomConfig("/path/to/default.conf"))
```

{% include "../features/common_functional_options_list.md" %}

### Container Methods

The Nginx container exposes the following methods:

#### HTTPEndpoint

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

`HTTPEndpoint` returns the HTTP endpoint of the running Nginx container in the form `http://host:mappedPort`, using the container's mapped port 80.

<!--codeinclude-->
[Get HTTP endpoint](../../modules/nginx/nginx_test.go) inside_block:httpEndpoint
<!--/codeinclude-->

#### HTTPSEndpoint

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

`HTTPSEndpoint` returns the HTTPS endpoint of the running Nginx container in the form `https://host:mappedPort`, using the container's mapped port 443.

```golang
httpsEndpoint, err := nginxContainer.HTTPSEndpoint(ctx)
```

