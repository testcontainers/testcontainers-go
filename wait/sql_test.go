package wait_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"

	"github.com/testcontainers/testcontainers-go/wait"
)

//TODO: fix
/*
func Test_waitForSql_WithQuery(t *testing.T) {
	t.Run("default query", func(t *testing.T) {
		w := wait.ForSQL("5432/tcp", "postgres", func(host string, port nat.Port) string {
			return "fake-url"
		})

		if got := w.query; got != defaultForSqlQuery {
			t.Fatalf("expected %s, got %s", defaultForSqlQuery, got)
		}
	})
	t.Run("custom query", func(t *testing.T) {
		const q = "SELECT 100;"

		w := wait.ForSQL("5432/tcp", "postgres", func(host string, port nat.Port) string {
			return "fake-url"
		}).WithQuery(q)

		if got := w.query; got != q {
			t.Fatalf("expected %s, got %s", q, got)
		}
	})
}
*/
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

// testForSQL tests the given strategy with different container
// state scenarios.
func testForSQL(t *testing.T, strategy wait.Strategy) {
	t.Helper()

	t.Run("running", func(t *testing.T) {
		newWaitBuilder().Run(t, strategy)
	})

	t.Run("oom", func(t *testing.T) {
		newWaitBuilder().State(oom).Run(t, strategy)
	})

	t.Run("exited", func(t *testing.T) {
		newWaitBuilder().State(exited).Run(t, strategy)
	})

	t.Run("dead", func(t *testing.T) {
		newWaitBuilder().State(dead).Run(t, strategy)
	})
}

func TestWaitForSQL(t *testing.T) {
	strategy := wait.ForSQL("3306", "mock", func(_ string, _ nat.Port) string { return "" }).
		WithStartupTimeout(500 * time.Millisecond).
		WithPollInterval(100 * time.Millisecond)

	testForSQL(t, strategy)
}
