package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"sync/atomic"
	"time"

	pb "github.com/example/grpc-service/proto/servicepb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

var (
	requestCount uint64
	serviceReady = true
)

type server struct {
	pb.UnimplementedDataServiceServer
}

func (s *server) GetData(ctx context.Context, req *pb.DataRequest) (*pb.DataResponse, error) {
	if !serviceReady {
		return nil, status.Error(codes.Unavailable, "Service is not ready")
	}

	id := atomic.AddUint64(&requestCount, 1)
	reqID := req.RequestId
	if reqID == "" {
		reqID = fmt.Sprintf("req-%d", id)
	}

	// Simulate delay if requested
	if req.DelayMs > 0 {
		time.Sleep(time.Duration(req.DelayMs) * time.Millisecond)
	}

	return &pb.DataResponse{
		Status:    "success",
		Message:   "Data retrieved successfully",
		RequestId: reqID,
		Items: []*pb.Item{
			{Id: "1", Name: "Item 1", Description: "Description 1"},
			{Id: "2", Name: "Item 2", Description: "Description 2"},
			{Id: "3", Name: "Item 3", Description: "Description 3"},
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}, nil
}

func (s *server) GetProtectedData(ctx context.Context, req *pb.DataRequest) (*pb.DataResponse, error) {
	if !serviceReady {
		return nil, status.Error(codes.Unavailable, "Service is not ready")
	}

	id := atomic.AddUint64(&requestCount, 1)
	reqID := req.RequestId
	if reqID == "" {
		reqID = fmt.Sprintf("req-%d", id)
	}

	// Check for authorization
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "No metadata found")
	}

	authTokens := md.Get("authorization")
	if len(authTokens) == 0 {
		return nil, status.Error(codes.Unauthenticated, "Authorization token required")
	}

	// Log metadata for debugging
	log.Printf("Request metadata: %v", md)

	return &pb.DataResponse{
		Status:    "success",
		Message:   "Access granted to protected data",
		RequestId: reqID,
		Items: []*pb.Item{
			{Id: "101", Name: "Protected Item 1", Description: "Confidential Description 1"},
			{Id: "102", Name: "Protected Item 2", Description: "Confidential Description 2"},
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}, nil
}

func (s *server) HealthCheck(ctx context.Context, req *pb.HealthRequest) (*pb.HealthResponse, error) {
	if !serviceReady {
		return nil, status.Error(codes.Unavailable, "Service is not ready")
	}

	return &pb.HealthResponse{
		Status:    "success",
		Message:   "Service is healthy",
		Timestamp: time.Now().Format(time.RFC3339),
	}, nil
}

func main() {
	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "9000"
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterDataServiceServer(s, &server{})

	// Register reflection service for easier testing
	reflection.Register(s)

	log.Printf("gRPC service starting on :%s", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
