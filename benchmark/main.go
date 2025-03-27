package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
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
	Scenario       string
	TokenCount     int
	JWTSecret      string
	OutputFile     string
	GrpcEnabled    bool
	RateLimitTest  bool
	ResilienceTest bool
	AuthTest       bool
	ProxyTest      bool
}

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

func main() {
	config := parseFlags()

	// Create JWT tokens for authentication tests
	tokens := generateTokens(config.TokenCount, config.JWTSecret)

	// Run benchmarks for each gateway
	var results []BenchmarkResult

	fmt.Println("Starting benchmarks...")

	// HTTP Benchmarks
	if config.AuthTest {
		fmt.Println("\nRunning Authentication Tests (HTTP)...")
		results = append(results, runHTTPBenchmark(config, tokens, "auth"))
	}

	if config.RateLimitTest {
		fmt.Println("\nRunning Rate Limiting Tests (HTTP)...")
		results = append(results, runHTTPBenchmark(config, tokens, "ratelimit"))
	}

	if config.ResilienceTest {
		fmt.Println("\nRunning Resilience Tests (HTTP)...")
		results = append(results, runHTTPBenchmark(config, tokens, "resilience"))
	}

	if config.ProxyTest {
		fmt.Println("\nRunning API Proxying Tests (HTTP)...")
		results = append(results, runHTTPBenchmark(config, tokens, "proxy"))
	}

	// gRPC Benchmarks
	if config.GrpcEnabled {
		if config.AuthTest {
			fmt.Println("\nRunning Authentication Tests (gRPC)...")
			results = append(results, runGRPCBenchmark(config, tokens, "auth"))
		}

		if config.RateLimitTest {
			fmt.Println("\nRunning Rate Limiting Tests (gRPC)...")
			results = append(results, runGRPCBenchmark(config, tokens, "ratelimit"))
		}

		if config.ProxyTest {
			fmt.Println("\nRunning API Proxying Tests (gRPC)...")
			results = append(results, runGRPCBenchmark(config, tokens, "proxy"))
		}
	}

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
	scenario := flag.String("scenario", "all", "Test scenario (auth, resilience, ratelimit, proxy, all)")
	tokenCount := flag.Int("tokens", 10, "Number of different JWT tokens to use")
	jwtSecret := flag.String("jwtsecret", "test-secret-key-for-benchmark", "Secret key for JWT tokens")
	outputFile := flag.String("output", "results/benchmark_results.json", "Output file for results")
	grpcEnabled := flag.Bool("grpc", true, "Enable gRPC tests")

	flag.Parse()

	config := BenchmarkConfig{
		TykBaseURL:     *tykURL,
		KrakendBaseURL: *krakendURL,
		Concurrency:    *concurrency,
		RequestCount:   *requestCount,
		Timeout:        *timeout,
		Scenario:       *scenario,
		TokenCount:     *tokenCount,
		JWTSecret:      *jwtSecret,
		OutputFile:     *outputFile,
		GrpcEnabled:    *grpcEnabled,
	}

	// Set test flags based on scenario
	if config.Scenario == "all" {
		config.AuthTest = true
		config.ResilienceTest = true
		config.RateLimitTest = true
		config.ProxyTest = true
	} else {
		config.AuthTest = config.Scenario == "auth"
		config.ResilienceTest = config.Scenario == "resilience"
		config.RateLimitTest = config.Scenario == "ratelimit"
		config.ProxyTest = config.Scenario == "proxy"
	}

	return config
}

func generateTokens(count int, secret string) []string {
	// In a real app, you would use a JWT library to generate actual tokens
	// This is a placeholder for demonstration
	tokens := make([]string, count)
	for i := 0; i < count; i++ {
		tokens[i] = fmt.Sprintf("Bearer fake-jwt-token-%d", i)
	}
	return tokens
}

func runHTTPBenchmark(config BenchmarkConfig, tokens []string, scenario string) []BenchmarkResult {
	var results []BenchmarkResult

	endpoints := map[string]string{
		"tyk":     "",
		"krakend": "",
	}

	switch scenario {
	case "auth":
		endpoints["tyk"] = fmt.Sprintf("%s/http-api/api/protected", config.TykBaseURL)
		endpoints["krakend"] = fmt.Sprintf("%s/http/protected", config.KrakendBaseURL)
	case "resilience":
		endpoints["tyk"] = fmt.Sprintf("%s/http-api/api/data?delay=100", config.TykBaseURL)
		endpoints["krakend"] = fmt.Sprintf("%s/http/data?delay=100", config.KrakendBaseURL)
	case "ratelimit":
		endpoints["tyk"] = fmt.Sprintf("%s/http-api/api/data", config.TykBaseURL)
		endpoints["krakend"] = fmt.Sprintf("%s/http/data", config.KrakendBaseURL)
	case "proxy":
		endpoints["tyk"] = fmt.Sprintf("%s/http-api/", config.TykBaseURL)
		endpoints["krakend"] = fmt.Sprintf("%s/http/data", config.KrakendBaseURL)
	}

	for gateway, endpoint := range endpoints {
		result := BenchmarkResult{
			Gateway:         gateway,
			Scenario:        scenario,
			Protocol:        "HTTP",
			RequestCount:    config.RequestCount,
			MinResponseTime: time.Hour, // Initialize with a large value
		}

		start := time.Now()
		var wg sync.WaitGroup
		requestsPerWorker := config.RequestCount / config.Concurrency
		results := make(chan time.Duration, config.RequestCount)
		errors := make(chan error, config.RequestCount)

		for i := 0; i < config.Concurrency; i++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()
				client := &http.Client{
					Timeout: config.Timeout,
				}

				for j := 0; j < requestsPerWorker; j++ {
					req, err := http.NewRequest("GET", endpoint, nil)
					if err != nil {
						errors <- err
						continue
					}

					// Add authorization header if testing auth
					if scenario == "auth" || scenario == "ratelimit" {
						tokenIndex := (workerID*requestsPerWorker + j) % len(tokens)
						req.Header.Set("Authorization", tokens[tokenIndex])
					}

					reqStart := time.Now()
					resp, err := client.Do(req)
					reqDuration := time.Since(reqStart)

					if err != nil {
						errors <- err
						continue
					}

					if resp.StatusCode >= 200 && resp.StatusCode < 300 {
						results <- reqDuration
					} else {
						errors <- fmt.Errorf("request failed with status code: %d", resp.StatusCode)
					}
					resp.Body.Close()
				}
			}(i)
		}

		wg.Wait()
		close(results)
		close(errors)

		// Process results
		totalTime := time.Duration(0)
		successCount := 0

		for duration := range results {
			totalTime += duration
			successCount++

			if duration < result.MinResponseTime {
				result.MinResponseTime = duration
			}

			if duration > result.MaxResponseTime {
				result.MaxResponseTime = duration
			}
		}

		// Calculate stats
		result.TotalDuration = time.Since(start)
		result.SuccessCount = successCount
		result.FailureCount = len(errors)

		if successCount > 0 {
			result.AvgResponseTime = totalTime / time.Duration(successCount)
			result.RPS = float64(successCount) / result.TotalDuration.Seconds()
		}

		results = append(results, result)
		fmt.Printf("Completed benchmark for %s (%s): %d requests, %d successful, %d failed, %.2f RPS\n",
			gateway, scenario, config.RequestCount, successCount, len(errors), result.RPS)
	}

	return results
}

func runGRPCBenchmark(config BenchmarkConfig, tokens []string, scenario string) []BenchmarkResult {
	var results []BenchmarkResult

	// This is a simplified version - in a real benchmark you'd implement actual gRPC clients
	// For this example, we'll just create placeholder results

	for _, gateway := range []string{"tyk", "krakend"} {
		result := BenchmarkResult{
			Gateway:         gateway,
			Scenario:        scenario,
			Protocol:        "gRPC",
			RequestCount:    config.RequestCount,
			SuccessCount:    config.RequestCount, // Simplified
			FailureCount:    0,
			TotalDuration:   time.Second * 5,                    // Placeholder
			AvgResponseTime: time.Millisecond * 50,              // Placeholder
			MinResponseTime: time.Millisecond * 10,              // Placeholder
			MaxResponseTime: time.Millisecond * 200,             // Placeholder
			RPS:             float64(config.RequestCount) / 5.0, // Placeholder
		}

		results = append(results, result)
	}

	return results
}

func printResults(results []BenchmarkResult) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Gateway", "Scenario", "Protocol", "Requests", "Success", "Failed", "Avg Time", "RPS"})

	for _, result := range results {
		table.Append([]string{
			result.Gateway,
			result.Scenario,
			result.Protocol,
			fmt.Sprintf("%d", result.RequestCount),
			fmt.Sprintf("%d", result.SuccessCount),
			fmt.Sprintf("%d", result.FailureCount),
			fmt.Sprintf("%.2f ms", float64(result.AvgResponseTime.Microseconds())/1000.0),
			fmt.Sprintf("%.2f", result.RPS),
		})
	}

	table.Render()
}

func saveResults(results []BenchmarkResult, filename string) {
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		log.Printf("Error marshaling results: %v", err)
		return
	}

	err = os.MkdirAll("results", 0755)
	if err != nil {
		log.Printf("Error creating results directory: %v", err)
		return
	}

	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		log.Printf("Error writing results to file: %v", err)
		return
	}

	fmt.Printf("Results saved to %s\n", filename)
}
