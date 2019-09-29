package wait

import (
	"context"
	"time"

	"database/sql"
	"database/sql/driver"

	"github.com/pkg/errors"
)

type SQLVariables map[string]interface{}

type SQLConnectorFromTarget func(ctx context.Context, target StrategyTarget, variables SQLVariables) (driver.Connector, error)

var _ Strategy = (*SQLStrategy)(nil)

type SQLStrategy struct {
	startupTimeout  time.Duration
	PollInterval    time.Duration
	ConnectorSource SQLConnectorFromTarget
	SQLVariables    SQLVariables
}

func NewSQLStrategy(ds SQLConnectorFromTarget, sv SQLVariables) *SQLStrategy {
	return &SQLStrategy{
		startupTimeout:  defaultStartupTimeout(),
		PollInterval:    500 * time.Millisecond,
		ConnectorSource: ds,
		SQLVariables:    sv,
	}
}

func ForSQL(ds SQLConnectorFromTarget, sv SQLVariables) *SQLStrategy {
	return NewSQLStrategy(ds, sv)
}

// WithStartupTimeout can be used to change the default startup timeout
func (ws *SQLStrategy) WithStartupTimeout(startupTimeout time.Duration) *SQLStrategy {
	ws.startupTimeout = startupTimeout
	return ws
}

// WithPollInterval can be used to override the default polling interval of 100 milliseconds
func (ws *SQLStrategy) WithPollInterval(pollInterval time.Duration) *SQLStrategy {
	ws.PollInterval = pollInterval
	return ws
}

func (ws *SQLStrategy) WaitUntilReady(ctx context.Context, target StrategyTarget) error {

	ctx, cancelContext := context.WithTimeout(ctx, ws.startupTimeout)
	defer cancelContext()

	conn, err := ws.ConnectorSource(ctx, target, ws.SQLVariables)
	if err != nil {
		return errors.Wrap(err, "could not retrieve the SQL connector from the provided function")
	}

	db := sql.OpenDB(conn)

LOOP:
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			_, err := db.ExecContext(ctx, "SELECT 1")
			if err != nil {
				time.Sleep(ws.PollInterval)
				continue
			}
			break LOOP
		}

	}

	return nil
}
