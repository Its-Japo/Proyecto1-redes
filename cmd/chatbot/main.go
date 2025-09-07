package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"proyecto-mcp-bolsa/internal/llm"
	"proyecto-mcp-bolsa/internal/mcp"
)

type ChatbotHost struct {
	claudeClient *llm.ClaudeClient
	mcpClients   map[string]*mcp.Client
	logger       *log.Logger
	conversation []llm.Message
}

func NewChatbotHost() *ChatbotHost {
	logger := log.New(os.Stderr, "[CHATBOT] ", log.LstdFlags)

	claudeAPIKey := os.Getenv("ANTHROPIC_API_KEY")
	if claudeAPIKey == "" {
		logger.Println("Warning: ANTHROPIC_API_KEY not set, Claude integration will not work")
	}

	claudeClient := llm.NewClaudeClient(claudeAPIKey, "", "claude-3-haiku-20240307")

	return &ChatbotHost{
		claudeClient: claudeClient,
		mcpClients:   make(map[string]*mcp.Client),
		logger:       logger,
		conversation: make([]llm.Message, 0),
	}
}

func (c *ChatbotHost) Start() error {
	c.logger.Println("Starting MCP Chatbot Host...")
	
	fmt.Println("MCP Stock Analysis Chatbot")
	fmt.Println("===============================")
	fmt.Println("Available commands:")
	fmt.Println("  /connect <server_path>  - Connect to MCP server")
	fmt.Println("  /disconnect <server>    - Disconnect from MCP server")
	fmt.Println("  /status                 - Show connection status")
	fmt.Println("  /list                   - List available tools")
	fmt.Println("  /analyze <symbols>      - Advanced portfolio analysis with reliability")
	fmt.Println("  /predict <symbol>       - Get price predictions with confidence intervals")
	fmt.Println("  /trends <symbol>        - Analyze historical trends and patterns")
	fmt.Println("  /price <symbol>         - Get enhanced stock analysis")
	fmt.Println("  /help                   - Show help")
	fmt.Println("  /quit                   - Exit chatbot")
	fmt.Println()

	noAutoConnect := flag.Bool("no-auto-connect", false, "Disable auto-connection to stock analyzer")
	flag.Parse()
	
	if !*noAutoConnect {
		fmt.Println("Auto-connecting to stock analyzer...")
		c.connectToStockServer()
	} else {
		fmt.Println("Auto-connect disabled. Use /connect to connect manually.")
	}

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("You: ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		if err := c.processInput(input); err != nil {
			fmt.Printf("Error: %v\n", err)
		}

		if input == "/quit" {
			break
		}
	}

	c.cleanup()
	return nil
}

func (c *ChatbotHost) processInput(input string) error {
	c.logInteraction("USER", input)

	if strings.HasPrefix(input, "/") {
		return c.handleCommand(input)
	}

	return c.handleConversation(input)
}

func (c *ChatbotHost) handleCommand(input string) error {
	parts := strings.Fields(input)
	command := parts[0]

	switch command {
	case "/connect":
		if len(parts) < 2 {
			fmt.Println("Usage: /connect <server_path>")
			return nil
		}
		return c.connectToMCPServer(parts[1])

	case "/disconnect":
		if len(parts) < 2 {
			fmt.Println("Usage: /disconnect <server_name>")
			return nil
		}
		return c.disconnectFromMCPServer(parts[1])

	case "/status":
		return c.showConnectionStatus()

	case "/list":
		return c.listAvailableTools()

	case "/analyze":
		if len(parts) < 2 {
			fmt.Println("Usage: /analyze AAPL,GOOGL,MSFT")
			return nil
		}
		symbols := strings.Split(parts[1], ",")
		return c.analyzePortfolioAdvanced(symbols)

	case "/predict":
		if len(parts) < 2 {
			fmt.Println("Usage: /predict AAPL")
			return nil
		}
		return c.getPricePrediction(parts[1])

	case "/trends":
		if len(parts) < 2 {
			fmt.Println("Usage: /trends AAPL")
			return nil
		}
		return c.analyzeHistoricalTrends(parts[1])

	case "/price":
		if len(parts) < 2 {
			fmt.Println("Usage: /price AAPL")
			return nil
		}
		return c.getEnhancedStockPrice(parts[1])

	case "/help":
		return c.showHelp()

	case "/quit":
		fmt.Println("Goodbye!")
		return nil

	default:
		fmt.Printf("Unknown command: %s\n", command)
		return nil
	}
}

func (c *ChatbotHost) handleConversation(input string) error {
	c.conversation = append(c.conversation, llm.Message{
		Role:    "user",
		Content: input,
	})

	if c.isStockRelatedQuery(input) {
		return c.handleStockQuery(input)
	}

	response, err := c.claudeClient.SendMessage(c.conversation)
	if err != nil {
		return fmt.Errorf("Claude API error: %w", err)
	}

	if len(response.Content) > 0 {
		reply := response.Content[0].Text
		fmt.Printf("Claude: %s\n", reply)
		
		c.conversation = append(c.conversation, llm.Message{
			Role:    "assistant",
			Content: reply,
		})

		c.logInteraction("CLAUDE", reply)
	}

	return nil
}

func (c *ChatbotHost) isStockRelatedQuery(input string) bool {
	stockKeywords := []string{
		"stock", "analyze", "price", "investment", "portfolio", "buy", "sell", "market", "trading", "share", "equity",
		"accion", "acciones", "analizar", "precio", "inversion", "cartera", "comprar", "vender", "mercado", "bolsa",
		"action", "actions", "analyser", "prix", "investissement", "portefeuille", "acheter", "vendre", "marché", "bourse",
		"aktie", "aktien", "analysieren", "preis", "investition", "portfolio", "kaufen", "verkaufen", "markt", "börse",
		"azione", "azioni", "analizzare", "prezzo", "investimento", "portafoglio", "comprare", "vendere", "mercato", "borsa",
		"ação", "ações", "analisar", "preço", "investimento", "portfólio", "comprar", "vender", "mercado", "bolsa",
	}
	
	lowerInput := strings.ToLower(input)
	
	for _, keyword := range stockKeywords {
		if strings.Contains(lowerInput, keyword) {
			return true
		}
	}
	
	if len(c.extractSymbols(input)) > 0 {
		return true
	}
	
	return false
}

func (c *ChatbotHost) handleStockQuery(input string) error {
	symbols := c.extractSymbols(input)
	
	if len(symbols) > 0 {
		fmt.Printf("Detected stock symbols: %v\n", symbols)
		
		if c.getStockAnalyzerClient() == nil {
			fmt.Println("Stock analyzer server not connected. Please connect using /connect ./bin/stock-analyzer")
			return nil
		}
		
		if strings.Contains(strings.ToLower(input), "analyze") || len(symbols) > 1 {
			return c.analyzePortfolio(symbols)
		} else {
			return c.getStockPrice(symbols[0])
		}
	}

	contextMessage := fmt.Sprintf(`%s

I have access to stock analysis tools through MCP servers. I can:
- Analyze portfolios of stocks with technical indicators
- Get current stock prices  
- Provide investment recommendations

If you'd like stock analysis, please specify company names or stock symbols (e.g., Apple, Microsoft, AAPL, GOOGL, MSFT).`, input)

	response, err := c.claudeClient.Chat(contextMessage)
	if err != nil {
		return fmt.Errorf("Claude API error: %w", err)
	}

	fmt.Printf("🤖 Claude: %s\n", response)
	c.logInteraction("CLAUDE", response)
	return nil
}

func (c *ChatbotHost) extractSymbols(input string) []string {
	words := strings.Fields(strings.ToUpper(input))
	symbols := make([]string, 0)
	
	commonSymbols := map[string]bool{
		"AAPL": true, "GOOGL": true, "MSFT": true, "AMZN": true, "TSLA": true,
		"META": true, "NVDA": true, "NFLX": true, "AMD": true, "INTC": true,
		"GOOG": true, "FB": true, "TSMC": true, "V": true, "JPM": true,
	}
	
	for _, word := range words {
		word = strings.Trim(word, ".,!?;:")
		if len(word) >= 1 && len(word) <= 5 && commonSymbols[word] {
			symbols = append(symbols, word)
		}
	}
	
	if len(symbols) == 0 {
		claudeSymbols := c.extractSymbolsWithClaude(input)
		symbols = append(symbols, claudeSymbols...)
	}
	
	return symbols
}

func (c *ChatbotHost) extractSymbolsWithClaude(input string) []string {
	if c.claudeClient == nil {
		return []string{}
	}
	
	prompt := fmt.Sprintf(`Analyze this text and extract stock ticker symbols for any companies mentioned. 

Text: "%s"

Instructions:
1. Identify any company names mentioned in any language
2. Convert them to their corresponding stock ticker symbols (e.g., "Apple" -> "AAPL", "Microsoft" -> "MSFT")
3. Return ONLY the ticker symbols, separated by commas
4. If no companies are mentioned, return "NONE"
5. Focus on publicly traded companies only

Examples:
- "Apple stock price" -> "AAPL"
- "Should I buy Microsoft and Google?" -> "MSFT,GOOGL"
- "Tesla and Amazon analysis" -> "TSLA,AMZN"
- "How is the weather?" -> "NONE"

Response:`, input)

	response, err := c.claudeClient.Chat(prompt)
	if err != nil {
		c.logger.Printf("Error extracting symbols with Claude: %v", err)
		return []string{}
	}
	
	response = strings.TrimSpace(response)
	if response == "NONE" || response == "" {
		return []string{}
	}
	
	symbols := make([]string, 0)
	symbolParts := strings.Split(response, ",")
	for _, symbol := range symbolParts {
		symbol = strings.TrimSpace(strings.ToUpper(symbol))
		if len(symbol) >= 1 && len(symbol) <= 5 {
			symbols = append(symbols, symbol)
		}
	}
	
	c.logger.Printf("Claude extracted symbols: %v from input: %s", symbols, input)
	return symbols
}

func (c *ChatbotHost) connectToStockServer() {
	stockServerBin := "./bin/stock-analyzer"
	if _, err := os.Stat(stockServerBin); err == nil {
		fmt.Printf("Launching MCP server: %s\n", stockServerBin)
		c.logger.Println("Attempting auto-connection to stock analyzer server...")
		if err := c.connectToMCPServer(stockServerBin); err != nil {
			c.logger.Printf("Auto-connect to built server failed: %v", err)
			fmt.Printf("Auto-connection failed: %v\n", err)
			fmt.Println("Use /connect ./bin/stock-analyzer to connect manually")
			return
		}
		fmt.Println("MCP server launched and connected successfully!")
		return
	}
	
	fmt.Println("Stock analyzer binary not found at ./bin/stock-analyzer")
	fmt.Println("Run: go build -o bin/stock-analyzer ./servers/stock-analyzer/")
}

func (c *ChatbotHost) connectToMCPServer(serverPath string) error {
	serverName := filepath.Base(serverPath)
	
	if _, exists := c.mcpClients[serverName]; exists {
		fmt.Printf("Already connected to %s\n", serverName)
		return nil
	}

	var cmd []string
	if strings.HasSuffix(serverPath, ".go") {
		cmd = []string{"go", "run", serverPath}
	} else {
		cmd = []string{serverPath}
	}

	client := mcp.NewClient(cmd, c.logger)
	
	if err := client.Connect(); err != nil {
		return fmt.Errorf("failed to connect to %s: %w", serverName, err)
	}

	initResponse, err := client.Initialize()
	if err != nil {
		client.Close()
		return fmt.Errorf("failed to initialize %s: %w", serverName, err)
	}

	c.mcpClients[serverName] = client
	c.logMCPInteraction("CONNECT", serverName, fmt.Sprintf("Connected to %s v%s", initResponse.ServerInfo.Name, initResponse.ServerInfo.Version))
	
	fmt.Printf("Connected to %s\n", initResponse.ServerInfo.Name)
	return nil
}

func (c *ChatbotHost) disconnectFromMCPServer(serverName string) error {
	client, exists := c.mcpClients[serverName]
	if !exists {
		fmt.Printf("Server '%s' not found. Use /status to see connected servers.\n", serverName)
		return nil
	}

	if err := client.Close(); err != nil {
		fmt.Printf("Error disconnecting from %s: %v\n", serverName, err)
	} else {
		fmt.Printf("Disconnected from %s\n", serverName)
	}

	delete(c.mcpClients, serverName)
	c.logMCPInteraction("DISCONNECT", serverName, "Disconnected")
	return nil
}

func (c *ChatbotHost) showConnectionStatus() error {
	fmt.Println("MCP Connection Status")
	fmt.Println("========================")
	
	if len(c.mcpClients) == 0 {
		fmt.Println("No MCP servers connected")
		fmt.Println()
		fmt.Println("How MCP Works:")
		fmt.Println("   • The chatbot launches and manages MCP server processes")
		fmt.Println("   • Each connection starts its own server instance")
		fmt.Println("   • Servers communicate via stdin/stdout (JSON-RPC 2.0)")
		fmt.Println()
		fmt.Println("Available servers to connect:")
		fmt.Println("  ./bin/stock-analyzer       - Built stock analyzer (recommended)")
		fmt.Println("  ./servers/stock-analyzer/  - Source stock analyzer (via go run)")
		return nil
	}

	fmt.Printf("Connected servers (%d):\n", len(c.mcpClients))
	for name, client := range c.mcpClients {
		_, err := client.ListTools()
		if err != nil {
			fmt.Printf("  %s (connection lost: %v)\n", name, err)
			client.Close()
			delete(c.mcpClients, name)
		} else {
			fmt.Printf("  %s (active - managed process)\n", name)
		}
	}
	
	fmt.Println()
	fmt.Println("  Each MCP server runs as a separate process managed by this client")

	return nil
}

func (c *ChatbotHost) listAvailableTools() error {
	if len(c.mcpClients) == 0 {
		fmt.Println("No MCP servers connected")
		return nil
	}

	fmt.Println("Available Tools:")
	for serverName, client := range c.mcpClients {
		tools, err := client.ListTools()
		if err != nil {
			fmt.Printf("Error listing tools for %s: %v\n", serverName, err)
			continue
		}

		fmt.Printf("\n%s:\n", serverName)
		for _, tool := range tools {
			fmt.Printf("  • %s: %s\n", tool.Name, tool.Description)
		}
	}

	return nil
}

func (c *ChatbotHost) analyzePortfolio(symbols []string) error {
	client := c.getStockAnalyzerClient()
	if client == nil {
		return fmt.Errorf("stock analyzer server not connected")
	}

	fmt.Printf("Analyzing portfolio: %v\n", symbols)

	args := map[string]interface{}{
		"symbols":   symbols,
		"timeframe": "1M",
	}

	c.logMCPInteraction("CALL_TOOL", "analyze_portfolio", fmt.Sprintf("Analyzing symbols: %v", symbols))

	response, err := client.CallTool("analyze_portfolio", args)
	if err != nil {
		return fmt.Errorf("portfolio analysis failed: %w", err)
	}

	if response.IsError {
		fmt.Println("Analysis failed:")
	} else {
		fmt.Println("Analysis complete:")
	}

	for _, content := range response.Content {
		fmt.Println(content.Text)
	}

	c.logMCPInteraction("TOOL_RESPONSE", "analyze_portfolio", "Analysis completed")
	return nil
}

func (c *ChatbotHost) getStockPrice(symbol string) error {
	client := c.getStockAnalyzerClient()
	if client == nil {
		return fmt.Errorf("stock analyzer server not connected")
	}

	fmt.Printf("Getting price for %s\n", symbol)

	args := map[string]interface{}{
		"symbol": symbol,
	}

	c.logMCPInteraction("CALL_TOOL", "get_stock_price", fmt.Sprintf("Getting price for: %s", symbol))

	response, err := client.CallTool("get_stock_price", args)
	if err != nil {
		return fmt.Errorf("price lookup failed: %w", err)
	}

	if response.IsError {
		fmt.Println("Price lookup failed:")
	}

	for _, content := range response.Content {
		fmt.Println(content.Text)
	}

	c.logMCPInteraction("TOOL_RESPONSE", "get_stock_price", "Price lookup completed")
	return nil
}


func (c *ChatbotHost) analyzePortfolioAdvanced(symbols []string) error {
	client := c.getStockAnalyzerClient()
	if client == nil {
		return fmt.Errorf("stock analyzer server not connected")
	}

	cleanSymbols := make([]string, len(symbols))
	for i, symbol := range symbols {
		cleanSymbols[i] = strings.ToUpper(strings.TrimSpace(symbol))
	}

	fmt.Printf("Advanced portfolio analysis: %v\n", cleanSymbols)

	args := map[string]interface{}{
		"symbols":   cleanSymbols,
		"timeframe": "1M",
	}

	c.logMCPInteraction("CALL_TOOL", "analyze_portfolio_advanced", fmt.Sprintf("Analyzing symbols: %v", cleanSymbols))

	response, err := client.CallTool("analyze_portfolio_advanced", args)
	if err != nil {
		return fmt.Errorf("advanced portfolio analysis failed: %w", err)
	}

	if response.IsError {
		fmt.Println("Advanced analysis failed:")
	} else {
		fmt.Println("Advanced analysis complete:")
	}

	for _, content := range response.Content {
		fmt.Println(content.Text)
	}

	c.logMCPInteraction("TOOL_RESPONSE", "analyze_portfolio_advanced", "Advanced analysis completed")
	return nil
}

func (c *ChatbotHost) getEnhancedStockPrice(symbol string) error {
	client := c.getStockAnalyzerClient()
	if client == nil {
		return fmt.Errorf("stock analyzer server not connected")
	}

	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	fmt.Printf("Enhanced analysis for %s\n", symbol)

	args := map[string]interface{}{
		"symbol":    symbol,
		"timeframe": "1M",
	}

	c.logMCPInteraction("CALL_TOOL", "analyze_stock_with_reliability", fmt.Sprintf("Enhanced analysis for: %s", symbol))

	response, err := client.CallTool("analyze_stock_with_reliability", args)
	if err != nil {
		return fmt.Errorf("enhanced stock analysis failed: %w", err)
	}

	if response.IsError {
		fmt.Println("Enhanced analysis failed:")
	}

	for _, content := range response.Content {
		fmt.Println(content.Text)
	}

	c.logMCPInteraction("TOOL_RESPONSE", "analyze_stock_with_reliability", "Enhanced analysis completed")
	return nil
}

func (c *ChatbotHost) getPricePrediction(symbol string) error {
	client := c.getStockAnalyzerClient()
	if client == nil {
		return fmt.Errorf("stock analyzer server not connected")
	}

	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	fmt.Printf("Price prediction for %s\n", symbol)

	args := map[string]interface{}{
		"symbol":    symbol,
		"timeframe": "1M",
	}

	c.logMCPInteraction("CALL_TOOL", "get_price_prediction", fmt.Sprintf("Price prediction for: %s", symbol))

	response, err := client.CallTool("get_price_prediction", args)
	if err != nil {
		return fmt.Errorf("price prediction failed: %w", err)
	}

	if response.IsError {
		fmt.Println("Price prediction failed:")
	}

	for _, content := range response.Content {
		fmt.Println(content.Text)
	}

	c.logMCPInteraction("TOOL_RESPONSE", "get_price_prediction", "Price prediction completed")
	return nil
}

func (c *ChatbotHost) analyzeHistoricalTrends(symbol string) error {
	client := c.getStockAnalyzerClient()
	if client == nil {
		return fmt.Errorf("stock analyzer server not connected")
	}

	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	fmt.Printf("Historical trend analysis for %s\n", symbol)

	args := map[string]interface{}{
		"symbol":    symbol,
		"timeframe": "3M", 
	}

	c.logMCPInteraction("CALL_TOOL", "analyze_historical_trends", fmt.Sprintf("Trend analysis for: %s", symbol))

	response, err := client.CallTool("analyze_historical_trends", args)
	if err != nil {
		return fmt.Errorf("historical trend analysis failed: %w", err)
	}

	if response.IsError {
		fmt.Println("Trend analysis failed:")
	}

	for _, content := range response.Content {
		fmt.Println(content.Text)
	}

	c.logMCPInteraction("TOOL_RESPONSE", "analyze_historical_trends", "Trend analysis completed")
	return nil
}

func (c *ChatbotHost) getStockAnalyzerClient() *mcp.Client {
	for name, client := range c.mcpClients {
		if strings.Contains(strings.ToLower(name), "stock") || strings.Contains(strings.ToLower(name), "main.go") {
			return client
		}
	}
	return nil
}

func (c *ChatbotHost) showHelp() error {
	fmt.Println(`
 MCP Stock Analysis Chatbot Help
==================================

Connection Commands:
  /connect <server>     Connect to MCP server (e.g., ./bin/stock-analyzer)
  /disconnect <server>  Disconnect from MCP server
  /status              Show connection status and health

Enhanced Analysis Commands:
  /list                List all available tools from connected servers
  /analyze <symbols>   Advanced portfolio analysis with reliability (e.g., /analyze AAPL,GOOGL,MSFT)
  /predict <symbol>    Get price predictions with confidence intervals (e.g., /predict AAPL)
  /trends <symbol>     Analyze historical trends and patterns (e.g., /trends AAPL)
  /price <symbol>      Enhanced stock analysis with reliability (e.g., /price AAPL)

General Commands:
  /help                Show this help message
  /quit                Exit the chatbot

Natural Language:
  You can also chat naturally! Try:
  - "Analyze Apple and Microsoft stocks"
  - "What's the price of Tesla?"
  - "Should I invest in tech stocks?"

The chatbot will detect stock-related queries and use the appropriate tools,
or forward general questions to Claude for conversation.
`)
	return nil
}

func (c *ChatbotHost) logInteraction(role, message string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	c.logger.Printf("[%s] %s: %s", timestamp, role, message)
}

func (c *ChatbotHost) logMCPInteraction(action, tool, details string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	c.logger.Printf("[%s] MCP_%s %s: %s", timestamp, action, tool, details)
}

func (c *ChatbotHost) cleanup() {
	c.logger.Println("Shutting down chatbot...")
	for name, client := range c.mcpClients {
		if err := client.Close(); err != nil {
			c.logger.Printf("Error closing %s: %v", name, err)
		}
	}
}

func main() {
	chatbot := NewChatbotHost()
	
	if err := chatbot.Start(); err != nil {
		log.Fatalf("Chatbot error: %v", err)
	}
}
