package dex

import (
	"fmt"
	"slices"

	"golang.org/x/crypto/bcrypt"
)

// ClientRegistration is a copy of the fields supported by Dex's runtime admin
// API. It contains a secret; treat it as credentials. Runtime registration does
// not support grant types; use WithClient for non-default grants.
type ClientRegistration struct {
	ID           string
	Secret       string
	Name         string
	RedirectURIs []string
	Public       bool
}

// Registration returns an independent snapshot for runtime client registration.
func (c Client) Registration() ClientRegistration {
	return ClientRegistration{
		ID: c.id, Secret: c.secret, Name: c.name,
		RedirectURIs: slices.Clone(c.redirectURIs), Public: c.public,
	}
}

// PasswordRegistration is a password DB entry for runtime administration.
// Hash is a newly generated bcrypt hash, never the plaintext password.
type PasswordRegistration struct {
	Email    string
	Username string
	UserID   string
	Hash     []byte
}

// PasswordRegistration hashes the password at the module's existing bcrypt
// cost (10) and generates a UUIDv4 if WithUserID was not supplied. Each call
// returns fresh data; pin WithUserID when a stable subject is required.
func (u User) PasswordRegistration() (PasswordRegistration, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(u.password), testBcryptCost)
	if err != nil {
		return PasswordRegistration{}, fmt.Errorf("dex: bcrypt: %w", err)
	}
	userID := u.userID
	if userID == "" {
		uid, err := newUUIDv4()
		if err != nil {
			return PasswordRegistration{}, fmt.Errorf("dex: generate user id: %w", err)
		}
		userID = uid
	}
	return PasswordRegistration{Email: u.email, Username: u.username, UserID: userID, Hash: hash}, nil
}
