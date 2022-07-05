package wait

import (
	"testing"

	"github.com/docker/go-connections/nat"
)

func Test_waitForSql_WithQuery(t *testing.T) {
	t.Run("default query", func(t *testing.T) {
		w := ForSQL("5432/tcp", "postgres", func(port nat.Port) string {
			return "fake-url"
		})

		if got := w.query; got != defaultForSqlQuery {
			t.Fatalf("expected %s, got %s", defaultForSqlQuery, got)
		}
	})
	t.Run("custom query", func(t *testing.T) {
		const q = "SELECT 100;"

		w := ForSQL("5432/tcp", "postgres", func(port nat.Port) string {
			return "fake-url"
		}).WithQuery(q)

		if got := w.query; got != q {
			t.Fatalf("expected %s, got %s", q, got)
		}
	})
}
