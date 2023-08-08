package testcontainersdocker

// DockerSocketPathWithSchema concatenates the Docker socket schema with the Docker socket path, removing the first slash.
var DockerSocketPathWithSchema = DockerSocketSchema + DockerSocketPath

// TCPSchema is the tcp schema.
var TCPSchema = "tcp://"
