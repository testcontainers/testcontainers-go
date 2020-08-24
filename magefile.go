//+build mage

package main

import "github.com/magefile/mage/sh"

// Runs go fmt
func Format() error {
	return sh.Run("go", "fmt", "./...")
}
