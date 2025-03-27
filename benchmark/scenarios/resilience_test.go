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

	pb "api-gateway-benchmark/backend-services/grpc-service/proto"
)

func TestHTTPResilience(t *testing.T) {
	config := BaseConfig{
		TykBaseURL:     "http://tyk-gateway:8080",
		KrakendBaseURL: "http://krakend:8081",
		Concurrency:    5,
		RequestCount:   50,
		Timeout:        5 * time.Second,
	}

	// Test timeout handling
	t.Run("Tyk-HTTP-Timeout", func(t *testing.T) {
		// Use a delay parameter to test timeout handling
		endpoint := fmt.Sprintf("%s/http-api/api/data?delay=2000", config.TykBaseURL)
		client := &http.Client{Timeout: 1 * time.Second} // Short timeout to force client-side timeout

		req, err := http.NewRequest("GET", endpoint, nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		_, err = RequestStats(t, client, req)

		// We expect a timeout error
		if err == nil {
			t.Errorf("Expected timeout error but request succeeded")
		} else {
			t.Logf("Got expected timeout error: %v", err)
		}
	})

	t.Run("KrakenD-HTTP-Timeout", func(t *testing.T) {
		// Use a delay parameter to test timeout handling
		endpoint := fmt.Sprintf("%s/http/data?delay=2000", config.KrakendBaseURL)
		client := &http.Client{Timeout: 1 * time.Second} // Short timeout to force client-side timeout

		req, err := http.NewRequest("GET", endpoint, nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		_, err = RequestStats(t, client, req)

		// We expect a timeout error
		if err == nil {
			t.Errorf("Expected timeout error but request succeeded")
		} else {
			t.Logf("Got expected timeout error: %v", err)
		}
	})

	// Test circuit breaker by making many concurrent requests with delays
	t.Run("Tyk-HTTP-CircuitBreaker", func(t *testing.T) {
		endpoint := fmt.Sprintf("%s/http-api/api/data?delay=500", config.TykBaseURL)
		successCount := 0
		failureCount := 0
		var mu sync.Mutex

		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				client := &http.Client{Timeout: config.Timeout}

				for j := 0; j < 5; j++ {
					req, err := http.NewRequest("GET", endpoint, nil)
					if err != nil {
						t.Logf("Failed to create request: %v", err)
						continue
					}

					_, err = RequestStats(t, client, req)

					mu.Lock()
					if err != nil {
						failureCount++
						t.Logf("Request %d-%d failed: %v", id, j, err)
					} else {
						successCount++
					}
					mu.Unlock()

					// Small delay between requests
					time.Sleep(50 * time.Millisecond)
				}
			}(i)
		}

		wg.Wait()

		t.Logf("Circuit breaker test completed with %d successes and %d failures", successCount, failureCount)

		// If circuit breaker is working, we expect to see some failures after the circuit opens
		if failureCount > 0 {
			t.Logf("Circuit breaker may have activated after %d failures", failureCount)
		}
	})

	t.Run("KrakenD-HTTP-CircuitBreaker", func(t *testing.T) {
		endpoint := fmt.Sprintf("%s/http/data?delay=500", config.KrakendBaseURL)
		successCount := 0
		failureCount := 0
		var mu sync.Mutex

		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				client := &http.Client{Timeout: config.Timeout}

				for j := 0; j < 5; j++ {
					req, err := http.NewRequest("GET", endpoint, nil)
					if err != nil {
						t.Logf("Failed to create request: %v", err)
						continue
					}

					_, err = RequestStats(t, client, req)

					mu.Lock()
					if err != nil {
						failureCount++
						t.Logf("Request %d-%d failed: %v", id, j, err)
					} else {
						successCount++
					}
					mu.Unlock()

					// Small delay between requests
					time.Sleep(50 * time.Millisecond)
				}
			}(i)
		}

		wg.Wait()

		t.Logf("Circuit breaker test completed with %d successes and %d failures", successCount, failureCount)

		// If circuit breaker is working, we expect to see some failures after the circuit opens
		if failureCount > 0 {
			t.Logf("Circuit breaker may have activated after %d failures", failureCount)
		}
	})

	// Test service unavailability handling
	t.Run("HTTP-ServiceUnavailable", func(t *testing.T) {
		// First toggle the service to "not ready" state
		toggleEndpoint := "http://http-service:8000/toggle-ready"
		client := &http.Client{Timeout: config.Timeout}

		req, err := http.NewRequest("GET", toggleEndpoint, nil)
		if err != nil {
			t.Fatalf("Failed to create toggle request: %v", err)
		}

		_, err = client.Do(req)
		if err != nil {
			t.Fatalf("Failed to toggle service readiness: %v", err)
		}

		// Now test both gateways with the service unavailable
		t.Run("Tyk-ServiceUnavailable", func(t *testing.T) {
			endpoint := fmt.Sprintf("%s/http-api/api/data", config.TykBaseURL)
			req, err := http.NewRequest("GET", endpoint, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			_, err = RequestStats(t, client, req)

			// We expect an error because the service is unavailable
			if err == nil {
				t.Errorf("Expected error but request succeeded")
			} else {
				t.Logf("Got expected error: %v", err)
			}
		})

		t.Run("KrakenD-ServiceUnavailable", func(t *testing.T) {
			endpoint := fmt.Sprintf("%s/http/data", config.KrakendBaseURL)
			req, err := http.NewRequest("GET", endpoint, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			_, err = RequestStats(t, client, req)

			// We expect an error because the service is unavailable
			if err == nil {
				t.Errorf("Expected error but request succeeded")
			} else {
				t.Logf("Got expected error: %v", err)
			}
		})

		// Toggle the service back to "ready" state for other tests
		req, err = http.NewRequest("GET", toggleEndpoint, nil)
		if err != nil {
			t.Fatalf("Failed to create toggle request: %v", err)
		}

		_, err = client.Do(req)
		if err != nil {
			t.Fatalf("Failed to toggle service readiness: %v", err)
		}

		// Wait for service to stabilize
		time.Sleep(1 * time.Second)
	})
}

func TestGRPCResilience(t *testing.T) {
	config := BaseConfig{
		TykBaseURL:     "tyk-gateway:8080",
		KrakendBaseURL: "http://krakend:8081",
		Timeout:        5 * time.Second,
	}

	t.Run("Tyk-gRPC-Delay", func(t *testing.T) {
		// Set up a connection to the Tyk gRPC server
		conn, err := grpc.Dial(config.TykBaseURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			t.Skipf("Failed to connect to gRPC server: %v - skipping test", err)
			return
		}
		defer conn.Close()

		client := pb.NewDataServiceClient(conn)

		// Use a short timeout to test handling of delayed responses
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		// Request with 2 second delay should timeout with 1 second client timeout
		_, err = client.GetData(ctx, &pb.DataRequest{
			RequestId: "test-req-1",
			DelayMs:   2000,
		})

		// We expect a timeout error
		if err == nil {
			t.Errorf("Expected timeout error but request succeeded")
		} else {
			t.Logf("Got expected error: %v", err)
		}
	})

	t.Run("KrakenD-gRPC-Delay", func(t *testing.T) {
		// KrakenD uses HTTP-to-gRPC gateway
		endpoint := config.KrakendBaseURL + "/grpc/data"
		client := &http.Client{Timeout: 1 * time.Second} // Short timeout

		// Create JSON request with delay
		jsonBody := []byte(`{"request_id": "test-req-1", "delay_ms": 2000}`)
		req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonBody))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		// Add headers
		req.Header.Set("Content-Type", "application/json")

		_, err = RequestStats(t, client, req)

		// We expect a timeout error
		if err == nil {
			t.Errorf("Expected timeout error but request succeeded")
		} else {
			t.Logf("Got expected error: %v", err)
		}
	})

	// Test circuit breaker by making many concurrent requests with delays
	t.Run("KrakenD-gRPC-CircuitBreaker", func(t *testing.T) {
		endpoint := config.KrakendBaseURL + "/grpc/data"
		successCount := 0
		failureCount := 0
		var mu sync.Mutex

		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				client := &http.Client{Timeout: config.Timeout}

				for j := 0; j < 5; j++ {
					// Create JSON request with delay
					jsonBody := []byte(fmt.Sprintf(`{"request_id": "test-req-%d-%d", "delay_ms": 500}`, id, j))
					req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonBody))
					if err != nil {
						t.Logf("Failed to create request: %v", err)
						continue
					}

					// Add headers
					req.Header.Set("Content-Type", "application/json")

					_, err = RequestStats(t, client, req)

					mu.Lock()
					if err != nil {
						failureCount++
					} else {
						successCount++
					}
					mu.Unlock()

					// Small delay between requests
					time.Sleep(50 * time.Millisecond)
				}
			}(i)
		}

		wg.Wait()

		t.Logf("gRPC circuit breaker test completed with %d successes and %d failures", successCount, failureCount)

		// If circuit breaker is working, we expect to see some failures after the circuit opens
		if failureCount > 0 {
			t.Logf("Circuit breaker may have activated after %d failures", failureCount)
		}
	})
}
