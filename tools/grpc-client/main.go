package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	pb "api-gateway-benchmark/backend-services/grpc-service/proto/servicepb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func main() {
	// Command line flags
	tykAddr := flag.String("tyk", "localhost:8080", "Tyk gRPC endpoint")
	krakendAddr := flag.String("krakend", "localhost:8081", "KrakenD gRPC endpoint")
	useGateway := flag.String("gateway", "tyk", "Which gateway to use (tyk or krakend)")
	jwt := flag.String("jwt", "", "JWT token for authentication")
	flag.Parse()

	var serverAddr string
	if *useGateway == "tyk" {
		serverAddr = *tykAddr
	} else {
		serverAddr = *krakendAddr
	}

	// Set up a connection to the server
	conn, err := grpc.Dial(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	c := pb.NewDataServiceClient(conn)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Add JWT token if provided
	if *jwt != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+*jwt)
	}

	// Call GetData
	fmt.Println("Calling GetData...")
	dataResp, err := c.GetData(ctx, &pb.DataRequest{RequestId: "client-req-1"})
	if err != nil {
		log.Fatalf("GetData failed: %v", err)
	}
	fmt.Printf("GetData Response: %v\n", dataResp)

	// Call GetProtectedData if token provided
	if *jwt != "" {
		fmt.Println("\nCalling GetProtectedData...")
		protectedResp, err := c.GetProtectedData(ctx, &pb.DataRequest{RequestId: "client-req-2"})
		if err != nil {
			log.Fatalf("GetProtectedData failed: %v", err)
		}
		fmt.Printf("GetProtectedData Response: %v\n", protectedResp)
	}
}
