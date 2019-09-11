package testcontainers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRandomStringHonorsLength(t *testing.T) {
	r := RandomString(10)
	assert.Equal(t, 10, len(r))

	r = RandomString(0)
	assert.Equal(t, 0, len(r))
}
