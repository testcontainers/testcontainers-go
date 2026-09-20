package postgres

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// The actual "database already exists" race reported in #3233 is
// non-deterministic and, per the maintainers' own note on the related
// PR #3264, "very hard to enforce" in a test. This instead asserts the
// exact commands Restore issues, so the terminate-backend addition for
// c.dbName (and its ordering relative to the DROP) is verified directly,
// without needing a live database.
//
// The added terminate-backend call doesn't reach any connection DROP
// DATABASE ... WITH (FORCE) couldn't already terminate itself (FORCE uses
// the same pg_terminate_backend mechanism internally, against the same
// target, with the same documented exceptions). Its value is timing: an
// explicit, separate call gives those backends more wall-clock time to
// actually exit before FORCE's own connection check runs immediately
// after -- narrowing, not eliminating, the race.
func TestRestoreCommands(t *testing.T) {
	c := &PostgresContainer{dbName: "target_db", user: "myuser"}

	cmds := c.restoreCommands("snap1")
	require.Len(t, cmds, 4)

	require.Contains(t, cmds[0], "pg_terminate_backend")
	require.Contains(t, cmds[0], "datname = 'snap1'")

	require.Contains(t, cmds[1], "pg_terminate_backend")
	require.Contains(t, cmds[1], "datname = 'target_db'")

	require.True(t, strings.HasPrefix(cmds[2], `DROP DATABASE IF EXISTS "target_db"`))
	require.True(t, strings.HasPrefix(cmds[3], `CREATE DATABASE "target_db" WITH TEMPLATE "snap1" OWNER "myuser"`))

	// The target database's connections must be terminated before it is
	// dropped, same as the template database's connections are terminated
	// before the CREATE ... TEMPLATE step that requires it.
	terminateTargetIdx, dropIdx := -1, -1
	for i, cmd := range cmds {
		if strings.Contains(cmd, "datname = 'target_db'") {
			terminateTargetIdx = i
		}
		if strings.HasPrefix(cmd, `DROP DATABASE`) {
			dropIdx = i
		}
	}
	require.Less(t, terminateTargetIdx, dropIdx, "expected target database connections to be terminated before it is dropped")
}
