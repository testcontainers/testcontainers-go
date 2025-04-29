package config

import (
	"encoding/base64"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecodeBase64Auth(t *testing.T) {
	for _, tc := range base64TestCases() {
		t.Run(tc.name, testBase64Case(tc, func() (string, string, error) {
			return DecodeBase64Auth(tc.config)
		}))
	}
}

func TestConfig_GetRegistryCredentials(t *testing.T) {
	t.Run("from base64 auth", func(t *testing.T) {
		for _, tc := range base64TestCases() {
			t.Run(tc.name, func(t *testing.T) {
				config := Config{
					AuthConfigs: map[string]AuthConfig{
						"some.domain": tc.config,
					},
				}
				testBase64Case(tc, func() (string, string, error) {
					return config.GetRegistryCredentials("some.domain")
				})(t)
			})
		}
	})
}

type base64TestCase struct {
	name    string
	config  AuthConfig
	expUser string
	expPass string
	expErr  bool
}

func base64TestCases() []base64TestCase {
	cases := []base64TestCase{
		{name: "empty"},
		{name: "not base64", expErr: true, config: AuthConfig{Auth: "not base64"}},
		{name: "invalid format", expErr: true, config: AuthConfig{
			Auth: base64.StdEncoding.EncodeToString([]byte("invalid format")),
		}},
		{name: "happy case", expUser: "user", expPass: "pass", config: AuthConfig{
			Auth: base64.StdEncoding.EncodeToString([]byte("user:pass")),
		}},
	}

	return cases
}

type testAuthFn func() (string, string, error)

func testBase64Case(tc base64TestCase, authFn testAuthFn) func(t *testing.T) {
	return func(t *testing.T) {
		t.Helper()

		u, p, err := authFn()
		if tc.expErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}

		require.Equal(t, tc.expUser, u)
		require.Equal(t, tc.expPass, p)
	}
}

// validateAuth is a helper function to validate the username and password for a given hostname.
func validateAuth(t *testing.T, hostname, expectedUser, expectedPass string) {
	t.Helper()

	username, password, err := GetRegistryCredentials(hostname)
	require.NoError(t, err)
	require.Equal(t, expectedUser, username)
	require.Equal(t, expectedPass, password)
}

// validateAuthError is a helper function to validate we get an error for the given hostname.
func validateAuthError(t *testing.T, hostname string, expectedErr error) {
	t.Helper()

	username, password, err := GetRegistryCredentials(hostname)
	require.Error(t, err)
	require.Equal(t, expectedErr.Error(), err.Error())
	require.Empty(t, username)
	require.Empty(t, password)
}

// mockExecCommand is a helper function to mock exec.LookPath and exec.Command for testing.
func mockExecCommand(t *testing.T, env ...string) {
	t.Helper()

	execLookPath = func(file string) (string, error) {
		switch file {
		case "docker-credential-helper":
			return os.Args[0], nil
		case "docker-credential-error":
			return "", errors.New("lookup error")
		}

		return "", exec.ErrNotFound
	}

	execCommand = func(name string, arg ...string) *exec.Cmd {
		cmd := exec.Command(name, arg...)
		cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")
		cmd.Env = append(cmd.Env, env...)
		return cmd
	}

	t.Cleanup(func() {
		execLookPath = exec.LookPath
		execCommand = exec.Command
	})
}

func TestGetRegistryCredentials(t *testing.T) {
	t.Setenv(EnvOverrideDir, filepath.Join("testdata", "credhelpers-config"))

	t.Run("auths/user-pass", func(t *testing.T) {
		validateAuth(t, "userpass.io", "user", "pass")
	})

	t.Run("auths/auth", func(t *testing.T) {
		validateAuth(t, "auth.io", "auth", "authsecret")
	})

	t.Run("credsStore", func(t *testing.T) {
		validateAuth(t, "credstore.io", "", "")
	})

	t.Run("credHelpers/user-pass", func(t *testing.T) {
		mockExecCommand(t, `HELPER_STDOUT={"Username":"credhelper","Secret":"credhelpersecret"}`)
		validateAuth(t, "helper.io", "credhelper", "credhelpersecret")
	})

	t.Run("credHelpers/token", func(t *testing.T) {
		mockExecCommand(t, `HELPER_STDOUT={"Username":"<token>", "Secret":"credhelpersecret"}`)
		validateAuth(t, "helper.io", "", "credhelpersecret")
	})

	t.Run("credHelpers/not-found", func(t *testing.T) {
		mockExecCommand(t, "HELPER_STDOUT="+ErrCredentialsNotFound.Error(), "HELPER_EXIT_CODE=1")
		validateAuth(t, "helper.io", "", "")
	})

	t.Run("credHelpers/missing-url", func(t *testing.T) {
		mockExecCommand(t, "HELPER_STDOUT="+ErrCredentialsMissingServerURL.Error(), "HELPER_EXIT_CODE=1")
		validateAuthError(t, "helper.io", ErrCredentialsMissingServerURL)
	})

	t.Run("credHelpers/other-error", func(t *testing.T) {
		mockExecCommand(t, "HELPER_STDOUT=output", "HELPER_STDERR=my error", "HELPER_EXIT_CODE=10")
		expectedErr := errors.New(`execute "docker-credential-helper" stdout: "output" stderr: "my error": exit status 10`)
		validateAuthError(t, "helper.io", expectedErr)
	})

	t.Run("credHelpers/lookup-not-found", func(t *testing.T) {
		mockExecCommand(t, "HELPER_STDOUT=output", "HELPER_STDERR=my error", "HELPER_EXIT_CODE=10")
		validateAuth(t, "other.io", "", "")
	})

	t.Run("credHelpers/lookup-error", func(t *testing.T) {
		mockExecCommand(t, "HELPER_STDOUT=output", "HELPER_STDERR=my error", "HELPER_EXIT_CODE=10")
		expectedErr := errors.New(`look up "docker-credential-error": lookup error`)
		validateAuthError(t, "error.io", expectedErr)
	})

	t.Run("credHelpers/decode-json", func(t *testing.T) {
		mockExecCommand(t, "HELPER_STDOUT=bad-json")
		expectedErr := errors.New(`unmarshal credentials from: "docker-credential-helper": invalid character 'b' looking for beginning of value`)
		validateAuthError(t, "helper.io", expectedErr)
	})

	t.Run("config/not-found", func(t *testing.T) {
		t.Setenv(EnvOverrideDir, filepath.Join("testdata", "missing"))
		validateAuth(t, "userpass.io", "", "")
	})
}

// TestMain is hijacked so we can run a test helper which can write
// cleanly to stdout and stderr.
func TestMain(m *testing.M) {
	pid := os.Getpid()
	if os.Getenv("GO_EXEC_TEST_PID") == "" {
		os.Setenv("GO_EXEC_TEST_PID", strconv.Itoa(pid))
		// Run the tests.
		os.Exit(m.Run())
	}

	// Run the helper which slurps stdin and writes to stdout and stderr.
	if _, err := io.Copy(io.Discard, os.Stdin); err != nil {
		if _, err = os.Stderr.WriteString(err.Error()); err != nil {
			panic(err)
		}
	}

	if out := os.Getenv("HELPER_STDOUT"); out != "" {
		if _, err := os.Stdout.WriteString(out); err != nil {
			panic(err)
		}
	}

	if out := os.Getenv("HELPER_STDERR"); out != "" {
		if _, err := os.Stderr.WriteString(out); err != nil {
			panic(err)
		}
	}

	if code := os.Getenv("HELPER_EXIT_CODE"); code != "" {
		code, err := strconv.Atoi(code)
		if err != nil {
			panic(err)
		}

		os.Exit(code)
	}
}
