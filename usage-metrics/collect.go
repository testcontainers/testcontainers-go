package main

import (
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
		fmt.Fprintln(os.Stderr, "Usage: collect <versions|modules> [flags]")
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

	default:
		fmt.Fprintf(os.Stderr, "Unknown subcommand %q. Use 'versions' or 'modules'.\n", subcommand)
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

func runGHSearch(query string) (int, error) {
	params := url.Values{}
	params.Add("q", query)
	endpoint := "/search/code?" + params.Encode()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	output, err := exec.CommandContext(ctx, "gh", "api",
		"-H", "Accept: application/vnd.github+json",
		"-H", "X-GitHub-Api-Version: 2022-11-28",
		endpoint,
	).Output()
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return 0, fmt.Errorf("gh api timeout after 30s: %w", ctx.Err())
		}
		exitErr := &exec.ExitError{}
		if errors.As(err, &exitErr) {
			return 0, fmt.Errorf("gh api failed: %s", string(exitErr.Stderr))
		}
		return 0, fmt.Errorf("gh api: %w", err)
	}

	var resp searchResponse
	if err := json.Unmarshal(output, &resp); err != nil {
		return 0, fmt.Errorf("unmarshal: %w", err)
	}

	return resp.TotalCount, nil
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
