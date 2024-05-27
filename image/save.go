package image

import (
	"context"
	"fmt"
	"os"

	"github.com/testcontainers/testcontainers-go/internal/core"
)

// SaveImages exports a list of images as an uncompressed tar
func SaveImages(ctx context.Context, output string, images ...string) error {
	outputFile, err := os.Create(output)
	if err != nil {
		return fmt.Errorf("opening output file %w", err)
	}
	defer func() {
		_ = outputFile.Close()
	}()

	cli, err := core.NewClient(ctx)
	if err != nil {
		return err
	}
	defer cli.Close()

	imageReader, err := cli.ImageSave(ctx, images)
	if err != nil {
		return fmt.Errorf("saving images %w", err)
	}
	defer func() {
		_ = imageReader.Close()
	}()

	// Attempt optimized readFrom, implemented in linux
	_, err = outputFile.ReadFrom(imageReader)
	if err != nil {
		return fmt.Errorf("writing images to output %w", err)
	}

	return nil
}
