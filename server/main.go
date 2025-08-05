package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"grpc-sample/proto/goodbye"
	"grpc-sample/proto/hello"

	"github.com/gorilla/mux"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
)

// helloServer is used to implement hello.GreeterServer.
type helloServer struct {
	hello.UnimplementedGreeterServer
}

// goodbyeServer is used to implement goodbye.FarewellServer.
type goodbyeServer struct {
	goodbye.UnimplementedFarewellServer
}

// HTTP request/response structs for REST API
type HelloRequest struct {
	Name string `json:"name"`
}

type HelloResponse struct {
	Message string `json:"message"`
}

type GoodbyeRequest struct {
	Name string `json:"name"`
}

type GoodbyeResponse struct {
	Message string `json:"message"`
}

// SayHello implements hello.GreeterServer
func (s *helloServer) SayHello(ctx context.Context, in *hello.HelloRequest) (*hello.HelloReply, error) {
	log.Printf("gRPC: Received SayHello request: %v", in.GetName())

	// Read incoming metadata
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		log.Printf("gRPC: Incoming metadata:")
		for key, values := range md {
			log.Printf("  %s: %v", key, values)
		}
	}

	// Set response headers
	header := metadata.Pairs(
		"server-name", "grpc-sample-server",
		"method", "SayHello",
		"timestamp", time.Now().Format(time.RFC3339),
		"response-id", fmt.Sprintf("hello-%d", time.Now().Unix()),
	)
	grpc.SendHeader(ctx, header)

	// Set response trailers
	trailer := metadata.Pairs(
		"processing-time", "fast",
		"server-version", "1.0.0",
	)
	grpc.SetTrailer(ctx, trailer)

	return &hello.HelloReply{Message: "Hello " + in.GetName()}, nil
}

// SayHelloStream implements hello.GreeterServer
func (s *helloServer) SayHelloStream(in *hello.HelloRequest, stream hello.Greeter_SayHelloStreamServer) error {
	log.Printf("gRPC: Received stream request for: %v", in.GetName())

	// Read incoming metadata
	if md, ok := metadata.FromIncomingContext(stream.Context()); ok {
		log.Printf("gRPC: Stream incoming metadata:")
		for key, values := range md {
			log.Printf("  %s: %v", key, values)
		}
	}

	// Set stream headers
	header := metadata.Pairs(
		"server-name", "grpc-sample-server",
		"method", "SayHelloStream",
		"stream-id", fmt.Sprintf("stream-%d", time.Now().Unix()),
		"expected-messages", "5",
	)
	stream.SendHeader(header)

	for i := 0; i < 5; i++ {
		reply := &hello.HelloReply{
			Message: fmt.Sprintf("Hello %s - Message %d", in.GetName(), i+1),
		}

		if err := stream.Send(reply); err != nil {
			return err
		}

		// Add a small delay between messages
		time.Sleep(1 * time.Second)
	}

	// Set stream trailers
	trailer := metadata.Pairs(
		"messages-sent", "5",
		"stream-duration", "5s",
		"stream-status", "completed",
	)
	stream.SetTrailer(trailer)

	return nil
}

// SayHelloClientStream implements hello.GreeterServer
func (s *helloServer) SayHelloClientStream(stream hello.Greeter_SayHelloClientStreamServer) error {
	log.Printf("gRPC: Received client stream request")

	// Read incoming metadata
	if md, ok := metadata.FromIncomingContext(stream.Context()); ok {
		log.Printf("gRPC: Client stream incoming metadata:")
		for key, values := range md {
			log.Printf("  %s: %v", key, values)
		}
	}

	// Set stream headers
	header := metadata.Pairs(
		"server-name", "grpc-sample-server",
		"method", "SayHelloClientStream",
		"stream-id", fmt.Sprintf("client-stream-%d", time.Now().Unix()),
		"stream-type", "client-streaming",
	)
	stream.SendHeader(header)

	var names []string
	messageCount := 0

	// Receive all messages from client
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			// Client finished sending
			break
		}
		if err != nil {
			return err
		}
		messageCount++
		names = append(names, req.GetName())
		log.Printf("gRPC: Received client stream message %d: %s", messageCount, req.GetName())
	}

	// Send single response with summary
	summary := fmt.Sprintf("Hello to all %d friends: %s!", len(names), strings.Join(names, ", "))

	// Set response trailers
	trailer := metadata.Pairs(
		"messages-received", fmt.Sprintf("%d", messageCount),
		"names-processed", strings.Join(names, ","),
		"stream-status", "completed",
		"processing-time", "batch",
	)
	stream.SetTrailer(trailer)

	return stream.SendAndClose(&hello.HelloReply{Message: summary})
}

// SayHelloBidirectional implements hello.GreeterServer
func (s *helloServer) SayHelloBidirectional(stream hello.Greeter_SayHelloBidirectionalServer) error {
	log.Printf("gRPC: Received bidirectional stream request")

	// Read incoming metadata
	if md, ok := metadata.FromIncomingContext(stream.Context()); ok {
		log.Printf("gRPC: Bidirectional stream incoming metadata:")
		for key, values := range md {
			log.Printf("  %s: %v", key, values)
		}
	}

	// Set stream headers
	header := metadata.Pairs(
		"server-name", "grpc-sample-server",
		"method", "SayHelloBidirectional",
		"stream-id", fmt.Sprintf("bidi-stream-%d", time.Now().Unix()),
		"stream-type", "bidirectional",
	)
	stream.SendHeader(header)

	messageCount := 0
	var processedNames []string

	// Handle bidirectional streaming
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			// Client finished sending
			break
		}
		if err != nil {
			return err
		}

		messageCount++
		name := req.GetName()
		processedNames = append(processedNames, name)
		log.Printf("gRPC: Received bidirectional message %d: %s", messageCount, name)

		// Send immediate response for each received message
		response := fmt.Sprintf("Hello %s! (Message %d received)", name, messageCount)
		if err := stream.Send(&hello.HelloReply{Message: response}); err != nil {
			return err
		}

		// Add a small delay to simulate processing
		time.Sleep(500 * time.Millisecond)
	}

	// Set stream trailers
	trailer := metadata.Pairs(
		"messages-exchanged", fmt.Sprintf("%d", messageCount),
		"names-processed", strings.Join(processedNames, ","),
		"stream-status", "completed",
		"stream-duration", fmt.Sprintf("%.1fs", float64(messageCount)*0.5),
	)
	stream.SetTrailer(trailer)

	return nil
}

// SayGoodbye implements goodbye.FarewellServer
func (s *goodbyeServer) SayGoodbye(ctx context.Context, in *goodbye.GoodbyeRequest) (*goodbye.GoodbyeReply, error) {
	log.Printf("gRPC: Received goodbye request for: %v", in.GetName())

	// Read incoming metadata
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		log.Printf("gRPC: Goodbye incoming metadata:")
		for key, values := range md {
			log.Printf("  %s: %v", key, values)
		}
	}

	// Set response headers
	header := metadata.Pairs(
		"server-name", "grpc-sample-server",
		"method", "SayGoodbye",
		"timestamp", time.Now().Format(time.RFC3339),
		"farewell-type", "friendly",
	)
	grpc.SendHeader(ctx, header)

	// Set response trailers
	trailer := metadata.Pairs(
		"goodbye-processed", "true",
		"session-ended", time.Now().Format(time.RFC3339),
	)
	grpc.SetTrailer(ctx, trailer)

	return &goodbye.GoodbyeReply{Message: "Goodbye " + in.GetName() + "! See you later!"}, nil
}

// SayGoodbyeStream implements goodbye.FarewellServer
func (s *goodbyeServer) SayGoodbyeStream(in *goodbye.GoodbyeRequest, stream goodbye.Farewell_SayGoodbyeStreamServer) error {
	log.Printf("gRPC: Received goodbye stream request for: %v", in.GetName())

	// Read incoming metadata
	if md, ok := metadata.FromIncomingContext(stream.Context()); ok {
		log.Printf("gRPC: Goodbye stream incoming metadata:")
		for key, values := range md {
			log.Printf("  %s: %v", key, values)
		}
	}

	// Set stream headers
	header := metadata.Pairs(
		"server-name", "grpc-sample-server",
		"method", "SayGoodbyeStream",
		"stream-id", fmt.Sprintf("goodbye-stream-%d", time.Now().Unix()),
		"expected-messages", "3",
		"farewell-type", "streaming",
	)
	stream.SendHeader(header)

	// Send multiple goodbye messages
	goodbyeMessages := []string{
		"Thanks for using our service, %s!",
		"It was great having you, %s!",
		"Until we meet again, %s! Farewell!",
	}

	for i, template := range goodbyeMessages {
		reply := &goodbye.GoodbyeReply{
			Message: fmt.Sprintf(template, in.GetName()),
		}

		if err := stream.Send(reply); err != nil {
			return err
		}

		log.Printf("gRPC: Sent goodbye message %d: %s", i+1, reply.Message)

		// Add a delay between messages
		time.Sleep(1500 * time.Millisecond)
	}

	// Set stream trailers
	trailer := metadata.Pairs(
		"messages-sent", "3",
		"stream-duration", "4.5s",
		"stream-status", "completed",
		"farewell-completed", time.Now().Format(time.RFC3339),
	)
	stream.SetTrailer(trailer)

	return nil
}

// SayGoodbyeClientStream implements goodbye.FarewellServer
func (s *goodbyeServer) SayGoodbyeClientStream(stream goodbye.Farewell_SayGoodbyeClientStreamServer) error {
	log.Printf("gRPC: Received goodbye client stream request")

	// Read incoming metadata
	if md, ok := metadata.FromIncomingContext(stream.Context()); ok {
		log.Printf("gRPC: Goodbye client stream incoming metadata:")
		for key, values := range md {
			log.Printf("  %s: %v", key, values)
		}
	}

	// Set stream headers
	header := metadata.Pairs(
		"server-name", "grpc-sample-server",
		"method", "SayGoodbyeClientStream",
		"stream-id", fmt.Sprintf("goodbye-client-stream-%d", time.Now().Unix()),
		"stream-type", "client-streaming",
		"farewell-type", "batch",
	)
	stream.SendHeader(header)

	var names []string
	messageCount := 0

	// Receive all messages from client
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			// Client finished sending
			break
		}
		if err != nil {
			return err
		}
		messageCount++
		names = append(names, req.GetName())
		log.Printf("gRPC: Received goodbye client stream message %d: %s", messageCount, req.GetName())
	}

	// Send single farewell response with summary
	summary := fmt.Sprintf("Farewell to all %d wonderful people: %s! May your paths be bright!", len(names), strings.Join(names, ", "))

	// Set response trailers
	trailer := metadata.Pairs(
		"messages-received", fmt.Sprintf("%d", messageCount),
		"names-processed", strings.Join(names, ","),
		"stream-status", "completed",
		"farewell-type", "collective",
		"session-ended", time.Now().Format(time.RFC3339),
	)
	stream.SetTrailer(trailer)

	return stream.SendAndClose(&goodbye.GoodbyeReply{Message: summary})
}

// SayGoodbyeBidirectional implements goodbye.FarewellServer
func (s *goodbyeServer) SayGoodbyeBidirectional(stream goodbye.Farewell_SayGoodbyeBidirectionalServer) error {
	log.Printf("gRPC: Received goodbye bidirectional stream request")

	// Read incoming metadata
	if md, ok := metadata.FromIncomingContext(stream.Context()); ok {
		log.Printf("gRPC: Goodbye bidirectional stream incoming metadata:")
		for key, values := range md {
			log.Printf("  %s: %v", key, values)
		}
	}

	// Set stream headers
	header := metadata.Pairs(
		"server-name", "grpc-sample-server",
		"method", "SayGoodbyeBidirectional",
		"stream-id", fmt.Sprintf("goodbye-bidi-stream-%d", time.Now().Unix()),
		"stream-type", "bidirectional",
		"farewell-type", "interactive",
	)
	stream.SendHeader(header)

	messageCount := 0
	var processedNames []string
	farewellMessages := []string{
		"Take care, %s!",
		"Safe travels, %s!",
		"Until next time, %s!",
		"Farewell, dear %s!",
		"Goodbye and good luck, %s!",
	}

	// Handle bidirectional streaming
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			// Client finished sending
			break
		}
		if err != nil {
			return err
		}

		messageCount++
		name := req.GetName()
		processedNames = append(processedNames, name)
		log.Printf("gRPC: Received goodbye bidirectional message %d: %s", messageCount, name)

		// Send personalized farewell response for each received message
		farewellTemplate := farewellMessages[(messageCount-1)%len(farewellMessages)]
		response := fmt.Sprintf(farewellTemplate+" (Farewell %d)", name, messageCount)
		if err := stream.Send(&goodbye.GoodbyeReply{Message: response}); err != nil {
			return err
		}

		// Add a delay to simulate thoughtful farewell processing
		time.Sleep(750 * time.Millisecond)
	}

	// Set stream trailers
	trailer := metadata.Pairs(
		"farewells-exchanged", fmt.Sprintf("%d", messageCount),
		"names-processed", strings.Join(processedNames, ","),
		"stream-status", "completed",
		"stream-duration", fmt.Sprintf("%.1fs", float64(messageCount)*0.75),
		"final-farewell", time.Now().Format(time.RFC3339),
	)
	stream.SetTrailer(trailer)

	return nil
}

// HTTP REST API handlers
func (s *helloServer) handleSayHelloHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("HTTP: Received SayHello request")

	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req HelloRequest
	var name string

	if r.Method == "GET" {
		// Handle GET request with query parameter
		name = r.URL.Query().Get("name")
		if name == "" {
			name = "World"
		}
	} else if r.Method == "POST" {
		// Handle POST request with JSON body
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		name = req.Name
		if name == "" {
			name = "World"
		}
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Printf("HTTP: Processing hello request for: %s", name)

	// Create gRPC request and call the gRPC method
	grpcReq := &hello.HelloRequest{Name: name}
	grpcResp, err := s.SayHello(context.Background(), grpcReq)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Convert to HTTP response
	resp := HelloResponse{Message: grpcResp.Message}

	// Add custom headers
	w.Header().Set("X-Server-Name", "grpc-sample-server")
	w.Header().Set("X-Method", "SayHello")
	w.Header().Set("X-Protocol", "HTTP")
	w.Header().Set("X-Timestamp", time.Now().Format(time.RFC3339))

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (s *goodbyeServer) handleSayGoodbyeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("HTTP: Received SayGoodbye request")

	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req GoodbyeRequest
	var name string

	if r.Method == "GET" {
		// Handle GET request with query parameter
		name = r.URL.Query().Get("name")
		if name == "" {
			name = "Friend"
		}
	} else if r.Method == "POST" {
		// Handle POST request with JSON body
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		name = req.Name
		if name == "" {
			name = "Friend"
		}
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Printf("HTTP: Processing goodbye request for: %s", name)

	// Create gRPC request and call the gRPC method
	grpcReq := &goodbye.GoodbyeRequest{Name: name}
	grpcResp, err := s.SayGoodbye(context.Background(), grpcReq)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Convert to HTTP response
	resp := GoodbyeResponse{Message: grpcResp.Message}

	// Add custom headers
	w.Header().Set("X-Server-Name", "grpc-sample-server")
	w.Header().Set("X-Method", "SayGoodbye")
	w.Header().Set("X-Protocol", "HTTP")
	w.Header().Set("X-Timestamp", time.Now().Format(time.RFC3339))

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// Health check endpoint
func handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"services": map[string]string{
			"grpc": "running on :50051",
			"http": "running on :50051 (same port)",
		},
		"version": "1.0.0",
		"note":    "Both gRPC and HTTP protocols are served on the same port",
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(health)
}

// API documentation endpoint
func handleAPIDoc(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	apiDoc := map[string]interface{}{
		"title":       "gRPC Sample Server API",
		"version":     "1.0.0",
		"description": "Unified server supporting both gRPC and HTTP REST APIs on the same port",
		"protocols":   []string{"gRPC", "HTTP"},
		"port":        "50051",
		"note":        "Both protocols are served on the same port using protocol multiplexing",
		"endpoints": map[string]interface{}{
			"grpc": map[string]interface{}{
				"address": ":50051",
				"services": []map[string]interface{}{
					{
						"name":    "grpc.hello.Greeter",
						"methods": []string{"SayHello", "SayHelloStream", "SayHelloClientStream", "SayHelloBidirectional"},
					},
					{
						"name":    "grpc.goodbye.Farewell",
						"methods": []string{"SayGoodbye", "SayGoodbyeStream", "SayGoodbyeClientStream", "SayGoodbyeBidirectional"},
					},
				},
			},
			"http": map[string]interface{}{
				"address": ":50051",
				"routes": []map[string]interface{}{
					{
						"path":        "/api/hello",
						"methods":     []string{"GET", "POST"},
						"description": "Say hello to someone",
						"parameters": map[string]string{
							"name": "Name of the person to greet (query param for GET, JSON body for POST)",
						},
					},
					{
						"path":        "/api/goodbye",
						"methods":     []string{"GET", "POST"},
						"description": "Say goodbye to someone",
						"parameters": map[string]string{
							"name": "Name of the person to bid farewell (query param for GET, JSON body for POST)",
						},
					},
					{
						"path":        "/health",
						"methods":     []string{"GET"},
						"description": "Health check endpoint",
					},
					{
						"path":        "/api/doc",
						"methods":     []string{"GET"},
						"description": "API documentation",
					},
				},
			},
		},
		"examples": map[string]interface{}{
			"grpc": map[string]string{
				"list_services": "grpcurl -plaintext localhost:50051 list",
				"say_hello":     "grpcurl -plaintext -d '{\"name\":\"World\"}' localhost:50051 grpc.hello.Greeter/SayHello",
				"say_goodbye":   "grpcurl -plaintext -d '{\"name\":\"Friend\"}' localhost:50051 grpc.goodbye.Farewell/SayGoodbye",
			},
			"http": map[string]string{
				"say_hello_get":  "curl 'http://localhost:50051/api/hello?name=World'",
				"say_hello_post": "curl -X POST -H 'Content-Type: application/json' -d '{\"name\":\"World\"}' http://localhost:50051/api/hello",
				"say_goodbye":    "curl 'http://localhost:50051/api/goodbye?name=Friend'",
				"health_check":   "curl http://localhost:50051/health",
			},
		},
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(apiDoc)
}

// Protocol multiplexer that can handle both gRPC and HTTP on the same port
func createMultiplexedHandler(grpcServer *grpc.Server, httpHandler http.Handler) http.Handler {
	return h2c.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if this is a gRPC request
		if r.ProtoMajor == 2 && strings.HasPrefix(r.Header.Get("Content-Type"), "application/grpc") {
			// This is a gRPC request
			grpcServer.ServeHTTP(w, r)
		} else {
			// This is an HTTP request
			httpHandler.ServeHTTP(w, r)
		}
	}), &http2.Server{})
}

// Setup HTTP router
func setupHTTPRouter(helloSrv *helloServer, goodbyeSrv *goodbyeServer) http.Handler {
	router := mux.NewRouter()

	// API routes
	router.HandleFunc("/api/hello", helloSrv.handleSayHelloHTTP).Methods("GET", "POST", "OPTIONS")
	router.HandleFunc("/api/goodbye", goodbyeSrv.handleSayGoodbyeHTTP).Methods("GET", "POST", "OPTIONS")

	// Utility routes
	router.HandleFunc("/health", handleHealthCheck).Methods("GET")
	router.HandleFunc("/api/doc", handleAPIDoc).Methods("GET")

	// Root route
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		welcome := map[string]interface{}{
			"message": "Welcome to gRPC Sample Server",
			"port":    "50051",
			"protocols": map[string]string{
				"grpc": "localhost:50051 (use grpcurl)",
				"http": "localhost:50051 (use curl)",
			},
			"note":          "Both protocols are served on the same port",
			"documentation": "/api/doc",
			"health":        "/health",
		}

		json.NewEncoder(w).Encode(welcome)
	}).Methods("GET")

	return router
}

func main() {
	// Get port from environment variable or use default
	port := os.Getenv("GRPC_PORT")
	if port == "" {
		port = "50051"
	}

	// Create server instances
	helloSrv := &helloServer{}
	goodbyeSrv := &goodbyeServer{}

	// Create gRPC server
	grpcServer := grpc.NewServer()

	// Register both services
	hello.RegisterGreeterServer(grpcServer, helloSrv)
	goodbye.RegisterFarewellServer(grpcServer, goodbyeSrv)

	// Register reflection service on gRPC server
	reflection.Register(grpcServer)

	// Setup HTTP router
	httpHandler := setupHTTPRouter(helloSrv, goodbyeSrv)

	// Create multiplexed handler that can serve both gRPC and HTTP
	multiplexedHandler := createMultiplexedHandler(grpcServer, httpHandler)

	// Create HTTP server with the multiplexed handler
	server := &http.Server{
		Addr:    ":" + port,
		Handler: multiplexedHandler,
	}

	// Create listener
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", port, err)
	}

	log.Printf("🚀 Unified server starting on port %s", port)
	log.Printf("📋 Protocols supported:")
	log.Printf("   🔧 gRPC: localhost:%s (use grpcurl)", port)
	log.Printf("   🌐 HTTP: localhost:%s (use curl)", port)
	log.Printf("📋 Available gRPC services: Greeter (hello), Farewell (goodbye)")
	log.Printf("📋 Available HTTP endpoints:")
	log.Printf("   GET/POST /api/hello - Say hello")
	log.Printf("   GET/POST /api/goodbye - Say goodbye")
	log.Printf("   GET /health - Health check")
	log.Printf("   GET /api/doc - API documentation")
	log.Printf("   GET / - Welcome message")
	log.Printf("🔍 gRPC reflection enabled for grpcurl support")
	log.Printf("📖 Visit http://localhost:%s/api/doc for API documentation", port)
	log.Printf("🎯 Both protocols are served on the same port using protocol multiplexing!")

	// Start the unified server
	if err := server.Serve(lis); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to serve: %v", err)
	}
}
