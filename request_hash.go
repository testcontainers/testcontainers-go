package testcontainers

import (
	"fmt"
	"io"
	"os"

	"github.com/testcontainers/testcontainers-go/internal/core"
)

// containerHash represents the hash of the container configuration
type containerHash struct {
	// Hash of the container configuration
	Hash uint64
	// Hash of the files copied to the container, to verify if they have changed
	FilesHash uint64
}

func (ch containerHash) String() string {
	return fmt.Sprintf("{Hash: %d, FilesHash: %d}", ch.Hash, ch.FilesHash)
}

func (c ContainerRequest) hash() containerHash {
	var ch containerHash
	hash, err := core.Hash(c)
	if err != nil {
		return ch
	}

	ch = containerHash{
		Hash: hash,
	}

	// The initial hash of the files copied to the container is zero.
	var filesHash uint64

	if len(c.Files) > 0 {
		for _, f := range c.Files {
			var fileContent []byte
			// Read the file content to calculate the hash, if there is an error reading the file,
			// the hash will be zero to avoid breaking the hash calculation.
			if f.Reader != nil {
				fileContent, err = io.ReadAll(f.Reader)
				if err != nil {
					continue
				}
			} else {
				ok, err := isDir(f.HostFilePath)
				if err != nil {
					continue
				}

				if !ok {
					// Calculate the hash of the file content only if it is a file.
					fileContent, err = os.ReadFile(f.HostFilePath)
					if err != nil {
						continue
					}
				}
				// The else of this condition is a NOOP, as calculating the hash of the directory content is not supported.
			}

			fh, err := core.Hash(fileContent)
			if err != nil {
				continue
			}
			filesHash += fh
		}

		ch.FilesHash = filesHash
	}

	return ch
}
