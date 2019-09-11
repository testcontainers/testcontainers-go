package testcontainers

import (
	"math/rand"
	"time"
)

var alfabet = []rune("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz")

// RandomString returns a random string
func RandomString(length int) string {
	rand.Seed(time.Now().UnixNano())

	b := make([]rune, length)
	for i := range b {
		b[i] = alfabet[rand.Intn(len(alfabet))]
	}
	return string(b)
}
