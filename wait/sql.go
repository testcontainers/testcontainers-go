package wait

import (
	"context"
	"net"
	"os"
	"syscall"
	"time"

	"database/sql"
	"database/sql/driver"

	"github.com/lib/pq"
	"github.com/pkg/errors"
)

type SQLVariables map[string]interface{}

type SQLConnectorFromTarget func(ctx context.Context, target StrategyTarget, variables SQLVariables) (driver.Connector, error)

var _ Strategy = (*SQLStrategy)(nil)

type SQLStrategy struct {
	startupTimeout  time.Duration
	ConnectorSource SQLConnectorFromTarget
	SQLVariables    SQLVariables
}

func NewSQLStrategy(ds SQLConnectorFromTarget, sv SQLVariables) *SQLStrategy {
	return &SQLStrategy{
		startupTimeout:  defaultStartupTimeout(),
		ConnectorSource: ds,
		SQLVariables:    sv,
	}
}

func ForSQL(ds SQLConnectorFromTarget, sv SQLVariables) *SQLStrategy {
	return NewSQLStrategy(ds, sv)
}

func (ws *SQLStrategy) WaitUntilReady(ctx context.Context, target StrategyTarget) error {

	conn, err := ws.ConnectorSource(ctx, target, ws.SQLVariables)
	if err != nil {
		return errors.Wrap(err, "could not retrieve the SQL connector from the provided function")
	}

	db := sql.OpenDB(conn)

	for {
		_, err := db.ExecContext(ctx, "SELECT 1")
		if err != nil {
			// First we wait a bit for TCP connections getting reset...
			if v, ok := err.(*net.OpError); ok {
				if v2, ok := (v.Err).(*os.SyscallError); ok {
					if v2.Err == syscall.ECONNRESET {
						time.Sleep(100 * time.Millisecond)
						continue
					}
				}
			}

			// Then for Postgres "starting up"
			if v, ok := err.(*pq.Error); ok {
				if v.Message == "the database system is starting up" {
					time.Sleep(500 * time.Millisecond)
					continue
				}
			}
			return errors.Wrap(err, "an unexpected error occurred while waiting for the DB to start")
		}
		break
	}

	return nil
}
