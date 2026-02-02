package services

// ==================== MCP SERVERS - EXTENDED PART 3 ====================
// Finance, Legal, Real Estate, Automotive, Entertainment, Design

// RegisterFinanceMCP registra servidores de Finanças adicionais
func (cc *ClaudeCode) RegisterFinanceMCP() {
	// Já tem: Stripe, Plaid, Coinbase

	// PayPal
	cc.mcpServers["paypal"] = &MCPServer{
		Name:        "PayPal",
		Description: "Pagamentos e transferências PayPal",
		Command:     "npx",
		Args:        []string{"-y", "paypal-mcp-server"},
		Env:         map[string]string{"PAYPAL_CLIENT_ID": "", "PAYPAL_CLIENT_SECRET": ""},
		IsEnabled:   false,
	}

	// Square
	cc.mcpServers["square"] = &MCPServer{
		Name:        "Square",
		Description: "Pagamentos Square",
		Command:     "npx",
		Args:        []string{"-y", "square-mcp-server"},
		Env:         map[string]string{"SQUARE_ACCESS_TOKEN": ""},
		IsEnabled:   false,
	}

	// Venmo
	cc.mcpServers["venmo"] = &MCPServer{
		Name:        "Venmo",
		Description: "Transferências Venmo",
		Command:     "npx",
		Args:        []string{"-y", "venmo-mcp-server"},
		IsEnabled:   false,
	}

	// Wise (TransferWise)
	cc.mcpServers["wise"] = &MCPServer{
		Name:        "Wise",
		Description: "Transferências internacionais",
		Command:     "npx",
		Args:        []string{"-y", "wise-mcp-server"},
		Env:         map[string]string{"WISE_API_TOKEN": ""},
		IsEnabled:   false,
	}

	// Revolut
	cc.mcpServers["revolut"] = &MCPServer{
		Name:        "Revolut",
		Description: "Banco digital Revolut",
		Command:     "npx",
		Args:        []string{"-y", "revolut-mcp-server"},
		Env:         map[string]string{"REVOLUT_ACCESS_TOKEN": ""},
		IsEnabled:   false,
	}

	// Nubank (Brasil)
	cc.mcpServers["nubank"] = &MCPServer{
		Name:        "Nubank",
		Description: "Banco Nubank",
		Command:     "npx",
		Args:        []string{"-y", "nubank-mcp-server"},
		IsEnabled:   false,
	}

	// Inter (Brasil)
	cc.mcpServers["banco-inter"] = &MCPServer{
		Name:        "Banco Inter",
		Description: "Banco Inter Brasil",
		Command:     "npx",
		Args:        []string{"-y", "banco-inter-mcp-server"},
		IsEnabled:   false,
	}

	// Alpha Vantage (stocks)
	cc.mcpServers["alpha-vantage"] = &MCPServer{
		Name:        "Alpha Vantage",
		Description: "Dados de ações e mercado",
		Command:     "npx",
		Args:        []string{"-y", "alpha-vantage-mcp-server"},
		Env:         map[string]string{"ALPHA_VANTAGE_API_KEY": ""},
		IsEnabled:   false,
	}

	// Yahoo Finance
	cc.mcpServers["yahoo-finance"] = &MCPServer{
		Name:        "Yahoo Finance",
		Description: "Cotações e notícias financeiras",
		Command:     "npx",
		Args:        []string{"-y", "yahoo-finance-mcp-server"},
		IsEnabled:   false,
	}

	// Polygon.io
	cc.mcpServers["polygon"] = &MCPServer{
		Name:        "Polygon.io",
		Description: "Market data em tempo real",
		Command:     "npx",
		Args:        []string{"-y", "polygon-mcp-server"},
		Env:         map[string]string{"POLYGON_API_KEY": ""},
		IsEnabled:   false,
	}

	// Alpaca
	cc.mcpServers["alpaca"] = &MCPServer{
		Name:        "Alpaca",
		Description: "Trading API comission-free",
		Command:     "npx",
		Args:        []string{"-y", "alpaca-mcp-server"},
		Env:         map[string]string{"ALPACA_API_KEY": "", "ALPACA_SECRET_KEY": ""},
		IsEnabled:   false,
	}

	// Interactive Brokers
	cc.mcpServers["ibkr"] = &MCPServer{
		Name:        "Interactive Brokers",
		Description: "Trading IBKR",
		Command:     "npx",
		Args:        []string{"-y", "ibkr-mcp-server"},
		IsEnabled:   false,
	}

	// Binance
	cc.mcpServers["binance"] = &MCPServer{
		Name:        "Binance",
		Description: "Exchange crypto Binance",
		Command:     "npx",
		Args:        []string{"-y", "binance-mcp-server"},
		Env:         map[string]string{"BINANCE_API_KEY": "", "BINANCE_SECRET_KEY": ""},
		IsEnabled:   false,
	}

	// Kraken
	cc.mcpServers["kraken"] = &MCPServer{
		Name:        "Kraken",
		Description: "Exchange crypto Kraken",
		Command:     "npx",
		Args:        []string{"-y", "kraken-mcp-server"},
		Env:         map[string]string{"KRAKEN_API_KEY": "", "KRAKEN_PRIVATE_KEY": ""},
		IsEnabled:   false,
	}

	// CoinGecko
	cc.mcpServers["coingecko"] = &MCPServer{
		Name:        "CoinGecko",
		Description: "Dados de criptomoedas",
		Command:     "npx",
		Args:        []string{"-y", "coingecko-mcp-server"},
		IsEnabled:   false,
	}

	// CoinMarketCap
	cc.mcpServers["coinmarketcap"] = &MCPServer{
		Name:        "CoinMarketCap",
		Description: "Market cap e preços crypto",
		Command:     "npx",
		Args:        []string{"-y", "coinmarketcap-mcp-server"},
		Env:         map[string]string{"CMC_API_KEY": ""},
		IsEnabled:   false,
	}

	// QuickBooks
	cc.mcpServers["quickbooks"] = &MCPServer{
		Name:        "QuickBooks",
		Description: "Contabilidade QuickBooks",
		Command:     "npx",
		Args:        []string{"-y", "quickbooks-mcp-server"},
		Env:         map[string]string{"QB_CLIENT_ID": "", "QB_CLIENT_SECRET": ""},
		IsEnabled:   false,
	}

	// Xero
	cc.mcpServers["xero"] = &MCPServer{
		Name:        "Xero",
		Description: "Contabilidade Xero",
		Command:     "npx",
		Args:        []string{"-y", "xero-mcp-server"},
		Env:         map[string]string{"XERO_CLIENT_ID": "", "XERO_CLIENT_SECRET": ""},
		IsEnabled:   false,
	}

	// Wave
	cc.mcpServers["wave"] = &MCPServer{
		Name:        "Wave",
		Description: "Contabilidade gratuita Wave",
		Command:     "npx",
		Args:        []string{"-y", "wave-mcp-server"},
		IsEnabled:   false,
	}

	// FreshBooks
	cc.mcpServers["freshbooks"] = &MCPServer{
		Name:        "FreshBooks",
		Description: "Faturamento FreshBooks",
		Command:     "npx",
		Args:        []string{"-y", "freshbooks-mcp-server"},
		IsEnabled:   false,
	}

	// YNAB
	cc.mcpServers["ynab"] = &MCPServer{
		Name:        "YNAB",
		Description: "Orçamento pessoal YNAB",
		Command:     "npx",
		Args:        []string{"-y", "ynab-mcp-server"},
		Env:         map[string]string{"YNAB_ACCESS_TOKEN": ""},
		IsEnabled:   false,
	}

	// Mint
	cc.mcpServers["mint"] = &MCPServer{
		Name:        "Mint",
		Description: "Finanças pessoais Mint",
		Command:     "npx",
		Args:        []string{"-y", "mint-mcp-server"},
		IsEnabled:   false,
	}

	// Personal Capital
	cc.mcpServers["personal-capital"] = &MCPServer{
		Name:        "Personal Capital",
		Description: "Wealth management",
		Command:     "npx",
		Args:        []string{"-y", "personal-capital-mcp-server"},
		IsEnabled:   false,
	}

	// Guiabolso (Brasil)
	cc.mcpServers["guiabolso"] = &MCPServer{
		Name:        "GuiaBolso",
		Description: "Finanças pessoais Brasil",
		Command:     "npx",
		Args:        []string{"-y", "guiabolso-mcp-server"},
		IsEnabled:   false,
	}

	// Organizze (Brasil)
	cc.mcpServers["organizze"] = &MCPServer{
		Name:        "Organizze",
		Description: "Controle financeiro",
		Command:     "npx",
		Args:        []string{"-y", "organizze-mcp-server"},
		IsEnabled:   false,
	}
}

// RegisterLegalMCP registra servidores Jurídicos
func (cc *ClaudeCode) RegisterLegalMCP() {
	// DocuSign
	cc.mcpServers["docusign"] = &MCPServer{
		Name:        "DocuSign",
		Description: "Assinatura eletrônica",
		Command:     "npx",
		Args:        []string{"-y", "docusign-mcp-server"},
		Env:         map[string]string{"DOCUSIGN_INTEGRATION_KEY": "", "DOCUSIGN_SECRET_KEY": ""},
		IsEnabled:   false,
	}

	// HelloSign
	cc.mcpServers["hellosign"] = &MCPServer{
		Name:        "HelloSign",
		Description: "Assinaturas Dropbox Sign",
		Command:     "npx",
		Args:        []string{"-y", "hellosign-mcp-server"},
		Env:         map[string]string{"HELLOSIGN_API_KEY": ""},
		IsEnabled:   false,
	}

	// PandaDoc
	cc.mcpServers["pandadoc"] = &MCPServer{
		Name:        "PandaDoc",
		Description: "Documentos e propostas",
		Command:     "npx",
		Args:        []string{"-y", "pandadoc-mcp-server"},
		Env:         map[string]string{"PANDADOC_API_KEY": ""},
		IsEnabled:   false,
	}

	// Clio
	cc.mcpServers["clio"] = &MCPServer{
		Name:        "Clio",
		Description: "Software jurídico",
		Command:     "npx",
		Args:        []string{"-y", "clio-mcp-server"},
		Env:         map[string]string{"CLIO_CLIENT_ID": "", "CLIO_CLIENT_SECRET": ""},
		IsEnabled:   false,
	}

	// LegalZoom
	cc.mcpServers["legalzoom"] = &MCPServer{
		Name:        "LegalZoom",
		Description: "Serviços legais online",
		Command:     "npx",
		Args:        []string{"-y", "legalzoom-mcp-server"},
		IsEnabled:   false,
	}

	// Rocket Lawyer
	cc.mcpServers["rocket-lawyer"] = &MCPServer{
		Name:        "Rocket Lawyer",
		Description: "Documentos legais",
		Command:     "npx",
		Args:        []string{"-y", "rocket-lawyer-mcp-server"},
		IsEnabled:   false,
	}

	// Notarize
	cc.mcpServers["notarize"] = &MCPServer{
		Name:        "Notarize",
		Description: "Notarização online",
		Command:     "npx",
		Args:        []string{"-y", "notarize-mcp-server"},
		IsEnabled:   false,
	}

	// Court Listener
	cc.mcpServers["courtlistener"] = &MCPServer{
		Name:        "CourtListener",
		Description: "Casos judiciais públicos",
		Command:     "npx",
		Args:        []string{"-y", "courtlistener-mcp-server"},
		IsEnabled:   false,
	}

	// PACER
	cc.mcpServers["pacer"] = &MCPServer{
		Name:        "PACER",
		Description: "Registros judiciais federais",
		Command:     "npx",
		Args:        []string{"-y", "pacer-mcp-server"},
		IsEnabled:   false,
	}
}

// RegisterRealEstateMCP registra servidores Imobiliários
func (cc *ClaudeCode) RegisterRealEstateMCP() {
	// Zillow
	cc.mcpServers["zillow"] = &MCPServer{
		Name:        "Zillow",
		Description: "Imóveis e avaliações Zillow",
		Command:     "npx",
		Args:        []string{"-y", "zillow-mcp-server"},
		Env:         map[string]string{"ZILLOW_API_KEY": ""},
		IsEnabled:   false,
	}

	// Redfin
	cc.mcpServers["redfin"] = &MCPServer{
		Name:        "Redfin",
		Description: "Imóveis Redfin",
		Command:     "npx",
		Args:        []string{"-y", "redfin-mcp-server"},
		IsEnabled:   false,
	}

	// Realtor.com
	cc.mcpServers["realtor"] = &MCPServer{
		Name:        "Realtor.com",
		Description: "Listagens de imóveis",
		Command:     "npx",
		Args:        []string{"-y", "realtor-mcp-server"},
		Env:         map[string]string{"REALTOR_API_KEY": ""},
		IsEnabled:   false,
	}

	// Trulia
	cc.mcpServers["trulia"] = &MCPServer{
		Name:        "Trulia",
		Description: "Busca de imóveis Trulia",
		Command:     "npx",
		Args:        []string{"-y", "trulia-mcp-server"},
		IsEnabled:   false,
	}

	// Apartments.com
	cc.mcpServers["apartments"] = &MCPServer{
		Name:        "Apartments.com",
		Description: "Aluguel de apartamentos",
		Command:     "npx",
		Args:        []string{"-y", "apartments-mcp-server"},
		IsEnabled:   false,
	}

	// Zap Imóveis (Brasil)
	cc.mcpServers["zap-imoveis"] = &MCPServer{
		Name:        "Zap Imóveis",
		Description: "Portal imobiliário Brasil",
		Command:     "npx",
		Args:        []string{"-y", "zap-imoveis-mcp-server"},
		IsEnabled:   false,
	}

	// QuintoAndar (Brasil)
	cc.mcpServers["quintoandar"] = &MCPServer{
		Name:        "QuintoAndar",
		Description: "Aluguel QuintoAndar",
		Command:     "npx",
		Args:        []string{"-y", "quintoandar-mcp-server"},
		IsEnabled:   false,
	}

	// Imovelweb (Brasil)
	cc.mcpServers["imovelweb"] = &MCPServer{
		Name:        "Imovelweb",
		Description: "Portal Imovelweb",
		Command:     "npx",
		Args:        []string{"-y", "imovelweb-mcp-server"},
		IsEnabled:   false,
	}

	// Idealista (Europa)
	cc.mcpServers["idealista"] = &MCPServer{
		Name:        "Idealista",
		Description: "Imóveis Europa",
		Command:     "npx",
		Args:        []string{"-y", "idealista-mcp-server"},
		IsEnabled:   false,
	}

	// Rightmove (UK)
	cc.mcpServers["rightmove"] = &MCPServer{
		Name:        "Rightmove",
		Description: "Imóveis Reino Unido",
		Command:     "npx",
		Args:        []string{"-y", "rightmove-mcp-server"},
		IsEnabled:   false,
	}
}

// RegisterAutomotiveMCP registra servidores Automotivos
func (cc *ClaudeCode) RegisterAutomotiveMCP() {
	// Tesla já registrado em Smart Home

	// Ford
	cc.mcpServers["ford"] = &MCPServer{
		Name:        "Ford",
		Description: "FordPass Connect",
		Command:     "npx",
		Args:        []string{"-y", "ford-mcp-server"},
		Env:         map[string]string{"FORD_CLIENT_ID": "", "FORD_CLIENT_SECRET": ""},
		IsEnabled:   false,
	}

	// GM OnStar
	cc.mcpServers["onstar"] = &MCPServer{
		Name:        "GM OnStar",
		Description: "Veículos GM/Chevrolet",
		Command:     "npx",
		Args:        []string{"-y", "onstar-mcp-server"},
		IsEnabled:   false,
	}

	// BMW Connected
	cc.mcpServers["bmw"] = &MCPServer{
		Name:        "BMW Connected",
		Description: "BMW ConnectedDrive",
		Command:     "npx",
		Args:        []string{"-y", "bmw-mcp-server"},
		IsEnabled:   false,
	}

	// Mercedes me
	cc.mcpServers["mercedes"] = &MCPServer{
		Name:        "Mercedes me",
		Description: "Mercedes-Benz connect",
		Command:     "npx",
		Args:        []string{"-y", "mercedes-mcp-server"},
		IsEnabled:   false,
	}

	// Audi Connect
	cc.mcpServers["audi"] = &MCPServer{
		Name:        "Audi Connect",
		Description: "Audi connect services",
		Command:     "npx",
		Args:        []string{"-y", "audi-mcp-server"},
		IsEnabled:   false,
	}

	// Porsche Connect
	cc.mcpServers["porsche"] = &MCPServer{
		Name:        "Porsche Connect",
		Description: "Porsche connect",
		Command:     "npx",
		Args:        []string{"-y", "porsche-mcp-server"},
		IsEnabled:   false,
	}

	// Volvo On Call
	cc.mcpServers["volvo"] = &MCPServer{
		Name:        "Volvo On Call",
		Description: "Volvo connect services",
		Command:     "npx",
		Args:        []string{"-y", "volvo-mcp-server"},
		IsEnabled:   false,
	}

	// Toyota Connected
	cc.mcpServers["toyota"] = &MCPServer{
		Name:        "Toyota Connected",
		Description: "Toyota/Lexus connect",
		Command:     "npx",
		Args:        []string{"-y", "toyota-mcp-server"},
		IsEnabled:   false,
	}

	// Honda Link
	cc.mcpServers["honda"] = &MCPServer{
		Name:        "HondaLink",
		Description: "Honda connect",
		Command:     "npx",
		Args:        []string{"-y", "honda-mcp-server"},
		IsEnabled:   false,
	}

	// Hyundai Bluelink
	cc.mcpServers["hyundai"] = &MCPServer{
		Name:        "Hyundai Bluelink",
		Description: "Hyundai/Kia connect",
		Command:     "npx",
		Args:        []string{"-y", "hyundai-mcp-server"},
		IsEnabled:   false,
	}

	// Rivian
	cc.mcpServers["rivian"] = &MCPServer{
		Name:        "Rivian",
		Description: "Veículos elétricos Rivian",
		Command:     "npx",
		Args:        []string{"-y", "rivian-mcp-server"},
		IsEnabled:   false,
	}

	// Lucid
	cc.mcpServers["lucid"] = &MCPServer{
		Name:        "Lucid",
		Description: "Lucid Motors",
		Command:     "npx",
		Args:        []string{"-y", "lucid-mcp-server"},
		IsEnabled:   false,
	}

	// ChargePoint
	cc.mcpServers["chargepoint"] = &MCPServer{
		Name:        "ChargePoint",
		Description: "Estações de carregamento EV",
		Command:     "npx",
		Args:        []string{"-y", "chargepoint-mcp-server"},
		IsEnabled:   false,
	}

	// PlugShare
	cc.mcpServers["plugshare"] = &MCPServer{
		Name:        "PlugShare",
		Description: "Mapa de carregadores EV",
		Command:     "npx",
		Args:        []string{"-y", "plugshare-mcp-server"},
		IsEnabled:   false,
	}

	// Kelley Blue Book
	cc.mcpServers["kbb"] = &MCPServer{
		Name:        "Kelley Blue Book",
		Description: "Avaliação de veículos",
		Command:     "npx",
		Args:        []string{"-y", "kbb-mcp-server"},
		IsEnabled:   false,
	}

	// Edmunds
	cc.mcpServers["edmunds"] = &MCPServer{
		Name:        "Edmunds",
		Description: "Reviews e preços de carros",
		Command:     "npx",
		Args:        []string{"-y", "edmunds-mcp-server"},
		Env:         map[string]string{"EDMUNDS_API_KEY": ""},
		IsEnabled:   false,
	}

	// Carfax
	cc.mcpServers["carfax"] = &MCPServer{
		Name:        "Carfax",
		Description: "Histórico de veículos",
		Command:     "npx",
		Args:        []string{"-y", "carfax-mcp-server"},
		IsEnabled:   false,
	}

	// FIPE (Brasil)
	cc.mcpServers["fipe"] = &MCPServer{
		Name:        "Tabela FIPE",
		Description: "Preços de veículos Brasil",
		Command:     "npx",
		Args:        []string{"-y", "fipe-mcp-server"},
		IsEnabled:   false,
	}

	// Webmotors (Brasil)
	cc.mcpServers["webmotors"] = &MCPServer{
		Name:        "Webmotors",
		Description: "Compra e venda de veículos",
		Command:     "npx",
		Args:        []string{"-y", "webmotors-mcp-server"},
		IsEnabled:   false,
	}

	// OLX Autos (Brasil)
	cc.mcpServers["olx-autos"] = &MCPServer{
		Name:        "OLX Autos",
		Description: "Veículos OLX",
		Command:     "npx",
		Args:        []string{"-y", "olx-autos-mcp-server"},
		IsEnabled:   false,
	}
}

// RegisterEntertainmentMCP registra servidores de Entretenimento
func (cc *ClaudeCode) RegisterEntertainmentMCP() {
	// Netflix
	cc.mcpServers["netflix"] = &MCPServer{
		Name:        "Netflix",
		Description: "Catálogo e watchlist Netflix",
		Command:     "npx",
		Args:        []string{"-y", "netflix-mcp-server"},
		IsEnabled:   false,
	}

	// Disney+
	cc.mcpServers["disney-plus"] = &MCPServer{
		Name:        "Disney+",
		Description: "Streaming Disney+",
		Command:     "npx",
		Args:        []string{"-y", "disney-plus-mcp-server"},
		IsEnabled:   false,
	}

	// HBO Max
	cc.mcpServers["hbo-max"] = &MCPServer{
		Name:        "HBO Max",
		Description: "Streaming HBO Max",
		Command:     "npx",
		Args:        []string{"-y", "hbo-max-mcp-server"},
		IsEnabled:   false,
	}

	// Amazon Prime Video
	cc.mcpServers["prime-video"] = &MCPServer{
		Name:        "Prime Video",
		Description: "Amazon Prime Video",
		Command:     "npx",
		Args:        []string{"-y", "prime-video-mcp-server"},
		IsEnabled:   false,
	}

	// Apple TV+
	cc.mcpServers["apple-tv"] = &MCPServer{
		Name:        "Apple TV+",
		Description: "Streaming Apple TV+",
		Command:     "npx",
		Args:        []string{"-y", "apple-tv-mcp-server"},
		IsEnabled:   false,
	}

	// Hulu
	cc.mcpServers["hulu"] = &MCPServer{
		Name:        "Hulu",
		Description: "Streaming Hulu",
		Command:     "npx",
		Args:        []string{"-y", "hulu-mcp-server"},
		IsEnabled:   false,
	}

	// Paramount+
	cc.mcpServers["paramount-plus"] = &MCPServer{
		Name:        "Paramount+",
		Description: "Streaming Paramount+",
		Command:     "npx",
		Args:        []string{"-y", "paramount-plus-mcp-server"},
		IsEnabled:   false,
	}

	// Peacock
	cc.mcpServers["peacock"] = &MCPServer{
		Name:        "Peacock",
		Description: "Streaming NBC Peacock",
		Command:     "npx",
		Args:        []string{"-y", "peacock-mcp-server"},
		IsEnabled:   false,
	}

	// Plex
	cc.mcpServers["plex"] = &MCPServer{
		Name:        "Plex",
		Description: "Media server Plex",
		Command:     "npx",
		Args:        []string{"-y", "plex-mcp-server"},
		Env:         map[string]string{"PLEX_TOKEN": ""},
		IsEnabled:   false,
	}

	// Jellyfin
	cc.mcpServers["jellyfin"] = &MCPServer{
		Name:        "Jellyfin",
		Description: "Media server open-source",
		Command:     "npx",
		Args:        []string{"-y", "jellyfin-mcp-server"},
		Env:         map[string]string{"JELLYFIN_URL": "", "JELLYFIN_API_KEY": ""},
		IsEnabled:   false,
	}

	// Emby
	cc.mcpServers["emby"] = &MCPServer{
		Name:        "Emby",
		Description: "Media server Emby",
		Command:     "npx",
		Args:        []string{"-y", "emby-mcp-server"},
		IsEnabled:   false,
	}

	// Kodi
	cc.mcpServers["kodi"] = &MCPServer{
		Name:        "Kodi",
		Description: "Media center Kodi",
		Command:     "npx",
		Args:        []string{"-y", "kodi-mcp-server"},
		Env:         map[string]string{"KODI_HOST": "", "KODI_PORT": ""},
		IsEnabled:   false,
	}

	// Trakt
	cc.mcpServers["trakt"] = &MCPServer{
		Name:        "Trakt",
		Description: "Tracking de filmes/séries",
		Command:     "npx",
		Args:        []string{"-y", "trakt-mcp-server"},
		Env:         map[string]string{"TRAKT_CLIENT_ID": "", "TRAKT_CLIENT_SECRET": ""},
		IsEnabled:   false,
	}

	// Letterboxd
	cc.mcpServers["letterboxd"] = &MCPServer{
		Name:        "Letterboxd",
		Description: "Diário de filmes",
		Command:     "npx",
		Args:        []string{"-y", "letterboxd-mcp-server"},
		IsEnabled:   false,
	}

	// IMDb
	cc.mcpServers["imdb"] = &MCPServer{
		Name:        "IMDb",
		Description: "Database de filmes",
		Command:     "npx",
		Args:        []string{"-y", "imdb-mcp-server"},
		IsEnabled:   false,
	}

	// TMDb
	cc.mcpServers["tmdb"] = &MCPServer{
		Name:        "TMDb",
		Description: "The Movie Database",
		Command:     "npx",
		Args:        []string{"-y", "tmdb-mcp-server"},
		Env:         map[string]string{"TMDB_API_KEY": ""},
		IsEnabled:   false,
	}

	// TVDb
	cc.mcpServers["tvdb"] = &MCPServer{
		Name:        "TheTVDB",
		Description: "Database de séries",
		Command:     "npx",
		Args:        []string{"-y", "tvdb-mcp-server"},
		Env:         map[string]string{"TVDB_API_KEY": ""},
		IsEnabled:   false,
	}

	// Goodreads
	cc.mcpServers["goodreads"] = &MCPServer{
		Name:        "Goodreads",
		Description: "Reviews de livros",
		Command:     "npx",
		Args:        []string{"-y", "goodreads-mcp-server"},
		IsEnabled:   false,
	}

	// Audible
	cc.mcpServers["audible"] = &MCPServer{
		Name:        "Audible",
		Description: "Audiobooks Audible",
		Command:     "npx",
		Args:        []string{"-y", "audible-mcp-server"},
		IsEnabled:   false,
	}

	// Kindle
	cc.mcpServers["kindle"] = &MCPServer{
		Name:        "Kindle",
		Description: "E-books Amazon Kindle",
		Command:     "npx",
		Args:        []string{"-y", "kindle-mcp-server"},
		IsEnabled:   false,
	}

	// Apple Music
	cc.mcpServers["apple-music"] = &MCPServer{
		Name:        "Apple Music",
		Description: "Streaming Apple Music",
		Command:     "npx",
		Args:        []string{"-y", "apple-music-mcp-server"},
		IsEnabled:   false,
	}

	// Deezer
	cc.mcpServers["deezer"] = &MCPServer{
		Name:        "Deezer",
		Description: "Streaming Deezer",
		Command:     "npx",
		Args:        []string{"-y", "deezer-mcp-server"},
		Env:         map[string]string{"DEEZER_APP_ID": "", "DEEZER_SECRET": ""},
		IsEnabled:   false,
	}

	// Tidal
	cc.mcpServers["tidal"] = &MCPServer{
		Name:        "Tidal",
		Description: "Streaming HiFi Tidal",
		Command:     "npx",
		Args:        []string{"-y", "tidal-mcp-server"},
		IsEnabled:   false,
	}

	// SoundCloud
	cc.mcpServers["soundcloud"] = &MCPServer{
		Name:        "SoundCloud",
		Description: "Músicas independentes",
		Command:     "npx",
		Args:        []string{"-y", "soundcloud-mcp-server"},
		Env:         map[string]string{"SOUNDCLOUD_CLIENT_ID": ""},
		IsEnabled:   false,
	}

	// Last.fm
	cc.mcpServers["lastfm"] = &MCPServer{
		Name:        "Last.fm",
		Description: "Scrobbling e descoberta",
		Command:     "npx",
		Args:        []string{"-y", "lastfm-mcp-server"},
		Env:         map[string]string{"LASTFM_API_KEY": ""},
		IsEnabled:   false,
	}

	// Bandcamp
	cc.mcpServers["bandcamp"] = &MCPServer{
		Name:        "Bandcamp",
		Description: "Música independente",
		Command:     "npx",
		Args:        []string{"-y", "bandcamp-mcp-server"},
		IsEnabled:   false,
	}

	// Podcasts (Apple)
	cc.mcpServers["apple-podcasts"] = &MCPServer{
		Name:        "Apple Podcasts",
		Description: "Podcasts Apple",
		Command:     "npx",
		Args:        []string{"-y", "apple-podcasts-mcp-server"},
		IsEnabled:   false,
	}

	// Pocket Casts
	cc.mcpServers["pocket-casts"] = &MCPServer{
		Name:        "Pocket Casts",
		Description: "Player de podcasts",
		Command:     "npx",
		Args:        []string{"-y", "pocket-casts-mcp-server"},
		IsEnabled:   false,
	}

	// Overcast
	cc.mcpServers["overcast"] = &MCPServer{
		Name:        "Overcast",
		Description: "Podcasts Overcast",
		Command:     "npx",
		Args:        []string{"-y", "overcast-mcp-server"},
		IsEnabled:   false,
	}

	// Globoplay (Brasil)
	cc.mcpServers["globoplay"] = &MCPServer{
		Name:        "Globoplay",
		Description: "Streaming Globo",
		Command:     "npx",
		Args:        []string{"-y", "globoplay-mcp-server"},
		IsEnabled:   false,
	}
}

// RegisterDesignMCP registra servidores de Design
func (cc *ClaudeCode) RegisterDesignMCP() {
	// Figma
	cc.mcpServers["figma"] = &MCPServer{
		Name:        "Figma",
		Description: "Design colaborativo Figma",
		Command:     "npx",
		Args:        []string{"-y", "figma-mcp-server"},
		Env:         map[string]string{"FIGMA_ACCESS_TOKEN": ""},
		IsEnabled:   false,
	}

	// Sketch
	cc.mcpServers["sketch"] = &MCPServer{
		Name:        "Sketch",
		Description: "Design Sketch",
		Command:     "npx",
		Args:        []string{"-y", "sketch-mcp-server"},
		IsEnabled:   false,
	}

	// Adobe Creative Cloud
	cc.mcpServers["adobe-cc"] = &MCPServer{
		Name:        "Adobe Creative Cloud",
		Description: "Photoshop, Illustrator, etc",
		Command:     "npx",
		Args:        []string{"-y", "adobe-cc-mcp-server"},
		Env:         map[string]string{"ADOBE_CLIENT_ID": "", "ADOBE_CLIENT_SECRET": ""},
		IsEnabled:   false,
	}

	// Canva
	cc.mcpServers["canva"] = &MCPServer{
		Name:        "Canva",
		Description: "Design online Canva",
		Command:     "npx",
		Args:        []string{"-y", "canva-mcp-server"},
		Env:         map[string]string{"CANVA_ACCESS_TOKEN": ""},
		IsEnabled:   false,
	}

	// InVision
	cc.mcpServers["invision"] = &MCPServer{
		Name:        "InVision",
		Description: "Protótipos InVision",
		Command:     "npx",
		Args:        []string{"-y", "invision-mcp-server"},
		IsEnabled:   false,
	}

	// Zeplin
	cc.mcpServers["zeplin"] = &MCPServer{
		Name:        "Zeplin",
		Description: "Handoff design-dev",
		Command:     "npx",
		Args:        []string{"-y", "zeplin-mcp-server"},
		Env:         map[string]string{"ZEPLIN_ACCESS_TOKEN": ""},
		IsEnabled:   false,
	}

	// Framer
	cc.mcpServers["framer"] = &MCPServer{
		Name:        "Framer",
		Description: "Protótipos interativos",
		Command:     "npx",
		Args:        []string{"-y", "framer-mcp-server"},
		IsEnabled:   false,
	}

	// Miro
	cc.mcpServers["miro"] = &MCPServer{
		Name:        "Miro",
		Description: "Whiteboard colaborativo",
		Command:     "npx",
		Args:        []string{"-y", "miro-mcp-server"},
		Env:         map[string]string{"MIRO_ACCESS_TOKEN": ""},
		IsEnabled:   false,
	}

	// Whimsical
	cc.mcpServers["whimsical"] = &MCPServer{
		Name:        "Whimsical",
		Description: "Flowcharts e wireframes",
		Command:     "npx",
		Args:        []string{"-y", "whimsical-mcp-server"},
		IsEnabled:   false,
	}

	// Dribbble
	cc.mcpServers["dribbble"] = &MCPServer{
		Name:        "Dribbble",
		Description: "Comunidade de designers",
		Command:     "npx",
		Args:        []string{"-y", "dribbble-mcp-server"},
		Env:         map[string]string{"DRIBBBLE_ACCESS_TOKEN": ""},
		IsEnabled:   false,
	}

	// Behance
	cc.mcpServers["behance"] = &MCPServer{
		Name:        "Behance",
		Description: "Portfólio criativo Adobe",
		Command:     "npx",
		Args:        []string{"-y", "behance-mcp-server"},
		IsEnabled:   false,
	}

	// Unsplash
	cc.mcpServers["unsplash"] = &MCPServer{
		Name:        "Unsplash",
		Description: "Fotos gratuitas",
		Command:     "npx",
		Args:        []string{"-y", "unsplash-mcp-server"},
		Env:         map[string]string{"UNSPLASH_ACCESS_KEY": ""},
		IsEnabled:   false,
	}

	// Pexels
	cc.mcpServers["pexels"] = &MCPServer{
		Name:        "Pexels",
		Description: "Fotos e vídeos gratuitos",
		Command:     "npx",
		Args:        []string{"-y", "pexels-mcp-server"},
		Env:         map[string]string{"PEXELS_API_KEY": ""},
		IsEnabled:   false,
	}

	// Pixabay
	cc.mcpServers["pixabay"] = &MCPServer{
		Name:        "Pixabay",
		Description: "Imagens gratuitas",
		Command:     "npx",
		Args:        []string{"-y", "pixabay-mcp-server"},
		Env:         map[string]string{"PIXABAY_API_KEY": ""},
		IsEnabled:   false,
	}

	// Shutterstock
	cc.mcpServers["shutterstock"] = &MCPServer{
		Name:        "Shutterstock",
		Description: "Stock images premium",
		Command:     "npx",
		Args:        []string{"-y", "shutterstock-mcp-server"},
		Env:         map[string]string{"SHUTTERSTOCK_API_KEY": ""},
		IsEnabled:   false,
	}

	// Getty Images
	cc.mcpServers["getty"] = &MCPServer{
		Name:        "Getty Images",
		Description: "Stock premium Getty",
		Command:     "npx",
		Args:        []string{"-y", "getty-mcp-server"},
		Env:         map[string]string{"GETTY_API_KEY": ""},
		IsEnabled:   false,
	}

	// Remove.bg
	cc.mcpServers["removebg"] = &MCPServer{
		Name:        "Remove.bg",
		Description: "Remover fundo de imagens",
		Command:     "npx",
		Args:        []string{"-y", "removebg-mcp-server"},
		Env:         map[string]string{"REMOVEBG_API_KEY": ""},
		IsEnabled:   false,
	}

	// Lottie
	cc.mcpServers["lottie"] = &MCPServer{
		Name:        "Lottie",
		Description: "Animações Lottie",
		Command:     "npx",
		Args:        []string{"-y", "lottie-mcp-server"},
		IsEnabled:   false,
	}

	// Noun Project
	cc.mcpServers["noun-project"] = &MCPServer{
		Name:        "Noun Project",
		Description: "Ícones e símbolos",
		Command:     "npx",
		Args:        []string{"-y", "noun-project-mcp-server"},
		Env:         map[string]string{"NOUN_API_KEY": "", "NOUN_SECRET": ""},
		IsEnabled:   false,
	}

	// Flaticon
	cc.mcpServers["flaticon"] = &MCPServer{
		Name:        "Flaticon",
		Description: "Ícones vetoriais",
		Command:     "npx",
		Args:        []string{"-y", "flaticon-mcp-server"},
		IsEnabled:   false,
	}

	// Font Awesome
	cc.mcpServers["fontawesome"] = &MCPServer{
		Name:        "Font Awesome",
		Description: "Biblioteca de ícones",
		Command:     "npx",
		Args:        []string{"-y", "fontawesome-mcp-server"},
		IsEnabled:   false,
	}

	// Google Fonts
	cc.mcpServers["google-fonts"] = &MCPServer{
		Name:        "Google Fonts",
		Description: "Fontes gratuitas Google",
		Command:     "npx",
		Args:        []string{"-y", "google-fonts-mcp-server"},
		Env:         map[string]string{"GOOGLE_FONTS_API_KEY": ""},
		IsEnabled:   false,
	}

	// Coolors
	cc.mcpServers["coolors"] = &MCPServer{
		Name:        "Coolors",
		Description: "Gerador de paletas",
		Command:     "npx",
		Args:        []string{"-y", "coolors-mcp-server"},
		IsEnabled:   false,
	}
}
