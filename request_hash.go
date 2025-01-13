package testcontainers

import (
	"fmt"
	"hash/fnv"
	"io"
	"os"
	"strconv"

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

func (c ContainerRequest) hash() (containerHash, error) {
	var ch containerHash
	hash, err := core.Hash(c)
	if err != nil {
		return ch, err
	}

	fileHashWriter := fnv.New64()
	defer fileHashWriter.Reset()

	for _, f := range c.Files {
		// Read the file content to calculate the hash, if there is an error reading the file,
		// the hash will be zero to avoid breaking the hash calculation.
		// It uses streaming to avoid loading the whole file in memory.
		if f.Reader != nil {
			fileHash := fnv.New64()
			_, err := io.Copy(fileHash, f.Reader)
			if err != nil {
				return ch, fmt.Errorf("copy file from reader: %w", err)
			}

			// Write the file hash into the combined hash
			_, err = fileHashWriter.Write([]byte(strconv.FormatUint(fileHash.Sum64(), 10)))
			if err != nil {
				return ch, fmt.Errorf("write hash: %w", err)
			}
			continue // move to the next file
		}

		// There is no reader, so we need to read the file content from the host file path.
		var fileContent []byte
		ok, err := isDir(f.HostFilePath)
		if err != nil {
			return ch, err
		} else if ok {
			// If the file is a directory, we skip it.
			continue
		}

		// Calculate the hash of the file content only if it is a file.
		fileContent, err = os.ReadFile(f.HostFilePath)
		if err != nil {
			return ch, fmt.Errorf("read file: %w", err)
		}

		// At this point, we have the file content in bytes, so we calculate its hash.
		fh, err := core.Hash(fileContent)
		if err != nil {
			return ch, fmt.Errorf("hash file: %w", err)
		}
		_, err = fileHashWriter.Write([]byte(strconv.FormatUint(fh, 10)))
		if err != nil {
			return ch, fmt.Errorf("write hash: %w", err)
		}
	}

	ch = containerHash{
		Hash: hash,
		// if there are no files, the filesHash will be zero because of the default value of the uint64 type.
		FilesHash: fileHashWriter.Sum64(),
	}

	return ch, nil
}
