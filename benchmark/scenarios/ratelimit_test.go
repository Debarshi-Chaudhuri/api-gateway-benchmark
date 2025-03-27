package scenarios

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	pb "api-gateway-benchmark/backend-services/grpc-service/proto"
)

func TestHTTPRateLimit(t *testing.T) {
	config := BaseConfig{
		TykBaseURL:     "http://tyk-gateway:8080",
		KrakendBaseURL: "http://krakend:8081",
		Concurrency:    5,
		RequestCount:   100,
		Timeout:        1 * time.Second, // Short timeout to detect rate limiting faster
	}

	// Generate JWT token
	token := generateJWTToken()

	t.Run("Tyk-HTTP-RateLimit", func(t *testing.T) {
		endpoint := config.TykBaseURL + "/http-api/api/data"
		successCount := 0
		failureCount := 0
		var mu sync.Mutex

		var wg sync.WaitGroup
		for i := 0; i < 1; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				client := &http.Client{Timeout: config.Timeout}

				for j := 0; j < 10; j++ {
					req, err := http.NewRequest("GET", endpoint, nil)
					if err != nil {
						t.Logf("Failed to create request: %v", err)
						continue
					}

					// Add JWT token
					req.Header.Set("Authorization", token)

					_, err = RequestStats(t, client, req)

					// Log error message if rate limiting is detected
					if err != nil {
						t.Logf("Request %d failed: %v", j, err)
					}
					mu.Lock()
					if err != nil {
						failureCount++
					} else {
						successCount++
					}
					mu.Unlock()

					// Don't wait between requests to trigger rate limiting
					time.Sleep(10 * time.Millisecond)
				}
			}(i)
		}

		wg.Wait()

		t.Logf("Rate limit test completed with %d successes and %d failures", successCount, failureCount)

		// We expect some failures due to rate limiting for a proper test
		// but if all requests are failing, something is wrong
		if failureCount == 0 {
			t.Logf("Warning: Expected some failures due to rate limiting but got none. Rate limiting may not be working.")
		} else if successCount == 0 {
			t.Errorf("All requests failed. Rate limiting may be too aggressive or there's another issue.")
		}
	})

	t.Run("KrakenD-HTTP-RateLimit", func(t *testing.T) {
		endpoint := config.KrakendBaseURL + "/http/data"
		successCount := 0
		failureCount := 0
		var mu sync.Mutex

		var wg sync.WaitGroup
		for i := 0; i < 20; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				client := &http.Client{Timeout: config.Timeout}

				for j := 0; j < 10; j++ {
					req, err := http.NewRequest("GET", endpoint, nil)
					if err != nil {
						t.Logf("Failed to create request: %v", err)
						continue
					}

					// Add JWT token
					req.Header.Set("Authorization", token)

					_, err = RequestStats(t, client, req)

					mu.Lock()
					if err != nil {
						failureCount++
					} else {
						successCount++
					}
					mu.Unlock()

					// Don't wait between requests to trigger rate limiting
					time.Sleep(10 * time.Millisecond)
				}
			}(i)
		}

		wg.Wait()

		t.Logf("Rate limit test completed with %d successes and %d failures", successCount, failureCount)

		// We expect some failures due to rate limiting for a proper test
		if failureCount == 0 {
			t.Logf("Warning: Expected some failures due to rate limiting but got none. Rate limiting may not be working.")
		} else if successCount == 0 {
			t.Errorf("All requests failed. Rate limiting may be too aggressive or there's another issue.")
		}
	})
}

func TestGRPCRateLimit(t *testing.T) {
	config := BaseConfig{
		TykBaseURL:     "tyk-gateway:8080",
		KrakendBaseURL: "http://krakend:8081",
		Timeout:        1 * time.Second,
	}

	// Generate JWT token
	token := generateJWTToken()

	t.Run("Tyk-gRPC-RateLimit", func(t *testing.T) {
		// Set up a connection to the Tyk gRPC server
		conn, err := grpc.Dial(config.TykBaseURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			t.Skipf("Failed to connect to gRPC server: %v - skipping test", err)
			return
		}
		defer conn.Close()

		client := pb.NewDataServiceClient(conn)

		successCount := 0
		failureCount := 0
		var mu sync.Mutex

		var wg sync.WaitGroup
		for i := 0; i < 20; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				for j := 0; j < 10; j++ {
					ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)

					// Add JWT token to metadata
					ctx = metadata.AppendToOutgoingContext(ctx, "Authorization", token)

					_, err := client.GetData(ctx, &pb.DataRequest{
						RequestId: fmt.Sprintf("test-req-%d-%d", id, j),
					})

					cancel()

					mu.Lock()
					if err != nil {
						failureCount++
					} else {
						successCount++
					}
					mu.Unlock()

					// Don't wait between requests to trigger rate limiting
					time.Sleep(10 * time.Millisecond)
				}
			}(i)
		}

		wg.Wait()

		t.Logf("gRPC rate limit test completed with %d successes and %d failures", successCount, failureCount)

		// We expect some failures due to rate limiting
		if failureCount == 0 {
			t.Logf("Warning: Expected some failures due to rate limiting but got none. Rate limiting may not be working.")
		} else if successCount == 0 {
			t.Errorf("All requests failed. Rate limiting may be too aggressive or there's another issue.")
		}
	})

	t.Run("KrakenD-gRPC-RateLimit", func(t *testing.T) {
		endpoint := config.KrakendBaseURL + "/grpc/data"
		successCount := 0
		failureCount := 0
		var mu sync.Mutex

		var wg sync.WaitGroup
		for i := 0; i < 20; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				client := &http.Client{Timeout: config.Timeout}

				for j := 0; j < 10; j++ {
					// Create JSON request
					jsonBody := []byte(fmt.Sprintf(`{"request_id": "test-req-%d-%d"}`, id, j))
					req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonBody))
					if err != nil {
						t.Logf("Failed to create request: %v", err)
						continue
					}

					// Add headers
					req.Header.Set("Content-Type", "application/json")
					req.Header.Set("Authorization", token)

					_, err = RequestStats(t, client, req)

					mu.Lock()
					if err != nil {
						failureCount++
					} else {
						successCount++
					}
					mu.Unlock()

					// Don't wait between requests to trigger rate limiting
					time.Sleep(10 * time.Millisecond)
				}
			}(i)
		}

		wg.Wait()

		t.Logf("gRPC over HTTP rate limit test completed with %d successes and %d failures", successCount, failureCount)

		// We expect some failures due to rate limiting
		if failureCount == 0 {
			t.Logf("Warning: Expected some failures due to rate limiting but got none. Rate limiting may not be working.")
		} else if successCount == 0 {
			t.Errorf("All requests failed. Rate limiting may be too aggressive or there's another issue.")
		}
	})
}
