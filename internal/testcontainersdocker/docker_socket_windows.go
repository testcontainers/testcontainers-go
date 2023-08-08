package testcontainersdocker

// DockerSocketPath The socket path for Windows contains two slashes, exactly the same as the Docker Desktop socket path.
var DockerSocketPath = "//var/run/docker.sock"

// DockerSocketPathWithSchema concatenates the Docker socket schema with the Docker socket path, removing the first slash.
var DockerSocketPathWithSchema = DockerSocketSchema + DockerSocketPath[1:]
