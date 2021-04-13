package wait

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/docker/go-connections/nat"
)

var _ Strategy = &waitForSqlWithHost{}

//ForSQL constructs a new waitForSql strategy for the given driver
func ForSQL(port nat.Port, driver string, url func(nat.Port) string) *waitForSql {
	return &waitForSql{
		Port:           port,
		URL:            url,
		Driver:         driver,
		startupTimeout: defaultStartupTimeout(),
		PollInterval:   defaultPollInterval(),
	}
}

//ForSQLWithHost constructs a new waitForSqlWithHost strategy for the given driver.ForSQLWithHost
//It is functionally identical to waitForSql, but the URL gives you access to the DB's host as well as the port.
func ForSQLWithHost(port nat.Port, driver string, url func(nat.Port, string) string) *waitForSqlWithHost {
	return &waitForSqlWithHost{
		Port:           port,
		URL:            url,
		Driver:         driver,
		startupTimeout: defaultStartupTimeout(),
		PollInterval:   defaultPollInterval(),
	}
}

type waitForSql struct {
	URL            func(port nat.Port) string
	Driver         string
	Port           nat.Port
	startupTimeout time.Duration
	PollInterval   time.Duration
}

//Timeout sets the maximum waiting time for the strategy after which it'll give up and return an error
func (w *waitForSql) Timeout(duration time.Duration) *waitForSql {
	w.startupTimeout = duration
	return w
}

//WithPollInterval can be used to override the default polling interval of 100 milliseconds
func (w *waitForSql) WithPollInterval(pollInterval time.Duration) *waitForSql {
	w.PollInterval = pollInterval
	return w
}

//WaitUntilReady repeatedly tries to run "SELECT 1" query on the given port using sql and driver.
// If the it doesn't succeed until the timeout value which defaults to 60 seconds, it will return an error
func (w *waitForSql) WaitUntilReady(ctx context.Context, target StrategyTarget) (err error) {
	ctx, cancel := context.WithTimeout(ctx, w.startupTimeout)
	defer cancel()

	ticker := time.NewTicker(w.PollInterval)
	defer ticker.Stop()

	var port nat.Port
	port, err = target.MappedPort(ctx, w.Port)

	for port == "" {
		select {
		case <-ctx.Done():
			return fmt.Errorf("%s:%w", ctx.Err(), err)
		case <-ticker.C:
			port, err = target.MappedPort(ctx, w.Port)
		}
	}

	db, err := sql.Open(w.Driver, w.URL(port))
	if err != nil {
		return fmt.Errorf("sql.Open: %v", err)
	}
	defer db.Close()
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

type waitForSqlWithHost struct {
	URL            func(port nat.Port, host string) string
	Driver         string
	Port           nat.Port
	startupTimeout time.Duration
	PollInterval   time.Duration
}

//Timeout sets the maximum waiting time for the strategy after which it'll give up and return an error
func (w *waitForSqlWithHost) Timeout(duration time.Duration) *waitForSqlWithHost {
	w.startupTimeout = duration
	return w
}

//WithPollInterval can be used to override the default polling interval
func (w *waitForSqlWithHost) WithPollInterval(pollInterval time.Duration) *waitForSqlWithHost {
	w.PollInterval = pollInterval
	return w
}

//WaitUntilReady repeatedly tries to run "SELECT 1" query on the given port using sql and driver.
//If it doesn't succeed before the context timeout, it will return an error.
func (w *waitForSqlWithHost) WaitUntilReady(ctx context.Context, target StrategyTarget) (err error) {
	ctx, cancel := context.WithTimeout(ctx, w.startupTimeout)
	defer cancel()

	ticker := time.NewTicker(w.PollInterval)
	defer ticker.Stop()

	var port nat.Port
	port, err = target.MappedPort(ctx, w.Port)

	for port == "" {
		select {
		case <-ctx.Done():
			return fmt.Errorf("%s:%w", ctx.Err(), err)
		case <-ticker.C:
			port, err = target.MappedPort(ctx, w.Port)
		}
	}

	var host string
	host, err = target.Host(ctx)

	for host == "" {
		select {
		case <-ctx.Done():
			return fmt.Errorf("%s:%w", ctx.Err(), err)
		case <-ticker.C:
			host, err = target.Host(ctx)
		}
	}

	db, err := sql.Open(w.Driver, w.URL(port, host))
	if err != nil {
		return fmt.Errorf("sql.Open: %v", err)
	}
	defer db.Close()
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
