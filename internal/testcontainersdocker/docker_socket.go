package testcontainersdocker

// DockerSocketPath is the path to the docker socket under unix systems.
var DockerSocketPath = "/var/run/docker.sock"

// DockerSocketPathWithSchema is the path to the docker socket under unix systems with the unix schema.
var DockerSocketPathWithSchema = "unix:///var/run/docker.sock"
