# Custom configuration

You can override some default properties if your environment requires that.

## Configuration locations
The configuration will be loaded from multiple locations. Properties are considered in the following order:

1. Environment variables
2. `.testcontainers.properties` in user's home folder. Example locations:  
**Linux:** `/home/myuser/.testcontainers.properties`  
**Windows:** `C:/Users/myuser/.testcontainers.properties`  
**macOS:** `/Users/myuser/.testcontainers.properties`

## Customizing Docker host detection

Testcontainers will attempt to detect the Docker environment and configure everything to work automatically.

However, sometimes customization is required. Testcontainers will respect the following **environment variables**:

> **DOCKER_HOST** = unix:///var/run/docker.sock  
> See [Docker environment variables](https://docs.docker.com/engine/reference/commandline/cli/#environment-variables)

For advanced users, the Docker host connection can be configured **via configuration** in `~/.testcontainers.properties`.
The example below illustrates usage:

```properties
docker.host=tcp\://my.docker.host\:1234     # Equivalent to the DOCKER_HOST environment variable. Colons should be escaped.
docker.tls.verify=1                         # Equivalent to the DOCKER_TLS_VERIFY environment variable
docker.cert.path=/some/path                 # Equivalent to the DOCKER_CERT_PATH environment variable
```
