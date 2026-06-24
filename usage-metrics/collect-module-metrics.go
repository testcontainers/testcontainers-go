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

type moduleSearchResponse struct {
	TotalCount int `json:"total_count"`
}

type moduleMetric struct {
	Date   string
	Module string
	Count  int
}

type moduleArrayFlags []string

func (a *moduleArrayFlags) String() string {
	return strings.Join(*a, ",")
}

func (a *moduleArrayFlags) Set(value string) error {
	*a = append(*a, value)
	return nil
}

func main() {
	var modules moduleArrayFlags
	csvPath := flag.String("csv", filepath.Join("..", "docs", "modules-usage-metrics.csv"), "Path to CSV file")
	flag.Var(&modules, "module", "Module to query (can be specified multiple times)")
	flag.Parse()

	if len(modules) == 0 {
		log.Fatal("At least one module is required. Use -module flag (can be repeated)")
	}

	if err := collectModuleMetrics(modules, *csvPath); err != nil {
		log.Fatalf("Failed to collect module metrics: %v", err)
	}
}

func collectModuleMetrics(modules []string, csvPath string) error {
	date := time.Now().Format("2006-01-02")
	metrics := make(map[string]moduleMetric)

	// Build a unique, non-empty list of modules to query
	pending := make([]string, 0, len(modules))
	seen := make(map[string]struct{}, len(modules))
	for _, module := range modules {
		module = strings.TrimSpace(module)
		if module == "" {
			continue
		}
		if _, ok := seen[module]; ok {
			continue
		}
		seen[module] = struct{}{}
		pending = append(pending, module)
	}
	if len(pending) == 0 {
		return errors.New("at least one non-empty module is required")
	}

	const (
		maxPasses         = 5
		interRequestWait  = 7 * time.Second   // 10 requests per 60 seconds = 6 seconds minimum
		rateLimitCooldown = 65 * time.Second  // cool down after a rate-limit hit within a pass
		passCooldown      = 120 * time.Second // wait for rate limit window to fully reset between passes
	)

	for pass := 0; pass < maxPasses && len(pending) > 0; pass++ {
		if pass > 0 {
			log.Printf("Pass %d: waiting %v for rate limit window to reset before retrying %d failed module(s)...",
				pass+1, passCooldown, len(pending))
			time.Sleep(passCooldown)
		} else {
			log.Printf("Pass 1: querying %d module(s)...", len(pending))
		}

		var failed []string
		queriesMade := 0
		rateLimitHit := false
		for _, module := range pending {
			// Add delay before querying to avoid rate limiting.
			// Use a longer delay if we recently hit a rate limit within this pass.
			if queriesMade > 0 {
				wait := interRequestWait
				if rateLimitHit {
					wait = rateLimitCooldown
					rateLimitHit = false
				}
				log.Printf("Waiting %v before querying next module...", wait)
				time.Sleep(wait)
			}

			count, err := queryGitHubModuleUsage(module)
			queriesMade++
			if err != nil {
				log.Printf("Pass %d: failed to query module %s: %v", pass+1, module, err)
				if isModuleRetryableError(err) {
					rateLimitHit = isModuleRateLimitError(err)
					failed = append(failed, module)
					continue
				}
				return fmt.Errorf("query %s: %w", module, err)
			}

			metrics[module] = moduleMetric{
				Date:   date,
				Module: module,
				Count:  count,
			}
			fmt.Printf("Successfully queried: %s has %d usages on %s\n", module, count, date)
		}

		pending = failed
		if len(pending) == 0 {
			log.Printf("All modules queried successfully after %d pass(es).", pass+1)
		}
	}

	if len(pending) > 0 {
		log.Printf("Warning: %d module(s) still failed after %d passes: %s", len(pending), maxPasses, strings.Join(pending, ", "))
	}

	// Append new metrics to CSV
	for _, metric := range metrics {
		if err := appendModuleToCSV(csvPath, metric); err != nil {
			log.Printf("Warning: Failed to write metric for %s: %v", metric.Module, err)
			continue
		}
		fmt.Printf("Successfully recorded: %s has %d usages on %s\n", metric.Module, metric.Count, metric.Date)
	}

	// Sort the entire CSV so rows are ordered by (date, module) regardless
	// of the order they were appended across multiple runs.
	if err := sortModuleCSV(csvPath); err != nil {
		return fmt.Errorf("sort csv: %w", err)
	}

	return nil
}

// isModuleRateLimitError returns true for rate-limit specific errors (403/429).
func isModuleRateLimitError(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "rate limit") ||
		strings.Contains(msg, "403") ||
		strings.Contains(msg, "429")
}

// isModuleRetryableError returns true for rate-limit and transient HTTP errors
// that are worth retrying in a subsequent pass.
func isModuleRetryableError(err error) bool {
	return isModuleRateLimitError(err) ||
		strings.Contains(err.Error(), "500") ||
		strings.Contains(err.Error(), "502") ||
		strings.Contains(err.Error(), "503")
}

func queryGitHubModuleUsage(module string) (int, error) {
	query := fmt.Sprintf(`"testcontainers/testcontainers-go/modules/%s" filename:go.mod -is:fork -org:testcontainers`, module)

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

	var resp moduleSearchResponse
	if err := json.Unmarshal(output, &resp); err != nil {
		return 0, fmt.Errorf("unmarshal: %w", err)
	}

	return resp.TotalCount, nil
}

func sortModuleCSV(csvPath string) error {
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
		return data[i][1] < data[j][1] // module ascending
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

func appendModuleToCSV(csvPath string, metric moduleMetric) error {
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
		if err := writer.Write([]string{"date", "module", "count"}); err != nil {
			return fmt.Errorf("write header: %w", err)
		}
	}

	record := []string{metric.Date, metric.Module, strconv.Itoa(metric.Count)}
	if err := writer.Write(record); err != nil {
		return fmt.Errorf("write record: %w", err)
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return fmt.Errorf("flush csv: %w", err)
	}

	return nil
}
