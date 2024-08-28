package core_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/internal/core"
)

type ExampleTag struct {
	Name  string
	Age   int
	Score float64
	Valid bool
	Fn    func() error `hash:"ignore"`
}

type ExampleNoTag struct {
	Name  string
	Age   int
	Score float64
	Valid bool
	Fn    func() error
}

func TestHashStruct_withTag(t *testing.T) {
	e := ExampleTag{"Alice", 30, 95.5, true, nil}

	hash, err := core.Hash(e)
	require.NoError(t, err)
	require.NotEqual(t, 0, hash)
}

func TestHashStruct_withNoTag(t *testing.T) {
	e := ExampleNoTag{"Alice", 30, 95.5, true, nil}

	hash, err := core.Hash(e)
	require.Error(t, err)
	require.Equal(t, uint64(0), hash)
}
