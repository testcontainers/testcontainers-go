package wait

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/require"
)

func Test_waitForSql_WithQuery(t *testing.T) {
	t.Run("default query", func(t *testing.T) {
		w := ForSQL("5432/tcp", "postgres", func(_ string, _ nat.Port) string {
			return "fake-url"
		})

		require.Equal(t, defaultForSQLQuery, w.query)
	})
	t.Run("custom query", func(t *testing.T) {
		const q = "SELECT 100;"

		w := ForSQL("5432/tcp", "postgres", func(_ string, _ nat.Port) string {
			return "fake-url"
		}).WithQuery(q)

		require.Equal(t, q, w.query)
	})
}

func init() {
	sql.Register("mock", &mockSQLDriver{})
}

type mockSQLDriver struct {
	driver.Driver
}

func (sd *mockSQLDriver) Open(_ string) (driver.Conn, error) {
	return &mockSQLConn{}, nil
}

type mockSQLConn struct {
	driver.Conn
	driver.ConnBeginTx
	driver.ConnPrepareContext
}

func (sc *mockSQLConn) Close() error {
	return nil
}

func (sc *mockSQLConn) PrepareContext(_ context.Context, _ string) (driver.Stmt, error) {
	return &mockSQLStmt{}, nil
}

type mockSQLStmt struct {
	driver.Stmt
	driver.StmtExecContext
	driver.StmtQueryContext
}

func (st *mockSQLStmt) Close() error {
	return nil
}

func (st *mockSQLStmt) NumInput() int {
	return 0
}

func (st *mockSQLStmt) ExecContext(_ context.Context, _ []driver.NamedValue) (driver.Result, error) {
	return nil, nil
}

func TestWaitForSQLSucceeds(t *testing.T) {
	var mappedPortCount int
	target := &MockStrategyTarget{
		HostImpl: func(_ context.Context) (string, error) {
			return "localhost", nil
		},
		MappedPortImpl: func(_ context.Context, _ nat.Port) (nat.Port, error) {
			defer func() { mappedPortCount++ }()
			if mappedPortCount == 0 {
				return "", ErrPortNotFound
			}
			return "49152", nil
		},
		StateImpl: func(_ context.Context) (*container.State, error) {
			return &container.State{
				Running: true,
			}, nil
		},
	}

	wg := ForSQL("3306", "mock", func(_ string, _ nat.Port) string { return "" }).
		WithStartupTimeout(500 * time.Millisecond).
		WithPollInterval(100 * time.Millisecond)

	err := wg.WaitUntilReady(context.Background(), target)
	require.NoError(t, err)
}

func TestWaitForSQLFailsWhileGettingPortDueToOOMKilledContainer(t *testing.T) {
	var mappedPortCount int
	target := &MockStrategyTarget{
		HostImpl: func(_ context.Context) (string, error) {
			return "localhost", nil
		},
		MappedPortImpl: func(_ context.Context, _ nat.Port) (nat.Port, error) {
			defer func() { mappedPortCount++ }()
			if mappedPortCount == 0 {
				return "", ErrPortNotFound
			}
			return "49152", nil
		},
		StateImpl: func(_ context.Context) (*container.State, error) {
			return &container.State{
				OOMKilled: true,
			}, nil
		},
	}

	wg := ForSQL("3306", "mock", func(_ string, _ nat.Port) string { return "" }).
		WithStartupTimeout(500 * time.Millisecond).
		WithPollInterval(100 * time.Millisecond)

	{
		err := wg.WaitUntilReady(context.Background(), target)
		require.EqualError(t, err, "container crashed with out-of-memory (OOMKilled)")
	}
}

func TestWaitForSQLFailsWhileGettingPortDueToExitedContainer(t *testing.T) {
	var mappedPortCount int
	target := &MockStrategyTarget{
		HostImpl: func(_ context.Context) (string, error) {
			return "localhost", nil
		},
		MappedPortImpl: func(_ context.Context, _ nat.Port) (nat.Port, error) {
			defer func() { mappedPortCount++ }()
			if mappedPortCount == 0 {
				return "", ErrPortNotFound
			}
			return "49152", nil
		},
		StateImpl: func(_ context.Context) (*container.State, error) {
			return &container.State{
				Status:   "exited",
				ExitCode: 1,
			}, nil
		},
	}

	wg := ForSQL("3306", "mock", func(_ string, _ nat.Port) string { return "" }).
		WithStartupTimeout(500 * time.Millisecond).
		WithPollInterval(100 * time.Millisecond)

	{
		err := wg.WaitUntilReady(context.Background(), target)
		require.EqualError(t, err, "container exited with code 1")
	}
}

func TestWaitForSQLFailsWhileGettingPortDueToUnexpectedContainerStatus(t *testing.T) {
	var mappedPortCount int
	target := &MockStrategyTarget{
		HostImpl: func(_ context.Context) (string, error) {
			return "localhost", nil
		},
		MappedPortImpl: func(_ context.Context, _ nat.Port) (nat.Port, error) {
			defer func() { mappedPortCount++ }()
			if mappedPortCount == 0 {
				return "", ErrPortNotFound
			}
			return "49152", nil
		},
		StateImpl: func(_ context.Context) (*container.State, error) {
			return &container.State{
				Status: "dead",
			}, nil
		},
	}

	wg := ForSQL("3306", "mock", func(_ string, _ nat.Port) string { return "" }).
		WithStartupTimeout(500 * time.Millisecond).
		WithPollInterval(100 * time.Millisecond)

	{
		err := wg.WaitUntilReady(context.Background(), target)
		require.EqualError(t, err, "unexpected container status \"dead\"")
	}
}

func TestWaitForSQLFailsWhileQueryExecutingDueToOOMKilledContainer(t *testing.T) {
	target := &MockStrategyTarget{
		HostImpl: func(_ context.Context) (string, error) {
			return "localhost", nil
		},
		MappedPortImpl: func(_ context.Context, _ nat.Port) (nat.Port, error) {
			return "49152", nil
		},
		StateImpl: func(_ context.Context) (*container.State, error) {
			return &container.State{
				OOMKilled: true,
			}, nil
		},
	}

	wg := ForSQL("3306", "mock", func(_ string, _ nat.Port) string { return "" }).
		WithStartupTimeout(500 * time.Millisecond).
		WithPollInterval(100 * time.Millisecond)

	{
		err := wg.WaitUntilReady(context.Background(), target)
		require.EqualError(t, err, "container crashed with out-of-memory (OOMKilled)")
	}
}

func TestWaitForSQLFailsWhileQueryExecutingDueToExitedContainer(t *testing.T) {
	target := &MockStrategyTarget{
		HostImpl: func(_ context.Context) (string, error) {
			return "localhost", nil
		},
		MappedPortImpl: func(_ context.Context, _ nat.Port) (nat.Port, error) {
			return "49152", nil
		},
		StateImpl: func(_ context.Context) (*container.State, error) {
			return &container.State{
				Status:   "exited",
				ExitCode: 1,
			}, nil
		},
	}

	wg := ForSQL("3306", "mock", func(_ string, _ nat.Port) string { return "" }).
		WithStartupTimeout(500 * time.Millisecond).
		WithPollInterval(100 * time.Millisecond)

	{
		err := wg.WaitUntilReady(context.Background(), target)
		require.EqualError(t, err, "container exited with code 1")
	}
}

func TestWaitForSQLFailsWhileQueryExecutingDueToUnexpectedContainerStatus(t *testing.T) {
	target := &MockStrategyTarget{
		HostImpl: func(_ context.Context) (string, error) {
			return "localhost", nil
		},
		MappedPortImpl: func(_ context.Context, _ nat.Port) (nat.Port, error) {
			return "49152", nil
		},
		StateImpl: func(_ context.Context) (*container.State, error) {
			return &container.State{
				Status: "dead",
			}, nil
		},
	}

	wg := ForSQL("3306", "mock", func(_ string, _ nat.Port) string { return "" }).
		WithStartupTimeout(500 * time.Millisecond).
		WithPollInterval(100 * time.Millisecond)

	{
		err := wg.WaitUntilReady(context.Background(), target)
		require.EqualError(t, err, "unexpected container status \"dead\"")
	}
}
