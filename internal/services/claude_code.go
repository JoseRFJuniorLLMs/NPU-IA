package services

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ==================== CLAUDE CODE INTEGRATION ====================

// ClaudeCode integração com Claude Code CLI
type ClaudeCode struct {
	binaryPath    string
	configPath    string
	isAvailable   bool
	sessions      map[string]*ClaudeSession
	mcpServers    map[string]*MCPServer
	currentSession *ClaudeSession
	mu            sync.RWMutex
}

// ClaudeSession sessão do Claude Code
type ClaudeSession struct {
	ID          string    `json:"id"`
	StartedAt   time.Time `json:"started_at"`
	WorkingDir  string    `json:"working_dir"`
	Messages    []Message `json:"messages"`
	TokensUsed  int       `json:"tokens_used"`
	IsActive    bool      `json:"is_active"`
}

// Message mensagem na sessão
type Message struct {
	Role      string    `json:"role"` // user, assistant
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// MCPServer servidor MCP
type MCPServer struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Command     string            `json:"command"`
	Args        []string          `json:"args"`
	Env         map[string]string `json:"env"`
	IsEnabled   bool              `json:"is_enabled"`
	Tools       []MCPTool         `json:"tools"`
}

// MCPTool ferramenta MCP
type MCPTool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	InputSchema string `json:"input_schema"`
}

// NewClaudeCode cria integração com Claude Code
func NewClaudeCode() *ClaudeCode {
	cc := &ClaudeCode{
		sessions:   make(map[string]*ClaudeSession),
		mcpServers: make(map[string]*MCPServer),
	}

	cc.detectInstallation()
	cc.loadMCPServers()
	cc.registerPublicMCPServers()

	return cc
}

// detectInstallation detecta instalação do Claude Code
func (cc *ClaudeCode) detectInstallation() {
	// Procura claude no PATH
	path, err := exec.LookPath("claude")
	if err == nil {
		cc.binaryPath = path
		cc.isAvailable = true
	}

	// Procura em locais comuns
	commonPaths := []string{
		filepath.Join(os.Getenv("LOCALAPPDATA"), "Programs", "claude-code", "claude.exe"),
		filepath.Join(os.Getenv("APPDATA"), "npm", "claude.cmd"),
		filepath.Join(os.Getenv("USERPROFILE"), ".local", "bin", "claude"),
	}

	for _, p := range commonPaths {
		if _, err := os.Stat(p); err == nil {
			cc.binaryPath = p
			cc.isAvailable = true
			break
		}
	}

	// Config path
	cc.configPath = filepath.Join(os.Getenv("USERPROFILE"), ".claude")
}

// loadMCPServers carrega servidores MCP configurados
func (cc *ClaudeCode) loadMCPServers() {
	configFile := filepath.Join(cc.configPath, "mcp_servers.json")
	data, err := os.ReadFile(configFile)
	if err != nil {
		return
	}

	json.Unmarshal(data, &cc.mcpServers)
}

// registerPublicMCPServers registra servidores MCP públicos conhecidos
func (cc *ClaudeCode) registerPublicMCPServers() {
	// ==================== OFICIAL/POPULARES ====================

	cc.mcpServers["filesystem"] = &MCPServer{
		Name:        "Filesystem",
		Description: "Acesso seguro ao sistema de arquivos com permissões configuráveis",
		Command:     "npx",
		Args:        []string{"-y", "@modelcontextprotocol/server-filesystem", "/path/to/allowed/dir"},
		IsEnabled:   false,
	}

	cc.mcpServers["github"] = &MCPServer{
		Name:        "GitHub",
		Description: "Integração completa com GitHub - repos, issues, PRs, actions",
		Command:     "npx",
		Args:        []string{"-y", "@modelcontextprotocol/server-github"},
		Env:         map[string]string{"GITHUB_TOKEN": ""},
		IsEnabled:   false,
	}

	cc.mcpServers["gitlab"] = &MCPServer{
		Name:        "GitLab",
		Description: "Integração com GitLab - projetos, merge requests, pipelines",
		Command:     "npx",
		Args:        []string{"-y", "@modelcontextprotocol/server-gitlab"},
		Env:         map[string]string{"GITLAB_TOKEN": ""},
		IsEnabled:   false,
	}

	cc.mcpServers["slack"] = &MCPServer{
		Name:        "Slack",
		Description: "Integração com Slack - mensagens, canais, arquivos",
		Command:     "npx",
		Args:        []string{"-y", "@modelcontextprotocol/server-slack"},
		Env:         map[string]string{"SLACK_BOT_TOKEN": "", "SLACK_TEAM_ID": ""},
		IsEnabled:   false,
	}

	cc.mcpServers["google-drive"] = &MCPServer{
		Name:        "Google Drive",
		Description: "Acesso ao Google Drive - arquivos, pastas, busca",
		Command:     "npx",
		Args:        []string{"-y", "@modelcontextprotocol/server-gdrive"},
		IsEnabled:   false,
	}

	cc.mcpServers["postgres"] = &MCPServer{
		Name:        "PostgreSQL",
		Description: "Conexão com banco PostgreSQL - queries, schema inspection",
		Command:     "npx",
		Args:        []string{"-y", "@modelcontextprotocol/server-postgres"},
		Env:         map[string]string{"POSTGRES_CONNECTION_STRING": ""},
		IsEnabled:   false,
	}

	cc.mcpServers["sqlite"] = &MCPServer{
		Name:        "SQLite",
		Description: "Operações em bancos SQLite locais",
		Command:     "npx",
		Args:        []string{"-y", "@modelcontextprotocol/server-sqlite"},
		IsEnabled:   false,
	}

	cc.mcpServers["puppeteer"] = &MCPServer{
		Name:        "Puppeteer",
		Description: "Automação de browser - navegação, screenshots, scraping",
		Command:     "npx",
		Args:        []string{"-y", "@modelcontextprotocol/server-puppeteer"},
		IsEnabled:   false,
	}

	cc.mcpServers["brave-search"] = &MCPServer{
		Name:        "Brave Search",
		Description: "Busca na web via Brave Search API",
		Command:     "npx",
		Args:        []string{"-y", "@modelcontextprotocol/server-brave-search"},
		Env:         map[string]string{"BRAVE_API_KEY": ""},
		IsEnabled:   false,
	}

	cc.mcpServers["google-maps"] = &MCPServer{
		Name:        "Google Maps",
		Description: "Geocoding, direções, places, distâncias",
		Command:     "npx",
		Args:        []string{"-y", "@modelcontextprotocol/server-google-maps"},
		Env:         map[string]string{"GOOGLE_MAPS_API_KEY": ""},
		IsEnabled:   false,
	}

	cc.mcpServers["fetch"] = &MCPServer{
		Name:        "Fetch",
		Description: "Buscar conteúdo de URLs - HTML, JSON, texto",
		Command:     "uvx",
		Args:        []string{"mcp-server-fetch"},
		IsEnabled:   false,
	}

	// ==================== PRODUTIVIDADE ====================

	cc.mcpServers["notion"] = &MCPServer{
		Name:        "Notion",
		Description: "Integração com Notion - páginas, databases, blocos",
		Command:     "npx",
		Args:        []string{"-y", "notion-mcp-server"},
		Env:         map[string]string{"NOTION_API_KEY": ""},
		IsEnabled:   false,
	}

	cc.mcpServers["todoist"] = &MCPServer{
		Name:        "Todoist",
		Description: "Gerenciamento de tarefas Todoist",
		Command:     "npx",
		Args:        []string{"-y", "todoist-mcp-server"},
		Env:         map[string]string{"TODOIST_API_TOKEN": ""},
		IsEnabled:   false,
	}

	cc.mcpServers["linear"] = &MCPServer{
		Name:        "Linear",
		Description: "Integração com Linear - issues, projetos, ciclos",
		Command:     "npx",
		Args:        []string{"-y", "@linear/mcp-server"},
		Env:         map[string]string{"LINEAR_API_KEY": ""},
		IsEnabled:   false,
	}

	cc.mcpServers["asana"] = &MCPServer{
		Name:        "Asana",
		Description: "Gerenciamento de projetos Asana",
		Command:     "npx",
		Args:        []string{"-y", "asana-mcp-server"},
		Env:         map[string]string{"ASANA_ACCESS_TOKEN": ""},
		IsEnabled:   false,
	}

	cc.mcpServers["trello"] = &MCPServer{
		Name:        "Trello",
		Description: "Quadros, listas e cards do Trello",
		Command:     "npx",
		Args:        []string{"-y", "trello-mcp-server"},
		Env:         map[string]string{"TRELLO_API_KEY": "", "TRELLO_TOKEN": ""},
		IsEnabled:   false,
	}

	// ==================== COMUNICAÇÃO ====================

	cc.mcpServers["discord"] = &MCPServer{
		Name:        "Discord",
		Description: "Integração com Discord - mensagens, canais, servidores",
		Command:     "npx",
		Args:        []string{"-y", "discord-mcp-server"},
		Env:         map[string]string{"DISCORD_BOT_TOKEN": ""},
		IsEnabled:   false,
	}

	cc.mcpServers["telegram"] = &MCPServer{
		Name:        "Telegram",
		Description: "Bot Telegram - mensagens, grupos, canais",
		Command:     "npx",
		Args:        []string{"-y", "telegram-mcp-server"},
		Env:         map[string]string{"TELEGRAM_BOT_TOKEN": ""},
		IsEnabled:   false,
	}

	cc.mcpServers["email"] = &MCPServer{
		Name:        "Email (IMAP/SMTP)",
		Description: "Acesso a email via IMAP/SMTP",
		Command:     "npx",
		Args:        []string{"-y", "email-mcp-server"},
		IsEnabled:   false,
	}

	// ==================== DESENVOLVIMENTO ====================

	cc.mcpServers["docker"] = &MCPServer{
		Name:        "Docker",
		Description: "Gerenciamento de containers Docker",
		Command:     "npx",
		Args:        []string{"-y", "docker-mcp-server"},
		IsEnabled:   false,
	}

	cc.mcpServers["kubernetes"] = &MCPServer{
		Name:        "Kubernetes",
		Description: "Gerenciamento de clusters K8s",
		Command:     "npx",
		Args:        []string{"-y", "kubernetes-mcp-server"},
		IsEnabled:   false,
	}

	cc.mcpServers["aws"] = &MCPServer{
		Name:        "AWS",
		Description: "Serviços AWS - S3, EC2, Lambda, etc",
		Command:     "npx",
		Args:        []string{"-y", "aws-mcp-server"},
		Env:         map[string]string{"AWS_ACCESS_KEY_ID": "", "AWS_SECRET_ACCESS_KEY": ""},
		IsEnabled:   false,
	}

	cc.mcpServers["azure"] = &MCPServer{
		Name:        "Azure",
		Description: "Serviços Microsoft Azure",
		Command:     "npx",
		Args:        []string{"-y", "azure-mcp-server"},
		IsEnabled:   false,
	}

	cc.mcpServers["gcp"] = &MCPServer{
		Name:        "Google Cloud",
		Description: "Serviços Google Cloud Platform",
		Command:     "npx",
		Args:        []string{"-y", "gcp-mcp-server"},
		IsEnabled:   false,
	}

	cc.mcpServers["vercel"] = &MCPServer{
		Name:        "Vercel",
		Description: "Deploy e gerenciamento Vercel",
		Command:     "npx",
		Args:        []string{"-y", "vercel-mcp-server"},
		Env:         map[string]string{"VERCEL_TOKEN": ""},
		IsEnabled:   false,
	}

	cc.mcpServers["netlify"] = &MCPServer{
		Name:        "Netlify",
		Description: "Deploy e gerenciamento Netlify",
		Command:     "npx",
		Args:        []string{"-y", "netlify-mcp-server"},
		Env:         map[string]string{"NETLIFY_AUTH_TOKEN": ""},
		IsEnabled:   false,
	}

	// ==================== DADOS & ANALYTICS ====================

	cc.mcpServers["bigquery"] = &MCPServer{
		Name:        "BigQuery",
		Description: "Queries no Google BigQuery",
		Command:     "npx",
		Args:        []string{"-y", "bigquery-mcp-server"},
		IsEnabled:   false,
	}

	cc.mcpServers["snowflake"] = &MCPServer{
		Name:        "Snowflake",
		Description: "Data warehouse Snowflake",
		Command:     "npx",
		Args:        []string{"-y", "snowflake-mcp-server"},
		IsEnabled:   false,
	}

	cc.mcpServers["mongodb"] = &MCPServer{
		Name:        "MongoDB",
		Description: "Operações em MongoDB",
		Command:     "npx",
		Args:        []string{"-y", "mongodb-mcp-server"},
		Env:         map[string]string{"MONGODB_URI": ""},
		IsEnabled:   false,
	}

	cc.mcpServers["redis"] = &MCPServer{
		Name:        "Redis",
		Description: "Operações em Redis",
		Command:     "npx",
		Args:        []string{"-y", "redis-mcp-server"},
		Env:         map[string]string{"REDIS_URL": ""},
		IsEnabled:   false,
	}

	// ==================== TRANSPORTE & DELIVERY ====================

	cc.mcpServers["uber"] = &MCPServer{
		Name:        "Uber",
		Description: "Integração Uber - solicitar corridas, estimativas, histórico",
		Command:     "npx",
		Args:        []string{"-y", "uber-mcp-server"},
		Env:         map[string]string{"UBER_CLIENT_ID": "", "UBER_CLIENT_SECRET": ""},
		IsEnabled:   false,
		Tools: []MCPTool{
			{Name: "request_ride", Description: "Solicitar corrida Uber"},
			{Name: "get_estimate", Description: "Obter estimativa de preço"},
			{Name: "get_ride_status", Description: "Status da corrida atual"},
			{Name: "cancel_ride", Description: "Cancelar corrida"},
			{Name: "get_ride_history", Description: "Histórico de corridas"},
		},
	}

	cc.mcpServers["lyft"] = &MCPServer{
		Name:        "Lyft",
		Description: "Integração Lyft - corridas e estimativas",
		Command:     "npx",
		Args:        []string{"-y", "lyft-mcp-server"},
		Env:         map[string]string{"LYFT_CLIENT_ID": "", "LYFT_CLIENT_SECRET": ""},
		IsEnabled:   false,
	}

	cc.mcpServers["doordash"] = &MCPServer{
		Name:        "DoorDash",
		Description: "Pedidos de comida DoorDash",
		Command:     "npx",
		Args:        []string{"-y", "doordash-mcp-server"},
		IsEnabled:   false,
	}

	cc.mcpServers["instacart"] = &MCPServer{
		Name:        "Instacart",
		Description: "Compras de supermercado Instacart",
		Command:     "npx",
		Args:        []string{"-y", "instacart-mcp-server"},
		IsEnabled:   false,
	}

	// ==================== FINANÇAS ====================

	cc.mcpServers["stripe"] = &MCPServer{
		Name:        "Stripe",
		Description: "Pagamentos Stripe - clientes, cobranças, assinaturas",
		Command:     "npx",
		Args:        []string{"-y", "stripe-mcp-server"},
		Env:         map[string]string{"STRIPE_SECRET_KEY": ""},
		IsEnabled:   false,
	}

	cc.mcpServers["plaid"] = &MCPServer{
		Name:        "Plaid",
		Description: "Conexão bancária via Plaid",
		Command:     "npx",
		Args:        []string{"-y", "plaid-mcp-server"},
		IsEnabled:   false,
	}

	cc.mcpServers["coinbase"] = &MCPServer{
		Name:        "Coinbase",
		Description: "Operações crypto Coinbase",
		Command:     "npx",
		Args:        []string{"-y", "coinbase-mcp-server"},
		Env:         map[string]string{"COINBASE_API_KEY": "", "COINBASE_API_SECRET": ""},
		IsEnabled:   false,
	}

	// ==================== AI/ML ====================

	cc.mcpServers["openai"] = &MCPServer{
		Name:        "OpenAI",
		Description: "API OpenAI - GPT, DALL-E, Whisper",
		Command:     "npx",
		Args:        []string{"-y", "openai-mcp-server"},
		Env:         map[string]string{"OPENAI_API_KEY": ""},
		IsEnabled:   false,
	}

	cc.mcpServers["huggingface"] = &MCPServer{
		Name:        "Hugging Face",
		Description: "Modelos e datasets Hugging Face",
		Command:     "npx",
		Args:        []string{"-y", "huggingface-mcp-server"},
		Env:         map[string]string{"HF_TOKEN": ""},
		IsEnabled:   false,
	}

	cc.mcpServers["replicate"] = &MCPServer{
		Name:        "Replicate",
		Description: "Modelos ML no Replicate",
		Command:     "npx",
		Args:        []string{"-y", "replicate-mcp-server"},
		Env:         map[string]string{"REPLICATE_API_TOKEN": ""},
		IsEnabled:   false,
	}

	// ==================== MÍDIA & CONTEÚDO ====================

	cc.mcpServers["spotify"] = &MCPServer{
		Name:        "Spotify",
		Description: "Controle Spotify - playlists, reprodução, busca",
		Command:     "npx",
		Args:        []string{"-y", "spotify-mcp-server"},
		Env:         map[string]string{"SPOTIFY_CLIENT_ID": "", "SPOTIFY_CLIENT_SECRET": ""},
		IsEnabled:   false,
	}

	cc.mcpServers["youtube"] = &MCPServer{
		Name:        "YouTube",
		Description: "YouTube Data API - vídeos, playlists, canais",
		Command:     "npx",
		Args:        []string{"-y", "youtube-mcp-server"},
		Env:         map[string]string{"YOUTUBE_API_KEY": ""},
		IsEnabled:   false,
	}

	cc.mcpServers["twitter"] = &MCPServer{
		Name:        "Twitter/X",
		Description: "API X/Twitter - tweets, timeline, busca",
		Command:     "npx",
		Args:        []string{"-y", "twitter-mcp-server"},
		Env:         map[string]string{"TWITTER_BEARER_TOKEN": ""},
		IsEnabled:   false,
	}

	// ==================== UTILIDADES ====================

	cc.mcpServers["time"] = &MCPServer{
		Name:        "Time",
		Description: "Operações de data/hora e timezone",
		Command:     "npx",
		Args:        []string{"-y", "@modelcontextprotocol/server-time"},
		IsEnabled:   false,
	}

	cc.mcpServers["memory"] = &MCPServer{
		Name:        "Memory",
		Description: "Memória persistente para o modelo",
		Command:     "npx",
		Args:        []string{"-y", "@modelcontextprotocol/server-memory"},
		IsEnabled:   false,
	}

	cc.mcpServers["sequential-thinking"] = &MCPServer{
		Name:        "Sequential Thinking",
		Description: "Raciocínio passo a passo estruturado",
		Command:     "npx",
		Args:        []string{"-y", "@modelcontextprotocol/server-sequential-thinking"},
		IsEnabled:   false,
	}

	cc.mcpServers["everything"] = &MCPServer{
		Name:        "Everything Search",
		Description: "Busca instantânea de arquivos (Windows)",
		Command:     "npx",
		Args:        []string{"-y", "everything-mcp-server"},
		IsEnabled:   false,
	}

	cc.mcpServers["clipboard"] = &MCPServer{
		Name:        "Clipboard",
		Description: "Acesso ao clipboard do sistema",
		Command:     "npx",
		Args:        []string{"-y", "clipboard-mcp-server"},
		IsEnabled:   false,
	}

	cc.mcpServers["screenshot"] = &MCPServer{
		Name:        "Screenshot",
		Description: "Captura de tela",
		Command:     "npx",
		Args:        []string{"-y", "screenshot-mcp-server"},
		IsEnabled:   false,
	}
}

// ==================== OPERAÇÕES ====================

// StartSession inicia sessão Claude Code
func (cc *ClaudeCode) StartSession(workingDir string) (*ClaudeSession, error) {
	if !cc.isAvailable {
		return nil, fmt.Errorf("Claude Code não está instalado")
	}

	session := &ClaudeSession{
		ID:         fmt.Sprintf("session_%d", time.Now().UnixNano()),
		StartedAt:  time.Now(),
		WorkingDir: workingDir,
		Messages:   make([]Message, 0),
		IsActive:   true,
	}

	cc.mu.Lock()
	cc.sessions[session.ID] = session
	cc.currentSession = session
	cc.mu.Unlock()

	return session, nil
}

// SendMessage envia mensagem para Claude Code
func (cc *ClaudeCode) SendMessage(ctx context.Context, message string) (string, error) {
	if !cc.isAvailable {
		return "", fmt.Errorf("Claude Code não está instalado")
	}

	// Executa claude com a mensagem
	cmd := exec.CommandContext(ctx, cc.binaryPath, "-p", message)
	if cc.currentSession != nil {
		cmd.Dir = cc.currentSession.WorkingDir
	}

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("erro ao executar Claude Code: %w", err)
	}

	response := string(output)

	// Registra na sessão
	if cc.currentSession != nil {
		cc.mu.Lock()
		cc.currentSession.Messages = append(cc.currentSession.Messages,
			Message{Role: "user", Content: message, Timestamp: time.Now()},
			Message{Role: "assistant", Content: response, Timestamp: time.Now()},
		)
		cc.mu.Unlock()
	}

	return response, nil
}

// SendMessageInteractive inicia sessão interativa
func (cc *ClaudeCode) SendMessageInteractive(ctx context.Context, message string, callback func(string)) error {
	if !cc.isAvailable {
		return fmt.Errorf("Claude Code não está instalado")
	}

	cmd := exec.CommandContext(ctx, cc.binaryPath)
	if cc.currentSession != nil {
		cmd.Dir = cc.currentSession.WorkingDir
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	// Envia mensagem
	stdin.Write([]byte(message + "\n"))

	// Lê resposta em tempo real
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		if callback != nil {
			callback(line)
		}
	}

	return cmd.Wait()
}

// ExecuteCommand executa comando específico
func (cc *ClaudeCode) ExecuteCommand(ctx context.Context, command string, args ...string) (string, error) {
	if !cc.isAvailable {
		return "", fmt.Errorf("Claude Code não está instalado")
	}

	fullArgs := append([]string{command}, args...)
	cmd := exec.CommandContext(ctx, cc.binaryPath, fullArgs...)

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(output), nil
}

// ==================== MCP SERVERS ====================

// EnableMCPServer habilita servidor MCP
func (cc *ClaudeCode) EnableMCPServer(serverID string) error {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	server, ok := cc.mcpServers[serverID]
	if !ok {
		return fmt.Errorf("servidor MCP não encontrado: %s", serverID)
	}

	server.IsEnabled = true

	// Atualiza configuração do Claude Code
	return cc.updateMCPConfig()
}

// DisableMCPServer desabilita servidor MCP
func (cc *ClaudeCode) DisableMCPServer(serverID string) error {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	server, ok := cc.mcpServers[serverID]
	if !ok {
		return fmt.Errorf("servidor MCP não encontrado: %s", serverID)
	}

	server.IsEnabled = false

	return cc.updateMCPConfig()
}

// ConfigureMCPServer configura servidor MCP
func (cc *ClaudeCode) ConfigureMCPServer(serverID string, env map[string]string) error {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	server, ok := cc.mcpServers[serverID]
	if !ok {
		return fmt.Errorf("servidor MCP não encontrado: %s", serverID)
	}

	for key, value := range env {
		server.Env[key] = value
	}

	return cc.updateMCPConfig()
}

// GetMCPServers retorna todos os servidores MCP
func (cc *ClaudeCode) GetMCPServers() map[string]*MCPServer {
	cc.mu.RLock()
	defer cc.mu.RUnlock()

	return cc.mcpServers
}

// GetEnabledMCPServers retorna servidores habilitados
func (cc *ClaudeCode) GetEnabledMCPServers() []*MCPServer {
	cc.mu.RLock()
	defer cc.mu.RUnlock()

	enabled := make([]*MCPServer, 0)
	for _, server := range cc.mcpServers {
		if server.IsEnabled {
			enabled = append(enabled, server)
		}
	}
	return enabled
}

// updateMCPConfig atualiza configuração MCP no Claude Code
func (cc *ClaudeCode) updateMCPConfig() error {
	configPath := filepath.Join(cc.configPath, "claude_desktop_config.json")

	// Estrutura do config
	config := struct {
		MCPServers map[string]struct {
			Command string            `json:"command"`
			Args    []string          `json:"args"`
			Env     map[string]string `json:"env,omitempty"`
		} `json:"mcpServers"`
	}{
		MCPServers: make(map[string]struct {
			Command string            `json:"command"`
			Args    []string          `json:"args"`
			Env     map[string]string `json:"env,omitempty"`
		}),
	}

	// Adiciona servidores habilitados
	for id, server := range cc.mcpServers {
		if server.IsEnabled {
			config.MCPServers[id] = struct {
				Command string            `json:"command"`
				Args    []string          `json:"args"`
				Env     map[string]string `json:"env,omitempty"`
			}{
				Command: server.Command,
				Args:    server.Args,
				Env:     server.Env,
			}
		}
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

// ==================== UBER INTEGRATION ====================

// UberIntegration integração específica com Uber
type UberIntegration struct {
	claudeCode *ClaudeCode
	clientID   string
	clientSecret string
	accessToken  string
}

// NewUberIntegration cria integração Uber
func NewUberIntegration(cc *ClaudeCode) *UberIntegration {
	return &UberIntegration{
		claudeCode: cc,
	}
}

// Configure configura credenciais Uber
func (ui *UberIntegration) Configure(clientID, clientSecret string) error {
	ui.clientID = clientID
	ui.clientSecret = clientSecret

	return ui.claudeCode.ConfigureMCPServer("uber", map[string]string{
		"UBER_CLIENT_ID":     clientID,
		"UBER_CLIENT_SECRET": clientSecret,
	})
}

// Enable habilita integração Uber
func (ui *UberIntegration) Enable() error {
	return ui.claudeCode.EnableMCPServer("uber")
}

// RequestRide solicita corrida (via Claude Code)
func (ui *UberIntegration) RequestRide(ctx context.Context, pickup, dropoff string) (string, error) {
	query := fmt.Sprintf("Use the Uber MCP server to request a ride from %s to %s", pickup, dropoff)
	return ui.claudeCode.SendMessage(ctx, query)
}

// GetEstimate obtém estimativa de preço
func (ui *UberIntegration) GetEstimate(ctx context.Context, pickup, dropoff string) (string, error) {
	query := fmt.Sprintf("Use the Uber MCP server to get a price estimate from %s to %s", pickup, dropoff)
	return ui.claudeCode.SendMessage(ctx, query)
}

// GetRideStatus obtém status da corrida
func (ui *UberIntegration) GetRideStatus(ctx context.Context) (string, error) {
	return ui.claudeCode.SendMessage(ctx, "Use the Uber MCP server to get the current ride status")
}

// ==================== STATUS ====================

// GetStatus retorna status do Claude Code
func (cc *ClaudeCode) GetStatus() map[string]interface{} {
	cc.mu.RLock()
	defer cc.mu.RUnlock()

	enabledServers := 0
	for _, s := range cc.mcpServers {
		if s.IsEnabled {
			enabledServers++
		}
	}

	return map[string]interface{}{
		"is_available":    cc.isAvailable,
		"binary_path":     cc.binaryPath,
		"config_path":     cc.configPath,
		"total_sessions":  len(cc.sessions),
		"total_mcp_servers": len(cc.mcpServers),
		"enabled_servers": enabledServers,
	}
}

// IsAvailable verifica disponibilidade
func (cc *ClaudeCode) IsAvailable() bool {
	return cc.isAvailable
}

// GetMCPServersByCategory retorna servidores por categoria
func (cc *ClaudeCode) GetMCPServersByCategory() map[string][]*MCPServer {
	cc.mu.RLock()
	defer cc.mu.RUnlock()

	categories := map[string][]*MCPServer{
		"Official":      {},
		"Productivity":  {},
		"Communication": {},
		"Development":   {},
		"Data":          {},
		"Transport":     {},
		"Finance":       {},
		"AI/ML":         {},
		"Media":         {},
		"Utilities":     {},
	}

	// Mapeamento
	categoryMap := map[string]string{
		"filesystem": "Official", "github": "Official", "gitlab": "Official",
		"slack": "Communication", "postgres": "Official", "sqlite": "Official",
		"puppeteer": "Official", "brave-search": "Official", "google-maps": "Official",
		"fetch": "Official", "notion": "Productivity", "todoist": "Productivity",
		"linear": "Productivity", "asana": "Productivity", "trello": "Productivity",
		"discord": "Communication", "telegram": "Communication", "email": "Communication",
		"docker": "Development", "kubernetes": "Development", "aws": "Development",
		"azure": "Development", "gcp": "Development", "vercel": "Development",
		"netlify": "Development", "bigquery": "Data", "snowflake": "Data",
		"mongodb": "Data", "redis": "Data", "uber": "Transport",
		"lyft": "Transport", "doordash": "Transport", "instacart": "Transport",
		"stripe": "Finance", "plaid": "Finance", "coinbase": "Finance",
		"openai": "AI/ML", "huggingface": "AI/ML", "replicate": "AI/ML",
		"spotify": "Media", "youtube": "Media", "twitter": "Media",
		"time": "Utilities", "memory": "Utilities", "sequential-thinking": "Utilities",
		"everything": "Utilities", "clipboard": "Utilities", "screenshot": "Utilities",
		"google-drive": "Official",
	}

	for id, server := range cc.mcpServers {
		if cat, ok := categoryMap[id]; ok {
			categories[cat] = append(categories[cat], server)
		}
	}

	return categories
}

// ListMCPServersFormatted retorna lista formatada de servidores
func (cc *ClaudeCode) ListMCPServersFormatted() string {
	cc.mu.RLock()
	defer cc.mu.RUnlock()

	var sb strings.Builder
	sb.WriteString("╔══════════════════════════════════════════════════════════════╗\n")
	sb.WriteString("║              MCP SERVERS DISPONÍVEIS                        ║\n")
	sb.WriteString("╚══════════════════════════════════════════════════════════════╝\n\n")

	categories := cc.GetMCPServersByCategory()
	for cat, servers := range categories {
		if len(servers) == 0 {
			continue
		}
		sb.WriteString(fmt.Sprintf("📦 %s:\n", cat))
		for _, s := range servers {
			status := "❌"
			if s.IsEnabled {
				status = "✅"
			}
			sb.WriteString(fmt.Sprintf("   %s %s - %s\n", status, s.Name, s.Description))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
