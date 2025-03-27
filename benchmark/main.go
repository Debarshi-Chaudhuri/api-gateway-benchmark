package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/olekukonko/tablewriter"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	pb "api-gateway-benchmark/backend-services/grpc-service/proto"
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
	RunTests       bool // Flag to run Go tests
	SkipAuth       bool // Skip authentication tests
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

	// Run tests if requested
	if config.RunTests {
		runGoTests()
		return
	}

	// Create JWT tokens for authentication tests (even if skipping, just in case)
	tokens := []string{}
	if !config.SkipAuth {
		tokens = generateTokens(config.TokenCount, config.JWTSecret)
	}

	// Run benchmarks for each gateway
	var results []BenchmarkResult

	fmt.Println("Starting benchmarks...")

	// HTTP Benchmarks
	if config.AuthTest && !config.SkipAuth {
		fmt.Println("\nRunning Authentication Tests (HTTP)...")
		results = append(results, runHTTPBenchmark(config, tokens, "auth")...)
	} else if config.AuthTest {
		fmt.Println("\nSkipping Authentication Tests (auth tests disabled)...")
	}

	if config.RateLimitTest {
		fmt.Println("\nRunning Rate Limiting Tests (HTTP)...")
		results = append(results, runHTTPBenchmark(config, tokens, "ratelimit")...)
	}

	if config.ResilienceTest {
		fmt.Println("\nRunning Resilience Tests (HTTP)...")
		results = append(results, runHTTPBenchmark(config, tokens, "resilience")...)
	}

	if config.ProxyTest {
		fmt.Println("\nRunning API Proxying Tests (HTTP)...")
		results = append(results, runHTTPBenchmark(config, tokens, "proxy")...)
	}

	// gRPC Benchmarks
	if config.GrpcEnabled {
		if config.AuthTest && !config.SkipAuth {
			fmt.Println("\nRunning Authentication Tests (gRPC)...")
			results = append(results, runGRPCBenchmark(config, tokens, "auth")...)
		} else if config.AuthTest {
			fmt.Println("\nSkipping Authentication Tests for gRPC (auth tests disabled)...")
		}

		if config.RateLimitTest {
			fmt.Println("\nRunning Rate Limiting Tests (gRPC)...")
			results = append(results, runGRPCBenchmark(config, tokens, "ratelimit")...)
		}

		if config.ProxyTest {
			fmt.Println("\nRunning API Proxying Tests (gRPC)...")
			results = append(results, runGRPCBenchmark(config, tokens, "proxy")...)
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
	runTests := flag.Bool("test", false, "Run Go tests instead of benchmarks")
	skipAuth := flag.Bool("skipauth", true, "Skip authentication tests") // Default to true

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
		RunTests:       *runTests,
		SkipAuth:       *skipAuth,
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
	// For simplicity and to avoid external dependencies, we'll use a placeholder
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
		// For auth tests, use public endpoints if skipping auth
		if config.SkipAuth {
			endpoints["tyk"] = fmt.Sprintf("%s/http-api/api/data", config.TykBaseURL)
			endpoints["krakend"] = fmt.Sprintf("%s/http/data", config.KrakendBaseURL)
		} else {
			endpoints["tyk"] = fmt.Sprintf("%s/http-api/api/protected", config.TykBaseURL)
			endpoints["krakend"] = fmt.Sprintf("%s/http/protected", config.KrakendBaseURL)
		}
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
		resultsChan := make(chan time.Duration, config.RequestCount)
		errorsChan := make(chan error, config.RequestCount)

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
						errorsChan <- err
						continue
					}

					// Add authorization header if testing auth and not skipping auth
					if (scenario == "auth" || scenario == "ratelimit") && !config.SkipAuth {
						tokenIndex := (workerID*requestsPerWorker + j) % len(tokens)
						req.Header.Set("Authorization", tokens[tokenIndex])
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
			}(i)
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
		fmt.Printf("Completed benchmark for %s (%s): %d requests, %d successful, %d failed, %.2f RPS\n",
			gateway, scenario, config.RequestCount, successCount, errorCount, result.RPS)
	}

	return results
}

func runGRPCBenchmark(config BenchmarkConfig, tokens []string, scenario string) []BenchmarkResult {
	var results []BenchmarkResult

	// Define endpoints for each gateway and scenario
	type endpointConfig struct {
		useHTTP bool // true for HTTP-to-gRPC (KrakenD), false for direct gRPC (Tyk)
		host    string
		path    string
		method  string
	}

	endpoints := map[string]endpointConfig{
		"tyk":     {useHTTP: false, host: "", path: "", method: ""},
		"krakend": {useHTTP: true, host: "", path: "", method: ""},
	}

	switch scenario {
	case "auth":
		if config.SkipAuth {
			// Use non-protected endpoints if skipping auth
			endpoints["tyk"] = endpointConfig{
				useHTTP: false,
				host:    config.TykBaseURL,
				path:    "grpc-api",
				method:  "GetData",
			}
			endpoints["krakend"] = endpointConfig{
				useHTTP: true,
				host:    config.KrakendBaseURL,
				path:    "/grpc/data",
				method:  "GetData",
			}
		} else {
			// Use protected endpoints for auth tests
			endpoints["tyk"] = endpointConfig{
				useHTTP: false,
				host:    config.TykBaseURL,
				path:    "grpc-api",
				method:  "GetProtectedData",
			}
			endpoints["krakend"] = endpointConfig{
				useHTTP: true,
				host:    config.KrakendBaseURL,
				path:    "/grpc/protected",
				method:  "GetProtectedData",
			}
		}
	case "ratelimit", "proxy":
		endpoints["tyk"] = endpointConfig{
			useHTTP: false,
			host:    config.TykBaseURL,
			path:    "grpc-api",
			method:  "GetData",
		}
		endpoints["krakend"] = endpointConfig{
			useHTTP: true,
			host:    config.KrakendBaseURL,
			path:    "/grpc/data",
			method:  "GetData",
		}
	}

	for gateway, endpoint := range endpoints {
		result := BenchmarkResult{
			Gateway:         gateway,
			Scenario:        scenario,
			Protocol:        "gRPC",
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
			go func(workerID int) {
				defer wg.Done()

				if endpoint.useHTTP {
					// HTTP-to-gRPC for KrakenD
					client := &http.Client{
						Timeout: config.Timeout,
					}

					for j := 0; j < requestsPerWorker; j++ {
						// Prepare JSON request body
						jsonBody := []byte(`{"request_id": "req-` + fmt.Sprintf("%d-%d", workerID, j) + `"}`)
						req, err := http.NewRequest("POST", endpoint.host+endpoint.path, bytes.NewBuffer(jsonBody))
						if err != nil {
							errorsChan <- err
							continue
						}

						req.Header.Set("Content-Type", "application/json")

						// Add authorization if testing auth and not skipping auth
						if (scenario == "auth" || scenario == "ratelimit") && !config.SkipAuth {
							tokenIndex := (workerID*requestsPerWorker + j) % len(tokens)
							req.Header.Set("Authorization", tokens[tokenIndex])
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
				} else {
					// Direct gRPC for Tyk
					// Parse host:port from URL
					url := strings.TrimPrefix(endpoint.host, "http://")
					url = strings.TrimPrefix(url, "https://")
					hostParts := strings.Split(url, ":")
					host := hostParts[0]
					port := "8080" // Default
					if len(hostParts) > 1 {
						port = hostParts[1]
					}

					serverAddr := fmt.Sprintf("%s:%s", host, port)
					conn, err := grpc.Dial(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
					if err != nil {
						for j := 0; j < requestsPerWorker; j++ {
							errorsChan <- fmt.Errorf("failed to connect to gRPC server: %v", err)
						}
						return
					}
					defer conn.Close()

					client := pb.NewDataServiceClient(conn)

					for j := 0; j < requestsPerWorker; j++ {
						ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)

						// Add authorization if testing auth and not skipping auth
						if (scenario == "auth" || scenario == "ratelimit") && !config.SkipAuth {
							tokenIndex := (workerID*requestsPerWorker + j) % len(tokens)
							ctx = metadata.AppendToOutgoingContext(ctx, "Authorization", tokens[tokenIndex])
						}

						reqStart := time.Now()
						var err error

						request := &pb.DataRequest{
							RequestId: fmt.Sprintf("req-%d-%d", workerID, j),
						}

						if endpoint.method == "GetProtectedData" {
							_, err = client.GetProtectedData(ctx, request)
						} else {
							_, err = client.GetData(ctx, request)
						}

						reqDuration := time.Since(reqStart)
						cancel()

						if err != nil {
							errorsChan <- err
						} else {
							resultsChan <- reqDuration
						}
					}
				}
			}(i)
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
		fmt.Printf("Completed benchmark for %s (%s): %d requests, %d successful, %d failed, %.2f RPS\n",
			gateway, scenario, config.RequestCount, successCount, errorCount, result.RPS)
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

func runGoTests() {
	fmt.Println("Running Go tests...")

	cmd := exec.Command("go", "test", "./scenarios/...", "-v")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("Tests failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Tests completed successfully")
}
