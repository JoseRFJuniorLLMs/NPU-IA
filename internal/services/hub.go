package services

import (
	"fmt"
	"sync"
)

// Hub centraliza todos os serviços
type Hub struct {
	// Google
	Google *GoogleServices

	// Microsoft
	Microsoft *MicrosoftServices

	// GitHub
	GitHub *GitHubServices

	// Social
	LinkedIn *LinkedInServices
	X        *XServices
	Discord  *DiscordServices
	Slack    *SlackServices
	Telegram *TelegramServices

	// Productivity
	Notion  *NotionServices
	Todoist *TodoistServices
	Spotify *SpotifyServices

	mu sync.RWMutex
}

// Config configuração de serviços
type Config struct {
	// Google
	GoogleCredentials string `yaml:"google_credentials"`

	// Microsoft
	MicrosoftClientID     string `yaml:"microsoft_client_id"`
	MicrosoftClientSecret string `yaml:"microsoft_client_secret"`
	MicrosoftTenantID     string `yaml:"microsoft_tenant_id"`

	// GitHub
	GitHubToken string `yaml:"github_token"`

	// LinkedIn
	LinkedInToken string `yaml:"linkedin_token"`

	// X (Twitter)
	XBearerToken string `yaml:"x_bearer_token"`

	// Discord
	DiscordBotToken string `yaml:"discord_bot_token"`

	// Slack
	SlackBotToken string `yaml:"slack_bot_token"`

	// Telegram
	TelegramBotToken string `yaml:"telegram_bot_token"`

	// Notion
	NotionAPIKey string `yaml:"notion_api_key"`

	// Todoist
	TodoistAPIKey string `yaml:"todoist_api_key"`

	// Spotify
	SpotifyToken string `yaml:"spotify_token"`
}

// NewHub cria hub de serviços
func NewHub(cfg Config) (*Hub, error) {
	hub := &Hub{}

	// Inicializa serviços configurados
	if cfg.GoogleCredentials != "" {
		google, err := NewGoogleServices(cfg.GoogleCredentials)
		if err == nil {
			hub.Google = google
			fmt.Println("  ✓ Google Services conectado")
		}
	}

	if cfg.MicrosoftClientID != "" {
		ms, err := NewMicrosoftServices(cfg.MicrosoftClientID, cfg.MicrosoftClientSecret, cfg.MicrosoftTenantID)
		if err == nil {
			hub.Microsoft = ms
			fmt.Println("  ✓ Microsoft Services conectado")
		}
	}

	if cfg.GitHubToken != "" {
		gh, err := NewGitHubServices(cfg.GitHubToken)
		if err == nil {
			hub.GitHub = gh
			fmt.Println("  ✓ GitHub conectado")
		}
	}

	if cfg.LinkedInToken != "" {
		hub.LinkedIn = NewLinkedInServices(cfg.LinkedInToken)
		fmt.Println("  ✓ LinkedIn conectado")
	}

	if cfg.XBearerToken != "" {
		hub.X = NewXServices(cfg.XBearerToken)
		fmt.Println("  ✓ X (Twitter) conectado")
	}

	if cfg.DiscordBotToken != "" {
		hub.Discord = NewDiscordServices(cfg.DiscordBotToken)
		fmt.Println("  ✓ Discord conectado")
	}

	if cfg.SlackBotToken != "" {
		hub.Slack = NewSlackServices(cfg.SlackBotToken)
		fmt.Println("  ✓ Slack conectado")
	}

	if cfg.TelegramBotToken != "" {
		hub.Telegram = NewTelegramServices(cfg.TelegramBotToken)
		fmt.Println("  ✓ Telegram conectado")
	}

	if cfg.NotionAPIKey != "" {
		hub.Notion = NewNotionServices(cfg.NotionAPIKey)
		fmt.Println("  ✓ Notion conectado")
	}

	if cfg.TodoistAPIKey != "" {
		hub.Todoist = NewTodoistServices(cfg.TodoistAPIKey)
		fmt.Println("  ✓ Todoist conectado")
	}

	if cfg.SpotifyToken != "" {
		hub.Spotify = NewSpotifyServices(cfg.SpotifyToken)
		fmt.Println("  ✓ Spotify conectado")
	}

	return hub, nil
}

// Execute executa ação em um serviço
func (h *Hub) Execute(service, action string, params map[string]interface{}) (interface{}, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	switch service {
	// ==================== GOOGLE ====================
	case "gmail":
		if h.Google == nil {
			return nil, fmt.Errorf("Google não configurado")
		}
		switch action {
		case "list":
			query, _ := params["query"].(string)
			return h.Google.ListEmails(query, 10)
		case "send":
			to, _ := params["to"].(string)
			subject, _ := params["subject"].(string)
			body, _ := params["body"].(string)
			return nil, h.Google.SendEmail(to, subject, body)
		}

	case "drive":
		if h.Google == nil {
			return nil, fmt.Errorf("Google não configurado")
		}
		switch action {
		case "list":
			query, _ := params["query"].(string)
			return h.Google.ListDriveFiles(query, 20)
		case "search":
			name, _ := params["name"].(string)
			return h.Google.SearchDrive(name)
		}

	case "calendar":
		if h.Google == nil {
			return nil, fmt.Errorf("Google não configurado")
		}
		switch action {
		case "today":
			return h.Google.GetTodayEvents()
		}

	// ==================== GITHUB ====================
	case "github":
		if h.GitHub == nil {
			return nil, fmt.Errorf("GitHub não configurado")
		}
		switch action {
		case "repos":
			return h.GitHub.ListRepos()
		case "issues":
			owner, _ := params["owner"].(string)
			repo, _ := params["repo"].(string)
			return h.GitHub.ListIssues(owner, repo)
		case "prs":
			owner, _ := params["owner"].(string)
			repo, _ := params["repo"].(string)
			return h.GitHub.ListPullRequests(owner, repo)
		case "notifications":
			return h.GitHub.ListNotifications()
		}

	// ==================== SPOTIFY ====================
	case "spotify":
		if h.Spotify == nil {
			return nil, fmt.Errorf("Spotify não configurado")
		}
		switch action {
		case "playing":
			return h.Spotify.GetCurrentlyPlaying()
		case "play":
			return nil, h.Spotify.Play()
		case "pause":
			return nil, h.Spotify.Pause()
		case "next":
			return nil, h.Spotify.Next()
		case "previous":
			return nil, h.Spotify.Previous()
		case "search":
			query, _ := params["query"].(string)
			return h.Spotify.Search(query, 5)
		}

	// ==================== TODOIST ====================
	case "todoist":
		if h.Todoist == nil {
			return nil, fmt.Errorf("Todoist não configurado")
		}
		switch action {
		case "list":
			return h.Todoist.GetTasks()
		case "create":
			content, _ := params["content"].(string)
			return h.Todoist.CreateTask(content, "", "", 1)
		case "complete":
			id, _ := params["id"].(string)
			return nil, h.Todoist.CompleteTask(id)
		}

	// ==================== NOTION ====================
	case "notion":
		if h.Notion == nil {
			return nil, fmt.Errorf("Notion não configurado")
		}
		switch action {
		case "search":
			query, _ := params["query"].(string)
			return h.Notion.SearchPages(query)
		}

	// ==================== DISCORD ====================
	case "discord":
		if h.Discord == nil {
			return nil, fmt.Errorf("Discord não configurado")
		}
		switch action {
		case "send":
			channel, _ := params["channel"].(string)
			message, _ := params["message"].(string)
			return h.Discord.SendMessage(channel, message)
		case "guilds":
			return h.Discord.GetGuilds()
		}

	// ==================== SLACK ====================
	case "slack":
		if h.Slack == nil {
			return nil, fmt.Errorf("Slack não configurado")
		}
		switch action {
		case "send":
			channel, _ := params["channel"].(string)
			message, _ := params["message"].(string)
			return nil, h.Slack.SendMessage(channel, message)
		case "channels":
			return h.Slack.ListChannels()
		}

	// ==================== TELEGRAM ====================
	case "telegram":
		if h.Telegram == nil {
			return nil, fmt.Errorf("Telegram não configurado")
		}
		switch action {
		case "send":
			chat, _ := params["chat"].(string)
			message, _ := params["message"].(string)
			return nil, h.Telegram.SendMessage(chat, message)
		}

	// ==================== X (TWITTER) ====================
	case "x", "twitter":
		if h.X == nil {
			return nil, fmt.Errorf("X não configurado")
		}
		switch action {
		case "post":
			text, _ := params["text"].(string)
			return h.X.PostTweet(text)
		case "search":
			query, _ := params["query"].(string)
			return h.X.SearchTweets(query, 10)
		}

	// ==================== LINKEDIN ====================
	case "linkedin":
		if h.LinkedIn == nil {
			return nil, fmt.Errorf("LinkedIn não configurado")
		}
		switch action {
		case "profile":
			return h.LinkedIn.GetProfile()
		case "post":
			text, _ := params["text"].(string)
			return nil, h.LinkedIn.SharePost(text)
		}
	}

	return nil, fmt.Errorf("serviço ou ação não encontrado: %s/%s", service, action)
}

// GetConnectedServices retorna serviços conectados
func (h *Hub) GetConnectedServices() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	services := []string{}

	if h.Google != nil {
		services = append(services, "google")
	}
	if h.Microsoft != nil {
		services = append(services, "microsoft")
	}
	if h.GitHub != nil {
		services = append(services, "github")
	}
	if h.LinkedIn != nil {
		services = append(services, "linkedin")
	}
	if h.X != nil {
		services = append(services, "x")
	}
	if h.Discord != nil {
		services = append(services, "discord")
	}
	if h.Slack != nil {
		services = append(services, "slack")
	}
	if h.Telegram != nil {
		services = append(services, "telegram")
	}
	if h.Notion != nil {
		services = append(services, "notion")
	}
	if h.Todoist != nil {
		services = append(services, "todoist")
	}
	if h.Spotify != nil {
		services = append(services, "spotify")
	}

	return services
}
