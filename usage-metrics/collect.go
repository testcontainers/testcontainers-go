package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"
)

type arrayFlags []string

func (a *arrayFlags) String() string { return strings.Join(*a, ",") }

func (a *arrayFlags) Set(value string) error {
	*a = append(*a, value)
	return nil
}

type searchResponse struct {
	TotalCount int `json:"total_count"`
}

// retryTimings groups all sleep durations used by collect so tests can zero them out.
type retryTimings struct {
	interRequestWait  time.Duration
	rateLimitCooldown time.Duration
	passCooldown      time.Duration
	maxPasses         int
}

var productionTimings = retryTimings{
	maxPasses:         5,
	interRequestWait:  7 * time.Second,
	rateLimitCooldown: 65 * time.Second,
	passCooldown:      120 * time.Second,
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: collect <versions|modules|clones> [flags]")
		os.Exit(1)
	}

	subcommand := os.Args[1]
	args := os.Args[2:]

	switch subcommand {
	case "versions":
		fs := flag.NewFlagSet("versions", flag.ExitOnError)
		var items arrayFlags
		csvPath := fs.String("csv", filepath.Join("..", "docs", "usage-metrics", "core.csv"), "Path to CSV file")
		fs.Var(&items, "version", "Version to query (can be specified multiple times)")
		if err := fs.Parse(args); err != nil {
			log.Fatalf("Failed to parse flags: %v", err)
		}

		if len(items) == 0 {
			log.Fatal("At least one version is required. Use -version flag (can be repeated)")
		}
		search := func(v string) (int, error) {
			q := fmt.Sprintf(`"testcontainers/testcontainers-go %s" filename:go.mod -is:fork -org:testcontainers`, v)
			return runGHSearch(q)
		}
		if err := collect(items, search, *csvPath, "version"); err != nil {
			log.Fatalf("Failed to collect version metrics: %v", err)
		}

	case "modules":
		fs := flag.NewFlagSet("modules", flag.ExitOnError)
		var items arrayFlags
		csvPath := fs.String("csv", filepath.Join("..", "docs", "usage-metrics", "modules.csv"), "Path to CSV file")
		fs.Var(&items, "module", "Module to query (can be specified multiple times)")
		if err := fs.Parse(args); err != nil {
			log.Fatalf("Failed to parse flags: %v", err)
		}

		if len(items) == 0 {
			log.Fatal("At least one module is required. Use -module flag (can be repeated)")
		}
		search := func(m string) (int, error) {
			q := fmt.Sprintf(`"testcontainers/testcontainers-go/modules/%s" filename:go.mod -is:fork -org:testcontainers`, m)
			return runGHSearch(q)
		}
		if err := collect(items, search, *csvPath, "module"); err != nil {
			log.Fatalf("Failed to collect module metrics: %v", err)
		}

	case "clones":
		fs := flag.NewFlagSet("clones", flag.ExitOnError)
		csvPath := fs.String("csv", filepath.Join("..", "docs", "usage-metrics", "clones.csv"), "Path to CSV file")
		repo := fs.String("repo", "testcontainers/testcontainers-go", "GitHub repository in owner/name form")
		until := fs.String("until", "", "Only record days up to and including this YYYY-MM-DD date (default: every day returned)")
		if err := fs.Parse(args); err != nil {
			log.Fatalf("Failed to parse flags: %v", err)
		}

		if err := collectClones(*repo, *csvPath, *until); err != nil {
			log.Fatalf("Failed to collect clone metrics: %v", err)
		}

	default:
		fmt.Fprintf(os.Stderr, "Unknown subcommand %q. Use 'versions', 'modules' or 'clones'.\n", subcommand)
		os.Exit(1)
	}
}

// collect runs with production timings.
func collect(items []string, search func(string) (int, error), csvPath, column string) error {
	return collectWithTimings(items, search, csvPath, column, productionTimings)
}

func collectWithTimings(items []string, search func(string) (int, error), csvPath, column string, t retryTimings) error {
	date := time.Now().Format("2006-01-02")

	// Deduplicate and sanitise
	pending := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		pending = append(pending, item)
	}
	if len(pending) == 0 {
		return fmt.Errorf("at least one non-empty %s is required", column)
	}

	results := make(map[string]int, len(pending))

	for pass := 0; pass < t.maxPasses && len(pending) > 0; pass++ {
		if pass > 0 {
			log.Printf("Pass %d: waiting %v for rate limit reset before retrying %d %s(s)...",
				pass+1, t.passCooldown, len(pending), column)
			time.Sleep(t.passCooldown)
		} else {
			log.Printf("Pass 1: querying %d %s(s)...", len(pending), column)
		}

		var failed []string
		queriesMade := 0
		rateLimitHit := false
		for _, item := range pending {
			if queriesMade > 0 {
				wait := t.interRequestWait
				if rateLimitHit {
					wait = t.rateLimitCooldown
					rateLimitHit = false
				}
				log.Printf("Waiting %v before querying next %s...", wait, column)
				time.Sleep(wait)
			}

			count, err := search(item)
			queriesMade++
			if err != nil {
				log.Printf("Pass %d: failed to query %s %s: %v", pass+1, column, item, err)
				if isRetryableError(err) {
					rateLimitHit = isRateLimitError(err)
					failed = append(failed, item)
					continue
				}
				return fmt.Errorf("query %s: %w", item, err)
			}

			results[item] = count
			fmt.Printf("Successfully queried: %s=%s has %d usages on %s\n", column, item, count, date)
		}

		pending = failed
		if len(pending) == 0 {
			log.Printf("All %s(s) queried successfully after %d pass(es).", column, pass+1)
		}
	}

	if len(pending) > 0 {
		log.Printf("Warning: %d %s(s) still failed after %d passes: %s",
			len(pending), column, t.maxPasses, strings.Join(pending, ", "))
	}

	if len(results) == 0 {
		return nil
	}

	for item, count := range results {
		if err := appendToCSV(csvPath, column, date, item, count); err != nil {
			return fmt.Errorf("write %s=%s: %w", column, item, err)
		}
		fmt.Printf("Successfully recorded: %s=%s has %d usages on %s\n", column, item, count, date)
	}

	if err := sortCSV(csvPath); err != nil {
		return fmt.Errorf("sort csv: %w", err)
	}

	return nil
}

func isRateLimitError(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "rate limit") ||
		strings.Contains(msg, "403") ||
		strings.Contains(msg, "429")
}

func isRetryableError(err error) bool {
	return isRateLimitError(err) ||
		errors.Is(err, context.DeadlineExceeded) ||
		strings.Contains(err.Error(), "500") ||
		strings.Contains(err.Error(), "502") ||
		strings.Contains(err.Error(), "503")
}

// isAuthError reports whether err means the token cannot access the endpoint at all.
// The traffic endpoints answer 403 for missing permissions far more often than for
// rate limiting, so only a 403 that mentions a rate limit is worth retrying.
func isAuthError(err error) bool {
	msg := err.Error()
	if strings.Contains(msg, "Must have push access") ||
		strings.Contains(msg, "Resource not accessible") ||
		strings.Contains(msg, "401") {
		return true
	}
	return strings.Contains(msg, "403") && !strings.Contains(strings.ToLower(msg), "rate limit")
}

// cloneEntry is one day of clone traffic as returned by the GitHub traffic API.
type cloneEntry struct {
	Date    string // YYYY-MM-DD, UTC
	Count   int
	Uniques int
}

// clonesCSVHeader is the only header upsertClonesCSV accepts.
var clonesCSVHeader = []string{"date", "count", "uniques"}

type clonesResponse struct {
	Clones []struct {
		Timestamp string `json:"timestamp"`
		Count     int    `json:"count"`
		Uniques   int    `json:"uniques"`
	} `json:"clones"`
}

// parseClonesResponse converts the JSON body of GET /repos/{owner}/{repo}/traffic/clones?per=day
// into one cloneEntry per day. Timestamps are RFC 3339; the date is their UTC YYYY-MM-DD.
func parseClonesResponse(body []byte) ([]cloneEntry, error) {
	var resp clonesResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	entries := make([]cloneEntry, 0, len(resp.Clones))
	for _, c := range resp.Clones {
		ts, err := time.Parse(time.RFC3339, c.Timestamp)
		if err != nil {
			return nil, fmt.Errorf("parse timestamp %q: %w", c.Timestamp, err)
		}
		entries = append(entries, cloneEntry{
			Date:    ts.UTC().Format("2006-01-02"),
			Count:   c.Count,
			Uniques: c.Uniques,
		})
	}

	return entries, nil
}

// upsertClonesCSV merges entries into the CSV at csvPath, replacing any existing row with the
// same date, and rewrites the file sorted by date ascending with the header date,count,uniques.
// The file is created if it does not exist.
func upsertClonesCSV(csvPath string, entries []cloneEntry) error {
	absPath, err := filepath.Abs(csvPath)
	if err != nil {
		return fmt.Errorf("resolve path: %w", err)
	}

	byDate := make(map[string]cloneEntry)

	data, err := os.ReadFile(absPath)
	switch {
	case err == nil:
		records, err := csv.NewReader(bytes.NewReader(data)).ReadAll()
		if err != nil {
			return fmt.Errorf("read csv: %w", err)
		}
		if len(records) == 0 || !slices.Equal(records[0], clonesCSVHeader) {
			return fmt.Errorf("invalid csv header: expected %v", clonesCSVHeader)
		}
		for i, row := range records[1:] {
			line := i + 2 // 1-based, after the header
			count, err := strconv.Atoi(row[1])
			if err != nil {
				return fmt.Errorf("invalid count on line %d: %w", line, err)
			}
			uniques, err := strconv.Atoi(row[2])
			if err != nil {
				return fmt.Errorf("invalid uniques on line %d: %w", line, err)
			}
			byDate[row[0]] = cloneEntry{Date: row[0], Count: count, Uniques: uniques}
		}
	case errors.Is(err, os.ErrNotExist):
		// first run: start from an empty set
	default:
		return fmt.Errorf("read file: %w", err)
	}

	for _, e := range entries {
		byDate[e.Date] = e
	}

	dates := make([]string, 0, len(byDate))
	for d := range byDate {
		dates = append(dates, d)
	}
	sort.Strings(dates)

	tmpPath := absPath + ".tmp"
	if err := writeClonesCSV(tmpPath, dates, byDate); err != nil {
		os.Remove(tmpPath)
		return err
	}
	if err := os.Rename(tmpPath, absPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("replace file: %w", err)
	}

	return nil
}

// writeClonesCSV writes the header and the given dates' entries to path, failing on any
// write or close error so a partial file is never mistaken for a complete one.
func writeClonesCSV(path string, dates []string, byDate map[string]cloneEntry) error {
	out, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}

	writer := csv.NewWriter(out)
	if err := writer.Write(clonesCSVHeader); err != nil {
		out.Close()
		return fmt.Errorf("write header: %w", err)
	}
	for _, d := range dates {
		e := byDate[d]
		if err := writer.Write([]string{e.Date, strconv.Itoa(e.Count), strconv.Itoa(e.Uniques)}); err != nil {
			out.Close()
			return fmt.Errorf("write record: %w", err)
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		out.Close()
		return fmt.Errorf("flush: %w", err)
	}
	if err := out.Close(); err != nil {
		return fmt.Errorf("close file: %w", err)
	}

	return nil
}

// fetchWithRetry calls fetch up to t.maxPasses times, waiting t.passCooldown between attempts.
// Only retryable, non-authentication errors are retried.
func fetchWithRetry(fetch func() ([]byte, error), t retryTimings) ([]byte, error) {
	if t.maxPasses < 1 {
		return nil, errors.New("maxPasses must be at least 1")
	}

	var lastErr error
	for pass := 0; pass < t.maxPasses; pass++ {
		if pass > 0 {
			log.Printf("Pass %d: waiting %v before retrying clones query...", pass+1, t.passCooldown)
			time.Sleep(t.passCooldown)
		}

		body, err := fetch()
		if err == nil {
			return body, nil
		}
		if isAuthError(err) || !isRetryableError(err) {
			return nil, fmt.Errorf("query clones: %w", err)
		}
		log.Printf("Pass %d: failed to query clones: %v", pass+1, err)
		lastErr = err
	}

	return nil, fmt.Errorf("query clones after %d passes: %w", t.maxPasses, lastErr)
}

// collectClonesWithTimings fetches the clone traffic and upserts it into csvPath. When until
// is non-empty (YYYY-MM-DD), only days up to and including that date are recorded; the
// workflow uses it to skip the in-progress UTC day. Unlike collectWithTimings, exhausting
// the retries is an error: a skipped day of clone traffic cannot be recovered later.
func collectClonesWithTimings(fetch func() ([]byte, error), csvPath, until string, t retryTimings) error {
	if until != "" {
		if _, err := time.Parse("2006-01-02", until); err != nil {
			return fmt.Errorf("invalid -until date %q: expected YYYY-MM-DD", until)
		}
	}

	body, err := fetchWithRetry(fetch, t)
	if err != nil {
		return err
	}

	entries, err := parseClonesResponse(body)
	if err != nil {
		return err
	}
	if until != "" {
		kept := entries[:0]
		for _, e := range entries {
			if e.Date <= until {
				kept = append(kept, e)
			}
		}
		entries = kept
	}
	if len(entries) == 0 {
		log.Printf("Warning: no clone entries to record; nothing written")
		return nil
	}

	if err := upsertClonesCSV(csvPath, entries); err != nil {
		return fmt.Errorf("write clones csv: %w", err)
	}

	for _, e := range entries {
		fmt.Printf("Successfully recorded: %s has %d clones (%d unique)\n", e.Date, e.Count, e.Uniques)
	}

	return nil
}

// clonesEndpoint is the REST path for the daily clone traffic of an owner/name repository.
func clonesEndpoint(repo string) string {
	return fmt.Sprintf("/repos/%s/traffic/clones?per=day", repo)
}

// collectClones runs with production timings against the given owner/name repository.
func collectClones(repo, csvPath, until string) error {
	fetch := func() ([]byte, error) {
		return runGHAPI(clonesEndpoint(repo))
	}
	return collectClonesWithTimings(fetch, csvPath, until, productionTimings)
}

func runGHSearch(query string) (int, error) {
	params := url.Values{}
	params.Add("q", query)

	output, err := runGHAPI("/search/code?" + params.Encode())
	if err != nil {
		return 0, err
	}

	var resp searchResponse
	if err := json.Unmarshal(output, &resp); err != nil {
		return 0, fmt.Errorf("unmarshal: %w", err)
	}

	return resp.TotalCount, nil
}

// runGHAPI calls a GitHub REST endpoint through the gh CLI and returns the raw response body.
// gh authenticates with GH_TOKEN (or GITHUB_TOKEN) from the environment.
func runGHAPI(endpoint string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	output, err := exec.CommandContext(ctx, "gh", "api",
		"-H", "Accept: application/vnd.github+json",
		"-H", "X-GitHub-Api-Version: 2022-11-28",
		endpoint,
	).Output()
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return nil, fmt.Errorf("gh api timeout after 30s: %w", ctx.Err())
		}
		exitErr := &exec.ExitError{}
		if errors.As(err, &exitErr) {
			return nil, fmt.Errorf("gh api failed: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("gh api: %w", err)
	}

	return output, nil
}

func sortCSV(csvPath string) error {
	absPath, err := filepath.Abs(csvPath)
	if err != nil {
		return fmt.Errorf("resolve path: %w", err)
	}

	file, err := os.Open(absPath)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	file.Close()
	if err != nil {
		return fmt.Errorf("read csv: %w", err)
	}

	if len(records) <= 1 {
		return nil
	}

	header := records[0]
	data := records[1:]

	for i, row := range data {
		if len(row) < 2 {
			return fmt.Errorf("invalid csv row %d: expected at least 2 columns, got %d", i+2, len(row))
		}
	}

	sort.SliceStable(data, func(i, j int) bool {
		if data[i][0] != data[j][0] {
			return data[i][0] < data[j][0]
		}
		return data[i][1] < data[j][1]
	})

	out, err := os.Create(absPath)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer out.Close()

	writer := csv.NewWriter(out)
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("write header: %w", err)
	}
	if err := writer.WriteAll(data); err != nil {
		return fmt.Errorf("write records: %w", err)
	}

	return nil
}

func appendToCSV(csvPath, column, date, item string, count int) error {
	absPath, err := filepath.Abs(csvPath)
	if err != nil {
		return fmt.Errorf("resolve path: %w", err)
	}

	_, err = os.Stat(absPath)
	fileExists := !os.IsNotExist(err)

	file, err := os.OpenFile(absPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)

	if !fileExists {
		if err := writer.Write([]string{"date", column, "count"}); err != nil {
			return fmt.Errorf("write header: %w", err)
		}
	}

	if err := writer.Write([]string{date, item, strconv.Itoa(count)}); err != nil {
		return fmt.Errorf("write record: %w", err)
	}

	writer.Flush()
	return writer.Error()
}
