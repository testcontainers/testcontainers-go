package testcontainers

import (
	"sync"

	"github.com/google/uuid"
)

var tcSessionID uuid.UUID
var tcSessionIDOnce sync.Once

func sessionID() uuid.UUID {
	tcSessionIDOnce.Do(func() {
		tcSessionID = uuid.New()
	})

	return tcSessionID
}
