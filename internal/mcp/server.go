package mcp

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"proyecto-mcp-bolsa/pkg/models"
)

type Server struct {
	name         string
	version      string
	capabilities models.ServerCapabilities
	tools        map[string]ToolHandler
	logger       *log.Logger
}

type ToolHandler interface {
	Handle(args map[string]interface{}) (*models.CallToolResponse, error)
}

type ToolHandlerFunc func(args map[string]interface{}) (*models.CallToolResponse, error)

func (f ToolHandlerFunc) Handle(args map[string]interface{}) (*models.CallToolResponse, error) {
	return f(args)
}

func NewServer(name, version string) *Server {
	return &Server{
		name:    name,
		version: version,
		capabilities: models.ServerCapabilities{
			Tools: &models.ToolsCapability{
				ListChanged: false,
			},
			Logging: &models.LoggingCapability{},
		},
		tools:  make(map[string]ToolHandler),
		logger: log.New(os.Stderr, fmt.Sprintf("[%s] ", name), log.LstdFlags),
	}
}

func (s *Server) RegisterTool(name, description string, inputSchema json.RawMessage, handler ToolHandler) {
	s.tools[name] = handler
	s.logger.Printf("Registered tool: %s", name)
}

func (s *Server) HandleRequest(input io.Reader, output io.Writer) error {
	decoder := json.NewDecoder(input)
	encoder := json.NewEncoder(output)

	for {
		var request models.JSONRPCRequest
		if err := decoder.Decode(&request); err != nil {
			if err == io.EOF {
				s.logger.Println("Client disconnected")
				return nil
			}
			return s.sendError(encoder, nil, -32700, "Parse error", err.Error())
		}

		s.logger.Printf("Received request: %s", request.Method)

		var err error
		switch request.Method {
		case "initialize":
			err = s.handleInitialize(encoder, request)
		case "tools/list":
			err = s.handleListTools(encoder, request)
		case "tools/call":
			err = s.handleCallTool(encoder, request)
		case "notifications/initialized":
			err = s.handleInitialized(encoder, request)
		default:
			err = s.sendError(encoder, request.ID, -32601, "Method not found", request.Method)
		}

		if err != nil {
			s.logger.Printf("Error handling request %s: %v", request.Method, err)
			return err
		}
	}
}

func (s *Server) handleInitialize(encoder *json.Encoder, request models.JSONRPCRequest) error {
	var initReq models.InitializeRequest
	if request.Params != nil {
		paramsBytes, _ := json.Marshal(request.Params)
		if err := json.Unmarshal(paramsBytes, &initReq); err != nil {
			return s.sendError(encoder, request.ID, -32602, "Invalid params", err.Error())
		}
	}

	response := models.JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      request.ID,
		Result: models.InitializeResponse{
			ProtocolVersion: "2024-11-05",
			Capabilities:    s.capabilities,
			ServerInfo: models.ServerInfo{
				Name:    s.name,
				Version: s.version,
			},
		},
	}

	return encoder.Encode(response)
}

func (s *Server) handleListTools(encoder *json.Encoder, request models.JSONRPCRequest) error {
	tools := make([]models.Tool, 0, len(s.tools))
	
	for name := range s.tools {
		var inputSchema json.RawMessage
		switch name {
		case "analyze_portfolio":
			inputSchema = json.RawMessage(`{
				"type": "object",
				"properties": {
					"symbols": {
						"type": "array",
						"items": {"type": "string"},
						"description": "Array of stock symbols to analyze"
					},
					"timeframe": {
						"type": "string",
						"description": "Timeframe for analysis (1D, 5D, 1M, 3M, 6M, 1Y)",
						"default": "1M"
					}
				},
				"required": ["symbols"]
			}`)
		case "get_stock_price":
			inputSchema = json.RawMessage(`{
				"type": "object",
				"properties": {
					"symbol": {
						"type": "string",
						"description": "Stock symbol to get price for"
					}
				},
				"required": ["symbol"]
			}`)
		case "export_analysis":
			inputSchema = json.RawMessage(`{
				"type": "object",
				"properties": {
					"format": {
						"type": "string",
						"enum": ["csv", "json"],
						"description": "Export format",
						"default": "json"
					},
					"filename": {
						"type": "string",
						"description": "Output filename"
					}
				},
				"required": ["filename"]
			}`)
		}

		tool := models.Tool{
			Name:        name,
			Description: s.getToolDescription(name),
			InputSchema: inputSchema,
		}
		tools = append(tools, tool)
	}

	response := models.JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      request.ID,
		Result:  models.ListToolsResponse{Tools: tools},
	}

	return encoder.Encode(response)
}

func (s *Server) handleCallTool(encoder *json.Encoder, request models.JSONRPCRequest) error {
	var callReq models.CallToolRequest
	if request.Params != nil {
		paramsBytes, _ := json.Marshal(request.Params)
		if err := json.Unmarshal(paramsBytes, &callReq); err != nil {
			return s.sendError(encoder, request.ID, -32602, "Invalid params", err.Error())
		}
	}

	handler, exists := s.tools[callReq.Name]
	if !exists {
		return s.sendError(encoder, request.ID, -32601, "Tool not found", callReq.Name)
	}

	s.logger.Printf("Calling tool: %s", callReq.Name)
	result, err := handler.Handle(callReq.Arguments)
	if err != nil {
		return s.sendError(encoder, request.ID, -32603, "Tool execution error", err.Error())
	}

	response := models.JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      request.ID,
		Result:  result,
	}

	return encoder.Encode(response)
}

func (s *Server) handleInitialized(encoder *json.Encoder, request models.JSONRPCRequest) error {
	s.logger.Println("Client initialized successfully")
	return nil
}

func (s *Server) sendError(encoder *json.Encoder, id interface{}, code int, message, data string) error {
	response := models.JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &models.JSONRPCError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}
	return encoder.Encode(response)
}

func (s *Server) getToolDescription(name string) string {
	descriptions := map[string]string{
		"analyze_portfolio": "Analyze a portfolio of stocks and provide investment recommendations",
		"get_stock_price":   "Get current stock price and basic information",
		"export_analysis":   "Export analysis results to CSV or JSON format",
	}
	return descriptions[name]
}

func (s *Server) Run() error {
	s.logger.Printf("Starting %s server version %s", s.name, s.version)
	return s.HandleRequest(os.Stdin, os.Stdout)
}

// RunOnPort starts the MCP server listening on a TCP port
func (s *Server) RunOnPort(port int) error {
	s.logger.Printf("Starting %s server version %s on port %d", s.name, s.version, port)
	
	// Bind to all interfaces (0.0.0.0:port)
	address := fmt.Sprintf("0.0.0.0:%d", port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", address, err)
	}
	defer listener.Close()

	s.logger.Printf("MCP server listening on %s (all interfaces)", listener.Addr().String())

	for {
		conn, err := listener.Accept()
		if err != nil {
			s.logger.Printf("Failed to accept connection: %v", err)
			continue
		}

		s.logger.Printf("New client connected from %s", conn.RemoteAddr())
		go s.handleConnection(conn)
	}
}

// handleConnection handles a single TCP connection
func (s *Server) handleConnection(conn net.Conn) {
	defer func() {
		s.logger.Printf("Client %s disconnected", conn.RemoteAddr())
		conn.Close()
	}()

	s.logger.Printf("Handling connection from %s", conn.RemoteAddr())
	
	if err := s.HandleRequest(conn, conn); err != nil {
		s.logger.Printf("Connection error from %s: %v", conn.RemoteAddr(), err)
	} else {
		s.logger.Printf("Connection from %s completed successfully", conn.RemoteAddr())
	}
}