package testcontainersdocker

// DockerSocketSchema is the unix schema.
var DockerSocketSchema = "unix://"

// DockerSocketPath is the path to the docker socket under unix systems.
var DockerSocketPath = "/var/run/docker.sock"

// DockerSocketPathWithSchema is the path to the docker socket under unix systems with the unix schema.
var DockerSocketPathWithSchema = DockerSocketSchema + DockerSocketPath

// TCPSchema is the tcp schema.
var TCPSchema = "tcp://"
