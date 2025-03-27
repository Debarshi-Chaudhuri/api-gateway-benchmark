package scenarios

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "api-gateway-benchmark/backend-services/grpc-service/proto"
)

// Response is the expected structure from the HTTP service
type Response struct {
	Status    string      `json:"status"`
	Message   string      `json:"message"`
	RequestID string      `json:"request_id"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

func TestServicesAvailability(t *testing.T) {
	// Test HTTP service directly
	t.Run("HTTP-Service", func(t *testing.T) {
		endpoint := "http://http-service:8000/health"
		client := &http.Client{Timeout: 5 * time.Second}

		req, err := http.NewRequest("GET", endpoint, nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Failed to reach HTTP service: %v", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		var response Response
		if err := json.Unmarshal(body, &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v\nBody: %s", err, string(body))
		}

		if response.Status != "success" {
			t.Errorf("Expected status 'success', got '%s'", response.Status)
		}

		t.Logf("HTTP service is available: %s", response.Message)
	})

	// Test gRPC service directly
	t.Run("gRPC-Service", func(t *testing.T) {
		conn, err := grpc.Dial("grpc-service:9000", grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			t.Fatalf("Failed to connect to gRPC service: %v", err)
		}
		defer conn.Close()

		client := pb.NewDataServiceClient(conn)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		resp, err := client.HealthCheck(ctx, &pb.HealthRequest{})
		if err != nil {
			t.Fatalf("Failed to call gRPC service health check: %v", err)
		}

		if resp.Status != "success" {
			t.Errorf("Expected status 'success', got '%s'", resp.Status)
		}

		t.Logf("gRPC service is available: %s", resp.Message)
	})

	// Test Tyk gateway
	t.Run("Tyk-Gateway", func(t *testing.T) {
		endpoint := "http://tyk-gateway:8080/http-api/health"
		client := &http.Client{Timeout: 5 * time.Second}

		req, err := http.NewRequest("GET", endpoint, nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Failed to reach Tyk gateway: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status code 200, got %d", resp.StatusCode)
		}

		t.Logf("Tyk gateway is available")
	})

	// Test KrakenD gateway
	t.Run("KrakenD-Gateway", func(t *testing.T) {
		endpoint := "http://krakend:8081/health"
		client := &http.Client{Timeout: 5 * time.Second}

		req, err := http.NewRequest("GET", endpoint, nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Failed to reach KrakenD gateway: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status code 200, got %d", resp.StatusCode)
		}

		t.Logf("KrakenD gateway is available")
	})

	// Test KrakenD gRPC gateway
	t.Run("KrakenD-gRPC-Gateway", func(t *testing.T) {
		endpoint := "http://krakend:8081/grpc/data"
		client := &http.Client{Timeout: 5 * time.Second}

		jsonBody := []byte(`{"request_id": "test-integration"}`)
		req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonBody))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Failed to reach KrakenD gRPC gateway: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status code 200, got %d", resp.StatusCode)
		}

		t.Logf("KrakenD gRPC gateway is available")
	})
}

func TestEndToEndFlow(t *testing.T) {
	// This test performs a series of requests through both gateways
	// to verify that the entire request flow works correctly

	// Generate JWT token
	token := generateJWTToken()

	// Test Tyk HTTP flow
	t.Run("Tyk-HTTP-Flow", func(t *testing.T) {
		baseURL := "http://tyk-gateway:8080"
		client := &http.Client{Timeout: 5 * time.Second}

		// 1. Health check
		t.Log("Step 1: Health check")
		req, _ := http.NewRequest("GET", baseURL+"/http-api/health", nil)
		resp, err := client.Do(req)
		if err != nil || resp.StatusCode != http.StatusOK {
			t.Fatalf("Health check failed: %v, status: %v", err, resp.StatusCode)
		}
		resp.Body.Close()

		// 2. Public endpoint
		t.Log("Step 2: Public endpoint")
		req, _ = http.NewRequest("GET", baseURL+"/http-api/api/data", nil)
		req.Header.Set("Authorization", token)
		resp, err = client.Do(req)
		if err != nil || resp.StatusCode != http.StatusOK {
			t.Fatalf("Public endpoint failed: %v, status: %v", err, resp.StatusCode)
		}

		body, _ := io.ReadAll(resp.Body)
		var publicResp Response
		json.Unmarshal(body, &publicResp)
		resp.Body.Close()

		t.Logf("Public response: %s - %s", publicResp.Status, publicResp.Message)

		// 3. Protected endpoint
		t.Log("Step 3: Protected endpoint")
		req, _ = http.NewRequest("GET", baseURL+"/http-api/api/protected", nil)
		req.Header.Set("Authorization", token)
		resp, err = client.Do(req)
		if err != nil || resp.StatusCode != http.StatusOK {
			t.Fatalf("Protected endpoint failed: %v, status: %v", err, resp.StatusCode)
		}

		body, _ = io.ReadAll(resp.Body)
		var protectedResp Response
		json.Unmarshal(body, &protectedResp)
		resp.Body.Close()

		t.Logf("Protected response: %s - %s", protectedResp.Status, protectedResp.Message)

		t.Log("Tyk HTTP flow test completed successfully")
	})

	// Test KrakenD HTTP flow
	t.Run("KrakenD-HTTP-Flow", func(t *testing.T) {
		baseURL := "http://krakend:8081"
		client := &http.Client{Timeout: 5 * time.Second}

		// 1. Health check
		t.Log("Step 1: Health check")
		req, _ := http.NewRequest("GET", baseURL+"/health", nil)
		resp, err := client.Do(req)
		if err != nil || resp.StatusCode != http.StatusOK {
			t.Fatalf("Health check failed: %v, status: %v", err, resp.StatusCode)
		}
		resp.Body.Close()

		// 2. Public endpoint
		t.Log("Step 2: Public endpoint")
		req, _ = http.NewRequest("GET", baseURL+"/http/data", nil)
		req.Header.Set("Authorization", token)
		resp, err = client.Do(req)
		if err != nil || resp.StatusCode != http.StatusOK {
			t.Fatalf("Public endpoint failed: %v, status: %v", err, resp.StatusCode)
		}
		resp.Body.Close()

		// 3. Protected endpoint
		t.Log("Step 3: Protected endpoint")
		req, _ = http.NewRequest("GET", baseURL+"/http/protected", nil)
		req.Header.Set("Authorization", token)
		resp, err = client.Do(req)
		if err != nil || resp.StatusCode != http.StatusOK {
			t.Fatalf("Protected endpoint failed: %v, status: %v", err, resp.StatusCode)
		}
		resp.Body.Close()

		t.Log("KrakenD HTTP flow test completed successfully")
	})

	// Test KrakenD gRPC flow
	t.Run("KrakenD-gRPC-Flow", func(t *testing.T) {
		baseURL := "http://krakend:8081"
		client := &http.Client{Timeout: 5 * time.Second}

		// 1. Public gRPC endpoint
		t.Log("Step 1: Public gRPC endpoint")
		jsonBody := []byte(`{"request_id": "test-flow-1"}`)
		req, _ := http.NewRequest("POST", baseURL+"/grpc/data", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := client.Do(req)
		if err != nil || resp.StatusCode != http.StatusOK {
			t.Fatalf("Public gRPC endpoint failed: %v, status: %v", err, resp.StatusCode)
		}
		resp.Body.Close()

		// 2. Protected gRPC endpoint
		t.Log("Step 2: Protected gRPC endpoint")
		jsonBody = []byte(`{"request_id": "test-flow-2"}`)
		req, _ = http.NewRequest("POST", baseURL+"/grpc/protected", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", token)
		resp, err = client.Do(req)
		if err != nil || resp.StatusCode != http.StatusOK {
			t.Fatalf("Protected gRPC endpoint failed: %v, status: %v", err, resp.StatusCode)
		}
		resp.Body.Close()

		t.Log("KrakenD gRPC flow test completed successfully")
	})
}
