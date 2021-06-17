package wait

import (
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/assert"
)

func TestWaitSql_TimeoutAccessors(t *testing.T) {
	strategy := ForSQL(nat.Port(8080), "", func(p nat.Port) string {
		return p.Port()
	})

	strategy.timeout = time.Second * 2
	assert.Equal(t, time.Second*2, strategy.timeout)

	strategy.WithTimeout(time.Second * 3)
	assert.Equal(t, time.Second*3, strategy.timeout)

	// Deprecated
	strategy.Timeout(time.Second * 4)
	assert.Equal(t, time.Second*4, strategy.timeout)
}
