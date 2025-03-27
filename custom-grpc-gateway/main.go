package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	pb "api-gateway-benchmark/backend-services/grpc-service/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type dataRequest struct {
	RequestID string `json:"request_id"`
	DelayMs   int32  `json:"delay_ms"`
}

func main() {
	// Get environment variables or use defaults
	port := os.Getenv("PORT")
	if port == "" {
		port = "8090"
	}

	grpcServer := os.Getenv("GRPC_SERVER")
	if grpcServer == "" {
		grpcServer = "grpc-service:9000"
	}

	// Set up gRPC connection to the server
	conn, err := grpc.Dial(grpcServer, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	client := pb.NewDataServiceClient(conn)

	// Handle GetData endpoint
	http.HandleFunc("/service.DataService/GetData", func(w http.ResponseWriter, r *http.Request) {
		handleGrpcRequest(w, r, func(ctx context.Context, req *dataRequest) (interface{}, error) {
			return client.GetData(ctx, &pb.DataRequest{
				RequestId: req.RequestID,
				DelayMs:   req.DelayMs,
			})
		})
	})

	// Handle GetProtectedData endpoint
	http.HandleFunc("/service.DataService/GetProtectedData", func(w http.ResponseWriter, r *http.Request) {
		handleGrpcRequest(w, r, func(ctx context.Context, req *dataRequest) (interface{}, error) {
			return client.GetProtectedData(ctx, &pb.DataRequest{
				RequestId: req.RequestID,
				DelayMs:   req.DelayMs,
			})
		})
	})

	// Handle HealthCheck endpoint
	http.HandleFunc("/service.DataService/HealthCheck", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Forward headers (especially authorization) to gRPC metadata
		if authHeader := r.Header.Get("Authorization"); authHeader != "" {
			ctx = metadata.AppendToOutgoingContext(ctx, "Authorization", authHeader)
		}

		resp, err := client.HealthCheck(ctx, &pb.HealthRequest{})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// Start HTTP server
	log.Printf("gRPC gateway starting on :%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func handleGrpcRequest(w http.ResponseWriter, r *http.Request, grpcCall func(context.Context, *dataRequest) (interface{}, error)) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req dataRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Forward headers (especially authorization) to gRPC metadata
	if authHeader := r.Header.Get("Authorization"); authHeader != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "Authorization", authHeader)
	}

	resp, err := grpcCall(ctx, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
