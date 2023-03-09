package testcontainerssession

import (
	"sync"

	"github.com/google/uuid"
)

var id uuid.UUID
var idOnce sync.Once

func ID() uuid.UUID {
	idOnce.Do(func() {
		id = uuid.New()
	})

	return id
}

func String() string {
	return ID().String()
}
