package testcontainersdocker

// DockerSocketPath is the path to the docker socket under unix systems.
const DockerSocketPath = "/var/run/docker.sock"

// DockerSocketPathWithSchema is the path to the docker socket under unix systems with the unix schema.
const DockerSocketPathWithSchema = "unix:///var/run/docker.sock"
