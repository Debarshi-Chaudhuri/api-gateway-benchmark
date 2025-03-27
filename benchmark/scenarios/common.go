package scenarios

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

// BenchmarkResult represents the results of a benchmark
type BenchmarkResult struct {
	Gateway         string
	Scenario        string
	Protocol        string
	RequestCount    int
	SuccessCount    int
	FailureCount    int
	TotalDuration   time.Duration
	AvgResponseTime time.Duration
	MinResponseTime time.Duration
	MaxResponseTime time.Duration
	RPS             float64
}

// BaseConfig contains common configuration for all benchmark scenarios
type BaseConfig struct {
	TykBaseURL     string
	KrakendBaseURL string
	Concurrency    int
	RequestCount   int
	Timeout        time.Duration
}

// RequestStats collects stats for HTTP benchmarks
func RequestStats(t *testing.T, client *http.Client, req *http.Request) (time.Duration, error) {
	start := time.Now()
	resp, err := client.Do(req)
	duration := time.Since(start)

	if err != nil {
		return duration, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return duration, fmt.Errorf("request failed with status code: %d", resp.StatusCode)
	}

	return duration, nil
}
