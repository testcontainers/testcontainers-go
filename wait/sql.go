package wait

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/docker/go-connections/nat"
)

var _ Strategy = (*HTTPStrategy)(nil)

//ForSQL constructs a new waitForSql strategy for the given driver
func ForSQL(port nat.Port, driver string, url func(nat.Port) string) *waitForSql {
	return &waitForSql{
		Port:   port,
		URL:    url,
		Driver: driver,
		// Not using the default duration here because it is too low. It will never work
		startupTimeout: 20 * time.Second,
	}
}

type waitForSql struct {
	URL            func(port nat.Port) string
	Driver         string
	Port           nat.Port
	startupTimeout time.Duration
}

//Timeout sets the maximum waiting time for the strategy after which it'll give up and return an error
//Deprecated: uses WaitStartupTimeout instead
func (w *waitForSql) Timeout(duration time.Duration) *waitForSql {
	w.startupTimeout = duration
	return w
}

func (ws *waitForSql) WithStartupTimeout(startupTimeout time.Duration) *waitForSql {
	ws.startupTimeout = startupTimeout
	return ws
}

//WaitUntilReady repeatedly tries to run "SELECT 1" query on the given port using sql and driver.
// If the it doesn't succeed until the timeout value which defaults to 60 seconds, it will return an error
func (w *waitForSql) WaitUntilReady(ctx context.Context, target StrategyTarget) (err error) {
	ctx, cancel := context.WithTimeout(ctx, w.startupTimeout)
	defer cancel()

	ticker := time.NewTicker(time.Second * 1)
	defer ticker.Stop()

	port, err := target.MappedPort(ctx, w.Port)
	if err != nil {
		return fmt.Errorf("target.MappedPort: %v", err)
	}

	db, err := sql.Open(w.Driver, w.URL(port))
	if err != nil {
		return fmt.Errorf("sql.Open: %v", err)
	}
	db.SetConnMaxLifetime(0)
	db.SetMaxIdleConns(3)
	db.SetMaxOpenConns(3)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:

			if _, err := db.ExecContext(ctx, "SELECT 1"); err != nil {
				continue
			}
			return nil
		}
	}
}
