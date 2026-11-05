package dex

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestClientRegistration_Copy(t *testing.T) {
	c, err := NewClient("app", WithClientSecret("secret"), WithClientName("App"),
		WithClientRedirectURIs("http://localhost/cb"), WithClientPublic())
	require.NoError(t, err)
	r := c.Registration()
	require.Equal(t, ClientRegistration{
		ID: "app", Secret: "secret", Name: "App",
		RedirectURIs: []string{"http://localhost/cb"}, Public: true,
	}, r)
	r.RedirectURIs[0] = "http://other/cb"
	require.Equal(t, []string{"http://localhost/cb"}, c.Registration().RedirectURIs)
}

func TestUserPasswordRegistration(t *testing.T) {
	u, err := NewUser("u@example.com", "user", "password")
	require.NoError(t, err)
	first, err := u.PasswordRegistration()
	require.NoError(t, err)
	second, err := u.PasswordRegistration()
	require.NoError(t, err)
	require.Equal(t, "u@example.com", first.Email)
	require.Equal(t, "user", first.Username)
	require.Regexp(t, `^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`, first.UserID)
	require.NotEqual(t, first.UserID, second.UserID)
	require.NotEqual(t, first.Hash, second.Hash)
	require.NoError(t, bcrypt.CompareHashAndPassword(first.Hash, []byte("password")))
	cost, err := bcrypt.Cost(first.Hash)
	require.NoError(t, err)
	require.Equal(t, 10, cost)

	pinned, err := NewUser("u@example.com", "user", "password", WithUserID("stable"))
	require.NoError(t, err)
	r, err := pinned.PasswordRegistration()
	require.NoError(t, err)
	require.Equal(t, "stable", r.UserID)

	longPassword, err := NewUser("u@example.com", "user", strings.Repeat("p", 73))
	require.NoError(t, err)
	_, err = longPassword.PasswordRegistration()
	require.ErrorIs(t, err, bcrypt.ErrPasswordTooLong)
}
