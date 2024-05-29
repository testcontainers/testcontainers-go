package testcontainers

import (
	"io"

	"github.com/docker/docker/api/types"
)

// FromDockerfile represents the parameters needed to build an image from a Dockerfile
// rather than using a pre-built one
type FromDockerfile struct {
	Context        string             // the path to the context of the docker build
	ContextArchive io.Reader          // the tar archive file to send to docker that contains the build context
	Dockerfile     string             // the path from the context to the Dockerfile for the image, defaults to "Dockerfile"
	Repo           string             // the repo label for image, defaults to UUID
	Tag            string             // the tag label for image, defaults to UUID
	BuildArgs      map[string]*string // enable user to pass build args to docker daemon
	PrintBuildLog  bool               // enable user to print build log

	// KeepImage describes whether DockerContainer.Terminate should not delete the
	// container image. Useful for images that are built from a Dockerfile and take a
	// long time to build. Keeping the image also Docker to reuse it.
	KeepImage bool
	// BuildOptionsModifier Modifier for the build options before image build. Use it for
	// advanced configurations while building the image. Please consider that the modifier
	// is called after the default build options are set.
	BuildOptionsModifier func(*types.ImageBuildOptions)
}
