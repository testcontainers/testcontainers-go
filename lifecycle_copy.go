package testcontainers

import (
	"context"
	"fmt"
	"io"
)

// defaultCopyFileToContainerHook is a hook that will copy files to the container after it's created
// but before it's started
var defaultCopyFileToContainerHook = func(files []ContainerFile) LifecycleHooks {
	return LifecycleHooks{
		PostCreates: []CreatedContainerHook{
			// copy files to container after it's created
			func(ctx context.Context, c CreatedContainer) error {
				for _, f := range files {
					if err := f.Validate(); err != nil {
						return fmt.Errorf("invalid file: %w", err)
					}

					var err error
					// Bytes takes precedence over HostFilePath
					if f.Reader != nil {
						bs, ioerr := io.ReadAll(f.Reader)
						if ioerr != nil {
							return fmt.Errorf("can't read from reader: %w", ioerr)
						}

						err = c.CopyToContainer(ctx, bs, f.ContainerFilePath, f.FileMode)
					} else {
						err = c.CopyFileToContainer(ctx, f.HostFilePath, f.ContainerFilePath, f.FileMode)
					}

					if err != nil {
						return fmt.Errorf("can't copy %s to container: %w", f.HostFilePath, err)
					}
				}

				return nil
			},
		},
	}
}
