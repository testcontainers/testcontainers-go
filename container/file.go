package container

import (
	"archive/tar"
	"errors"
	"io"
)

type ContainerFile struct {
	HostFilePath      string    // If Reader is present, HostFilePath is ignored
	Reader            io.Reader // If Reader is present, HostFilePath is ignored
	ContainerFilePath string
	FileMode          int64
}

// validate validates the ContainerFile
func (c *ContainerFile) Validate() error {
	if c.HostFilePath == "" && c.Reader == nil {
		return errors.New("either HostFilePath or Reader must be specified")
	}

	if c.ContainerFilePath == "" {
		return errors.New("ContainerFilePath must be specified")
	}

	return nil
}

type FileFromContainer struct {
	Underlying *io.ReadCloser
	Tarreader  *tar.Reader
}

func (fc *FileFromContainer) Read(b []byte) (int, error) {
	return (*fc.Tarreader).Read(b)
}

func (fc *FileFromContainer) Close() error {
	return (*fc.Underlying).Close()
}
