package testcontainersdocker

// DockerSocketPath is the path to the docker socket under windows systems.
const DockerSocketPath = "//./pipe/docker_engine"

// DockerSocketPathWithSchema is the path to the docker socket under windows systems with the npipe schema.
const DockerSocketPathWithSchema = "npipe:////./pipe/docker_engine"
