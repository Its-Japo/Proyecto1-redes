#!/bin/bash

# MCP Stock Analysis Project - Build Script
# Creates both chatbot host and stock analyzer MCP server executables

set -e  # Exit on any error

echo "Building MCP Stock Analysis Project..."
echo "========================================"

echo "Creating bin directory..."
mkdir -p bin

echo "Installing dependencies..."
go mod download

echo "Building chatbot host..."
go build -o bin/chatbot ./cmd/chatbot/
echo "Chatbot built successfully: bin/chatbot"

echo "Building stock analyzer MCP server..."
go build -o bin/stock-analyzer ./servers/stock-analyzer/
echo "Stock analyzer built successfully: bin/stock-analyzer"

echo ""
echo "Build complete!"
echo "==============="
echo ""
echo "Usage Options:"
echo ""
echo "1. LOCAL MODE (default):"
echo "   ./bin/chatbot                    # Auto-connects to local stock analyzer"
echo "   ./bin/chatbot -no-auto-connect   # Manual connection mode"
echo ""
echo "2. NETWORK MODE (new):"
echo "   ./bin/stock-analyzer 8080        # Start server on port 8080"
echo "   ./bin/chatbot -no-auto-connect    # Then: /connect tcp://host:port"
echo ""
echo "3. CONNECTION OPTIONS:"
echo "   /connect ./bin/stock-analyzer           # Local process"
echo "   /connect tcp://localhost:8080           # Local network"
echo "   /connect tcp://192.168.1.100:8080       # Remote network"
echo "   /connect ./servers/stock-analyzer/      # Run from source"
echo ""
echo "4. ENVIRONMENT SETUP:"
echo "   export ALPHA_VANTAGE_API_KEY=your_key   # For real data"
echo "   export ANTHROPIC_API_KEY=your_key       # For Claude AI"
echo ""
echo "Network server binds to 0.0.0.0:port for remote access"
echo "Default mode uses stdin/stdout IPC (not visible in Wireshark)"
