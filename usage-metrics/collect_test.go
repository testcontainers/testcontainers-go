package main

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// zeroTimings eliminates all sleeps so tests run instantly.
var zeroTimings = retryTimings{
	maxPasses:         3,
	interRequestWait:  0,
	rateLimitCooldown: 0,
	passCooldown:      0,
}

// --- isRateLimitError ---

func TestIsRateLimitError(t *testing.T) {
	tests := []struct {
		msg  string
		want bool
	}{
		{"rate limit exceeded", true},
		{"secondary rate limit", true},
		{"403 Forbidden", true},
		{"429 Too Many Requests", true},
		{"500 Internal Server Error", false},
		{"502 Bad Gateway", false},
		{"connection refused", false},
		{"404 Not Found", false},
	}
	for _, tt := range tests {
		t.Run(tt.msg, func(t *testing.T) {
			got := isRateLimitError(errors.New(tt.msg))
			if got != tt.want {
				t.Errorf("isRateLimitError(%q) = %v, want %v", tt.msg, got, tt.want)
			}
		})
	}
}

// --- isRetryableError ---

func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		msg  string
		want bool
	}{
		{"rate limit exceeded", true},
		{"403 Forbidden", true},
		{"429 Too Many Requests", true},
		{"500 Internal Server Error", true},
		{"502 Bad Gateway", true},
		{"503 Service Unavailable", true},
		{"404 Not Found", false},
		{"401 Unauthorized", false},
		{"connection refused", false},
	}
	for _, tt := range tests {
		t.Run(tt.msg, func(t *testing.T) {
			got := isRetryableError(errors.New(tt.msg))
			if got != tt.want {
				t.Errorf("isRetryableError(%q) = %v, want %v", tt.msg, got, tt.want)
			}
		})
	}

	t.Run("context.DeadlineExceeded is retryable", func(t *testing.T) {
		if !isRetryableError(context.DeadlineExceeded) {
			t.Error("isRetryableError(context.DeadlineExceeded) = false, want true")
		}
	})
}

// --- appendToCSV ---

func TestAppendToCSV_CreatesHeaderOnNewFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.csv")

	if err := appendToCSV(path, "module", "2026-01-01", "kafka", 42); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	records := readCSV(t, path)
	if len(records) != 2 {
		t.Fatalf("want 2 rows (header + data), got %d", len(records))
	}
	assertRow(t, records[0], "date", "module", "count")
	assertRow(t, records[1], "2026-01-01", "kafka", "42")
}

func TestAppendToCSV_NoHeaderOnExistingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.csv")

	// seed with one record
	if err := appendToCSV(path, "module", "2026-01-01", "kafka", 10); err != nil {
		t.Fatalf("seed error: %v", err)
	}
	if err := appendToCSV(path, "module", "2026-01-01", "redis", 20); err != nil {
		t.Fatalf("append error: %v", err)
	}

	records := readCSV(t, path)
	if len(records) != 3 {
		t.Fatalf("want 3 rows, got %d", len(records))
	}
	assertRow(t, records[0], "date", "module", "count")
	assertRow(t, records[1], "2026-01-01", "kafka", "10")
	assertRow(t, records[2], "2026-01-01", "redis", "20")
}

func TestAppendToCSV_VersionColumn(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.csv")

	if err := appendToCSV(path, "version", "2026-01-01", "v0.35.0", 99); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	records := readCSV(t, path)
	assertRow(t, records[0], "date", "version", "count")
	assertRow(t, records[1], "2026-01-01", "v0.35.0", "99")
}

// --- sortCSV ---

func TestSortCSV_SortsByDateThenColumn(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.csv")

	rows := [][]string{
		{"date", "module", "count"},
		{"2026-03-01", "redis", "10"},
		{"2026-01-01", "kafka", "5"},
		{"2026-01-01", "aerospike", "3"},
		{"2026-02-01", "postgres", "8"},
	}
	writeCSV(t, path, rows)

	if err := sortCSV(path); err != nil {
		t.Fatalf("sortCSV error: %v", err)
	}

	records := readCSV(t, path)
	want := [][]string{
		{"date", "module", "count"},
		{"2026-01-01", "aerospike", "3"},
		{"2026-01-01", "kafka", "5"},
		{"2026-02-01", "postgres", "8"},
		{"2026-03-01", "redis", "10"},
	}
	if len(records) != len(want) {
		t.Fatalf("row count: got %d, want %d", len(records), len(want))
	}
	for i, row := range want {
		assertRow(t, records[i], row...)
	}
}

func TestSortCSV_HeaderOnly(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.csv")
	writeCSV(t, path, [][]string{{"date", "module", "count"}})

	if err := sortCSV(path); err != nil {
		t.Fatalf("unexpected error on header-only file: %v", err)
	}
}

// --- collectWithTimings ---

func TestCollect_AllSucceed(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.csv")

	counts := map[string]int{"kafka": 10, "redis": 20}
	search := func(item string) (int, error) { return counts[item], nil }

	err := collectWithTimings([]string{"kafka", "redis"}, search, path, "module", zeroTimings)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	records := readCSV(t, path)
	// header + 2 data rows
	if len(records) != 3 {
		t.Fatalf("want 3 rows, got %d", len(records))
	}
}

func TestCollect_DeduplicatesItems(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.csv")

	calls := 0
	search := func(_ string) (int, error) { calls++; return 1, nil }

	err := collectWithTimings([]string{"kafka", "kafka", " kafka "}, search, path, "module", zeroTimings)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 1 {
		t.Errorf("search called %d times, want 1 (dedup)", calls)
	}
}

func TestCollect_EmptyItemsAfterTrim(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.csv")
	search := func(_ string) (int, error) { return 0, nil }

	err := collectWithTimings([]string{"", "  ", ""}, search, path, "module", zeroTimings)
	if err == nil {
		t.Fatal("want error for all-empty items, got nil")
	}
}

func TestCollect_NonRetryableErrorAborts(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.csv")
	fatal := errors.New("404 Not Found")
	search := func(_ string) (int, error) { return 0, fatal }

	err := collectWithTimings([]string{"kafka"}, search, path, "module", zeroTimings)
	if err == nil {
		t.Fatal("want error, got nil")
	}
}

func TestCollect_RetryableErrorRetriesAndSucceeds(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.csv")

	attempts := 0
	search := func(_ string) (int, error) {
		attempts++
		if attempts < 2 {
			return 0, errors.New("429 Too Many Requests")
		}
		return 42, nil
	}

	err := collectWithTimings([]string{"kafka"}, search, path, "module", zeroTimings)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if attempts != 2 {
		t.Errorf("want 2 attempts, got %d", attempts)
	}

	records := readCSV(t, path)
	assertRow(t, records[1], records[1][0], "kafka", "42")
}

func TestCollect_ExhaustsRetriesWithoutError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.csv")
	search := func(_ string) (int, error) { return 0, errors.New("503 Service Unavailable") }

	// Should NOT return an error — just logs a warning and continues.
	err := collectWithTimings([]string{"kafka"}, search, path, "module", zeroTimings)
	if err != nil {
		t.Fatalf("exhausted retries should warn, not error — got: %v", err)
	}
}

func TestCollect_WritesCSVSortedByDateAndItem(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.csv")

	// Seed with an older entry so we can verify sorting across runs.
	if err := appendToCSV(path, "module", "2025-12-01", "zookeeper", 5); err != nil {
		t.Fatalf("seed error: %v", err)
	}

	counts := map[string]int{"aerospike": 1, "kafka": 100}
	search := func(item string) (int, error) { return counts[item], nil }

	err := collectWithTimings([]string{"kafka", "aerospike"}, search, path, "module", zeroTimings)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	records := readCSV(t, path)
	// Rows after header should be sorted: 2025-12-01/zookeeper, then today/aerospike, today/kafka
	if len(records) < 4 {
		t.Fatalf("want at least 4 rows, got %d", len(records))
	}
	// First data row is the oldest date
	if records[1][0] != "2025-12-01" {
		t.Errorf("first data row date = %q, want 2025-12-01", records[1][0])
	}
	// Within same date, aerospike < kafka alphabetically
	if records[2][1] != "aerospike" || records[3][1] != "kafka" {
		t.Errorf("within same date: got %q, %q; want aerospike, kafka", records[2][1], records[3][1])
	}
}

// --- clones ---

const twoDayClonesBody = `{"count":11465,"uniques":2939,"clones":[` +
	`{"timestamp":"2026-09-16T00:00:00Z","count":5392,"uniques":1372},` +
	`{"timestamp":"2026-09-17T00:00:00Z","count":6073,"uniques":1567}]}`

const emptyClonesBody = `{"count":0,"uniques":0,"clones":[]}`

func TestIsAuthError(t *testing.T) {
	tests := []struct {
		msg  string
		want bool
	}{
		{"HTTP 403: Must have push access to view repository traffic", true},
		{"Resource not accessible by integration", true},
		{"401 Unauthorized", true},
		{"429 Too Many Requests", false},
		{"503 Service Unavailable", false},
	}
	for _, tt := range tests {
		t.Run(tt.msg, func(t *testing.T) {
			if got := isAuthError(errors.New(tt.msg)); got != tt.want {
				t.Errorf("isAuthError(%q) = %v, want %v", tt.msg, got, tt.want)
			}
		})
	}
}

func TestParseClonesResponse_ParsesDays(t *testing.T) {
	entries, err := parseClonesResponse([]byte(twoDayClonesBody))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []cloneEntry{
		{Date: "2026-09-16", Count: 5392, Uniques: 1372},
		{Date: "2026-09-17", Count: 6073, Uniques: 1567},
	}
	if len(entries) != len(want) {
		t.Fatalf("expected %d entries, got %d", len(want), len(entries))
	}
	for i := range want {
		if entries[i] != want[i] {
			t.Errorf("entry %d: got %+v, want %+v", i, entries[i], want[i])
		}
	}
}

func TestParseClonesResponse_EmptyClones(t *testing.T) {
	entries, err := parseClonesResponse([]byte(emptyClonesBody))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}
}

func TestParseClonesResponse_InvalidJSON(t *testing.T) {
	if _, err := parseClonesResponse([]byte(`{not json`)); err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestParseClonesResponse_InvalidTimestamp(t *testing.T) {
	body := `{"clones":[{"timestamp":"yesterday","count":1,"uniques":1}]}`
	if _, err := parseClonesResponse([]byte(body)); err == nil {
		t.Fatal("expected error for invalid timestamp, got nil")
	}
}

func TestUpsertClonesCSV_CreatesFileWithHeader(t *testing.T) {
	path := filepath.Join(t.TempDir(), "clones.csv")
	entries := []cloneEntry{
		{Date: "2026-09-17", Count: 6073, Uniques: 1567},
		{Date: "2026-09-16", Count: 5392, Uniques: 1372},
	}
	if err := upsertClonesCSV(path, entries); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	rows := readCSV(t, path)
	if len(rows) != 3 {
		t.Fatalf("expected 3 rows, got %d: %v", len(rows), rows)
	}
	assertRow(t, rows[0], "date", "count", "uniques")
	assertRow(t, rows[1], "2026-09-16", "5392", "1372")
	assertRow(t, rows[2], "2026-09-17", "6073", "1567")
}

func TestUpsertClonesCSV_ReplacesExistingDate(t *testing.T) {
	path := filepath.Join(t.TempDir(), "clones.csv")
	writeCSV(t, path, [][]string{
		{"date", "count", "uniques"},
		{"2026-09-16", "100", "10"},
		{"2026-09-17", "5", "1"},
	})
	entries := []cloneEntry{
		{Date: "2026-09-17", Count: 200, Uniques: 20},
		{Date: "2026-09-18", Count: 7, Uniques: 2},
	}
	if err := upsertClonesCSV(path, entries); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	rows := readCSV(t, path)
	if len(rows) != 4 {
		t.Fatalf("expected 4 rows, got %d: %v", len(rows), rows)
	}
	assertRow(t, rows[0], "date", "count", "uniques")
	assertRow(t, rows[1], "2026-09-16", "100", "10")
	assertRow(t, rows[2], "2026-09-17", "200", "20")
	assertRow(t, rows[3], "2026-09-18", "7", "2")
}

func TestUpsertClonesCSV_RejectsMalformedRow(t *testing.T) {
	path := filepath.Join(t.TempDir(), "clones.csv")
	writeCSV(t, path, [][]string{
		{"date", "count", "uniques"},
		{"2026-09-16", "100"},
	})
	err := upsertClonesCSV(path, []cloneEntry{{Date: "2026-09-17", Count: 1, Uniques: 1}})
	if err == nil {
		t.Fatal("expected error for malformed row, got nil")
	}
}

func TestUpsertClonesCSV_RejectsNonNumericCount(t *testing.T) {
	path := filepath.Join(t.TempDir(), "clones.csv")
	writeCSV(t, path, [][]string{
		{"date", "count", "uniques"},
		{"2026-09-16", "many", "10"},
	})
	err := upsertClonesCSV(path, []cloneEntry{{Date: "2026-09-17", Count: 1, Uniques: 1}})
	if err == nil {
		t.Fatal("expected error for non-numeric count, got nil")
	}
}

func TestCollectClones_Success(t *testing.T) {
	path := filepath.Join(t.TempDir(), "clones.csv")
	fetch := func() ([]byte, error) { return []byte(twoDayClonesBody), nil }

	if err := collectClonesWithTimings(fetch, path, zeroTimings); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	rows := readCSV(t, path)
	if len(rows) != 3 {
		t.Fatalf("expected 3 rows, got %d: %v", len(rows), rows)
	}
	assertRow(t, rows[1], "2026-09-16", "5392", "1372")
}

func TestCollectClones_RetryableErrorRetriesAndSucceeds(t *testing.T) {
	path := filepath.Join(t.TempDir(), "clones.csv")
	calls := 0
	fetch := func() ([]byte, error) {
		calls++
		if calls == 1 {
			return nil, errors.New("429 Too Many Requests")
		}
		return []byte(twoDayClonesBody), nil
	}

	if err := collectClonesWithTimings(fetch, path, zeroTimings); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 2 {
		t.Errorf("expected 2 fetch calls, got %d", calls)
	}
}

func TestCollectClones_NonRetryableErrorAborts(t *testing.T) {
	path := filepath.Join(t.TempDir(), "clones.csv")
	calls := 0
	fetch := func() ([]byte, error) {
		calls++
		return nil, errors.New("404 Not Found")
	}

	if err := collectClonesWithTimings(fetch, path, zeroTimings); err == nil {
		t.Fatal("expected error, got nil")
	}
	if calls != 1 {
		t.Errorf("expected 1 fetch call, got %d", calls)
	}
}

func TestCollectClones_AuthErrorAbortsWithoutRetry(t *testing.T) {
	path := filepath.Join(t.TempDir(), "clones.csv")
	calls := 0
	fetch := func() ([]byte, error) {
		calls++
		return nil, errors.New("gh api failed: HTTP 403: Must have push access to view repository traffic")
	}

	if err := collectClonesWithTimings(fetch, path, zeroTimings); err == nil {
		t.Fatal("expected error, got nil")
	}
	if calls != 1 {
		t.Errorf("expected 1 fetch call, got %d", calls)
	}
}

func TestCollectClones_ExhaustsRetriesReturnsError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "clones.csv")
	calls := 0
	fetch := func() ([]byte, error) {
		calls++
		return nil, errors.New("503 Service Unavailable")
	}

	if err := collectClonesWithTimings(fetch, path, zeroTimings); err == nil {
		t.Fatal("expected error after exhausting retries, got nil")
	}
	if calls != zeroTimings.maxPasses {
		t.Errorf("expected %d fetch calls, got %d", zeroTimings.maxPasses, calls)
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("expected CSV not to exist, Stat err = %v", err)
	}
}

func TestCollectClones_EmptyResponseWritesNothing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "clones.csv")
	fetch := func() ([]byte, error) { return []byte(emptyClonesBody), nil }

	if err := collectClonesWithTimings(fetch, path, zeroTimings); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("expected CSV not to exist, Stat err = %v", err)
	}
}

// --- helpers ---

func readCSV(t *testing.T, path string) [][]string {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open csv: %v", err)
	}
	defer f.Close()
	records, err := csv.NewReader(f).ReadAll()
	if err != nil {
		t.Fatalf("read csv: %v", err)
	}
	return records
}

func writeCSV(t *testing.T, path string, rows [][]string) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create csv: %v", err)
	}
	defer f.Close()
	w := csv.NewWriter(f)
	if err := w.WriteAll(rows); err != nil {
		t.Fatalf("write csv: %v", err)
	}
}

func assertRow(t *testing.T, got []string, want ...string) {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf("row length: got %d, want %d (%v)", len(got), len(want), want)
		return
	}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("col %d: got %q, want %q (row: %v)", i, got[i], w, got)
		}
	}
}

// Ensure the query strings used in main() are not accidentally broken.
func TestQueryStrings(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "version",
			input: "v0.35.0",
			want:  `"testcontainers/testcontainers-go v0.35.0" filename:go.mod -is:fork -org:testcontainers`,
		},
		{
			name:  "module",
			input: "kafka",
			want:  `"testcontainers/testcontainers-go/modules/kafka" filename:go.mod -is:fork -org:testcontainers`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got string
			switch tt.name {
			case "version":
				got = fmt.Sprintf(`"testcontainers/testcontainers-go %s" filename:go.mod -is:fork -org:testcontainers`, tt.input)
			case "module":
				got = fmt.Sprintf(`"testcontainers/testcontainers-go/modules/%s" filename:go.mod -is:fork -org:testcontainers`, tt.input)
			}
			if got != tt.want {
				t.Errorf("query = %q, want %q", got, tt.want)
			}
		})
	}
}
