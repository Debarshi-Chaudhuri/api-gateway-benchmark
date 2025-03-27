package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/olekukonko/tablewriter"
)

type BenchmarkConfig struct {
	TykBaseURL     string
	KrakendBaseURL string
	Concurrency    int
	RequestCount   int
	Timeout        time.Duration
	OutputFile     string
}

type BenchmarkResult struct {
	Gateway         string
	RequestCount    int
	SuccessCount    int
	FailureCount    int
	TotalDuration   time.Duration
	AvgResponseTime time.Duration
	MinResponseTime time.Duration
	MaxResponseTime time.Duration
	RPS             float64
}

func main() {
	config := parseFlags()

	fmt.Println("Starting benchmarks...")

	var results []BenchmarkResult

	// Run benchmark for each gateway
	fmt.Println("\nRunning HTTP benchmarks...")
	results = append(results, runHTTPBenchmark(config)...)

	// Print results
	printResults(results)

	// Save results to file
	if config.OutputFile != "" {
		saveResults(results, config.OutputFile)
	}
}

func parseFlags() BenchmarkConfig {
	tykURL := flag.String("tyk", "http://tyk-gateway:8080", "Tyk Gateway URL")
	krakendURL := flag.String("krakend", "http://krakend:8081", "KrakenD Gateway URL")
	concurrency := flag.Int("concurrency", 10, "Number of concurrent clients")
	requestCount := flag.Int("requests", 1000, "Total number of requests")
	timeout := flag.Duration("timeout", 30*time.Second, "Request timeout")
	outputFile := flag.String("output", "results/benchmark_results.txt", "Output file for results")

	flag.Parse()

	return BenchmarkConfig{
		TykBaseURL:     *tykURL,
		KrakendBaseURL: *krakendURL,
		Concurrency:    *concurrency,
		RequestCount:   *requestCount,
		Timeout:        *timeout,
		OutputFile:     *outputFile,
	}
}

func runHTTPBenchmark(config BenchmarkConfig) []BenchmarkResult {
	var results []BenchmarkResult

	endpoints := map[string]string{
		"tyk":     fmt.Sprintf("%s/http-api/api/data", config.TykBaseURL),
		"krakend": fmt.Sprintf("%s/http/data", config.KrakendBaseURL),
	}

	for gateway, endpoint := range endpoints {
		result := BenchmarkResult{
			Gateway:         gateway,
			RequestCount:    config.RequestCount,
			MinResponseTime: time.Hour, // Initialize with a large value
		}

		start := time.Now()
		var wg sync.WaitGroup
		requestsPerWorker := config.RequestCount / config.Concurrency
		resultsChan := make(chan time.Duration, config.RequestCount)
		errorsChan := make(chan error, config.RequestCount)

		for i := 0; i < config.Concurrency; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				client := &http.Client{
					Timeout: config.Timeout,
				}

				for j := 0; j < requestsPerWorker; j++ {
					req, err := http.NewRequest("GET", endpoint, nil)
					if err != nil {
						errorsChan <- err
						continue
					}

					reqStart := time.Now()
					resp, err := client.Do(req)
					reqDuration := time.Since(reqStart)

					if err != nil {
						errorsChan <- err
						continue
					}

					if resp.StatusCode >= 200 && resp.StatusCode < 300 {
						resultsChan <- reqDuration
					} else {
						errorsChan <- fmt.Errorf("request failed with status code: %d", resp.StatusCode)
					}
					resp.Body.Close()
				}
			}()
		}

		wg.Wait()
		close(resultsChan)
		close(errorsChan)

		// Process results
		totalTime := time.Duration(0)
		successCount := 0
		errorCount := 0

		for duration := range resultsChan {
			totalTime += duration
			successCount++

			if duration < result.MinResponseTime {
				result.MinResponseTime = duration
			}

			if duration > result.MaxResponseTime {
				result.MaxResponseTime = duration
			}
		}

		for range errorsChan {
			errorCount++
		}

		// Calculate stats
		result.TotalDuration = time.Since(start)
		result.SuccessCount = successCount
		result.FailureCount = errorCount

		if successCount > 0 {
			result.AvgResponseTime = totalTime / time.Duration(successCount)
			result.RPS = float64(successCount) / result.TotalDuration.Seconds()
		}

		results = append(results, result)
		fmt.Printf("Completed benchmark for %s: %d requests, %d successful, %d failed, %.2f RPS\n",
			gateway, config.RequestCount, successCount, errorCount, result.RPS)
	}

	return results
}

func printResults(results []BenchmarkResult) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Gateway", "Requests", "Success", "Failed", "Avg Time", "Min Time", "Max Time", "RPS"})

	for _, result := range results {
		table.Append([]string{
			result.Gateway,
			fmt.Sprintf("%d", result.RequestCount),
			fmt.Sprintf("%d", result.SuccessCount),
			fmt.Sprintf("%d", result.FailureCount),
			fmt.Sprintf("%.2f ms", float64(result.AvgResponseTime.Microseconds())/1000.0),
			fmt.Sprintf("%.2f ms", float64(result.MinResponseTime.Microseconds())/1000.0),
			fmt.Sprintf("%.2f ms", float64(result.MaxResponseTime.Microseconds())/1000.0),
			fmt.Sprintf("%.2f", result.RPS),
		})
	}

	table.Render()
}

func saveResults(results []BenchmarkResult, filePath string) {
	// Create directory if it doesn't exist
	dir := "results"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.Mkdir(dir, 0755)
	}

	// Format results as text
	var content string
	content += "API Gateway Benchmark Results\n"
	content += "============================\n\n"

	for _, result := range results {
		content += fmt.Sprintf("Gateway: %s\n", result.Gateway)
		content += fmt.Sprintf("Requests: %d\n", result.RequestCount)
		content += fmt.Sprintf("Success: %d\n", result.SuccessCount)
		content += fmt.Sprintf("Failed: %d\n", result.FailureCount)
		content += fmt.Sprintf("Average Response Time: %.2f ms\n", float64(result.AvgResponseTime.Microseconds())/1000.0)
		content += fmt.Sprintf("Minimum Response Time: %.2f ms\n", float64(result.MinResponseTime.Microseconds())/1000.0)
		content += fmt.Sprintf("Maximum Response Time: %.2f ms\n", float64(result.MaxResponseTime.Microseconds())/1000.0)
		content += fmt.Sprintf("Requests Per Second: %.2f\n", result.RPS)
		content += "\n"
	}

	// Write to file
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		fmt.Printf("Error writing to file: %v\n", err)
		return
	}

	fmt.Printf("Results saved to %s\n", filePath)
}
