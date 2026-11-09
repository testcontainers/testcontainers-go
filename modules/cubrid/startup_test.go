package cubrid

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	startedOutput = `Starting CUBRID All Services...
Initializing database 'testdb'
Creating database with 128.0M size using locale en_US. The total amount of disk space needed is 1.1G.
CUBRID 11.4.6
Execute OK. (0.018000 sec) Committed. (0.004000 sec)
@ cubrid broker start
++ cubrid broker start: success
@ cubrid master start
++ cubrid master start: success
@ cubrid server start: testdb

This may take a long time depending on the amount of recovery works to do.
CUBRID 11.4.6
++ cubrid server start: success
CUBRID 11.4.6 (11.4.6.1963-0e7d3c1) (64bit release build for Linux) (Sep  7 2026 17:45:11)
`

	brokerOnlyOutput = `Starting CUBRID All Services...
Initializing database 'testdb'
@ cubrid broker start
++ cubrid broker start: success
@ cubrid server start: testdb
`

	localeOutput = `Starting CUBRID All Services...
Initializing database 'cubdb'
Couldn't create database.
Locale initialization: Locale xx_XX.utf8 was not loaded.
 xx_XX not found in cubrid_locales.txt.
Creating database with 128.0M size using locale xx_XX.utf8. The total amount of disk space needed is 1.1G.
@ cubrid broker start
++ cubrid broker start: success
@ cubrid server start: cubdb
++ cubrid server start: fail
`

	csqlOutput = `CUBRID 11.4.6
In line 1, column 31,
ERROR: In line 1, column 31 before '';'
Syntax error: unexpected 's', expecting ';'
CUBRID 11.4.6 (11.4.6.1963-0e7d3c1) (64bit release build for Linux) (Sep  7 2026 17:45:11)
`

	serverFailOutput = `@ cubrid server start: testdb
++ cubrid server start: fail
`
)

func TestCheckStartup(t *testing.T) {
	t.Run("started", func(t *testing.T) {
		require.NoError(t, checkStartup(startedOutput))
	})

	t.Run("still-starting", func(t *testing.T) {
		err := checkStartup(brokerOnlyOutput)
		require.EqualError(t, err, "cubrid broker and server have not started yet")

		var permanent *wait.PermanentError
		require.NotErrorAs(t, err, &permanent)
	})

	t.Run("unsupported-locale", func(t *testing.T) {
		requirePermanent(t, checkStartup(localeOutput), `cubrid failed to start:
Couldn't create database.
Locale initialization: Locale xx_XX.utf8 was not loaded.
xx_XX not found in cubrid_locales.txt.
Creating database with 128.0M size using locale xx_XX.utf8. The total amount of disk space needed is 1.1G.
@ cubrid broker start`)
	})

	t.Run("user-not-created", func(t *testing.T) {
		requirePermanent(t, checkStartup(csqlOutput), `cubrid failed to start:
ERROR: In line 1, column 31 before '';'
Syntax error: unexpected 's', expecting ';'
CUBRID 11.4.6 (11.4.6.1963-0e7d3c1) (64bit release build for Linux) (Sep  7 2026 17:45:11)`)
	})

	t.Run("server-failed", func(t *testing.T) {
		requirePermanent(t, checkStartup(serverFailOutput), "cubrid failed to start:\n++ cubrid server start: fail")
	})
}

func requirePermanent(t *testing.T, err error, msg string) {
	t.Helper()

	var permanent *wait.PermanentError
	require.ErrorAs(t, err, &permanent)
	require.EqualError(t, err, msg)
}
