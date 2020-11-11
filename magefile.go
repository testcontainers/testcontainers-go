//+build mage

package main

import (
	"fmt"

	"github.com/magefile/mage/sh"
)

// Runs go fmt
func Format() error {
	output, err := sh.Output("go", "fmt", "./...")
	if output != "" {
		return fmt.Errorf("Found unformatted files: %s\n", output)
	}
	return err
}
