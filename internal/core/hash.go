package core

import (
	"fmt"

	"github.com/mitchellh/hashstructure/v2"
)

// Hash calculates a hash of a struct by hashing all of its fields.
func Hash(v interface{}) (uint64, error) {
	hash, err := hashstructure.Hash(v, hashstructure.FormatV2, nil)
	if err != nil {
		return 0, fmt.Errorf("hashing struct: %w", err)
	}

	return hash, nil
}
