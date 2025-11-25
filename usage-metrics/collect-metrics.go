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
	csvPath := flag.String("csv", "../../docs/usage-metrics.csv", "Path to CSV file")
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
	metrics := make([]usageMetric, 0, len(versions))

	// Query all versions sequentially
	for _, version := range versions {
		version = strings.TrimSpace(version)
		if version == "" {
			continue
		}

		// Add delay BEFORE querying to avoid rate limiting
		if len(metrics) > 0 {
			log.Printf("Waiting 7 seconds before querying next version...")
			time.Sleep(7 * time.Second) // 10 requests per 60 seconds = 6 seconds minimum
		}

		count, err := queryGitHubUsageWithRetry(version)
		if err != nil {
			log.Printf("Warning: Failed to query version %s after retries: %v", version, err)
			continue
		}

		metric := usageMetric{
			Date:    date,
			Version: version,
			Count:   count,
		}

		metrics = append(metrics, metric)
		fmt.Printf("Successfully queried: %s has %d usages on %s\n", version, count, metric.Date)
	}

	// Sort metrics by version
	sort.Slice(metrics, func(i, j int) bool {
		return metrics[i].Version < metrics[j].Version
	})

	// Write all metrics to CSV
	for _, metric := range metrics {
		if err := appendToCSV(csvPath, metric); err != nil {
			log.Printf("Warning: Failed to write metric for %s: %v", metric.Version, err)
			continue
		}
		fmt.Printf("Successfully recorded: %s has %d usages on %s\n", metric.Version, metric.Count, metric.Date)
	}

	return nil
}

func queryGitHubUsageWithRetry(version string) (int, error) {
	var lastErr error
	// Backoff intervals: wait longer for rate limit to reset (rolling window)
	backoffIntervals := []time.Duration{
		60 * time.Second, // Wait for rolling window
		60 * time.Second,
		60 * time.Second,
	}

	// maxRetries includes the initial attempt plus one retry per backoff interval
	maxRetries := len(backoffIntervals) + 1

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			// Use predefined backoff intervals
			waitTime := backoffIntervals[attempt-1]
			log.Printf("Retrying version %s in %v (attempt %d/%d)", version, waitTime, attempt+1, maxRetries)
			time.Sleep(waitTime)
		}

		count, err := queryGitHubUsage(version)
		if err == nil {
			return count, nil
		}

		lastErr = err

		// Check if it's a rate limit error
		if strings.Contains(err.Error(), "rate limit") ||
			strings.Contains(err.Error(), "403") ||
			strings.Contains(err.Error(), "429") {
			log.Printf("Rate limit hit for version %s, will retry with backoff", version)
			continue
		}

		// For non-rate-limit errors, retry but with shorter backoff
		log.Printf("Error querying version %s: %v", version, err)
	}

	return 0, fmt.Errorf("max retries reached: %w", lastErr)
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
