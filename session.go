package testcontainers

import (
	"sync"

	"github.com/google/uuid"
)

var sessionID uuid.UUID
var sessionIDOnce sync.Once

func SessionID() uuid.UUID {
	sessionIDOnce.Do(func() {
		sessionID = uuid.New()
	})

	return sessionID
}
