package wait

import (
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/assert"
)

func TestWaitHostPort_TimeoutAccessors(t *testing.T) {
	strategy := ForListeningPort(nat.Port(8080))

	strategy.timeout = time.Second * 2
	assert.Equal(t, time.Second*2, strategy.timeout)

	strategy.WithTimeout(time.Second * 3)
	assert.Equal(t, time.Second*3, strategy.timeout)

	// Deprecated
	strategy.WithStartupTimeout(time.Second * 4)
	assert.Equal(t, time.Second*4, strategy.timeout)
}
