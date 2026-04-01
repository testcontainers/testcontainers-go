package main

import (
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

type searchResponse struct {
	TotalCount int `json:"total_count"`
}

type usageMetric struct {
	Date    string
	Version string
	Count   int
}

type arrayFlags []string

func (a *arrayFlags) String() string {
	return strings.Join(*a, ",")
}

func (a *arrayFlags) Set(value string) error {
	*a = append(*a, value)
	return nil
}

func main() {
	var versions arrayFlags
	csvPath := flag.String("csv", filepath.Join("..", "docs", "usage-metrics.csv"), "Path to CSV file")
	flag.Var(&versions, "version", "Version to query (can be specified multiple times)")
	flag.Parse()

	if len(versions) == 0 {
		log.Fatal("At least one version is required. Use -version flag (can be repeated)")
	}

	if err := collectMetrics(versions, *csvPath); err != nil {
		log.Fatalf("Failed to collect metrics: %v", err)
	}
}

func collectMetrics(versions []string, csvPath string) error {
	date := time.Now().Format("2006-01-02")
	metrics := make(map[string]usageMetric)

	// Build a unique, non-empty list of versions to query
	pending := make([]string, 0, len(versions))
	seen := make(map[string]struct{}, len(versions))
	for _, version := range versions {
		version = strings.TrimSpace(version)
		if version == "" {
			continue
		}
		if _, ok := seen[version]; ok {
			continue
		}
		seen[version] = struct{}{}
		pending = append(pending, version)
	}
	if len(pending) == 0 {
		return errors.New("at least one non-empty version is required")
	}

	const (
		maxPasses         = 5
		interRequestWait  = 7 * time.Second   // 10 requests per 60 seconds = 6 seconds minimum
		rateLimitCooldown = 65 * time.Second  // cool down after a rate-limit hit within a pass
		passCooldown      = 120 * time.Second // wait for rate limit window to fully reset between passes
	)

	for pass := 0; pass < maxPasses && len(pending) > 0; pass++ {
		if pass > 0 {
			log.Printf("Pass %d: waiting %v for rate limit window to reset before retrying %d failed version(s)...",
				pass+1, passCooldown, len(pending))
			time.Sleep(passCooldown)
		} else {
			log.Printf("Pass 1: querying %d version(s)...", len(pending))
		}

		var failed []string
		queriesMade := 0
		rateLimitHit := false
		for _, version := range pending {
			// Add delay before querying to avoid rate limiting.
			// Use a longer delay if we recently hit a rate limit within this pass.
			if queriesMade > 0 {
				wait := interRequestWait
				if rateLimitHit {
					wait = rateLimitCooldown
					rateLimitHit = false
				}
				log.Printf("Waiting %v before querying next version...", wait)
				time.Sleep(wait)
			}

			count, err := queryGitHubUsage(version)
			queriesMade++
			if err != nil {
				log.Printf("Pass %d: failed to query version %s: %v", pass+1, version, err)
				if isRetryableError(err) {
					rateLimitHit = isRateLimitError(err)
					failed = append(failed, version)
					continue
				}
				return fmt.Errorf("query %s: %w", version, err)
			}

			metrics[version] = usageMetric{
				Date:    date,
				Version: version,
				Count:   count,
			}
			fmt.Printf("Successfully queried: %s has %d usages on %s\n", version, count, date)
		}

		pending = failed
		if len(pending) == 0 {
			log.Printf("All versions queried successfully after %d pass(es).", pass+1)
		}
	}

	if len(pending) > 0 {
		log.Printf("Warning: %d version(s) still failed after %d passes: %s", len(pending), maxPasses, strings.Join(pending, ", "))
	}

	// Append new metrics to CSV
	for _, metric := range metrics {
		if err := appendToCSV(csvPath, metric); err != nil {
			log.Printf("Warning: Failed to write metric for %s: %v", metric.Version, err)
			continue
		}
		fmt.Printf("Successfully recorded: %s has %d usages on %s\n", metric.Version, metric.Count, metric.Date)
	}

	// Sort the entire CSV so rows are ordered by (date, version) regardless
	// of the order they were appended across multiple runs.
	if err := sortCSV(csvPath); err != nil {
		return fmt.Errorf("sort csv: %w", err)
	}

	return nil
}

// isRateLimitError returns true for rate-limit specific errors (403/429).
func isRateLimitError(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "rate limit") ||
		strings.Contains(msg, "403") ||
		strings.Contains(msg, "429")
}

// isRetryableError returns true for rate-limit and transient HTTP errors
// that are worth retrying in a subsequent pass.
func isRetryableError(err error) bool {
	return isRateLimitError(err) ||
		strings.Contains(err.Error(), "500") ||
		strings.Contains(err.Error(), "502") ||
		strings.Contains(err.Error(), "503")
}

func queryGitHubUsage(version string) (int, error) {
	query := fmt.Sprintf(`"testcontainers/testcontainers-go %s" filename:go.mod -is:fork -org:testcontainers`, version)

	params := url.Values{}
	params.Add("q", query)
	endpoint := "/search/code?" + params.Encode()

	output, err := exec.Command("gh", "api",
		"-H", "Accept: application/vnd.github+json",
		"-H", "X-GitHub-Api-Version: 2022-11-28",
		endpoint,
	).Output()
	if err != nil {
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
		return nil // nothing to sort (header only or empty)
	}

	header := records[0]
	data := records[1:]

	sort.SliceStable(data, func(i, j int) bool {
		if data[i][0] != data[j][0] {
			return data[i][0] < data[j][0] // date ascending
		}
		return data[i][1] < data[j][1] // version ascending
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

func appendToCSV(csvPath string, metric usageMetric) error {
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
		if err := writer.Write([]string{"date", "version", "count"}); err != nil {
			return fmt.Errorf("write header: %w", err)
		}
	}

	record := []string{metric.Date, metric.Version, strconv.Itoa(metric.Count)}
	if err := writer.Write(record); err != nil {
		return fmt.Errorf("write record: %w", err)
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return fmt.Errorf("flush csv: %w", err)
	}

	return nil
}
