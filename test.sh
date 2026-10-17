diff --git a/modules/postgres/wait_strategies_test.go b/modules/postgres/wait_strategies_test.go
new file mode 100644
index 0000000..b6e9ff2
--- /dev/null
+++ b/modules/postgres/wait_strategies_test.go
@@ -0,0 +1,107 @@
+package postgres_test
+
+import (
+	"context"
+	"database/sql"
+	"fmt"
+	"io"
+	"strings"
+	"testing"
+	"time"
+
+	_ "github.com/jackc/pgx/v5/stdlib"
+	"github.com/stretchr/testify/require"
+
+	"github.com/testcontainers/testcontainers-go"
+	"github.com/testcontainers/testcontainers-go/modules/postgres"
+)
+
+// TestBasicWaitStrategies_reusedContainer reproduces the false-positive ready
+// signal reported in https://github.com/testcontainers/testcontainers-go/issues/3671:
+// container logs survive restarts, so on a reused container a log-based wait
+// strategy can be satisfied by the output of a previous run and unblock before
+// the current postgres process accepts connections.
+func TestBasicWaitStrategies_reusedContainer(t *testing.T) {
+	ctx := context.Background()
+
+	reuseName := fmt.Sprintf("postgres-reused-wait-%d", time.Now().UnixNano())
+
+	run := func() *postgres.PostgresContainer {
+		t.Helper()
+		ctr, err := postgres.Run(ctx, "postgres:16-alpine",
+			postgres.WithDatabase(dbname),
+			postgres.WithUsername(user),
+			postgres.WithPassword(password),
+			postgres.BasicWaitStrategies(),
+			testcontainers.WithReuseByName(reuseName),
+		)
+		testcontainers.CleanupContainer(t, ctr)
+		require.NoError(t, err)
+		return ctr
+	}
+
+	connect := func(c *postgres.PostgresContainer) *sql.DB {
+		t.Helper()
+		connStr, err := c.ConnectionString(ctx, "sslmode=disable")
+		require.NoError(t, err)
+		db, err := sql.Open("pgx", connStr)
+		require.NoError(t, err)
+		t.Cleanup(func() { db.Close() })
+		return db
+	}
+
+	// recoveries counts how many crash recoveries the container has logged.
+	// Logs accumulate across restarts of the same container, which is the very
+	// property that makes log-based waits unsafe with reuse.
+	recoveries := func(c *postgres.PostgresContainer) int {
+		t.Helper()
+		rc, err := c.Logs(ctx)
+		require.NoError(t, err)
+		defer rc.Close()
+		logs, err := io.ReadAll(rc)
+		require.NoError(t, err)
+		return strings.Count(string(logs), "database system was not properly shut down")
+	}
+
+	const rowCount = 3_000_000
+
+	ctr := run()
+
+	db := connect(ctr)
+	_, err := db.ExecContext(ctx, "CREATE TABLE reuse_wait (v int)")
+	require.NoError(t, err)
+
+	for i := 0; i < 2; i++ {
+		// Generate WAL so that the unclean restart below forces crash recovery,
+		// during which postgres does not accept connections yet.
+		_, err = db.ExecContext(ctx, "TRUNCATE reuse_wait")
+		require.NoError(t, err)
+		_, err = db.ExecContext(ctx, fmt.Sprintf("INSERT INTO reuse_wait SELECT generate_series(1, %d)", rowCount))
+		require.NoError(t, err)
+
+		// Stop with a zero timeout follows SIGTERM with an immediate SIGKILL.
+		// The session above is kept open on purpose: postgres waits for it on
+		// SIGTERM, so the SIGKILL always interrupts an unclean shutdown and the
+		// next start is guaranteed to run crash recovery.
+		noGrace := time.Duration(0)
+		require.NoError(t, ctr.Stop(ctx, &noGrace))
+
+		ctr = run()
+
+		// The wait strategy must not unblock before postgres accepts connections:
+		// a single attempt with no retries must succeed.
+		db = connect(ctr)
+		var rows int
+		require.NoErrorf(t,
+			db.QueryRowContext(ctx, "SELECT count(*) FROM reuse_wait").Scan(&rows),
+			"reuse cycle %d: container reported ready before postgres accepted connections", i,
+		)
+		require.Equal(t, rowCount, rows)
+
+		// Guard the reproducer itself: every cycle must have gone through crash
+		// recovery, otherwise the scenario above degraded silently.
+		require.Equalf(t, i+1, recoveries(ctr),
+			"reuse cycle %d: expected the restart to run crash recovery", i,
+		)
+	}
+}
diff --git a/test.sh b/test.sh
new file mode 100755
index 0000000..2fb53a5
--- /dev/null
+++ b/test.sh
@@ -0,0 +1,53 @@
+#!/usr/bin/env bash
+set -uo pipefail
+
+OUTPUT_PATH=""
+MODE=""
+
+while [[ $# -gt 0 ]]; do
+    case "$1" in
+        --output_path)
+            OUTPUT_PATH="$2"
+            shift 2
+            ;;
+        base|new)
+            MODE="$1"
+            shift
+            ;;
+        *)
+            shift
+            ;;
+    esac
+done
+
+if [[ -z "$OUTPUT_PATH" ]]; then
+    echo "Error: --output_path is required" >&2
+    exit 1
+fi
+
+if [[ -z "$MODE" ]]; then
+    echo "Error: mode (base or new) is required" >&2
+    exit 1
+fi
+
+REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
+cd "$REPO_ROOT/modules/postgres"
+
+case "$MODE" in
+    base)
+        go test -v -count=1 -timeout 10m \
+            -run "^(TestContainerWithWaitForSQL|TestWithConfigFile|TestWithInitScript|TestWithOrderedInitScript)$" \
+            ./... 2>&1 \
+            | go-junit-report -set-exit-code > "$OUTPUT_PATH"
+        ;;
+    new)
+        go test -v -count=1 -timeout 10m \
+            -run "^TestBasicWaitStrategies_reusedContainer$" \
+            ./... 2>&1 \
+            | go-junit-report -set-exit-code > "$OUTPUT_PATH"
+        ;;
+    *)
+        echo "Error: mode must be 'base' or 'new'" >&2
+        exit 1
+        ;;
+esac
