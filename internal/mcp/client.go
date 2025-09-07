package mcp

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os/exec"
	"sync"
	"time"

	"proyecto-mcp-bolsa/pkg/models"
)

type Client struct {
	serverCommand []string
	serverCmd     *exec.Cmd
	stdin         io.WriteCloser
	stdout        io.ReadCloser
	stderr        io.ReadCloser
	encoder       *json.Encoder
	decoder       *json.Decoder
	nextID        int
	mu            sync.Mutex
	logger        *log.Logger
}

func NewClient(serverCommand []string, logger *log.Logger) *Client {
	return &Client{
		serverCommand: serverCommand,
		nextID:        1,
		logger:        logger,
	}
}

func (c *Client) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.serverCmd != nil {
		return fmt.Errorf("client already connected")
	}

	var cmd *exec.Cmd
	if len(c.serverCommand) == 1 {
		cmd = exec.Command(c.serverCommand[0])
	} else {
		cmd = exec.Command(c.serverCommand[0], c.serverCommand[1:]...)
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		stdin.Close()
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		stdin.Close()
		stdout.Close()
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		stdin.Close()
		stdout.Close()
		stderr.Close()
		return fmt.Errorf("failed to start server: %w", err)
	}


	time.Sleep(100 * time.Millisecond)
	
	if cmd.ProcessState != nil && cmd.ProcessState.Exited() {
		stdin.Close()
		stdout.Close()
		stderr.Close()
		return fmt.Errorf("server process exited immediately: %s", cmd.ProcessState.String())
	}

	c.serverCmd = cmd
	c.stdin = stdin
	c.stdout = stdout
	c.stderr = stderr
	c.encoder = json.NewEncoder(stdin)
	c.decoder = json.NewDecoder(stdout)

	c.logger.Printf("Connected to MCP server: %s", c.serverCommand[0])
	return nil
}

func (c *Client) Initialize() (*models.InitializeResponse, error) {
	request := models.JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      c.getNextID(),
		Method:  "initialize",
		Params: models.InitializeRequest{
			ProtocolVersion: "2024-11-05",
			Capabilities: models.ClientCapabilities{
				Experimental: make(map[string]interface{}),
			},
			ClientInfo: models.ClientInfo{
				Name:    "MCP Stock Analysis Client",
				Version: "1.0.0",
			},
		},
	}

	response, err := c.sendRequest(request)
	if err != nil {
		return nil, err
	}

	if response.Error != nil {
		return nil, fmt.Errorf("initialize error: %s", response.Error.Message)
	}

	var initResponse models.InitializeResponse
	resultBytes, _ := json.Marshal(response.Result)
	if err := json.Unmarshal(resultBytes, &initResponse); err != nil {
		return nil, fmt.Errorf("failed to parse initialize response: %w", err)
	}

	notification := models.JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "notifications/initialized",
	}

	if err := c.encoder.Encode(notification); err != nil {
		return nil, fmt.Errorf("failed to send initialized notification: %w", err)
	}

	c.logger.Printf("Successfully initialized with server version %s", initResponse.ProtocolVersion)
	return &initResponse, nil
}

func (c *Client) ListTools() ([]models.Tool, error) {
	request := models.JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      c.getNextID(),
		Method:  "tools/list",
	}

	response, err := c.sendRequest(request)
	if err != nil {
		return nil, err
	}

	if response.Error != nil {
		return nil, fmt.Errorf("list tools error: %s", response.Error.Message)
	}

	var listResponse models.ListToolsResponse
	resultBytes, _ := json.Marshal(response.Result)
	if err := json.Unmarshal(resultBytes, &listResponse); err != nil {
		return nil, fmt.Errorf("failed to parse list tools response: %w", err)
	}

	return listResponse.Tools, nil
}

func (c *Client) CallTool(name string, arguments map[string]interface{}) (*models.CallToolResponse, error) {
	request := models.JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      c.getNextID(),
		Method:  "tools/call",
		Params: models.CallToolRequest{
			Name:      name,
			Arguments: arguments,
		},
	}

	response, err := c.sendRequest(request)
	if err != nil {
		return nil, err
	}

	if response.Error != nil {
		return nil, fmt.Errorf("call tool error: %s", response.Error.Message)
	}

	var callResponse models.CallToolResponse
	resultBytes, _ := json.Marshal(response.Result)
	if err := json.Unmarshal(resultBytes, &callResponse); err != nil {
		return nil, fmt.Errorf("failed to parse call tool response: %w", err)
	}

	return &callResponse, nil
}

func (c *Client) sendRequest(request models.JSONRPCRequest) (*models.JSONRPCResponse, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.encoder == nil {
		return nil, fmt.Errorf("client not connected")
	}

	if err := c.encoder.Encode(request); err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	var response models.JSONRPCResponse
	if err := c.decoder.Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

func (c *Client) getNextID() int {
	id := c.nextID
	c.nextID++
	return id
}

func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.serverCmd == nil {
		return nil
	}

	c.stdin.Close()
	c.stdout.Close()
	c.stderr.Close()

	done := make(chan error, 1)
	go func() {
		done <- c.serverCmd.Wait()
	}()

	select {
	case err := <-done:
		c.logger.Printf("Server shut down: %v", err)
	case <-time.After(5 * time.Second):
		c.logger.Println("Server did not shut down gracefully, killing process")
		c.serverCmd.Process.Kill()
		<-done
	}

	c.serverCmd = nil
	c.logger.Printf("Disconnected from MCP server: %s", c.serverCommand[0])
	return nil
}
