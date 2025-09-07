# MCP Stock Analysis Project

A comprehensive implementation of the Model Context Protocol (MCP) in Go, featuring a stock market analysis system with real-time data integration and AI-powered chatbot interface.

## 🎯 Project Overview

This project demonstrates the implementation of MCP (Model Context Protocol) for a university networks course (CC3067 Redes - UVG). It consists of:

- **MCP Chatbot Host**: Interactive chatbot with Claude AI integration
- **Stock Analyzer MCP Server**: Local server providing stock analysis tools
- **Financial API Integration**: Real-time market data from Alpha Vantage
- **Technical Analysis Engine**: RSI, moving averages, MACD, Bollinger Bands
- **Investment Recommendations**: AI-driven buy/sell/hold recommendations

## 🏗️ Architecture

```
┌─────────────────┐    JSON-RPC     ┌──────────────────┐    HTTP API    ┌─────────────────┐
│   Chatbot Host  │ ◄────────────► │ Stock Analyzer   │ ◄───────────► │ Alpha Vantage   │
│   (Claude AI)   │                │   MCP Server     │                │  Financial API  │
└─────────────────┘                └──────────────────┘                └─────────────────┘
```

## 🚀 Quick Start

### Prerequisites

```bash
# Install Go 1.21+
go version

# Get API keys (free)
# 1. Alpha Vantage: https://www.alphavantage.co/support/#api-key
# 2. Anthropic Claude: https://console.anthropic.com/
```

### Setup

```bash
# Set environment variables
export ALPHA_VANTAGE_API_KEY="your_alpha_vantage_key"
export ANTHROPIC_API_KEY="your_anthropic_key"

# Install dependencies
go mod download

# Build executables
./setup.sh

# Run the chatbot (auto-connects to MCP server)
./bin/chatbot

# Or run without auto-connect for manual control
./bin/chatbot -no-auto-connect
```

## 💡 Usage Examples

### Interactive Commands

```bash
🤖 MCP Stock Analysis Chatbot
===============================

💬 You: /analyze AAPL,GOOGL,MSFT
📊 Analyzing portfolio: [AAPL GOOGL MSFT]
✅ Analysis complete:

📊 PORTFOLIO ANALYSIS REPORT
========================================
Portfolio: Analysis Portfolio
Overall Score: 72.3/100
Overall Risk: MEDIUM

🏢 AAPL
  Price: $185.64 (-1.23%)
  Recommendation: BUY (Score: 75.0/100)
  Risk Level: LOW
```

### Natural Language Queries

```bash
💬 You: Should I invest in Tesla stock?
🔍 Detected stock symbols: [TSLA]
📈 STOCK ANALYSIS: TSLA
==============================
Current Price: $248.42
Recommendation: HOLD (Score: 58.0/100)
Risk Level: HIGH
```

## 🛠️ MCP Tools Available

| Tool | Description | Parameters |
|------|-------------|------------|
| `analyze_portfolio` | Analyze multiple stocks with recommendations | `symbols[]`, `timeframe` |
| `get_stock_price` | Get current price and technical analysis | `symbol` |
| `export_analysis` | Export results to CSV/JSON | `format`, `filename` |

### Connection Management Commands

| Command | Description |
|---------|-------------|
| `/status` | Show connection status and health check |
| `/connect <path>` | Connect to MCP server manually |
| `/disconnect <name>` | Disconnect from MCP server |
| `/list` | List available tools from connected servers |

## 📊 Technical Features

### Financial Analysis
- **Technical Indicators**: RSI, SMA, EMA, MACD, Bollinger Bands
- **Risk Assessment**: Volatility analysis and risk scoring
- **Recommendation Engine**: Multi-factor scoring system
- **Portfolio Analytics**: Diversification analysis

### MCP Implementation
- **Pure JSON-RPC 2.0**: No external MCP SDK dependencies
- **Streaming Protocol**: Real-time bidirectional communication
- **Error Handling**: Comprehensive error responses
- **Tool Discovery**: Dynamic tool registration and listing

## 🔧 Development

### Project Structure

```
proyecto-mcp-bolsa/
├── cmd/chatbot/           # Chatbot host application
├── internal/
│   ├── mcp/              # MCP protocol implementation
│   ├── stock/            # Stock analysis logic
│   └── llm/              # Claude AI client
├── pkg/models/           # Data structures
├── servers/stock-analyzer/  # MCP server implementation
├── examples/scenarios/   # Usage examples
└── config.yaml          # Configuration
```

### Building

```bash
# Create bin directory
mkdir -p bin

# Build chatbot
go build -o bin/chatbot ./cmd/chatbot/

# Build stock analyzer server
go build -o bin/stock-analyzer ./servers/stock-analyzer/

# Run tests
go test ./...
```

## 🌐 Network Analysis

### Protocol Layers (OSI Model)

1. **Application Layer (7)**: MCP protocol, JSON-RPC 2.0
2. **Presentation Layer (6)**: JSON serialization, UTF-8 encoding
3. **Session Layer (5)**: HTTP/HTTPS sessions
4. **Transport Layer (4)**: TCP for reliable communication
5. **Network Layer (3)**: IP routing for API calls
6. **Data Link Layer (2)**: Ethernet framing
7. **Physical Layer (1)**: Network hardware

## 📚 MCP Protocol Compliance

This implementation follows the official MCP specification:
- Protocol version: 2024-11-05
- JSON-RPC 2.0 transport
- Standard initialization flow
- Tool discovery and execution
- Error handling conventions

## 🔗 References

- [MCP Specification](https://modelcontextprotocol.io/specification/2025-06-18)
- [JSON-RPC 2.0 Specification](https://www.jsonrpc.org/specification)
- [Alpha Vantage API Documentation](https://www.alphavantage.co/documentation/)
- [Anthropic Claude API](https://docs.anthropic.com/)

---

**Course**: CC3067 Redes - Universidad del Valle de Guatemala