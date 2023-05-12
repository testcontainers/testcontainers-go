package testcontainersdocker

// DockerSocketSchema is the npipe schema.
const DockerSocketSchema = "npipe://"

// DockerSocketPath is the path to the docker socket under windows systems.
var DockerSocketPath = "//./pipe/docker_engine"

// DockerSocketPathWithSchema is the path to the docker socket under windows systems with the npipe schema.
var DockerSocketPathWithSchema = DockerSocketSchema + DockerSocketPath
