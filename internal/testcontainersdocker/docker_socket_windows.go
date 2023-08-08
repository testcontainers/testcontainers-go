package testcontainersdocker

// DockerSocketMountPath is the Docker socket mount path on Windows
var DockerSocketMountPath = "//var/run/docker.sock"

// DockerSocketSchema is the Docker socket schema on Windows
var DockerSocketSchema = "npipe://"

// DockerSocketPath The socket path for Windows contains two slashes, exactly the same as the Docker Desktop socket path.
var DockerSocketPath = "//./pipe/docker_engine"

// DockerSocketPathWithSchema concatenates the Docker socket schema with the Docker socket path, removing the first slash.
var DockerSocketPathWithSchema = DockerSocketSchema + DockerSocketPath[1:]
