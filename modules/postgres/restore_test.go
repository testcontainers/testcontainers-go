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

// A database name containing an apostrophe (a valid, if unusual, Postgres
// identifier when created via a quoted CREATE DATABASE "o'brien_db") must
// not break out of the single-quoted datname string literal used by the
// pg_terminate_backend calls.
func TestRestoreCommandsQuotesApostropheInLiteral(t *testing.T) {
	c := &PostgresContainer{dbName: "o'brien_db", user: "myuser"}

	cmds := c.restoreCommands("snap'shot")
	require.Len(t, cmds, 4)

	require.Contains(t, cmds[0], "datname = 'snap''shot'")
	require.Contains(t, cmds[1], "datname = 'o''brien_db'")
}

// A database name containing a backslash must round-trip as a single literal
// backslash regardless of the server's standard_conforming_strings setting.
// Doubling only the quote (and leaving the string as a plain '...' literal)
// is unsafe when standard_conforming_strings=off, since a trailing backslash
// then escapes the closing quote instead of terminating the string -- see
// https://www.postgresql.org/docs/current/sql-syntax-lexical.html#SQL-SYNTAX-STRINGS-ESCAPE.
func TestRestoreCommandsQuotesBackslashInLiteral(t *testing.T) {
	c := &PostgresContainer{dbName: `evil\`, user: "myuser"}

	cmds := c.restoreCommands("snap1")

	require.Contains(t, cmds[1], `E'evil\\'`)
	require.NotContains(t, cmds[1], `datname = 'evil\'`)
}

// Dedicated unit tests for quoteLiteral itself, requested in review
// (https://github.com/testcontainers/testcontainers-go/pull/3907) as full
// coverage in addition to the exercising already done indirectly via
// TestRestoreCommandsQuotesApostropheInLiteral/Backslash above.
func TestQuoteLiteral(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "empty string",
			input: "",
			want:  `''`,
		},
		{
			name:  "plain identifier, no special characters",
			input: "mydb",
			want:  `'mydb'`,
		},
		{
			name:  "single apostrophe",
			input: "o'brien",
			want:  `'o''brien'`,
		},
		{
			name:  "multiple apostrophes",
			input: "a'b'c",
			want:  `'a''b''c'`,
		},
		{
			name:  "leading and trailing apostrophes",
			input: "'wrapped'",
			want:  `'''wrapped'''`,
		},
		{
			name:  "single backslash",
			input: `evil\`,
			want:  `E'evil\\'`,
		},
		{
			name:  "multiple backslashes",
			input: `a\b\c`,
			want:  `E'a\\b\\c'`,
		},
		{
			name:  "backslash and apostrophe together",
			input: `o'brien\`,
			// The apostrophe is doubled regardless of the backslash branch,
			// since the doubling happens unconditionally before the
			// backslash check.
			want: `E'o''brien\\'`,
		},
		{
			name: "backslash immediately followed by apostrophe",
			// A naive doubled-quote-only implementation would let this
			// backslash escape the following quote instead of the literal's
			// closing quote, under standard_conforming_strings=off.
			input: `\'`,
			want:  `E'\\'''`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := quoteLiteral(tt.input)
			require.Equal(t, tt.want, got)

			// The literal must always be wrapped in a matching pair of
			// single quotes (optionally E-prefixed), never leave one
			// dangling open.
			require.True(t, strings.HasPrefix(got, "'") || strings.HasPrefix(got, "E'"))
			require.True(t, strings.HasSuffix(got, "'"))
		})
	}
}
