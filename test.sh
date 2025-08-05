#!/bin/bash

# gRPC Sample Server Test Script
# This script demonstrates all gRPC methods using grpcurl

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Server configuration
SERVER="localhost:50051"

echo -e "${BLUE}=== gRPC Sample Server Test Script ===${NC}"
echo -e "${YELLOW}Testing server at: $SERVER${NC}"
echo ""

# Function to check if server is running
check_server() {
    echo -e "${YELLOW}Checking if server is running...${NC}"
    if ! grpcurl -plaintext $SERVER list > /dev/null 2>&1; then
        echo -e "${RED}Error: gRPC server is not running at $SERVER${NC}"
        echo -e "${YELLOW}Please start the server first: make server${NC}"
        exit 1
    fi
    echo -e "${GREEN}✓ Server is running${NC}"
    echo ""
}

# Function to run a test with description
run_test() {
    local description="$1"
    local command="$2"
    
    echo -e "${BLUE}--- $description ---${NC}"
    echo -e "${YELLOW}Command: $command${NC}"
    echo ""
    
    eval "$command"
    
    echo ""
    echo -e "${GREEN}✓ Test completed${NC}"
    echo ""
    sleep 1
}

# Check if grpcurl is installed
if ! command -v grpcurl &> /dev/null; then
    echo -e "${RED}Error: grpcurl is not installed${NC}"
    echo -e "${YELLOW}Install it with: go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest${NC}"
    exit 1
fi

# Check server
check_server

echo -e "${BLUE}=== Service Discovery ===${NC}"

# List all services
run_test "List all available services" \
    "grpcurl -plaintext $SERVER list"

# List methods for Hello service
run_test "List methods for Hello service" \
    "grpcurl -plaintext $SERVER list grpc.hello.Greeter"

# List methods for Goodbye service
run_test "List methods for Goodbye service" \
    "grpcurl -plaintext $SERVER list grpc.goodbye.Farewell"

# Describe Hello service
run_test "Describe Hello service" \
    "grpcurl -plaintext $SERVER describe grpc.hello.Greeter"

# Describe Goodbye service
run_test "Describe Goodbye service" \
    "grpcurl -plaintext $SERVER describe grpc.goodbye.Farewell"

echo -e "${BLUE}=== Hello Service Tests ===${NC}"

# Test Hello Service - Unary RPC
run_test "Hello Service - Unary RPC" \
    "grpcurl -plaintext -d '{\"name\":\"grpcurl\"}' $SERVER grpc.hello.Greeter/SayHello"

# Test Hello Service - Server Streaming RPC
run_test "Hello Service - Server Streaming RPC" \
    "grpcurl -plaintext -d '{\"name\":\"grpcurl\"}' $SERVER grpc.hello.Greeter/SayHelloStream"

# Test Hello Service - Client Streaming RPC
echo -e "${BLUE}--- Hello Service - Client Streaming RPC ---${NC}"
echo -e "${YELLOW}Command: echo with multiple names | grpcurl client streaming${NC}"
echo ""

echo '{"name":"Alice"}
{"name":"Bob"}
{"name":"Charlie"}' | grpcurl -plaintext -d @ $SERVER grpc.hello.Greeter/SayHelloClientStream

echo ""
echo -e "${GREEN}✓ Test completed${NC}"
echo ""
sleep 1

# Test Hello Service - Bidirectional Streaming RPC
echo -e "${BLUE}--- Hello Service - Bidirectional Streaming RPC ---${NC}"
echo -e "${YELLOW}Command: echo with multiple names | grpcurl bidirectional streaming${NC}"
echo ""

echo '{"name":"Emma"}
{"name":"Frank"}
{"name":"Grace"}' | grpcurl -plaintext -d @ $SERVER grpc.hello.Greeter/SayHelloBidirectional

echo ""
echo -e "${GREEN}✓ Test completed${NC}"
echo ""
sleep 1

echo -e "${BLUE}=== Goodbye Service Tests ===${NC}"

# Test Goodbye Service - Unary RPC
run_test "Goodbye Service - Unary RPC" \
    "grpcurl -plaintext -d '{\"name\":\"grpcurl\"}' $SERVER grpc.goodbye.Farewell/SayGoodbye"

# Test Goodbye Service - Server Streaming RPC
run_test "Goodbye Service - Server Streaming RPC" \
    "grpcurl -plaintext -d '{\"name\":\"grpcurl\"}' $SERVER grpc.goodbye.Farewell/SayGoodbyeStream"

# Test Goodbye Service - Client Streaming RPC
echo -e "${BLUE}--- Goodbye Service - Client Streaming RPC ---${NC}"
echo -e "${YELLOW}Command: echo with multiple names | grpcurl client streaming${NC}"
echo ""

echo '{"name":"Helen"}
{"name":"Ivan"}
{"name":"Julia"}
{"name":"Kevin"}
{"name":"Luna"}' | grpcurl -plaintext -d @ $SERVER grpc.goodbye.Farewell/SayGoodbyeClientStream

echo ""
echo -e "${GREEN}✓ Test completed${NC}"
echo ""
sleep 1

# Test Goodbye Service - Bidirectional Streaming RPC
echo -e "${BLUE}--- Goodbye Service - Bidirectional Streaming RPC ---${NC}"
echo -e "${YELLOW}Command: echo with multiple names | grpcurl bidirectional streaming${NC}"
echo ""

echo '{"name":"Maya"}
{"name":"Noah"}
{"name":"Olivia"}
{"name":"Paul"}' | grpcurl -plaintext -d @ $SERVER grpc.goodbye.Farewell/SayGoodbyeBidirectional

echo ""
echo -e "${GREEN}✓ Test completed${NC}"
echo ""
sleep 1

echo -e "${BLUE}=== Advanced Tests ===${NC}"

# Test with custom metadata
run_test "Test with custom metadata" \
    "grpcurl -plaintext -H 'client-id: grpcurl-test' -d '{\"name\":\"grpcurl\"}' $SERVER grpc.hello.Greeter/SayHello"

# Test with verbose output
run_test "Test with verbose output (headers and trailers)" \
    "grpcurl -plaintext -v -d '{\"name\":\"grpcurl\"}' $SERVER grpc.hello.Greeter/SayHello"

echo -e "${BLUE}=== File-based Input Tests ===${NC}"

# Create input files for testing
echo -e "${YELLOW}Creating test input files...${NC}"

# Create input file for hello client streaming
cat > hello_names.json << EOF
{"name":"Alice"}
{"name":"Bob"}
{"name":"Charlie"}
{"name":"Diana"}
EOF

# Create input file for goodbye client streaming
cat > goodbye_names.json << EOF
{"name":"Helen"}
{"name":"Ivan"}
{"name":"Julia"}
{"name":"Kevin"}
{"name":"Luna"}
EOF

# Create input file for bidirectional streaming
cat > bidi_names.json << EOF
{"name":"Emma"}
{"name":"Frank"}
{"name":"Grace"}
EOF

echo -e "${GREEN}✓ Input files created${NC}"
echo ""

# Test with input files
run_test "Hello Client Streaming with input file" \
    "grpcurl -plaintext -d @ $SERVER grpc.hello.Greeter/SayHelloClientStream < hello_names.json"

run_test "Goodbye Client Streaming with input file" \
    "grpcurl -plaintext -d @ $SERVER grpc.goodbye.Farewell/SayGoodbyeClientStream < goodbye_names.json"

run_test "Hello Bidirectional Streaming with input file" \
    "grpcurl -plaintext -d @ $SERVER grpc.hello.Greeter/SayHelloBidirectional < bidi_names.json"

# Clean up input files
echo -e "${YELLOW}Cleaning up test files...${NC}"
rm -f hello_names.json goodbye_names.json bidi_names.json
echo -e "${GREEN}✓ Test files cleaned up${NC}"
echo ""

echo -e "${GREEN}=== All Tests Completed Successfully! ===${NC}"
echo -e "${YELLOW}Summary:${NC}"
echo -e "  • Service Discovery: ✓"
echo -e "  • Hello Service (4 methods): ✓"
echo -e "  • Goodbye Service (4 methods): ✓"
echo -e "  • Advanced Features: ✓"
echo -e "  • File-based Input: ✓"
echo ""
echo -e "${BLUE}Total: 8 RPC methods tested across 2 services${NC}"
