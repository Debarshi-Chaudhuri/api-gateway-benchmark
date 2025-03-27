package scenarios

import (
	"bytes"
	"context"
	"net/http"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "api-gateway-benchmark/backend-services/grpc-service/proto"
)

func TestHTTPProxy(t *testing.T) {
	config := BaseConfig{
		TykBaseURL:     "http://tyk-gateway:8080",
		KrakendBaseURL: "http://krakend:8081",
		Timeout:        30 * time.Second,
	}

	t.Run("Tyk-HTTP-Proxy", func(t *testing.T) {
		// Test Tyk HTTP proxying
		endpoint := config.TykBaseURL + "/http-api/"
		client := &http.Client{Timeout: config.Timeout}

		req, err := http.NewRequest("GET", endpoint, nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		duration, err := RequestStats(t, client, req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		t.Logf("Request completed in %v", duration)
	})

	t.Run("KrakenD-HTTP-Proxy", func(t *testing.T) {
		// Test KrakenD HTTP proxying
		endpoint := config.KrakendBaseURL + "/http/data"
		client := &http.Client{Timeout: config.Timeout}

		req, err := http.NewRequest("GET", endpoint, nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		duration, err := RequestStats(t, client, req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		t.Logf("Request completed in %v", duration)
	})

	t.Run("Tyk-HTTP-Health", func(t *testing.T) {
		// Test health endpoint
		endpoint := config.TykBaseURL + "/http-api/health"
		client := &http.Client{Timeout: config.Timeout}

		req, err := http.NewRequest("GET", endpoint, nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		duration, err := RequestStats(t, client, req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		t.Logf("Health check completed in %v", duration)
	})

	t.Run("KrakenD-HTTP-Health", func(t *testing.T) {
		// Test health endpoint
		endpoint := config.KrakendBaseURL + "/health"
		client := &http.Client{Timeout: config.Timeout}

		req, err := http.NewRequest("GET", endpoint, nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		duration, err := RequestStats(t, client, req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		t.Logf("Health check completed in %v", duration)
	})
}

func TestGRPCProxy(t *testing.T) {
	config := BaseConfig{
		TykBaseURL:     "tyk-gateway:8080",
		KrakendBaseURL: "http://krakend:8081",
		Timeout:        30 * time.Second,
	}

	t.Run("Tyk-gRPC-Proxy", func(t *testing.T) {
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

		_, err = client.GetData(ctx, &pb.DataRequest{
			RequestId: "test-req-1",
		})

		if err != nil {
			t.Fatalf("gRPC request failed: %v", err)
		}

		t.Log("gRPC request completed successfully")
	})

	t.Run("KrakenD-gRPC-Proxy", func(t *testing.T) {
		// KrakenD uses HTTP-to-gRPC gateway
		endpoint := config.KrakendBaseURL + "/grpc/data"
		client := &http.Client{Timeout: config.Timeout}

		// Create JSON request
		jsonBody := []byte(`{"request_id": "test-req-1"}`)
		req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonBody))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		// Add headers
		req.Header.Set("Content-Type", "application/json")

		duration, err := RequestStats(t, client, req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		t.Logf("Request completed in %v", duration)
	})
}
