package wait

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/docker/go-connections/nat"
	"time"
)

//ForSQL constructs a new waitForSql strategy for the given driver
func ForSQL(port nat.Port, driver string, url func(nat.Port) string) *waitForSql {
	return &waitForSql{
		Port:   port,
		URL:    url,
		Driver: driver,
	}
}

type waitForSql struct {
	URL            func(port nat.Port) string
	Driver         string
	Port           nat.Port
	startupTimeout time.Duration
}

//Timeout sets the maximum waiting time for the strategy after which it'll give up and return an error
func (w *waitForSql) Timeout(duration time.Duration) *waitForSql {
	w.startupTimeout = duration
	return w
}

//WaitUntilReady repeatedly tries to run "SELECT 1" query on the given port using sql and driver.
// If the it doesn't succeed until the timeout value which defaults to 10 seconds, it will return an error
func (w *waitForSql) WaitUntilReady(ctx context.Context, target StrategyTarget) (err error) {
	if w.startupTimeout == 0 {
		w.startupTimeout = time.Second * 10
	}
	ctx, cancel := context.WithTimeout(ctx, w.startupTimeout)
	defer cancel()

	ticker := time.NewTicker(time.Millisecond * 100)
	defer ticker.Stop()

	port, err := target.MappedPort(ctx, w.Port)
	if err != nil {
		return fmt.Errorf("target.MappedPort: %v", err)
	}

	db, err := sql.Open(w.Driver, w.URL(port))
	if err != nil {
		return fmt.Errorf("sql.Open: %v", err)
	}
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
