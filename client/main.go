package main

import (
	"context"
	"io"
	"log"
	"os"
	"time"

	"grpc-sample/proto/goodbye"
	"grpc-sample/proto/hello"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	defaultAddress = "localhost:50051"
	defaultName    = "World"
)

// getServerAddress returns the server address from environment variable or default
func getServerAddress() string {
	if addr := os.Getenv("GRPC_SERVER_ADDRESS"); addr != "" {
		return addr
	}
	return defaultAddress
}

// printResponseInfo prints gRPC response metadata and status information
func printResponseInfo(ctx context.Context, err error, methodName string) {
	log.Printf("=== %s Response Info ===", methodName)

	// Print status information
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			log.Printf("Status Code: %v", st.Code())
			log.Printf("Status Message: %s", st.Message())
			log.Printf("Status Details: %v", st.Details())
		} else {
			log.Printf("Error (not gRPC status): %v", err)
		}
	} else {
		log.Printf("Status Code: OK")
		log.Printf("Status Message: Success")
	}

	// Print response headers/trailers if available
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		log.Printf("Response Headers/Trailers:")
		for key, values := range md {
			log.Printf("  %s: %v", key, values)
		}
	} else {
		log.Printf("No response metadata available")
	}

	log.Printf("========================\n")
}

func main() {
	// Get server address from environment or use default
	serverAddress := getServerAddress()
	log.Printf("Connecting to gRPC server at: %s", serverAddress)

	// Set up a connection to the server.
	conn, err := grpc.Dial(serverAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	// Create clients for both services
	helloClient := hello.NewGreeterClient(conn)
	goodbyeClient := goodbye.NewFarewellClient(conn)

	// Test SayHello with response info
	log.Printf("Calling SayHello with name: %s", defaultName)

	// Create context with metadata to capture response headers
	ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("client-id", "grpc-sample-client"))
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	var header, trailer metadata.MD
	r, err := helloClient.SayHello(ctx, &hello.HelloRequest{Name: defaultName}, grpc.Header(&header), grpc.Trailer(&trailer))

	if err != nil {
		printResponseInfo(ctx, err, "SayHello")
		log.Fatalf("could not greet: %v", err)
	} else {
		log.Printf("Greeting: %s", r.GetMessage())

		// Print detailed response information
		log.Printf("=== SayHello Response Info ===")
		log.Printf("Status Code: OK")
		log.Printf("Status Message: Success")
		log.Printf("Response Headers:")
		for key, values := range header {
			log.Printf("  %s: %v", key, values)
		}
		log.Printf("Response Trailers:")
		for key, values := range trailer {
			log.Printf("  %s: %v", key, values)
		}
		log.Printf("Response Size: %d bytes", len(r.GetMessage()))
		log.Printf("========================\n")
	}

	// Test streaming RPC with response info
	log.Printf("Calling SayHelloStream with name: %s", defaultName)

	streamCtx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("stream-client-id", "grpc-sample-stream"))
	stream, err := helloClient.SayHelloStream(streamCtx, &hello.HelloRequest{Name: defaultName})

	if err != nil {
		printResponseInfo(streamCtx, err, "SayHelloStream")
		log.Fatalf("could not call SayHelloStream: %v", err)
	}

	// Get stream headers
	streamHeader, err := stream.Header()
	if err != nil {
		log.Printf("Failed to get stream headers: %v", err)
	} else {
		log.Printf("=== SayHelloStream Headers ===")
		for key, values := range streamHeader {
			log.Printf("  %s: %v", key, values)
		}
		log.Printf("==============================")
	}

	messageCount := 0
	for {
		reply, err := stream.Recv()
		if err == io.EOF {
			// Get stream trailers
			streamTrailer := stream.Trailer()
			log.Printf("=== SayHelloStream Trailers ===")
			for key, values := range streamTrailer {
				log.Printf("  %s: %v", key, values)
			}
			log.Printf("Total messages received: %d", messageCount)
			log.Printf("Stream Status: Completed successfully")
			log.Printf("===============================\n")
			break
		}
		if err != nil {
			st, ok := status.FromError(err)
			if ok {
				log.Printf("Stream error - Code: %v, Message: %s", st.Code(), st.Message())
			}
			log.Fatalf("could not receive: %v", err)
		}
		messageCount++
		log.Printf("Stream message %d: %s", messageCount, reply.GetMessage())
	}

	// Test client streaming RPC
	log.Printf("Calling SayHelloClientStream")

	clientStreamCtx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("client-stream-id", "grpc-sample-client-stream"))
	clientStream, err := helloClient.SayHelloClientStream(clientStreamCtx)

	if err != nil {
		log.Fatalf("could not call SayHelloClientStream: %v", err)
	}

	// Get client stream headers
	clientStreamHeader, err := clientStream.Header()
	if err != nil {
		log.Printf("Failed to get client stream headers: %v", err)
	} else {
		log.Printf("=== SayHelloClientStream Headers ===")
		for key, values := range clientStreamHeader {
			log.Printf("  %s: %v", key, values)
		}
		log.Printf("====================================")
	}

	// Send multiple names to server
	names := []string{"Alice", "Bob", "Charlie", "Diana"}
	for i, name := range names {
		if err := clientStream.Send(&hello.HelloRequest{Name: name}); err != nil {
			log.Fatalf("could not send: %v", err)
		}
		log.Printf("Sent client stream message %d: %s", i+1, name)
		time.Sleep(500 * time.Millisecond)
	}

	// Close and receive response
	reply, err := clientStream.CloseAndRecv()
	if err != nil {
		log.Fatalf("could not receive: %v", err)
	}

	log.Printf("Client stream response: %s", reply.GetMessage())

	// Get client stream trailers
	clientStreamTrailer := clientStream.Trailer()
	log.Printf("=== SayHelloClientStream Trailers ===")
	for key, values := range clientStreamTrailer {
		log.Printf("  %s: %v", key, values)
	}
	log.Printf("=====================================\n")

	// Test bidirectional streaming RPC
	log.Printf("Calling SayHelloBidirectional")

	bidiCtx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("bidi-client-id", "grpc-sample-bidi"))
	bidiStream, err := helloClient.SayHelloBidirectional(bidiCtx)

	if err != nil {
		log.Fatalf("could not call SayHelloBidirectional: %v", err)
	}

	// Get bidirectional stream headers
	bidiHeader, err := bidiStream.Header()
	if err != nil {
		log.Printf("Failed to get bidirectional stream headers: %v", err)
	} else {
		log.Printf("=== SayHelloBidirectional Headers ===")
		for key, values := range bidiHeader {
			log.Printf("  %s: %v", key, values)
		}
		log.Printf("=====================================")
	}

	// Start a goroutine to send messages
	go func() {
		bidiNames := []string{"Emma", "Frank", "Grace"}
		for i, name := range bidiNames {
			if err := bidiStream.Send(&hello.HelloRequest{Name: name}); err != nil {
				log.Printf("could not send bidirectional: %v", err)
				return
			}
			log.Printf("Sent bidirectional message %d: %s", i+1, name)
			time.Sleep(1 * time.Second)
		}
		bidiStream.CloseSend()
	}()

	// Receive responses
	bidiMessageCount := 0
	for {
		reply, err := bidiStream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("could not receive bidirectional: %v", err)
		}
		bidiMessageCount++
		log.Printf("Bidirectional response %d: %s", bidiMessageCount, reply.GetMessage())
	}

	// Get bidirectional stream trailers
	bidiTrailer := bidiStream.Trailer()
	log.Printf("=== SayHelloBidirectional Trailers ===")
	for key, values := range bidiTrailer {
		log.Printf("  %s: %v", key, values)
	}
	log.Printf("Total bidirectional messages received: %d", bidiMessageCount)
	log.Printf("======================================\n")

	// Test goodbye RPC with response info
	log.Printf("Calling SayGoodbye with name: %s", defaultName)

	goodbyeCtx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("goodbye-client-id", "grpc-sample-goodbye"))
	goodbyeCtx, goodbyeCancel := context.WithTimeout(goodbyeCtx, time.Second)
	defer goodbyeCancel()

	var goodbyeHeader, goodbyeTrailer metadata.MD
	goodbyeReply, err := goodbyeClient.SayGoodbye(goodbyeCtx, &goodbye.GoodbyeRequest{Name: defaultName}, grpc.Header(&goodbyeHeader), grpc.Trailer(&goodbyeTrailer))

	if err != nil {
		printResponseInfo(goodbyeCtx, err, "SayGoodbye")
		log.Fatalf("could not say goodbye: %v", err)
	} else {
		log.Printf("Goodbye message: %s", goodbyeReply.GetMessage())

		// Print detailed response information
		log.Printf("=== SayGoodbye Response Info ===")
		log.Printf("Status Code: OK")
		log.Printf("Status Message: Success")
		log.Printf("Response Headers:")
		for key, values := range goodbyeHeader {
			log.Printf("  %s: %v", key, values)
		}
		log.Printf("Response Trailers:")
		for key, values := range goodbyeTrailer {
			log.Printf("  %s: %v", key, values)
		}
		log.Printf("Response Size: %d bytes", len(goodbyeReply.GetMessage()))
		log.Printf("===============================\n")
	}

	// Test goodbye streaming RPC with response info
	log.Printf("Calling SayGoodbyeStream with name: %s", defaultName)

	goodbyeStreamCtx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("goodbye-stream-client-id", "grpc-sample-goodbye-stream"))
	goodbyeStream, err := goodbyeClient.SayGoodbyeStream(goodbyeStreamCtx, &goodbye.GoodbyeRequest{Name: defaultName})

	if err != nil {
		printResponseInfo(goodbyeStreamCtx, err, "SayGoodbyeStream")
		log.Fatalf("could not call SayGoodbyeStream: %v", err)
	}

	// Get goodbye stream headers
	goodbyeStreamHeader, err := goodbyeStream.Header()
	if err != nil {
		log.Printf("Failed to get goodbye stream headers: %v", err)
	} else {
		log.Printf("=== SayGoodbyeStream Headers ===")
		for key, values := range goodbyeStreamHeader {
			log.Printf("  %s: %v", key, values)
		}
		log.Printf("================================")
	}

	goodbyeMessageCount := 0
	for {
		reply, err := goodbyeStream.Recv()
		if err == io.EOF {
			// Get goodbye stream trailers
			goodbyeStreamTrailer := goodbyeStream.Trailer()
			log.Printf("=== SayGoodbyeStream Trailers ===")
			for key, values := range goodbyeStreamTrailer {
				log.Printf("  %s: %v", key, values)
			}
			log.Printf("Total goodbye messages received: %d", goodbyeMessageCount)
			log.Printf("Goodbye Stream Status: Completed successfully")
			log.Printf("=================================\n")
			break
		}
		if err != nil {
			st, ok := status.FromError(err)
			if ok {
				log.Printf("Goodbye stream error - Code: %v, Message: %s", st.Code(), st.Message())
			}
			log.Fatalf("could not receive goodbye: %v", err)
		}
		goodbyeMessageCount++
		log.Printf("Goodbye stream message %d: %s", goodbyeMessageCount, reply.GetMessage())
	}

	// Test goodbye client streaming RPC
	log.Printf("Calling SayGoodbyeClientStream")

	goodbyeClientStreamCtx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("goodbye-client-stream-id", "grpc-sample-goodbye-client-stream"))
	goodbyeClientStream, err := goodbyeClient.SayGoodbyeClientStream(goodbyeClientStreamCtx)

	if err != nil {
		log.Fatalf("could not call SayGoodbyeClientStream: %v", err)
	}

	// Get goodbye client stream headers
	goodbyeClientStreamHeader, err := goodbyeClientStream.Header()
	if err != nil {
		log.Printf("Failed to get goodbye client stream headers: %v", err)
	} else {
		log.Printf("=== SayGoodbyeClientStream Headers ===")
		for key, values := range goodbyeClientStreamHeader {
			log.Printf("  %s: %v", key, values)
		}
		log.Printf("======================================")
	}

	// Send multiple names for goodbye
	goodbyeNames := []string{"Helen", "Ivan", "Julia", "Kevin", "Luna"}
	for i, name := range goodbyeNames {
		if err := goodbyeClientStream.Send(&goodbye.GoodbyeRequest{Name: name}); err != nil {
			log.Fatalf("could not send goodbye: %v", err)
		}
		log.Printf("Sent goodbye client stream message %d: %s", i+1, name)
		time.Sleep(400 * time.Millisecond)
	}

	// Close and receive response
	goodbyeReply2, err := goodbyeClientStream.CloseAndRecv()
	if err != nil {
		log.Fatalf("could not receive goodbye: %v", err)
	}

	log.Printf("Goodbye client stream response: %s", goodbyeReply2.GetMessage())

	// Get goodbye client stream trailers
	goodbyeClientStreamTrailer := goodbyeClientStream.Trailer()
	log.Printf("=== SayGoodbyeClientStream Trailers ===")
	for key, values := range goodbyeClientStreamTrailer {
		log.Printf("  %s: %v", key, values)
	}
	log.Printf("=======================================\n")

	// Test goodbye bidirectional streaming RPC
	log.Printf("Calling SayGoodbyeBidirectional")

	goodbyeBidiCtx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("goodbye-bidi-client-id", "grpc-sample-goodbye-bidi"))
	goodbyeBidiStream, err := goodbyeClient.SayGoodbyeBidirectional(goodbyeBidiCtx)

	if err != nil {
		log.Fatalf("could not call SayGoodbyeBidirectional: %v", err)
	}

	// Get goodbye bidirectional stream headers
	goodbyeBidiHeader, err := goodbyeBidiStream.Header()
	if err != nil {
		log.Printf("Failed to get goodbye bidirectional stream headers: %v", err)
	} else {
		log.Printf("=== SayGoodbyeBidirectional Headers ===")
		for key, values := range goodbyeBidiHeader {
			log.Printf("  %s: %v", key, values)
		}
		log.Printf("=======================================")
	}

	// Start a goroutine to send goodbye messages
	go func() {
		goodbyeBidiNames := []string{"Maya", "Noah", "Olivia", "Paul"}
		for i, name := range goodbyeBidiNames {
			if err := goodbyeBidiStream.Send(&goodbye.GoodbyeRequest{Name: name}); err != nil {
				log.Printf("could not send goodbye bidirectional: %v", err)
				return
			}
			log.Printf("Sent goodbye bidirectional message %d: %s", i+1, name)
			time.Sleep(1200 * time.Millisecond)
		}
		goodbyeBidiStream.CloseSend()
	}()

	// Receive goodbye responses
	goodbyeBidiMessageCount := 0
	for {
		reply, err := goodbyeBidiStream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("could not receive goodbye bidirectional: %v", err)
		}
		goodbyeBidiMessageCount++
		log.Printf("Goodbye bidirectional response %d: %s", goodbyeBidiMessageCount, reply.GetMessage())
	}

	// Get goodbye bidirectional stream trailers
	goodbyeBidiTrailer := goodbyeBidiStream.Trailer()
	log.Printf("=== SayGoodbyeBidirectional Trailers ===")
	for key, values := range goodbyeBidiTrailer {
		log.Printf("  %s: %v", key, values)
	}
	log.Printf("Total goodbye bidirectional messages received: %d", goodbyeBidiMessageCount)
	log.Printf("========================================\n")

	// Print connection state information
	log.Printf("=== Connection Info ===")
	log.Printf("Target: %s", conn.Target())
	log.Printf("Connection State: %v", conn.GetState())
	log.Printf("======================")
}
