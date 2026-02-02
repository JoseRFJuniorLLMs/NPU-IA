package services

// ==================== MCP SERVERS - EXTENDED PART 2 ====================
// Social Media, Gaming, Education, News, Weather, Sports

// RegisterSocialMediaMCP registra servidores de Redes Sociais
func (cc *ClaudeCode) RegisterSocialMediaMCP() {
	// Instagram
	cc.mcpServers["instagram"] = &MCPServer{
		Name:        "Instagram",
		Description: "Posts, stories, mensagens Instagram",
		Command:     "npx",
		Args:        []string{"-y", "instagram-mcp-server"},
		Env:         map[string]string{"INSTAGRAM_ACCESS_TOKEN": ""},
		IsEnabled:   false,
	}

	// Facebook
	cc.mcpServers["facebook"] = &MCPServer{
		Name:        "Facebook",
		Description: "Posts, páginas, grupos Facebook",
		Command:     "npx",
		Args:        []string{"-y", "facebook-mcp-server"},
		Env:         map[string]string{"FACEBOOK_ACCESS_TOKEN": ""},
		IsEnabled:   false,
	}

	// TikTok
	cc.mcpServers["tiktok"] = &MCPServer{
		Name:        "TikTok",
		Description: "Vídeos e analytics TikTok",
		Command:     "npx",
		Args:        []string{"-y", "tiktok-mcp-server"},
		Env:         map[string]string{"TIKTOK_ACCESS_TOKEN": ""},
		IsEnabled:   false,
	}

	// LinkedIn
	cc.mcpServers["linkedin"] = &MCPServer{
		Name:        "LinkedIn",
		Description: "Perfil, posts, conexões LinkedIn",
		Command:     "npx",
		Args:        []string{"-y", "linkedin-mcp-server"},
		Env:         map[string]string{"LINKEDIN_ACCESS_TOKEN": ""},
		IsEnabled:   false,
	}

	// Pinterest
	cc.mcpServers["pinterest"] = &MCPServer{
		Name:        "Pinterest",
		Description: "Pins e boards Pinterest",
		Command:     "npx",
		Args:        []string{"-y", "pinterest-mcp-server"},
		Env:         map[string]string{"PINTEREST_ACCESS_TOKEN": ""},
		IsEnabled:   false,
	}

	// Reddit
	cc.mcpServers["reddit"] = &MCPServer{
		Name:        "Reddit",
		Description: "Posts, comentários, subreddits",
		Command:     "npx",
		Args:        []string{"-y", "reddit-mcp-server"},
		Env:         map[string]string{"REDDIT_CLIENT_ID": "", "REDDIT_CLIENT_SECRET": ""},
		IsEnabled:   false,
	}

	// Snapchat
	cc.mcpServers["snapchat"] = &MCPServer{
		Name:        "Snapchat",
		Description: "Integração Snapchat",
		Command:     "npx",
		Args:        []string{"-y", "snapchat-mcp-server"},
		IsEnabled:   false,
	}

	// Mastodon
	cc.mcpServers["mastodon"] = &MCPServer{
		Name:        "Mastodon",
		Description: "Rede social descentralizada",
		Command:     "npx",
		Args:        []string{"-y", "mastodon-mcp-server"},
		Env:         map[string]string{"MASTODON_INSTANCE": "", "MASTODON_ACCESS_TOKEN": ""},
		IsEnabled:   false,
	}

	// Bluesky
	cc.mcpServers["bluesky"] = &MCPServer{
		Name:        "Bluesky",
		Description: "Rede social Bluesky/AT Protocol",
		Command:     "npx",
		Args:        []string{"-y", "bluesky-mcp-server"},
		Env:         map[string]string{"BLUESKY_HANDLE": "", "BLUESKY_PASSWORD": ""},
		IsEnabled:   false,
	}

	// Threads
	cc.mcpServers["threads"] = &MCPServer{
		Name:        "Threads",
		Description: "Meta Threads",
		Command:     "npx",
		Args:        []string{"-y", "threads-mcp-server"},
		Env:         map[string]string{"THREADS_ACCESS_TOKEN": ""},
		IsEnabled:   false,
	}

	// WhatsApp Business
	cc.mcpServers["whatsapp"] = &MCPServer{
		Name:        "WhatsApp Business",
		Description: "API WhatsApp Business",
		Command:     "npx",
		Args:        []string{"-y", "whatsapp-mcp-server"},
		Env:         map[string]string{"WHATSAPP_TOKEN": "", "WHATSAPP_PHONE_ID": ""},
		IsEnabled:   false,
	}

	// WeChat
	cc.mcpServers["wechat"] = &MCPServer{
		Name:        "WeChat",
		Description: "Mensagens e mini-programs WeChat",
		Command:     "npx",
		Args:        []string{"-y", "wechat-mcp-server"},
		Env:         map[string]string{"WECHAT_APP_ID": "", "WECHAT_APP_SECRET": ""},
		IsEnabled:   false,
	}

	// Tumblr
	cc.mcpServers["tumblr"] = &MCPServer{
		Name:        "Tumblr",
		Description: "Blog e posts Tumblr",
		Command:     "npx",
		Args:        []string{"-y", "tumblr-mcp-server"},
		Env:         map[string]string{"TUMBLR_CONSUMER_KEY": "", "TUMBLR_CONSUMER_SECRET": ""},
		IsEnabled:   false,
	}

	// Twitch
	cc.mcpServers["twitch"] = &MCPServer{
		Name:        "Twitch",
		Description: "Streams, chat, clips Twitch",
		Command:     "npx",
		Args:        []string{"-y", "twitch-mcp-server"},
		Env:         map[string]string{"TWITCH_CLIENT_ID": "", "TWITCH_CLIENT_SECRET": ""},
		IsEnabled:   false,
	}

	// Kick
	cc.mcpServers["kick"] = &MCPServer{
		Name:        "Kick",
		Description: "Streaming Kick",
		Command:     "npx",
		Args:        []string{"-y", "kick-mcp-server"},
		IsEnabled:   false,
	}

	// Clubhouse
	cc.mcpServers["clubhouse"] = &MCPServer{
		Name:        "Clubhouse",
		Description: "Áudio social Clubhouse",
		Command:     "npx",
		Args:        []string{"-y", "clubhouse-mcp-server"},
		IsEnabled:   false,
	}

	// BeReal
	cc.mcpServers["bereal"] = &MCPServer{
		Name:        "BeReal",
		Description: "App BeReal",
		Command:     "npx",
		Args:        []string{"-y", "bereal-mcp-server"},
		IsEnabled:   false,
	}
}

// RegisterGamingMCP registra servidores de Gaming
func (cc *ClaudeCode) RegisterGamingMCP() {
	// Steam
	cc.mcpServers["steam"] = &MCPServer{
		Name:        "Steam",
		Description: "Biblioteca, amigos, achievements Steam",
		Command:     "npx",
		Args:        []string{"-y", "steam-mcp-server"},
		Env:         map[string]string{"STEAM_API_KEY": "", "STEAM_ID": ""},
		IsEnabled:   false,
	}

	// PlayStation Network
	cc.mcpServers["playstation"] = &MCPServer{
		Name:        "PlayStation Network",
		Description: "PSN - troféus, amigos, jogos",
		Command:     "npx",
		Args:        []string{"-y", "psn-mcp-server"},
		Env:         map[string]string{"PSN_NPSSO": ""},
		IsEnabled:   false,
	}

	// Xbox Live
	cc.mcpServers["xbox"] = &MCPServer{
		Name:        "Xbox Live",
		Description: "Xbox - achievements, amigos, Game Pass",
		Command:     "npx",
		Args:        []string{"-y", "xbox-mcp-server"},
		Env:         map[string]string{"XBOX_CLIENT_ID": "", "XBOX_CLIENT_SECRET": ""},
		IsEnabled:   false,
	}

	// Nintendo
	cc.mcpServers["nintendo"] = &MCPServer{
		Name:        "Nintendo",
		Description: "Nintendo Switch Online",
		Command:     "npx",
		Args:        []string{"-y", "nintendo-mcp-server"},
		IsEnabled:   false,
	}

	// Epic Games
	cc.mcpServers["epic-games"] = &MCPServer{
		Name:        "Epic Games",
		Description: "Epic Games Store e biblioteca",
		Command:     "npx",
		Args:        []string{"-y", "epic-games-mcp-server"},
		IsEnabled:   false,
	}

	// GOG
	cc.mcpServers["gog"] = &MCPServer{
		Name:        "GOG",
		Description: "GOG Galaxy biblioteca",
		Command:     "npx",
		Args:        []string{"-y", "gog-mcp-server"},
		IsEnabled:   false,
	}

	// Battle.net
	cc.mcpServers["battlenet"] = &MCPServer{
		Name:        "Battle.net",
		Description: "Blizzard Battle.net",
		Command:     "npx",
		Args:        []string{"-y", "battlenet-mcp-server"},
		Env:         map[string]string{"BNET_CLIENT_ID": "", "BNET_CLIENT_SECRET": ""},
		IsEnabled:   false,
	}

	// Riot Games
	cc.mcpServers["riot-games"] = &MCPServer{
		Name:        "Riot Games",
		Description: "LoL, Valorant, TFT stats",
		Command:     "npx",
		Args:        []string{"-y", "riot-mcp-server"},
		Env:         map[string]string{"RIOT_API_KEY": ""},
		IsEnabled:   false,
	}

	// RAWG
	cc.mcpServers["rawg"] = &MCPServer{
		Name:        "RAWG",
		Description: "Database de jogos RAWG",
		Command:     "npx",
		Args:        []string{"-y", "rawg-mcp-server"},
		Env:         map[string]string{"RAWG_API_KEY": ""},
		IsEnabled:   false,
	}

	// IGDB
	cc.mcpServers["igdb"] = &MCPServer{
		Name:        "IGDB",
		Description: "Internet Game Database",
		Command:     "npx",
		Args:        []string{"-y", "igdb-mcp-server"},
		Env:         map[string]string{"IGDB_CLIENT_ID": "", "IGDB_ACCESS_TOKEN": ""},
		IsEnabled:   false,
	}

	// Roblox
	cc.mcpServers["roblox"] = &MCPServer{
		Name:        "Roblox",
		Description: "Plataforma Roblox",
		Command:     "npx",
		Args:        []string{"-y", "roblox-mcp-server"},
		IsEnabled:   false,
	}

	// Minecraft
	cc.mcpServers["minecraft"] = &MCPServer{
		Name:        "Minecraft",
		Description: "Servidores e perfis Minecraft",
		Command:     "npx",
		Args:        []string{"-y", "minecraft-mcp-server"},
		IsEnabled:   false,
	}
}

// RegisterEducationMCP registra servidores de Educação
func (cc *ClaudeCode) RegisterEducationMCP() {
	// Coursera
	cc.mcpServers["coursera"] = &MCPServer{
		Name:        "Coursera",
		Description: "Cursos online Coursera",
		Command:     "npx",
		Args:        []string{"-y", "coursera-mcp-server"},
		IsEnabled:   false,
	}

	// Udemy
	cc.mcpServers["udemy"] = &MCPServer{
		Name:        "Udemy",
		Description: "Cursos Udemy",
		Command:     "npx",
		Args:        []string{"-y", "udemy-mcp-server"},
		IsEnabled:   false,
	}

	// edX
	cc.mcpServers["edx"] = &MCPServer{
		Name:        "edX",
		Description: "Cursos universitários edX",
		Command:     "npx",
		Args:        []string{"-y", "edx-mcp-server"},
		IsEnabled:   false,
	}

	// Khan Academy
	cc.mcpServers["khan-academy"] = &MCPServer{
		Name:        "Khan Academy",
		Description: "Educação gratuita Khan Academy",
		Command:     "npx",
		Args:        []string{"-y", "khan-academy-mcp-server"},
		IsEnabled:   false,
	}

	// Duolingo
	cc.mcpServers["duolingo"] = &MCPServer{
		Name:        "Duolingo",
		Description: "Aprendizado de idiomas",
		Command:     "npx",
		Args:        []string{"-y", "duolingo-mcp-server"},
		IsEnabled:   false,
	}

	// Anki
	cc.mcpServers["anki"] = &MCPServer{
		Name:        "Anki",
		Description: "Flashcards Anki",
		Command:     "npx",
		Args:        []string{"-y", "anki-mcp-server"},
		IsEnabled:   false,
	}

	// Quizlet
	cc.mcpServers["quizlet"] = &MCPServer{
		Name:        "Quizlet",
		Description: "Flashcards e estudos Quizlet",
		Command:     "npx",
		Args:        []string{"-y", "quizlet-mcp-server"},
		IsEnabled:   false,
	}

	// Canvas LMS
	cc.mcpServers["canvas"] = &MCPServer{
		Name:        "Canvas LMS",
		Description: "Sistema de gestão de aprendizado",
		Command:     "npx",
		Args:        []string{"-y", "canvas-mcp-server"},
		Env:         map[string]string{"CANVAS_URL": "", "CANVAS_TOKEN": ""},
		IsEnabled:   false,
	}

	// Google Classroom
	cc.mcpServers["google-classroom"] = &MCPServer{
		Name:        "Google Classroom",
		Description: "Salas de aula Google",
		Command:     "npx",
		Args:        []string{"-y", "google-classroom-mcp-server"},
		IsEnabled:   false,
	}

	// Moodle
	cc.mcpServers["moodle"] = &MCPServer{
		Name:        "Moodle",
		Description: "Plataforma Moodle",
		Command:     "npx",
		Args:        []string{"-y", "moodle-mcp-server"},
		Env:         map[string]string{"MOODLE_URL": "", "MOODLE_TOKEN": ""},
		IsEnabled:   false,
	}

	// Skillshare
	cc.mcpServers["skillshare"] = &MCPServer{
		Name:        "Skillshare",
		Description: "Cursos criativos Skillshare",
		Command:     "npx",
		Args:        []string{"-y", "skillshare-mcp-server"},
		IsEnabled:   false,
	}

	// LinkedIn Learning
	cc.mcpServers["linkedin-learning"] = &MCPServer{
		Name:        "LinkedIn Learning",
		Description: "Cursos LinkedIn Learning",
		Command:     "npx",
		Args:        []string{"-y", "linkedin-learning-mcp-server"},
		IsEnabled:   false,
	}

	// Pluralsight
	cc.mcpServers["pluralsight"] = &MCPServer{
		Name:        "Pluralsight",
		Description: "Cursos de tecnologia Pluralsight",
		Command:     "npx",
		Args:        []string{"-y", "pluralsight-mcp-server"},
		IsEnabled:   false,
	}

	// Codecademy
	cc.mcpServers["codecademy"] = &MCPServer{
		Name:        "Codecademy",
		Description: "Aprender programação",
		Command:     "npx",
		Args:        []string{"-y", "codecademy-mcp-server"},
		IsEnabled:   false,
	}

	// LeetCode
	cc.mcpServers["leetcode"] = &MCPServer{
		Name:        "LeetCode",
		Description: "Problemas de programação",
		Command:     "npx",
		Args:        []string{"-y", "leetcode-mcp-server"},
		IsEnabled:   false,
	}

	// HackerRank
	cc.mcpServers["hackerrank"] = &MCPServer{
		Name:        "HackerRank",
		Description: "Desafios de código",
		Command:     "npx",
		Args:        []string{"-y", "hackerrank-mcp-server"},
		IsEnabled:   false,
	}

	// Wolfram Alpha
	cc.mcpServers["wolfram"] = &MCPServer{
		Name:        "Wolfram Alpha",
		Description: "Computação matemática",
		Command:     "npx",
		Args:        []string{"-y", "wolfram-mcp-server"},
		Env:         map[string]string{"WOLFRAM_APP_ID": ""},
		IsEnabled:   false,
	}

	// Wikipedia
	cc.mcpServers["wikipedia"] = &MCPServer{
		Name:        "Wikipedia",
		Description: "Enciclopédia Wikipedia",
		Command:     "npx",
		Args:        []string{"-y", "wikipedia-mcp-server"},
		IsEnabled:   false,
	}

	// arXiv
	cc.mcpServers["arxiv"] = &MCPServer{
		Name:        "arXiv",
		Description: "Papers científicos arXiv",
		Command:     "npx",
		Args:        []string{"-y", "arxiv-mcp-server"},
		IsEnabled:   false,
	}

	// Google Scholar
	cc.mcpServers["google-scholar"] = &MCPServer{
		Name:        "Google Scholar",
		Description: "Artigos acadêmicos",
		Command:     "npx",
		Args:        []string{"-y", "google-scholar-mcp-server"},
		IsEnabled:   false,
	}

	// Semantic Scholar
	cc.mcpServers["semantic-scholar"] = &MCPServer{
		Name:        "Semantic Scholar",
		Description: "Pesquisa acadêmica AI",
		Command:     "npx",
		Args:        []string{"-y", "semantic-scholar-mcp-server"},
		Env:         map[string]string{"S2_API_KEY": ""},
		IsEnabled:   false,
	}
}

// RegisterNewsMCP registra servidores de Notícias
func (cc *ClaudeCode) RegisterNewsMCP() {
	// NewsAPI
	cc.mcpServers["newsapi"] = &MCPServer{
		Name:        "NewsAPI",
		Description: "Agregador de notícias NewsAPI",
		Command:     "npx",
		Args:        []string{"-y", "newsapi-mcp-server"},
		Env:         map[string]string{"NEWSAPI_KEY": ""},
		IsEnabled:   false,
	}

	// Google News
	cc.mcpServers["google-news"] = &MCPServer{
		Name:        "Google News",
		Description: "Notícias Google",
		Command:     "npx",
		Args:        []string{"-y", "google-news-mcp-server"},
		IsEnabled:   false,
	}

	// Hacker News
	cc.mcpServers["hackernews"] = &MCPServer{
		Name:        "Hacker News",
		Description: "Tech news YC",
		Command:     "npx",
		Args:        []string{"-y", "hackernews-mcp-server"},
		IsEnabled:   false,
	}

	// Product Hunt
	cc.mcpServers["producthunt"] = &MCPServer{
		Name:        "Product Hunt",
		Description: "Novos produtos tech",
		Command:     "npx",
		Args:        []string{"-y", "producthunt-mcp-server"},
		Env:         map[string]string{"PH_ACCESS_TOKEN": ""},
		IsEnabled:   false,
	}

	// TechCrunch
	cc.mcpServers["techcrunch"] = &MCPServer{
		Name:        "TechCrunch",
		Description: "Notícias TechCrunch",
		Command:     "npx",
		Args:        []string{"-y", "techcrunch-mcp-server"},
		IsEnabled:   false,
	}

	// The Verge
	cc.mcpServers["theverge"] = &MCPServer{
		Name:        "The Verge",
		Description: "Tech news The Verge",
		Command:     "npx",
		Args:        []string{"-y", "theverge-mcp-server"},
		IsEnabled:   false,
	}

	// Ars Technica
	cc.mcpServers["arstechnica"] = &MCPServer{
		Name:        "Ars Technica",
		Description: "Tech journalism Ars",
		Command:     "npx",
		Args:        []string{"-y", "arstechnica-mcp-server"},
		IsEnabled:   false,
	}

	// RSS Feed
	cc.mcpServers["rss"] = &MCPServer{
		Name:        "RSS/Atom Feeds",
		Description: "Leitor de feeds RSS/Atom",
		Command:     "npx",
		Args:        []string{"-y", "rss-mcp-server"},
		IsEnabled:   false,
	}

	// Feedly
	cc.mcpServers["feedly"] = &MCPServer{
		Name:        "Feedly",
		Description: "Agregador Feedly",
		Command:     "npx",
		Args:        []string{"-y", "feedly-mcp-server"},
		Env:         map[string]string{"FEEDLY_ACCESS_TOKEN": ""},
		IsEnabled:   false,
	}

	// Pocket
	cc.mcpServers["pocket"] = &MCPServer{
		Name:        "Pocket",
		Description: "Salvar artigos Pocket",
		Command:     "npx",
		Args:        []string{"-y", "pocket-mcp-server"},
		Env:         map[string]string{"POCKET_CONSUMER_KEY": "", "POCKET_ACCESS_TOKEN": ""},
		IsEnabled:   false,
	}

	// Instapaper
	cc.mcpServers["instapaper"] = &MCPServer{
		Name:        "Instapaper",
		Description: "Salvar para ler depois",
		Command:     "npx",
		Args:        []string{"-y", "instapaper-mcp-server"},
		IsEnabled:   false,
	}

	// Medium
	cc.mcpServers["medium"] = &MCPServer{
		Name:        "Medium",
		Description: "Artigos Medium",
		Command:     "npx",
		Args:        []string{"-y", "medium-mcp-server"},
		Env:         map[string]string{"MEDIUM_ACCESS_TOKEN": ""},
		IsEnabled:   false,
	}

	// Substack
	cc.mcpServers["substack"] = &MCPServer{
		Name:        "Substack",
		Description: "Newsletters Substack",
		Command:     "npx",
		Args:        []string{"-y", "substack-mcp-server"},
		IsEnabled:   false,
	}

	// Dev.to
	cc.mcpServers["devto"] = &MCPServer{
		Name:        "Dev.to",
		Description: "Comunidade Dev.to",
		Command:     "npx",
		Args:        []string{"-y", "devto-mcp-server"},
		Env:         map[string]string{"DEVTO_API_KEY": ""},
		IsEnabled:   false,
	}

	// Hashnode
	cc.mcpServers["hashnode"] = &MCPServer{
		Name:        "Hashnode",
		Description: "Blogs de desenvolvedores",
		Command:     "npx",
		Args:        []string{"-y", "hashnode-mcp-server"},
		IsEnabled:   false,
	}
}

// RegisterWeatherMCP registra servidores de Clima
func (cc *ClaudeCode) RegisterWeatherMCP() {
	// OpenWeatherMap
	cc.mcpServers["openweathermap"] = &MCPServer{
		Name:        "OpenWeatherMap",
		Description: "Previsão do tempo OWM",
		Command:     "npx",
		Args:        []string{"-y", "openweathermap-mcp-server"},
		Env:         map[string]string{"OWM_API_KEY": ""},
		IsEnabled:   false,
	}

	// Weather.com
	cc.mcpServers["weather-com"] = &MCPServer{
		Name:        "Weather.com",
		Description: "The Weather Company",
		Command:     "npx",
		Args:        []string{"-y", "weather-com-mcp-server"},
		Env:         map[string]string{"TWC_API_KEY": ""},
		IsEnabled:   false,
	}

	// AccuWeather
	cc.mcpServers["accuweather"] = &MCPServer{
		Name:        "AccuWeather",
		Description: "Previsões AccuWeather",
		Command:     "npx",
		Args:        []string{"-y", "accuweather-mcp-server"},
		Env:         map[string]string{"ACCUWEATHER_API_KEY": ""},
		IsEnabled:   false,
	}

	// Tomorrow.io
	cc.mcpServers["tomorrow-io"] = &MCPServer{
		Name:        "Tomorrow.io",
		Description: "Weather API Tomorrow.io",
		Command:     "npx",
		Args:        []string{"-y", "tomorrow-io-mcp-server"},
		Env:         map[string]string{"TOMORROW_API_KEY": ""},
		IsEnabled:   false,
	}

	// Visual Crossing
	cc.mcpServers["visualcrossing"] = &MCPServer{
		Name:        "Visual Crossing",
		Description: "Historical weather data",
		Command:     "npx",
		Args:        []string{"-y", "visualcrossing-mcp-server"},
		Env:         map[string]string{"VC_API_KEY": ""},
		IsEnabled:   false,
	}

	// Weatherbit
	cc.mcpServers["weatherbit"] = &MCPServer{
		Name:        "Weatherbit",
		Description: "Weather and air quality",
		Command:     "npx",
		Args:        []string{"-y", "weatherbit-mcp-server"},
		Env:         map[string]string{"WEATHERBIT_API_KEY": ""},
		IsEnabled:   false,
	}

	// Windy
	cc.mcpServers["windy"] = &MCPServer{
		Name:        "Windy",
		Description: "Mapas de vento e clima",
		Command:     "npx",
		Args:        []string{"-y", "windy-mcp-server"},
		IsEnabled:   false,
	}

	// AirVisual
	cc.mcpServers["airvisual"] = &MCPServer{
		Name:        "AirVisual/IQAir",
		Description: "Qualidade do ar mundial",
		Command:     "npx",
		Args:        []string{"-y", "airvisual-mcp-server"},
		Env:         map[string]string{"AIRVISUAL_API_KEY": ""},
		IsEnabled:   false,
	}
}

// RegisterSportsMCP registra servidores de Esportes
func (cc *ClaudeCode) RegisterSportsMCP() {
	// ESPN
	cc.mcpServers["espn"] = &MCPServer{
		Name:        "ESPN",
		Description: "Notícias e resultados ESPN",
		Command:     "npx",
		Args:        []string{"-y", "espn-mcp-server"},
		IsEnabled:   false,
	}

	// NBA
	cc.mcpServers["nba"] = &MCPServer{
		Name:        "NBA",
		Description: "Estatísticas NBA",
		Command:     "npx",
		Args:        []string{"-y", "nba-mcp-server"},
		IsEnabled:   false,
	}

	// NFL
	cc.mcpServers["nfl"] = &MCPServer{
		Name:        "NFL",
		Description: "Estatísticas NFL",
		Command:     "npx",
		Args:        []string{"-y", "nfl-mcp-server"},
		IsEnabled:   false,
	}

	// MLB
	cc.mcpServers["mlb"] = &MCPServer{
		Name:        "MLB",
		Description: "Estatísticas MLB",
		Command:     "npx",
		Args:        []string{"-y", "mlb-mcp-server"},
		IsEnabled:   false,
	}

	// NHL
	cc.mcpServers["nhl"] = &MCPServer{
		Name:        "NHL",
		Description: "Estatísticas NHL",
		Command:     "npx",
		Args:        []string{"-y", "nhl-mcp-server"},
		IsEnabled:   false,
	}

	// Football-Data (Soccer)
	cc.mcpServers["football-data"] = &MCPServer{
		Name:        "Football-Data",
		Description: "Futebol mundial - ligas, times, jogos",
		Command:     "npx",
		Args:        []string{"-y", "football-data-mcp-server"},
		Env:         map[string]string{"FOOTBALL_DATA_API_KEY": ""},
		IsEnabled:   false,
	}

	// API-Football
	cc.mcpServers["api-football"] = &MCPServer{
		Name:        "API-Football",
		Description: "Dados de futebol completos",
		Command:     "npx",
		Args:        []string{"-y", "api-football-mcp-server"},
		Env:         map[string]string{"RAPIDAPI_KEY": ""},
		IsEnabled:   false,
	}

	// UFC
	cc.mcpServers["ufc"] = &MCPServer{
		Name:        "UFC",
		Description: "MMA e UFC stats",
		Command:     "npx",
		Args:        []string{"-y", "ufc-mcp-server"},
		IsEnabled:   false,
	}

	// Formula 1
	cc.mcpServers["f1"] = &MCPServer{
		Name:        "Formula 1",
		Description: "Ergast F1 API",
		Command:     "npx",
		Args:        []string{"-y", "f1-mcp-server"},
		IsEnabled:   false,
	}

	// SofaScore
	cc.mcpServers["sofascore"] = &MCPServer{
		Name:        "SofaScore",
		Description: "Resultados ao vivo",
		Command:     "npx",
		Args:        []string{"-y", "sofascore-mcp-server"},
		IsEnabled:   false,
	}

	// FlashScore
	cc.mcpServers["flashscore"] = &MCPServer{
		Name:        "FlashScore",
		Description: "Resultados esportivos",
		Command:     "npx",
		Args:        []string{"-y", "flashscore-mcp-server"},
		IsEnabled:   false,
	}

	// FPL (Fantasy Premier League)
	cc.mcpServers["fpl"] = &MCPServer{
		Name:        "Fantasy Premier League",
		Description: "Fantasy futebol FPL",
		Command:     "npx",
		Args:        []string{"-y", "fpl-mcp-server"},
		IsEnabled:   false,
	}

	// Cartola FC (Brasil)
	cc.mcpServers["cartola"] = &MCPServer{
		Name:        "Cartola FC",
		Description: "Fantasy Brasileirão",
		Command:     "npx",
		Args:        []string{"-y", "cartola-mcp-server"},
		IsEnabled:   false,
	}
}
