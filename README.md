# gRPC Sample Server with HTTP REST API

This is a comprehensive dual-protocol server written in Go that supports both gRPC and HTTP REST API, demonstrating multiple services, various RPC patterns, and detailed response information handling.

## Project Structure

```
grpc-sample/
├── proto/
│   ├── hello/
│   │   ├── hello.proto         # Hello service Protocol Buffer definition
│   │   ├── hello.pb.go         # Generated Go code for hello messages
│   │   └── hello_grpc.pb.go    # Generated Go code for hello gRPC service
│   └── goodbye/
│       ├── goodbye.proto       # Goodbye service Protocol Buffer definition
│       ├── goodbye.pb.go       # Generated Go code for goodbye messages
│       └── goodbye_grpc.pb.go  # Generated Go code for goodbye gRPC service
├── server/
│   └── main.go                 # gRPC server implementation (both services)
├── client/
│   └── main.go                 # gRPC client implementation (both services)
├── go.mod                      # Go module file
├── Makefile                    # Build automation
└── README.md                   # This file
```

## Features

This dual-protocol server supports both gRPC and HTTP REST API protocols running concurrently:

### Unified Protocol Support
- **Single Port**: Both gRPC and HTTP protocols run on port 50051
- **Protocol Multiplexing**: Automatic detection of gRPC vs HTTP requests
- **gRPC Server**: Full gRPC functionality with all streaming patterns
- **HTTP REST API**: JSON request/response with GET/POST support
- **Shared Business Logic**: HTTP endpoints internally call gRPC methods
- **CORS Support**: Cross-origin requests enabled for web applications

### HTTP REST API Endpoints
- **GET/POST /api/hello**: Say hello (query param or JSON body)
- **GET/POST /api/goodbye**: Say goodbye (query param or JSON body)
- **GET /health**: Health check endpoint
- **GET /api/doc**: API documentation
- **GET /**: Welcome message with server information

The sample includes two separate gRPC services:

### Hello Service (Greeter)
1. **Unary RPC**: `SayHello` - Simple request/response
2. **Server Streaming RPC**: `SayHelloStream` - Server sends 5 messages with 1-second intervals
3. **Client Streaming RPC**: `SayHelloClientStream` - Client sends multiple names, server responds with summary
4. **Bidirectional Streaming RPC**: `SayHelloBidirectional` - Real-time exchange of greetings

### Goodbye Service (Farewell)
1. **Unary RPC**: `SayGoodbye` - Simple goodbye message
2. **Server Streaming RPC**: `SayGoodbyeStream` - Server sends 3 farewell messages with 1.5-second intervals
3. **Client Streaming RPC**: `SayGoodbyeClientStream` - Client sends multiple names, server responds with collective farewell
4. **Bidirectional Streaming RPC**: `SayGoodbyeBidirectional` - Interactive farewell exchange with personalized messages

### Enhanced Response Information
- **gRPC Status Codes**: Complete status information including error details
- **Response Headers**: Custom metadata sent by server (server-name, method, timestamps, etc.)
- **Response Trailers**: Trailing metadata (processing info, timing, completion status)
- **Response Size**: Byte count of response messages
- **Connection State**: Target address and connection status
- **Stream Information**: Headers, trailers, message counts, and completion tracking

## Prerequisites

- Go 1.21 or later
- Protocol Buffers compiler (protoc) - optional, as generated files are included
- Docker (for containerization)

## Running the Server

1. Navigate to the project directory:
   ```bash
   cd grpc-sample
   ```

2. Install dependencies:
   ```bash
   make deps
   ```

3. Run the server:
   ```bash
   make server
   ```

   The server will start listening on `0.0.0.0:50051` (all interfaces) and register both services.

## Running the Client

In a separate terminal:

```bash
cd grpc-sample
make client
```

The client will demonstrate all four gRPC streaming patterns:
1. **Unary RPC**: `SayHello` and `SayGoodbye` - Simple request/response
2. **Server Streaming**: `SayHelloStream` (5 messages) and `SayGoodbyeStream` (3 messages)
3. **Client Streaming**: `SayHelloClientStream` (sends 4 names) and `SayGoodbyeClientStream` (sends 5 names)
4. **Bidirectional Streaming**: `SayHelloBidirectional` (3 exchanges) and `SayGoodbyeBidirectional` (4 exchanges)

## Expected Output

**Server output:**
```
2024/01/01 12:00:00 gRPC server listening at [::]:50051
2024/01/01 12:00:00 Available services: Greeter (hello), Farewell (goodbye)
2024/01/01 12:00:01 Received: World
2024/01/01 12:00:01 Incoming metadata:
2024/01/01 12:00:01   client-id: [grpc-sample-client]
2024/01/01 12:00:01 Received stream request for: World
2024/01/01 12:00:01 Stream incoming metadata:
2024/01/01 12:00:01   stream-client-id: [grpc-sample-stream]
2024/01/01 12:00:06 Received goodbye request for: World
2024/01/01 12:00:06 Goodbye incoming metadata:
2024/01/01 12:00:06   goodbye-client-id: [grpc-sample-goodbye]
2024/01/01 12:00:06 Received goodbye stream request for: World
2024/01/01 12:00:06 Goodbye stream incoming metadata:
2024/01/01 12:00:06   goodbye-stream-client-id: [grpc-sample-goodbye-stream]
2024/01/01 12:00:06 Sent goodbye message 1: Thanks for using our service, World!
2024/01/01 12:00:07 Sent goodbye message 2: It was great having you, World!
2024/01/01 12:00:09 Sent goodbye message 3: Until we meet again, World! Farewell!
```

**Client output:**
```
2024/01/01 12:00:01 Calling SayHello with name: World
2024/01/01 12:00:01 Greeting: Hello World
2024/01/01 12:00:01 === SayHello Response Info ===
2024/01/01 12:00:01 Status Code: OK
2024/01/01 12:00:01 Status Message: Success
2024/01/01 12:00:01 Response Headers:
2024/01/01 12:00:01   server-name: [grpc-sample-server]
2024/01/01 12:00:01   method: [SayHello]
2024/01/01 12:00:01   timestamp: [2024-01-01T12:00:01Z]
2024/01/01 12:00:01   response-id: [hello-1234567890]
2024/01/01 12:00:01 Response Trailers:
2024/01/01 12:00:01   processing-time: [fast]
2024/01/01 12:00:01   server-version: [1.0.0]
2024/01/01 12:00:01 Response Size: 11 bytes

2024/01/01 12:00:01 Calling SayHelloStream with name: World
2024/01/01 12:00:01 === SayHelloStream Headers ===
2024/01/01 12:00:01   server-name: [grpc-sample-server]
2024/01/01 12:00:01   method: [SayHelloStream]
2024/01/01 12:00:01   stream-id: [stream-1234567890]
2024/01/01 12:00:01   expected-messages: [5]
2024/01/01 12:00:01 Stream message 1: Hello World - Message 1
2024/01/01 12:00:02 Stream message 2: Hello World - Message 2
2024/01/01 12:00:03 Stream message 3: Hello World - Message 3
2024/01/01 12:00:04 Stream message 4: Hello World - Message 4
2024/01/01 12:00:05 Stream message 5: Hello World - Message 5
2024/01/01 12:00:05 === SayHelloStream Trailers ===
2024/01/01 12:00:05   messages-sent: [5]
2024/01/01 12:00:05   stream-duration: [5s]
2024/01/01 12:00:05   stream-status: [completed]

2024/01/01 12:00:06 Calling SayGoodbye with name: World
2024/01/01 12:00:06 Goodbye message: Goodbye World! See you later!
2024/01/01 12:00:06 === SayGoodbye Response Info ===
2024/01/01 12:00:06 Status Code: OK
2024/01/01 12:00:06 Status Message: Success
2024/01/01 12:00:06 Response Headers:
2024/01/01 12:00:06   server-name: [grpc-sample-server]
2024/01/01 12:00:06   method: [SayGoodbye]
2024/01/01 12:00:06   farewell-type: [friendly]
2024/01/01 12:00:06 Response Trailers:
2024/01/01 12:00:06   goodbye-processed: [true]
2024/01/01 12:00:06   session-ended: [2024-01-01T12:00:06Z]

2024/01/01 12:00:06 Calling SayGoodbyeStream with name: World
2024/01/01 12:00:06 === SayGoodbyeStream Headers ===
2024/01/01 12:00:06   server-name: [grpc-sample-server]
2024/01/01 12:00:06   method: [SayGoodbyeStream]
2024/01/01 12:00:06   stream-id: [goodbye-stream-1234567890]
2024/01/01 12:00:06   expected-messages: [3]
2024/01/01 12:00:06   farewell-type: [streaming]
2024/01/01 12:00:06 Goodbye stream message 1: Thanks for using our service, World!
2024/01/01 12:00:07 Goodbye stream message 2: It was great having you, World!
2024/01/01 12:00:09 Goodbye stream message 3: Until we meet again, World! Farewell!
2024/01/01 12:00:09 === SayGoodbyeStream Trailers ===
2024/01/01 12:00:09   messages-sent: [3]
2024/01/01 12:00:09   stream-duration: [4.5s]
2024/01/01 12:00:09   stream-status: [completed]
2024/01/01 12:00:09   farewell-completed: [2024-01-01T12:00:09Z]

=== Connection Info ===
Target: localhost:50051
Connection State: READY
```

## Available Make Targets

```bash
make help         # Show available targets
make deps         # Download dependencies
make server       # Run the gRPC server
make client       # Run the gRPC client
make test         # Run comprehensive grpcurl tests
make proto        # Regenerate protocol buffer code
make build        # Build binaries
make clean        # Clean build artifacts
```

## Regenerating Protocol Buffer Code

If you modify the `.proto` files, you can regenerate the Go code using:

```bash
make proto
```

Or manually:
```bash
# Install protoc-gen-go and protoc-gen-go-grpc
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Generate Go code
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/hello/hello.proto proto/goodbye/goodbye.proto
```

## Testing Both Protocols

The server supports comprehensive testing of both gRPC and HTTP REST API protocols.

### Dual-Protocol Test Script

Run the comprehensive test script that tests both protocols:

```bash
# Make the script executable (if needed)
chmod +x test-dual-protocol.sh

# Run comprehensive tests for both gRPC and HTTP
./test-dual-protocol.sh
```

The `test-dual-protocol.sh` script provides:

- **Protocol Verification**: Tests both gRPC and HTTP endpoints
- **Service Discovery**: Lists all gRPC services and HTTP routes
- **All RPC Patterns**: Tests all gRPC streaming patterns
- **HTTP REST API**: Tests all HTTP endpoints with GET/POST methods
- **CORS Testing**: Verifies cross-origin request support
- **Performance Testing**: Basic performance comparison between protocols
- **Colored Output**: Easy-to-read test results with status indicators

### Testing gRPC Protocol

```bash
# Install grpcurl (if not already installed)
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# List all services
grpcurl -plaintext localhost:50051 list

# Test Hello service
grpcurl -plaintext -d '{"name":"gRPC-Test"}' localhost:50051 grpc.hello.Greeter/SayHello

# Test Goodbye service
grpcurl -plaintext -d '{"name":"gRPC-Friend"}' localhost:50051 grpc.goodbye.Farewell/SayGoodbye

# Test server streaming
grpcurl -plaintext -d '{"name":"Stream-Test"}' localhost:50051 grpc.hello.Greeter/SayHelloStream

# Test with verbose output to see headers and trailers
grpcurl -plaintext -v -d '{"name":"Verbose-Test"}' localhost:50051 grpc.hello.Greeter/SayHello
```

### Testing HTTP REST API

```bash
# Test root endpoint
curl http://localhost:50051/

# Test health check
curl http://localhost:50051/health

# Test API documentation
curl http://localhost:50051/api/doc

# Test Hello endpoint (GET)
curl "http://localhost:50051/api/hello?name=HTTP-Test"

# Test Hello endpoint (POST)
curl -X POST -H "Content-Type: application/json" \
     -d '{"name":"HTTP-POST-Test"}' \
     http://localhost:50051/api/hello

# Test Goodbye endpoint (GET)
curl "http://localhost:50051/api/goodbye?name=HTTP-Friend"

# Test Goodbye endpoint (POST)
curl -X POST -H "Content-Type: application/json" \
     -d '{"name":"HTTP-POST-Friend"}' \
     http://localhost:50051/api/goodbye

# Test with formatted JSON output (requires jq)
curl -s http://localhost:50051/api/hello?name=World | jq '.'
```

### Expected HTTP Responses

**Hello endpoint response:**
```json
{
  "message": "Hello World"
}
```

**Goodbye endpoint response:**
```json
{
  "message": "Goodbye Friend! See you later!"
}
```

**Health check response:**
```json
{
  "status": "healthy",
  "timestamp": "2024-01-01T12:00:00Z",
  "services": {
    "grpc": "running on :50051",
    "http": "running on :50051 (same port)"
  },
  "version": "1.0.0",
  "note": "Both gRPC and HTTP protocols are served on the same port"
}
```

**API documentation response:**
```json
{
  "title": "gRPC Sample Server API",
  "version": "1.0.0",
  "description": "Unified server supporting both gRPC and HTTP REST APIs on the same port",
  "protocols": ["gRPC", "HTTP"],
  "port": "50051",
  "note": "Both protocols are served on the same port using protocol multiplexing",
  "endpoints": {
    "grpc": {
      "address": ":50051",
      "services": [...]
    },
    "http": {
      "address": ":50051",
      "routes": [...]
    }
  },
  "examples": {
    "grpc": {...},
    "http": {...}
  }
}
```

### Testing CORS Support

```bash
# Test CORS preflight request
curl -X OPTIONS -H "Origin: http://example.com" \
     -H "Access-Control-Request-Method: POST" \
     -H "Access-Control-Request-Headers: Content-Type" \
     -I http://localhost:50051/api/hello

# Expected CORS headers in response:
# Access-Control-Allow-Origin: *
# Access-Control-Allow-Methods: GET, POST, OPTIONS
# Access-Control-Allow-Headers: Content-Type
```

### Performance Testing

```bash
# Test HTTP endpoint performance
time for i in {1..10}; do
  curl -s "http://localhost:50051/api/hello?name=Perf-Test-$i" > /dev/null
done

# Test gRPC endpoint performance
time for i in {1..10}; do
  grpcurl -plaintext -d "{\"name\":\"gRPC-Perf-Test-$i\"}" \
    localhost:50051 grpc.hello.Greeter/SayHello > /dev/null
done
```

## Architecture Highlights

### Separate Services
- **Modular Design**: Hello and Goodbye services are completely separate with their own proto files
- **Independent Packages**: Each service has its own Go package for better organization
- **Service Registration**: Both services are registered on the same gRPC server

### Enhanced Metadata Handling
- **Request Metadata**: Client sends custom metadata with each request
- **Response Headers**: Server sends custom headers with method info, timestamps, and identifiers
- **Response Trailers**: Server sends trailing metadata with processing info and completion status
- **Stream Metadata**: Special handling for streaming RPCs with stream-specific metadata

### Comprehensive Response Information
- **Status Tracking**: Complete gRPC status code and message handling
- **Size Reporting**: Response payload size tracking
- **Connection State**: Real-time connection state monitoring
- **Error Handling**: Proper gRPC error status extraction and display

## Streaming Patterns Demonstrated

This sample showcases all four gRPC streaming patterns:

### 1. Unary RPC
- **Pattern**: Single request → Single response
- **Examples**: `SayHello`, `SayGoodbye`
- **Use Cases**: Simple operations, authentication, configuration retrieval

### 2. Server Streaming RPC
- **Pattern**: Single request → Multiple responses
- **Examples**: `SayHelloStream` (5 messages), `SayGoodbyeStream` (3 messages)
- **Use Cases**: Data feeds, real-time updates, file downloads

### 3. Client Streaming RPC
- **Pattern**: Multiple requests → Single response
- **Examples**: `SayHelloClientStream`, `SayGoodbyeClientStream`
- **Use Cases**: File uploads, batch processing, data aggregation

### 4. Bidirectional Streaming RPC
- **Pattern**: Multiple requests ↔ Multiple responses
- **Examples**: `SayHelloBidirectional`, `SayGoodbyeBidirectional`
- **Use Cases**: Chat applications, real-time collaboration, live data processing

