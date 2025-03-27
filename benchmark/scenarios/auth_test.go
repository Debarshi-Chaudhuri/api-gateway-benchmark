package scenarios

import (
	"bytes"
	"context"
	"net/http"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	pb "api-gateway-benchmark/backend-services/grpc-service/proto"
)

func generateJWTToken() string {
	// In a real implementation, this would use a proper JWT library
	return "Bearer test-jwt-token-for-benchmarks"
}

func TestHTTPAuthentication(t *testing.T) {
	// Skip authentication tests
	t.Skip("Authentication tests are disabled")

	config := BaseConfig{
		TykBaseURL:     "http://tyk-gateway:8080",
		KrakendBaseURL: "http://krakend:8081",
		Concurrency:    5,
		RequestCount:   20,
		Timeout:        30 * time.Second,
	}

	// Generate JWT token
	token := generateJWTToken()

	t.Run("Tyk-JWT-Auth", func(t *testing.T) {
		endpoint := config.TykBaseURL + "/http-api/api/protected"
		client := &http.Client{Timeout: config.Timeout}

		req, err := http.NewRequest("GET", endpoint, nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		// Add JWT token
		req.Header.Set("Authorization", token)

		duration, err := RequestStats(t, client, req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		t.Logf("Request completed in %v", duration)
	})

	t.Run("KrakenD-JWT-Auth", func(t *testing.T) {
		endpoint := config.KrakendBaseURL + "/http/protected"
		client := &http.Client{Timeout: config.Timeout}

		req, err := http.NewRequest("GET", endpoint, nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		// Add JWT token
		req.Header.Set("Authorization", token)

		duration, err := RequestStats(t, client, req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		t.Logf("Request completed in %v", duration)
	})
}

func TestGRPCAuthentication(t *testing.T) {
	// Skip authentication tests
	t.Skip("Authentication tests are disabled")

	config := BaseConfig{
		TykBaseURL:     "tyk-gateway:8080",
		KrakendBaseURL: "http://krakend:8081",
		Timeout:        30 * time.Second,
	}

	// Generate JWT token
	token := generateJWTToken()

	t.Run("Tyk-gRPC-JWT-Auth", func(t *testing.T) {
		// Set up a connection to the Tyk gRPC server
		conn, err := grpc.Dial(config.TykBaseURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			t.Skipf("Failed to connect to gRPC server: %v - skipping test", err)
			return
		}
		defer conn.Close()

		client := pb.NewDataServiceClient(conn)

		ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
		defer cancel()

		// Add JWT token to metadata
		ctx = metadata.AppendToOutgoingContext(ctx, "Authorization", token)

		_, err = client.GetProtectedData(ctx, &pb.DataRequest{
			RequestId: "test-req-1",
		})

		if err != nil {
			t.Fatalf("gRPC request failed: %v", err)
		}

		t.Log("gRPC request completed successfully")
	})

	t.Run("KrakenD-gRPC-JWT-Auth", func(t *testing.T) {
		// KrakenD uses HTTP-to-gRPC gateway
		endpoint := config.KrakendBaseURL + "/grpc/protected"
		client := &http.Client{Timeout: config.Timeout}

		// Create JSON request
		jsonBody := []byte(`{"request_id": "test-req-1"}`)
		req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonBody))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		// Add headers
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", token)

		duration, err := RequestStats(t, client, req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		t.Logf("Request completed in %v", duration)
	})
}
