package cockroachdb

import (
	"context"
	"errors"
	"io"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/moby/moby/api/types/network"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/exec"
	"github.com/testcontainers/testcontainers-go/wait"
)

type connectionTarget struct {
	testcontainers.Container
	hostErr error
	portErr error
	command []string
	exits   []int
}

func (c *connectionTarget) Host(context.Context) (string, error) {
	return "::1", c.hostErr
}

func (c *connectionTarget) MappedPort(context.Context, string) (network.Port, error) {
	return network.MustParsePort("30000/tcp"), c.portErr
}

func (c *connectionTarget) Exec(_ context.Context, cmd []string, _ ...exec.ProcessOption) (int, io.Reader, error) {
	c.command = cmd
	exit := c.exits[0]
	if len(c.exits) > 1 {
		c.exits = c.exits[1:]
	}
	return exit, strings.NewReader("1\n1\n"), nil
}

func TestConnectionInfo(t *testing.T) {
	target := &connectionTarget{}
	c := &CockroachDBContainer{Container: target, options: options{
		user: "a@b", password: "p@:/?#%", database: "db/name ?#%",
		tlsStrategy: wait.ForTLSCert("cert", "key").WithServerName("127.0.0.1"),
	}}
	info, err := c.ConnectionInfo(context.Background())
	require.NoError(t, err)
	u, err := url.Parse(info.URL)
	require.NoError(t, err)
	require.Equal(t, "[::1]:30000", u.Host)
	require.Equal(t, "a@b", u.User.Username())
	password, ok := u.User.Password()
	require.True(t, ok)
	require.Equal(t, "p@:/?#%", password)
	require.Equal(t, "/db/name ?#%", u.Path)
	require.Equal(t, "verify-full", u.Query().Get("sslmode"))
	require.Equal(t, "127.0.0.1", info.TLSConfig.ServerName)
	info.TLSConfig.ServerName = "changed"
	require.Equal(t, "127.0.0.1", c.tlsStrategy.TLSConfig().ServerName)

	c.tlsStrategy = nil
	info, err = c.ConnectionInfo(context.Background())
	require.NoError(t, err)
	require.Nil(t, info.TLSConfig)
	require.Contains(t, info.URL, "sslmode=disable")

	errLookup := errors.New("lookup")
	target.hostErr = errLookup
	_, err = c.ConnectionInfo(context.Background())
	require.ErrorIs(t, err, errLookup)
	require.ErrorContains(t, err, "host:")
	target.portErr = errLookup
	_, err = c.ConnectionInfo(context.Background())
	require.ErrorIs(t, err, errLookup)
	require.ErrorContains(t, err, "mapped port:")
	require.Panics(t, func() { c.MustConnectionString(context.Background()) })
}

func TestSQLReadiness(t *testing.T) {
	for _, insecure := range []bool{false, true} {
		name := "secure"
		if insecure {
			name = "insecure"
		}
		t.Run(name, func(t *testing.T) {
			command := []string{"cockroach", "sql", "--url", "", "--execute", "SELECT 1"}
			probe := wait.ForExec(command).WithPollInterval(time.Millisecond)
			req := &testcontainers.GenericContainerRequest{ContainerRequest: testcontainers.ContainerRequest{
				Env:        map[string]string{envUser: "testuser", envDatabase: "testdatabase", envPassword: "p@:/?"},
				WaitingFor: wait.ForAll(wait.ForTLSCert("cert", "key"), probe),
			}}
			if insecure {
				req.Cmd = []string{insecureFlag}
				req.Env[envPassword] = ""
			}
			c := &CockroachDBContainer{}
			require.NoError(t, c.configure(command)(req))
			target := &connectionTarget{exits: []int{1, 0}}
			require.NoError(t, probe.WaitUntilReady(context.Background(), target))
			require.Equal(t, []string{"cockroach", "sql", "--url"}, target.command[:3])
			require.Equal(t, []string{"--execute", "SELECT 1"}, target.command[4:])
			u, err := url.Parse(target.command[3])
			require.NoError(t, err)
			require.Equal(t, "testuser", u.User.Username())
			require.Equal(t, "/testdatabase", u.Path)
			require.Equal(t, "127.0.0.1:26257", u.Host)
			if insecure {
				require.Equal(t, "disable", u.Query().Get("sslmode"))
				require.Empty(t, u.Query().Get("sslcert"))
			} else {
				require.Equal(t, "verify-full", u.Query().Get("sslmode"))
				require.Equal(t, "/cockroach/certs/client.testuser.crt", u.Query().Get("sslcert"))
				require.Equal(t, "/cockroach/certs/client.testuser.key", u.Query().Get("sslkey"))
				require.Equal(t, "/cockroach/certs/ca.crt", u.Query().Get("sslrootcert"))
				password, _ := u.User.Password()
				require.Equal(t, "p@:/?", password)
			}
			// A failed SQL command must keep waiting and respect cancellation.
			target.exits = []int{1}
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
			defer cancel()
			require.ErrorIs(t, probe.WaitUntilReady(ctx, target), context.DeadlineExceeded)
		})
	}
}
