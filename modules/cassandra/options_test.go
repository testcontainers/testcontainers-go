package cassandra_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go/modules/cassandra"
)

func TestWithSSL(t *testing.T) {
	// Test that WithSSL sets tlsEnabled to true
	opts := &cassandra.Options{}
	err := cassandra.WithSSL()(opts)
	require.NoError(t, err)
	assert.True(t, opts.TLSEnabled())
}

func TestGenerateJKSKeystore(t *testing.T) {
	// Test keystore generation
	keystorePath, certPath, err := cassandra.GenerateJKSKeystore()
	require.NoError(t, err)

	// Verify that both files exist
	_, err = os.Stat(keystorePath)
	require.NoError(t, err, "keystore file should exist")

	_, err = os.Stat(certPath)
	require.NoError(t, err, "certificate file should exist")

	// Verify file extensions
	assert.Equal(t, ".jks", filepath.Ext(keystorePath), "keystore should have .jks extension")
	assert.Equal(t, ".pem", filepath.Ext(certPath), "certificate should have .pem extension")

	// Clean up
	os.Remove(keystorePath)
	os.Remove(certPath)
}

func TestGenerateJKSKeystoreOverwrite(t *testing.T) {
	// Test that existing keystore is overwritten
	keystorePath, certPath, err := cassandra.GenerateJKSKeystore()
	require.NoError(t, err)

	// Get initial file info
	initialKeystoreInfo, err := os.Stat(keystorePath)
	require.NoError(t, err)

	// Generate new keystore
	newKeystorePath, newCertPath, err := cassandra.GenerateJKSKeystore()
	require.NoError(t, err)

	// Verify paths are the same
	assert.Equal(t, keystorePath, newKeystorePath)
	assert.Equal(t, certPath, newCertPath)

	// Get new file info
	newKeystoreInfo, err := os.Stat(keystorePath)
	require.NoError(t, err)

	// Verify that the file was modified
	assert.NotEqual(t, initialKeystoreInfo.ModTime(), newKeystoreInfo.ModTime(), "keystore should be overwritten")

	// Clean up
	os.Remove(keystorePath)
	os.Remove(certPath)
}

func TestGenerateJKSKeystoreInvalidKeytool(t *testing.T) {
	// Save original PATH
	originalPath := os.Getenv("PATH")
	defer os.Setenv("PATH", originalPath)

	// Set invalid PATH to make keytool unavailable
	os.Setenv("PATH", "/nonexistent")

	// Test that keystore generation fails when keytool is not available
	_, _, err := cassandra.GenerateJKSKeystore()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to generate keystore")
}
