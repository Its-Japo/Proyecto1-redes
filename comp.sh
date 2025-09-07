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
