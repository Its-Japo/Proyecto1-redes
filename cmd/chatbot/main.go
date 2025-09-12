package main

import (
	"bufio"
	"encoding/json"
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
	fmt.Println("  /connect <server_path>  - Connect to local MCP server")
	fmt.Println("  /connect tcp://<host:port> - Connect to remote MCP server")
	fmt.Println("  /connect-filesystem     - Connect to official Filesystem MCP server")
	fmt.Println("  /connect-git           - Connect to official Git MCP server")
	fmt.Println("  /disconnect <server>    - Disconnect from MCP server")
	fmt.Println("  /status                 - Show connection status")
	fmt.Println("  /list                   - List available tools")
	fmt.Println("  /analyze <symbols>      - Advanced portfolio analysis with reliability")
	fmt.Println("  /predict <symbol>       - Get price predictions with confidence intervals")
	fmt.Println("  /trends <symbol>        - Analyze historical trends and patterns")
	fmt.Println("  /price <symbol>         - Get enhanced stock analysis")
	fmt.Println("  /demo-mcp              - Run MCP servers demo (create repo, README, commit)")
	fmt.Println("  /help                   - Show help")
	fmt.Println("  /quit                   - Exit chatbot")
	fmt.Println()
	fmt.Println("Natural Language MCP Operations:")
	fmt.Println("  You can use natural language to execute MCP functions!")
	fmt.Println("  Examples:")
	fmt.Println("    'Read the README file'")
	fmt.Println("    'Create a new directory called test'")
	fmt.Println("    'Show me the git status'")
	fmt.Println("    'List all files in the current directory'")
	fmt.Println("    'Commit my changes with message: Initial commit'")
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

	case "/connect-filesystem":
		return c.connectToFilesystemServer()

	case "/connect-git":
		return c.connectToGitServer()

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

	case "/demo-mcp":
		return c.runMCPDemo()

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

	if c.isMCPRelatedQuery(input) {
		return c.handleMCPQuery(input)
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

func (c *ChatbotHost) isMCPRelatedQuery(input string) bool {
	mcpKeywords := []string{
		// Filesystem operations
		"read", "write", "create", "delete", "file", "directory", "folder", "list", "show", "display", "open", "edit", "modify",
		"leer", "escribir", "crear", "eliminar", "archivo", "directorio", "carpeta", "mostrar", "abrir", "editar", "modificar",
		"lire", "écrire", "créer", "supprimer", "fichier", "répertoire", "dossier", "afficher", "ouvrir", "modifier",
		"lesen", "schreiben", "erstellen", "löschen", "datei", "verzeichnis", "ordner", "anzeigen", "öffnen", "bearbeiten",
		"leggere", "scrivere", "creare", "eliminare", "file", "directory", "cartella", "mostrare", "aprire", "modificare",
		"ler", "escrever", "criar", "excluir", "arquivo", "diretório", "pasta", "mostrar", "abrir", "editar", "modificar",
		
		// Git operations
		"git", "commit", "push", "pull", "branch", "merge", "status", "log", "diff", "add", "reset", "checkout", "clone",
		"repositorio", "commitear", "subir", "bajar", "rama", "fusionar", "estado", "registro", "diferencia", "agregar",
		"dépôt", "commiter", "pousser", "tirer", "branche", "fusionner", "statut", "journal", "différence", "ajouter",
		"repository", "committen", "pushen", "pullen", "zweig", "mergen", "status", "protokoll", "unterschied", "hinzufügen",
		"repository", "committare", "spingere", "tirare", "ramo", "unire", "stato", "registro", "differenza", "aggiungere",
		"repositório", "commitar", "empurrar", "puxar", "ramo", "mesclar", "status", "registro", "diferença", "adicionar",
		
		// General MCP operations
		"mcp", "tool", "function", "execute", "run", "call", "use", "perform", "operation", "action",
		"herramienta", "función", "ejecutar", "correr", "llamar", "usar", "realizar", "operación", "acción",
		"outil", "fonction", "exécuter", "courir", "appeler", "utiliser", "effectuer", "opération", "action",
		"werkzeug", "funktion", "ausführen", "laufen", "aufrufen", "verwenden", "durchführen", "operation", "aktion",
		"strumento", "funzione", "eseguire", "correre", "chiamare", "usare", "eseguire", "operazione", "azione",
		"ferramenta", "função", "executar", "correr", "chamar", "usar", "realizar", "operação", "ação",
	}
	
	lowerInput := strings.ToLower(input)
	
	for _, keyword := range mcpKeywords {
		if strings.Contains(lowerInput, keyword) {
			return true
		}
	}
	
	return false
}

func (c *ChatbotHost) handleMCPQuery(input string) error {
	// Check if we have MCP servers connected
	if len(c.mcpClients) == 0 {
		fmt.Println("No MCP servers connected. Use /connect-filesystem or /connect-git to connect to servers.")
		return nil
	}

	// Try simple pattern matching first (works without Claude)
	if operation := c.parseSimpleMCPOperation(input); operation != nil {
		return c.executeSingleMCPOperation(operation)
	}

	// If Claude is available, use it for more complex parsing
	if c.claudeClient != nil && c.claudeClient.IsAvailable() {
		contextMessage := fmt.Sprintf(`%s

I have access to MCP (Model Context Protocol) servers that provide tools for:

1. **Filesystem Operations** (if filesystem server is connected):
   - read_text_file: Read file contents (args: {"path": "filename"})
   - write_file: Create or overwrite files (args: {"path": "filename", "content": "text content"})
   - create_directory: Create directories (args: {"path": "directory_name"})
   - list_directory: List directory contents (args: {"path": "directory_path"})
   - search_files: Find files by pattern (args: {"path": "search_path", "pattern": "*.ext"})
   - move_file: Move or rename files (args: {"source": "old_path", "destination": "new_path"})
   - get_file_info: Get file metadata (args: {"path": "filename"})

2. **Git Operations** (if git server is connected):
   - git_status: Show repository status (args: {"repo_path": "."})
   - git_add: Stage files for commit (args: {"repo_path": ".", "files": ["file1", "file2"]})
   - git_commit: Create commits (args: {"repo_path": ".", "message": "commit message"})
   - git_log: Show commit history (args: {"repo_path": ".", "max_count": 10})
   - git_diff: Show differences (args: {"repo_path": ".", "target": "branch_or_commit"})
   - git_branch: List branches (args: {"repo_path": ".", "branch_type": "local|remote|all"})
   - git_checkout: Switch branches (args: {"repo_path": ".", "branch_name": "branch_name"})
   - git_init: Initialize repository (args: {"repo_path": "path"})

IMPORTANT: Use the exact parameter names shown above. For example:
- For write_file, use "path" not "file_path"
- For git operations, always include "repo_path" parameter

Please analyze the user's request and determine:
1. Which MCP server(s) should be used
2. Which specific tool(s) should be called
3. What arguments should be passed to the tool(s) using the EXACT parameter names above

Respond with a JSON object in this format:
{
  "server": "filesystem|git",
  "tool": "tool_name",
  "arguments": {"param_name": "value"},
  "explanation": "Brief explanation of what will be done"
}

If multiple operations are needed, provide an array of such objects.
If the request is unclear or cannot be fulfilled with available MCP tools, explain what the user should ask for instead.`, input)

		response, err := c.claudeClient.Chat(contextMessage)
		if err != nil {
			fmt.Printf("Claude API error: %v\n", err)
			fmt.Println("Falling back to simple pattern matching...")
			return c.handleMCPQueryFallback(input)
		}

		fmt.Printf("🤖 Claude: %s\n", response)
		c.logInteraction("CLAUDE", response)

		// Try to parse the response as JSON and execute MCP tools
		return c.executeMCPFromClaudeResponse(response)
	}

	// Fallback to simple pattern matching
	return c.handleMCPQueryFallback(input)
}

func (c *ChatbotHost) executeMCPFromClaudeResponse(response string) error {
	// Look for JSON in the response
	jsonStart := strings.Index(response, "{")
	if jsonStart == -1 {
		jsonStart = strings.Index(response, "[")
	}
	
	if jsonStart == -1 {
		// No JSON found, just return the response
		return nil
	}

	jsonEnd := strings.LastIndex(response, "}")
	if jsonEnd == -1 {
		jsonEnd = strings.LastIndex(response, "]")
	}
	
	if jsonEnd == -1 || jsonEnd <= jsonStart {
		// Invalid JSON, just return the response
		return nil
	}

	jsonStr := response[jsonStart : jsonEnd+1]
	
	// Try to parse as single operation
	var operation map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &operation); err == nil {
		return c.executeSingleMCPOperation(operation)
	}
	
	// Try to parse as array of operations
	var operations []map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &operations); err == nil {
		for _, op := range operations {
			if err := c.executeSingleMCPOperation(op); err != nil {
				fmt.Printf("Error executing operation: %v\n", err)
			}
		}
		return nil
	}

	// JSON parsing failed, just return the response
	return nil
}

func (c *ChatbotHost) executeSingleMCPOperation(operation map[string]interface{}) error {
	serverName, ok := operation["server"].(string)
	if !ok {
		return fmt.Errorf("invalid server name in operation")
	}

	toolName, ok := operation["tool"].(string)
	if !ok {
		return fmt.Errorf("invalid tool name in operation")
	}

	arguments, ok := operation["arguments"].(map[string]interface{})
	if !ok {
		arguments = make(map[string]interface{})
	}

	explanation, _ := operation["explanation"].(string)
	if explanation != "" {
		fmt.Printf("📋 %s\n", explanation)
	}

	// Handle multi-step operations
	if serverName == "multi-step" {
		return c.executeMultiStepOperation(operation)
	}

	// Get the appropriate MCP client
	client, exists := c.mcpClients[serverName]
	if !exists {
		return fmt.Errorf("MCP server '%s' not connected", serverName)
	}

	// Execute the tool
	fmt.Printf("🔧 Executing %s.%s...\n", serverName, toolName)
	return c.executeMCPTool(client, toolName, arguments)
}

func (c *ChatbotHost) parseSimpleMCPOperation(input string) map[string]interface{} {
	lowerInput := strings.ToLower(input)
	
	// Complex multi-step operations
	if strings.Contains(lowerInput, "create") && strings.Contains(lowerInput, "repository") {
		return c.parseRepositoryCreation(input)
	}
	
	// Filesystem operations
	if strings.Contains(lowerInput, "read") && strings.Contains(lowerInput, "file") {
		// Extract filename from input
		filename := c.extractFilenameFromInput(input)
		if filename == "" {
			filename = "README.md" // Default
		}
		return map[string]interface{}{
			"server": "filesystem",
			"tool": "read_text_file",
			"arguments": map[string]interface{}{
				"path": filename,
			},
			"explanation": fmt.Sprintf("Reading file: %s", filename),
		}
	}
	
	if strings.Contains(lowerInput, "list") && (strings.Contains(lowerInput, "directory") || strings.Contains(lowerInput, "folder") || strings.Contains(lowerInput, "files")) {
		path := c.extractPathFromInput(input)
		if path == "" || path == "list" {
			path = "." // Current directory
		}
		return map[string]interface{}{
			"server": "filesystem",
			"tool": "list_directory",
			"arguments": map[string]interface{}{
				"path": path,
			},
			"explanation": fmt.Sprintf("Listing directory: %s", path),
		}
	}
	
	if strings.Contains(lowerInput, "create") && strings.Contains(lowerInput, "directory") {
		dirname := c.extractDirectoryNameFromInput(input)
		if dirname == "" {
			return nil
		}
		return map[string]interface{}{
			"server": "filesystem",
			"tool": "create_directory",
			"arguments": map[string]interface{}{
				"path": dirname,
			},
			"explanation": fmt.Sprintf("Creating directory: %s", dirname),
		}
	}
	
	if strings.Contains(lowerInput, "write") && strings.Contains(lowerInput, "file") {
		filename, content := c.extractWriteFileFromInput(input)
		if filename == "" {
			return nil
		}
		return map[string]interface{}{
			"server": "filesystem",
			"tool": "write_file",
			"arguments": map[string]interface{}{
				"path": filename,
				"content": content,
			},
			"explanation": fmt.Sprintf("Writing file: %s", filename),
		}
	}
	
	// Git operations
	if strings.Contains(lowerInput, "git status") || strings.Contains(lowerInput, "show status") {
		return map[string]interface{}{
			"server": "git",
			"tool": "git_status",
			"arguments": map[string]interface{}{
				"repo_path": ".",
			},
			"explanation": "Showing git repository status",
		}
	}
	
	if strings.Contains(lowerInput, "git log") || strings.Contains(lowerInput, "show log") || strings.Contains(lowerInput, "commit history") {
		return map[string]interface{}{
			"server": "git",
			"tool": "git_log",
			"arguments": map[string]interface{}{
				"repo_path": ".",
				"max_count": 10,
			},
			"explanation": "Showing git commit history",
		}
	}
	
	if strings.Contains(lowerInput, "add") && strings.Contains(lowerInput, "git") {
		return map[string]interface{}{
			"server": "git",
			"tool": "git_add",
			"arguments": map[string]interface{}{
				"repo_path": ".",
				"files": []string{"."},
			},
			"explanation": "Adding all files to git staging area",
		}
	}
	
	return nil
}

func (c *ChatbotHost) handleMCPQueryFallback(input string) error {
	fmt.Println("🤖 I understand you want to perform an MCP operation, but I need more specific instructions.")
	fmt.Println("Available operations:")
	fmt.Println("  Filesystem: 'read file', 'list directory', 'create directory'")
	fmt.Println("  Git: 'git status', 'git log', 'add files'")
	fmt.Println("  Use /list to see all available MCP tools")
	return nil
}

func (c *ChatbotHost) extractFilenameFromInput(input string) string {
	// Look for quoted strings first
	if start := strings.Index(input, "\""); start != -1 {
		if end := strings.Index(input[start+1:], "\""); end != -1 {
			return input[start+1 : start+1+end]
		}
	}
	
	// Look for common file patterns
	words := strings.Fields(input)
	for _, word := range words {
		if strings.Contains(word, ".") && len(word) > 1 {
			return word
		}
	}
	
	return ""
}

func (c *ChatbotHost) extractPathFromInput(input string) string {
	// Look for quoted strings first
	if start := strings.Index(input, "\""); start != -1 {
		if end := strings.Index(input[start+1:], "\""); end != -1 {
			return input[start+1 : start+1+end]
		}
	}
	
	// Look for common directory patterns
	words := strings.Fields(input)
	commandWords := map[string]bool{
		"list": true, "directory": true, "folder": true, "files": true,
		"show": true, "display": true, "contents": true,
		"all": true, "current": true, "the": true, "in": true,
		"of": true, "with": true, "from": true, "to": true,
	}
	
	for _, word := range words {
		if word == "." || word == ".." {
			return word
		}
		// Only return words that are likely directory names (not command words)
		if !strings.Contains(word, ".") && len(word) > 1 && !commandWords[strings.ToLower(word)] {
			return word
		}
	}
	
	return ""
}

func (c *ChatbotHost) extractDirectoryNameFromInput(input string) string {
	// Look for "directory called X" or "folder called X"
	lowerInput := strings.ToLower(input)
	if strings.Contains(lowerInput, "called") {
		parts := strings.Split(input, "called")
		if len(parts) > 1 {
			name := strings.TrimSpace(parts[1])
			// Remove quotes if present
			name = strings.Trim(name, "\"'")
			return name
		}
	}
	
	// Look for quoted strings
	if start := strings.Index(input, "\""); start != -1 {
		if end := strings.Index(input[start+1:], "\""); end != -1 {
			return input[start+1 : start+1+end]
		}
	}
	
	return ""
}

func (c *ChatbotHost) extractWriteFileFromInput(input string) (string, string) {
	// Look for patterns like "Write a file called 'filename' with content 'content'"
	// or "Write 'content' to 'filename'"
	
	// Pattern 1: "Write a file called 'filename' with content 'content'"
	if strings.Contains(strings.ToLower(input), "called") && strings.Contains(strings.ToLower(input), "with content") {
		parts := strings.Split(input, "called")
		if len(parts) > 1 {
			secondPart := parts[1]
			contentParts := strings.Split(secondPart, "with content")
			if len(contentParts) >= 2 {
				filename := strings.TrimSpace(contentParts[0])
				filename = strings.Trim(filename, "\"'")
				content := strings.TrimSpace(contentParts[1])
				content = strings.Trim(content, "\"'")
				return filename, content
			}
		}
	}
	
	// Pattern 2: "Write 'content' to 'filename'"
	if strings.Contains(strings.ToLower(input), "to") {
		parts := strings.Split(input, "to")
		if len(parts) >= 2 {
			contentPart := strings.TrimSpace(parts[0])
			filenamePart := strings.TrimSpace(parts[1])
			
			// Extract content (remove "write" and quotes)
			contentWords := strings.Fields(contentPart)
			if len(contentWords) > 1 {
				content := strings.Join(contentWords[1:], " ")
				content = strings.Trim(content, "\"'")
				
				// Extract filename (remove quotes)
				filename := strings.Trim(filenamePart, "\"'")
				
				return filename, content
			}
		}
	}
	
	// Pattern 3: Look for quoted strings
	quotedStrings := c.extractQuotedStrings(input)
	if len(quotedStrings) >= 2 {
		// Assume first is filename, second is content
		return quotedStrings[0], quotedStrings[1]
	}
	
	return "", ""
}

func (c *ChatbotHost) extractQuotedStrings(input string) []string {
	var results []string
	start := -1
	
	for i, char := range input {
		if char == '"' || char == '\'' {
			if start == -1 {
				start = i + 1
			} else {
				results = append(results, input[start:i])
				start = -1
			}
		}
	}
	
	return results
}

func (c *ChatbotHost) parseRepositoryCreation(input string) map[string]interface{} {
	// Parse patterns like:
	// "Create a repository called Mom with the README.md file with the content 'Hey mom'"
	// "Create repository Mom with README.md containing 'Hey mom'"
	
	lowerInput := strings.ToLower(input)
	
	// Extract repository name
	repoName := ""
	if strings.Contains(lowerInput, "called") {
		parts := strings.Split(input, "called")
		if len(parts) > 1 {
			secondPart := parts[1]
			// Look for the next word or phrase before "with"
			words := strings.Fields(secondPart)
			if len(words) > 0 {
				repoName = words[0]
			}
		}
	}
	
	// Extract README content
	readmeContent := ""
	if strings.Contains(lowerInput, "content") {
		contentParts := strings.Split(input, "content")
		if len(contentParts) > 1 {
			content := strings.TrimSpace(contentParts[1])
			content = strings.Trim(content, "\"'")
			readmeContent = content
		}
	}
	
	if repoName == "" {
		repoName = "new-repo" // Default
	}
	if readmeContent == "" {
		readmeContent = "Hello World!" // Default
	}
	
	// Return a special multi-step operation
	return map[string]interface{}{
		"server": "multi-step",
		"tool": "create_repository",
		"arguments": map[string]interface{}{
			"repo_name": repoName,
			"readme_content": readmeContent,
		},
		"explanation": fmt.Sprintf("Creating repository '%s' with README.md containing '%s'", repoName, readmeContent),
	}
}

func (c *ChatbotHost) executeMultiStepOperation(operation map[string]interface{}) error {
	toolName := operation["tool"].(string)
	arguments := operation["arguments"].(map[string]interface{})
	
	switch toolName {
	case "create_repository":
		return c.executeCreateRepository(arguments)
	default:
		return fmt.Errorf("unknown multi-step operation: %s", toolName)
	}
}

func (c *ChatbotHost) executeCreateRepository(args map[string]interface{}) error {
	repoName := args["repo_name"].(string)
	readmeContent := args["readme_content"].(string)
	
	fmt.Printf("📋 Creating repository '%s' with README.md...\n", repoName)
	
	// Step 1: Create directory
	filesystemClient := c.mcpClients["filesystem"]
	if filesystemClient == nil {
		return fmt.Errorf("filesystem MCP server not connected")
	}
	
	fmt.Printf("🔧 Step 1: Creating directory '%s'...\n", repoName)
	if err := c.executeMCPTool(filesystemClient, "create_directory", map[string]interface{}{
		"path": repoName,
	}); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	
	// Step 2: Create README.md file
	fmt.Printf("🔧 Step 2: Creating README.md file...\n")
	if err := c.executeMCPTool(filesystemClient, "write_file", map[string]interface{}{
		"path": fmt.Sprintf("%s/README.md", repoName),
		"content": readmeContent,
	}); err != nil {
		return fmt.Errorf("failed to create README.md: %w", err)
	}
	
	// Step 3: Initialize git repository
	gitClient := c.mcpClients["git"]
	if gitClient != nil {
		fmt.Printf("🔧 Step 3: Initializing git repository...\n")
		// Use absolute path to ensure git init happens in the correct location
		absPath, err := filepath.Abs(repoName)
		if err != nil {
			fmt.Printf("⚠️  Could not get absolute path: %v (continuing anyway)\n", err)
		} else {
			if err := c.executeMCPTool(gitClient, "git_init", map[string]interface{}{
				"repo_path": absPath,
			}); err != nil {
				fmt.Printf("⚠️  Git initialization failed: %v (continuing anyway)\n", err)
			}
		}
	} else {
		fmt.Printf("⚠️  Git server not connected, skipping git initialization\n")
	}
	
	fmt.Printf("✅ Repository '%s' created successfully!\n", repoName)
	return nil
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

	// Check if this is a TCP connection (format: tcp://host:port)
	if strings.HasPrefix(serverPath, "tcp://") {
		return c.connectToTCPServer(serverPath, serverName)
	}

	// Local process connection (existing logic)
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

func (c *ChatbotHost) connectToTCPServer(serverURL string, serverName string) error {
	// Extract address from tcp://host:port format
	address := strings.TrimPrefix(serverURL, "tcp://")
	
	client := mcp.NewClient(nil, c.logger) // nil command for TCP connections
	
	if err := client.ConnectTCP(address); err != nil {
		return fmt.Errorf("failed to connect to %s: %w", address, err)
	}

	initResponse, err := client.Initialize()
	if err != nil {
		client.Close()
		return fmt.Errorf("failed to initialize %s: %w", serverName, err)
	}

	c.mcpClients[serverName] = client
	c.logMCPInteraction("CONNECT_TCP", serverName, fmt.Sprintf("Connected to remote %s v%s at %s", initResponse.ServerInfo.Name, initResponse.ServerInfo.Version, address))
	
	fmt.Printf("Connected to remote %s at %s\n", initResponse.ServerInfo.Name, address)
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
		// Check for stock analyzer by name patterns
		if strings.Contains(strings.ToLower(name), "stock") || 
		   strings.Contains(strings.ToLower(name), "main.go") ||
		   strings.Contains(name, ":8080") ||  // TCP connections on port 8080
		   strings.Contains(name, ".") {       // IP address patterns
			return client
		}
	}
	
	// If no specific match, return the first available client
	// (useful for TCP connections with IP names)
	for _, client := range c.mcpClients {
		return client
	}
	
	return nil
}

func (c *ChatbotHost) showHelp() error {
	fmt.Println(`
 MCP Stock Analysis Chatbot Help
==================================

Connection Commands:
  /connect <server>     Connect to local MCP server (e.g., ./bin/stock-analyzer)
  /connect tcp://host:port Connect to remote MCP server (e.g., tcp://localhost:8080)
  /connect-filesystem   Connect to official Filesystem MCP server
  /connect-git         Connect to official Git MCP server
  /disconnect <server>  Disconnect from MCP server
  /status              Show connection status and health

Enhanced Analysis Commands:
  /list                List all available tools from connected servers
  /analyze <symbols>   Advanced portfolio analysis with reliability (e.g., /analyze AAPL,GOOGL,MSFT)
  /predict <symbol>    Get price predictions with confidence intervals (e.g., /predict AAPL)
  /trends <symbol>     Analyze historical trends and patterns (e.g., /trends AAPL)
  /price <symbol>      Enhanced stock analysis with reliability (e.g., /price AAPL)

MCP Demo:
  /demo-mcp            Run MCP servers demo (create repo, README, commit)

Natural Language MCP Operations:
  You can use natural language to execute MCP functions! Examples:
  
  Filesystem Operations:
  - "Read the README file"
  - "Create a new directory called test"
  - "List all files in the current directory"
  - "Write 'Hello World' to hello.txt"
  - "Search for all .go files"
  - "Show me the contents of the src folder"
  
  Git Operations:
  - "Show me the git status"
  - "Commit my changes with message: Initial commit"
  - "Add all files to git"
  - "Show me the git log"
  - "Create a new branch called feature"
  - "Switch to the main branch"

Stock Analysis (Natural Language):
  - "Analyze Apple and Microsoft stocks"
  - "What's the price of Tesla?"
  - "Should I invest in tech stocks?"

General Commands:
  /help                Show this help message
  /quit                Exit the chatbot

The chatbot will automatically detect the type of query and use the appropriate
MCP tools or forward general questions to Claude for conversation.
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

func (c *ChatbotHost) connectToFilesystemServer() error {
	fmt.Println("Connecting to official Filesystem MCP server...")
	
	// Obtener el directorio actual de trabajo
	workDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}
	
	// Comando para lanzar el servidor filesystem
	cmd := []string{"./scripts/start-filesystem-mcp.sh"}
	
	client := mcp.NewClient(cmd, c.logger)
	
	if err := client.Connect(); err != nil {
		return fmt.Errorf("failed to connect to filesystem server: %w", err)
	}

	initResponse, err := client.Initialize()
	if err != nil {
		client.Close()
		return fmt.Errorf("failed to initialize filesystem server: %w", err)
	}

	c.mcpClients["filesystem"] = client
	c.logMCPInteraction("CONNECT", "filesystem", fmt.Sprintf("Connected to %s v%s", initResponse.ServerInfo.Name, initResponse.ServerInfo.Version))
	
	fmt.Printf("✅ Connected to %s (allowed directory: %s)\n", initResponse.ServerInfo.Name, workDir)
	return nil
}

func (c *ChatbotHost) connectToGitServer() error {
	fmt.Println("Connecting to official Git MCP server...")
	
	// Verificar que estamos en un repositorio Git
	if _, err := os.Stat(".git"); os.IsNotExist(err) {
		return fmt.Errorf("not in a git repository (no .git directory found)")
	}
	
	// Comando para lanzar el servidor git
	cmd := []string{"./scripts/start-git-mcp.sh"}
	
	client := mcp.NewClient(cmd, c.logger)
	
	if err := client.Connect(); err != nil {
		return fmt.Errorf("failed to connect to git server: %w", err)
	}

	initResponse, err := client.Initialize()
	if err != nil {
		client.Close()
		return fmt.Errorf("failed to initialize git server: %w", err)
	}

	c.mcpClients["git"] = client
	c.logMCPInteraction("CONNECT", "git", fmt.Sprintf("Connected to %s v%s", initResponse.ServerInfo.Name, initResponse.ServerInfo.Version))
	
	fmt.Printf("✅ Connected to %s\n", initResponse.ServerInfo.Name)
	return nil
}

func (c *ChatbotHost) runMCPDemo() error {
	fmt.Println("🎬 Running MCP Servers Demo...")
	fmt.Println("This will demonstrate filesystem and git operations using MCP servers")
	fmt.Println()

	// Verificar conexiones a servidores MCP
	filesystemClient := c.mcpClients["filesystem"]
	gitClient := c.mcpClients["git"]

	if filesystemClient == nil {
		fmt.Println("❌ Filesystem MCP server not connected. Use /connect-filesystem first.")
		return nil
	}

	if gitClient == nil {
		fmt.Println("❌ Git MCP server not connected. Use /connect-git first.")
		return nil
	}

	fmt.Println("✅ Both MCP servers connected. Starting demo...")
	fmt.Println()

	// Paso 1: Crear un repositorio de prueba
	fmt.Println("📁 Step 1: Creating a test workspace...")
	if err := c.executeMCPTool(filesystemClient, "create_directory", map[string]interface{}{
		"path": "demo-mcp-workspace",
	}); err != nil {
		fmt.Printf("Failed to create directory: %v\n", err)
	}

	// Paso 2: Crear un archivo README
	fmt.Println("📝 Step 2: Creating README.md file...")
	readmeContent := `# MCP Demo Repository

This repository was created automatically using Model Context Protocol (MCP) servers.

## What is MCP?

Model Context Protocol (MCP) is an open protocol that enables seamless integration between LLM applications and external data sources and tools.

## This Demo

This demo demonstrates:
- ✅ Filesystem operations using official Filesystem MCP server
- ✅ Git operations using official Git MCP server
- ✅ Automated repository creation and management

## Servers Used

1. **Filesystem MCP Server**: Provides secure file operations
2. **Git MCP Server**: Provides Git repository management

Created on: ` + time.Now().Format("2006-01-02 15:04:05") + `
`

	if err := c.executeMCPTool(filesystemClient, "write_file", map[string]interface{}{
		"path":    "demo-mcp-workspace/README.md",
		"content": readmeContent,
	}); err != nil {
		fmt.Printf("Failed to create README.md: %v\n", err)
	}

	// Paso 3: Inicializar repositorio Git
	fmt.Println("🔧 Step 3: Initializing Git repository...")
	if err := c.executeMCPTool(gitClient, "git_init", map[string]interface{}{
		"path": "demo-mcp-workspace",
	}); err != nil {
		fmt.Printf("Git init may have failed: %v (this might be expected)\n", err)
	}

	// Paso 4: Agregar archivo al repositorio
	fmt.Println("➕ Step 4: Adding README.md to git...")
	if err := c.executeMCPTool(gitClient, "git_add", map[string]interface{}{
		"paths": []string{"demo-mcp-workspace/README.md"},
	}); err != nil {
		fmt.Printf("Git add may have failed: %v\n", err)
	}

	// Paso 5: Hacer commit
	fmt.Println("💾 Step 5: Creating initial commit...")
	if err := c.executeMCPTool(gitClient, "git_commit", map[string]interface{}{
		"message": "Initial commit - MCP Demo Repository\n\nThis commit was created automatically using MCP servers:\n- Filesystem MCP server for file operations\n- Git MCP server for repository management",
	}); err != nil {
		fmt.Printf("Git commit may have failed: %v\n", err)
	}

	// Paso 6: Mostrar status del repositorio
	fmt.Println("📊 Step 6: Checking repository status...")
	if err := c.executeMCPTool(gitClient, "git_status", map[string]interface{}{}); err != nil {
		fmt.Printf("Git status failed: %v\n", err)
	}

	fmt.Println()
	fmt.Println("🎉 MCP Demo completed!")
	fmt.Println("Check the 'demo-mcp-workspace' directory to see the created files.")
	fmt.Println("You can explore the repository using the connected MCP servers.")

	return nil
}

func (c *ChatbotHost) executeMCPTool(client *mcp.Client, toolName string, args map[string]interface{}) error {
	c.logMCPInteraction("CALL_TOOL", toolName, fmt.Sprintf("Executing with args: %v", args))

	response, err := client.CallTool(toolName, args)
	if err != nil {
		return err
	}

	if response.IsError {
		fmt.Printf("❌ Tool %s failed:\n", toolName)
	} else {
		fmt.Printf("✅ Tool %s succeeded:\n", toolName)
	}

	for _, content := range response.Content {
		if content.Text != "" {
			fmt.Printf("   %s\n", content.Text)
		}
	}

	c.logMCPInteraction("TOOL_RESPONSE", toolName, fmt.Sprintf("Completed with %d content items", len(response.Content)))
	return nil
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
