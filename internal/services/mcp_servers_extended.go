package services

// ==================== MCP SERVERS - EXTENDED ====================
// Parte 1: Smart Home, Health, Travel, Shopping

// RegisterSmartHomeMCP registra servidores de Smart Home
func (cc *ClaudeCode) RegisterSmartHomeMCP() {
	// Home Assistant
	cc.mcpServers["home-assistant"] = &MCPServer{
		Name:        "Home Assistant",
		Description: "Controle completo de automação residencial",
		Command:     "npx",
		Args:        []string{"-y", "homeassistant-mcp-server"},
		Env:         map[string]string{"HASS_URL": "", "HASS_TOKEN": ""},
		IsEnabled:   false,
		Tools: []MCPTool{
			{Name: "turn_on", Description: "Ligar dispositivo"},
			{Name: "turn_off", Description: "Desligar dispositivo"},
			{Name: "set_temperature", Description: "Ajustar temperatura"},
			{Name: "get_state", Description: "Obter estado do dispositivo"},
			{Name: "run_scene", Description: "Executar cena"},
			{Name: "run_automation", Description: "Executar automação"},
		},
	}

	// Philips Hue
	cc.mcpServers["philips-hue"] = &MCPServer{
		Name:        "Philips Hue",
		Description: "Controle de iluminação Philips Hue",
		Command:     "npx",
		Args:        []string{"-y", "hue-mcp-server"},
		Env:         map[string]string{"HUE_BRIDGE_IP": "", "HUE_USERNAME": ""},
		IsEnabled:   false,
	}

	// Google Home / Nest
	cc.mcpServers["google-home"] = &MCPServer{
		Name:        "Google Home/Nest",
		Description: "Controle de dispositivos Google Home e Nest",
		Command:     "npx",
		Args:        []string{"-y", "google-home-mcp-server"},
		Env:         map[string]string{"GOOGLE_HOME_TOKEN": ""},
		IsEnabled:   false,
	}

	// Amazon Alexa
	cc.mcpServers["alexa"] = &MCPServer{
		Name:        "Amazon Alexa",
		Description: "Integração com Alexa e dispositivos Echo",
		Command:     "npx",
		Args:        []string{"-y", "alexa-mcp-server"},
		Env:         map[string]string{"ALEXA_CLIENT_ID": "", "ALEXA_CLIENT_SECRET": ""},
		IsEnabled:   false,
	}

	// SmartThings
	cc.mcpServers["smartthings"] = &MCPServer{
		Name:        "Samsung SmartThings",
		Description: "Hub SmartThings da Samsung",
		Command:     "npx",
		Args:        []string{"-y", "smartthings-mcp-server"},
		Env:         map[string]string{"SMARTTHINGS_TOKEN": ""},
		IsEnabled:   false,
	}

	// Ring
	cc.mcpServers["ring"] = &MCPServer{
		Name:        "Ring",
		Description: "Câmeras e campainhas Ring",
		Command:     "npx",
		Args:        []string{"-y", "ring-mcp-server"},
		Env:         map[string]string{"RING_EMAIL": "", "RING_PASSWORD": ""},
		IsEnabled:   false,
	}

	// Wyze
	cc.mcpServers["wyze"] = &MCPServer{
		Name:        "Wyze",
		Description: "Dispositivos Wyze - câmeras, sensores",
		Command:     "npx",
		Args:        []string{"-y", "wyze-mcp-server"},
		Env:         map[string]string{"WYZE_EMAIL": "", "WYZE_PASSWORD": ""},
		IsEnabled:   false,
	}

	// Tuya / Smart Life
	cc.mcpServers["tuya"] = &MCPServer{
		Name:        "Tuya/Smart Life",
		Description: "Dispositivos Tuya e Smart Life",
		Command:     "npx",
		Args:        []string{"-y", "tuya-mcp-server"},
		Env:         map[string]string{"TUYA_ACCESS_ID": "", "TUYA_ACCESS_SECRET": ""},
		IsEnabled:   false,
	}

	// IFTTT
	cc.mcpServers["ifttt"] = &MCPServer{
		Name:        "IFTTT",
		Description: "Automações If This Then That",
		Command:     "npx",
		Args:        []string{"-y", "ifttt-mcp-server"},
		Env:         map[string]string{"IFTTT_WEBHOOK_KEY": ""},
		IsEnabled:   false,
	}

	// Zigbee2MQTT
	cc.mcpServers["zigbee2mqtt"] = &MCPServer{
		Name:        "Zigbee2MQTT",
		Description: "Controle de dispositivos Zigbee",
		Command:     "npx",
		Args:        []string{"-y", "zigbee2mqtt-mcp-server"},
		Env:         map[string]string{"MQTT_HOST": "", "MQTT_USER": "", "MQTT_PASS": ""},
		IsEnabled:   false,
	}

	// Tesla (veículos e Powerwall)
	cc.mcpServers["tesla"] = &MCPServer{
		Name:        "Tesla",
		Description: "Veículos Tesla e Powerwall",
		Command:     "npx",
		Args:        []string{"-y", "tesla-mcp-server"},
		Env:         map[string]string{"TESLA_ACCESS_TOKEN": ""},
		IsEnabled:   false,
	}

	// ecobee (termostatos)
	cc.mcpServers["ecobee"] = &MCPServer{
		Name:        "ecobee",
		Description: "Termostatos ecobee",
		Command:     "npx",
		Args:        []string{"-y", "ecobee-mcp-server"},
		Env:         map[string]string{"ECOBEE_API_KEY": ""},
		IsEnabled:   false,
	}

	// Roomba / iRobot
	cc.mcpServers["irobot"] = &MCPServer{
		Name:        "iRobot/Roomba",
		Description: "Robôs aspiradores iRobot",
		Command:     "npx",
		Args:        []string{"-y", "irobot-mcp-server"},
		Env:         map[string]string{"IROBOT_EMAIL": "", "IROBOT_PASSWORD": ""},
		IsEnabled:   false,
	}

	// August (fechaduras)
	cc.mcpServers["august"] = &MCPServer{
		Name:        "August",
		Description: "Fechaduras inteligentes August",
		Command:     "npx",
		Args:        []string{"-y", "august-mcp-server"},
		Env:         map[string]string{"AUGUST_EMAIL": "", "AUGUST_PASSWORD": ""},
		IsEnabled:   false,
	}

	// Sonos
	cc.mcpServers["sonos"] = &MCPServer{
		Name:        "Sonos",
		Description: "Sistema de som Sonos",
		Command:     "npx",
		Args:        []string{"-y", "sonos-mcp-server"},
		IsEnabled:   false,
	}
}

// RegisterHealthMCP registra servidores de Saúde/Fitness
func (cc *ClaudeCode) RegisterHealthMCP() {
	// Apple Health (via export)
	cc.mcpServers["apple-health"] = &MCPServer{
		Name:        "Apple Health",
		Description: "Dados de saúde do Apple Health",
		Command:     "npx",
		Args:        []string{"-y", "apple-health-mcp-server"},
		IsEnabled:   false,
	}

	// Fitbit
	cc.mcpServers["fitbit"] = &MCPServer{
		Name:        "Fitbit",
		Description: "Dados de atividade e sono Fitbit",
		Command:     "npx",
		Args:        []string{"-y", "fitbit-mcp-server"},
		Env:         map[string]string{"FITBIT_CLIENT_ID": "", "FITBIT_CLIENT_SECRET": ""},
		IsEnabled:   false,
	}

	// Garmin
	cc.mcpServers["garmin"] = &MCPServer{
		Name:        "Garmin",
		Description: "Dados Garmin Connect",
		Command:     "npx",
		Args:        []string{"-y", "garmin-mcp-server"},
		Env:         map[string]string{"GARMIN_EMAIL": "", "GARMIN_PASSWORD": ""},
		IsEnabled:   false,
	}

	// Strava
	cc.mcpServers["strava"] = &MCPServer{
		Name:        "Strava",
		Description: "Atividades e segmentos Strava",
		Command:     "npx",
		Args:        []string{"-y", "strava-mcp-server"},
		Env:         map[string]string{"STRAVA_CLIENT_ID": "", "STRAVA_CLIENT_SECRET": ""},
		IsEnabled:   false,
	}

	// Whoop
	cc.mcpServers["whoop"] = &MCPServer{
		Name:        "WHOOP",
		Description: "Dados de recuperação e strain WHOOP",
		Command:     "npx",
		Args:        []string{"-y", "whoop-mcp-server"},
		Env:         map[string]string{"WHOOP_EMAIL": "", "WHOOP_PASSWORD": ""},
		IsEnabled:   false,
	}

	// Oura Ring
	cc.mcpServers["oura"] = &MCPServer{
		Name:        "Oura Ring",
		Description: "Dados de sono e atividade Oura",
		Command:     "npx",
		Args:        []string{"-y", "oura-mcp-server"},
		Env:         map[string]string{"OURA_ACCESS_TOKEN": ""},
		IsEnabled:   false,
	}

	// Withings
	cc.mcpServers["withings"] = &MCPServer{
		Name:        "Withings",
		Description: "Balanças e dispositivos Withings",
		Command:     "npx",
		Args:        []string{"-y", "withings-mcp-server"},
		Env:         map[string]string{"WITHINGS_CLIENT_ID": "", "WITHINGS_CLIENT_SECRET": ""},
		IsEnabled:   false,
	}

	// MyFitnessPal
	cc.mcpServers["myfitnesspal"] = &MCPServer{
		Name:        "MyFitnessPal",
		Description: "Nutrição e calorias MyFitnessPal",
		Command:     "npx",
		Args:        []string{"-y", "myfitnesspal-mcp-server"},
		Env:         map[string]string{"MFP_USERNAME": "", "MFP_PASSWORD": ""},
		IsEnabled:   false,
	}

	// Cronometer
	cc.mcpServers["cronometer"] = &MCPServer{
		Name:        "Cronometer",
		Description: "Tracking nutricional detalhado",
		Command:     "npx",
		Args:        []string{"-y", "cronometer-mcp-server"},
		IsEnabled:   false,
	}

	// Headspace
	cc.mcpServers["headspace"] = &MCPServer{
		Name:        "Headspace",
		Description: "Meditação e mindfulness",
		Command:     "npx",
		Args:        []string{"-y", "headspace-mcp-server"},
		IsEnabled:   false,
	}

	// Calm
	cc.mcpServers["calm"] = &MCPServer{
		Name:        "Calm",
		Description: "App de meditação Calm",
		Command:     "npx",
		Args:        []string{"-y", "calm-mcp-server"},
		IsEnabled:   false,
	}

	// Peloton
	cc.mcpServers["peloton"] = &MCPServer{
		Name:        "Peloton",
		Description: "Treinos e métricas Peloton",
		Command:     "npx",
		Args:        []string{"-y", "peloton-mcp-server"},
		Env:         map[string]string{"PELOTON_EMAIL": "", "PELOTON_PASSWORD": ""},
		IsEnabled:   false,
	}

	// Eight Sleep
	cc.mcpServers["eight-sleep"] = &MCPServer{
		Name:        "Eight Sleep",
		Description: "Colchão inteligente Eight Sleep",
		Command:     "npx",
		Args:        []string{"-y", "eight-sleep-mcp-server"},
		IsEnabled:   false,
	}
}

// RegisterTravelMCP registra servidores de Viagem
func (cc *ClaudeCode) RegisterTravelMCP() {
	// Uber (já registrado no principal)

	// Lyft (já registrado)

	// Airbnb
	cc.mcpServers["airbnb"] = &MCPServer{
		Name:        "Airbnb",
		Description: "Reservas e hospedagens Airbnb",
		Command:     "npx",
		Args:        []string{"-y", "airbnb-mcp-server"},
		IsEnabled:   false,
	}

	// Booking.com
	cc.mcpServers["booking"] = &MCPServer{
		Name:        "Booking.com",
		Description: "Reservas de hotéis Booking",
		Command:     "npx",
		Args:        []string{"-y", "booking-mcp-server"},
		IsEnabled:   false,
	}

	// Expedia
	cc.mcpServers["expedia"] = &MCPServer{
		Name:        "Expedia",
		Description: "Voos, hotéis e pacotes Expedia",
		Command:     "npx",
		Args:        []string{"-y", "expedia-mcp-server"},
		IsEnabled:   false,
	}

	// Kayak
	cc.mcpServers["kayak"] = &MCPServer{
		Name:        "Kayak",
		Description: "Comparador de voos Kayak",
		Command:     "npx",
		Args:        []string{"-y", "kayak-mcp-server"},
		IsEnabled:   false,
	}

	// Skyscanner
	cc.mcpServers["skyscanner"] = &MCPServer{
		Name:        "Skyscanner",
		Description: "Busca de voos Skyscanner",
		Command:     "npx",
		Args:        []string{"-y", "skyscanner-mcp-server"},
		Env:         map[string]string{"SKYSCANNER_API_KEY": ""},
		IsEnabled:   false,
	}

	// Google Flights
	cc.mcpServers["google-flights"] = &MCPServer{
		Name:        "Google Flights",
		Description: "Busca de voos Google",
		Command:     "npx",
		Args:        []string{"-y", "google-flights-mcp-server"},
		IsEnabled:   false,
	}

	// TripAdvisor
	cc.mcpServers["tripadvisor"] = &MCPServer{
		Name:        "TripAdvisor",
		Description: "Reviews e recomendações de viagem",
		Command:     "npx",
		Args:        []string{"-y", "tripadvisor-mcp-server"},
		Env:         map[string]string{"TRIPADVISOR_API_KEY": ""},
		IsEnabled:   false,
	}

	// Yelp
	cc.mcpServers["yelp"] = &MCPServer{
		Name:        "Yelp",
		Description: "Busca de restaurantes e negócios",
		Command:     "npx",
		Args:        []string{"-y", "yelp-mcp-server"},
		Env:         map[string]string{"YELP_API_KEY": ""},
		IsEnabled:   false,
	}

	// Foursquare
	cc.mcpServers["foursquare"] = &MCPServer{
		Name:        "Foursquare",
		Description: "Descoberta de lugares",
		Command:     "npx",
		Args:        []string{"-y", "foursquare-mcp-server"},
		Env:         map[string]string{"FOURSQUARE_API_KEY": ""},
		IsEnabled:   false,
	}

	// Citymapper
	cc.mcpServers["citymapper"] = &MCPServer{
		Name:        "Citymapper",
		Description: "Transporte público e rotas urbanas",
		Command:     "npx",
		Args:        []string{"-y", "citymapper-mcp-server"},
		IsEnabled:   false,
	}

	// Rome2Rio
	cc.mcpServers["rome2rio"] = &MCPServer{
		Name:        "Rome2Rio",
		Description: "Rotas multimodais de viagem",
		Command:     "npx",
		Args:        []string{"-y", "rome2rio-mcp-server"},
		Env:         map[string]string{"ROME2RIO_API_KEY": ""},
		IsEnabled:   false,
	}

	// Flightradar24
	cc.mcpServers["flightradar24"] = &MCPServer{
		Name:        "Flightradar24",
		Description: "Rastreamento de voos em tempo real",
		Command:     "npx",
		Args:        []string{"-y", "flightradar24-mcp-server"},
		IsEnabled:   false,
	}

	// FlightAware
	cc.mcpServers["flightaware"] = &MCPServer{
		Name:        "FlightAware",
		Description: "Status de voos e tracking",
		Command:     "npx",
		Args:        []string{"-y", "flightaware-mcp-server"},
		Env:         map[string]string{"FLIGHTAWARE_API_KEY": ""},
		IsEnabled:   false,
	}

	// Waze
	cc.mcpServers["waze"] = &MCPServer{
		Name:        "Waze",
		Description: "Navegação e tráfego em tempo real",
		Command:     "npx",
		Args:        []string{"-y", "waze-mcp-server"},
		IsEnabled:   false,
	}

	// Moovit
	cc.mcpServers["moovit"] = &MCPServer{
		Name:        "Moovit",
		Description: "Transporte público mundial",
		Command:     "npx",
		Args:        []string{"-y", "moovit-mcp-server"},
		Env:         map[string]string{"MOOVIT_API_KEY": ""},
		IsEnabled:   false,
	}

	// 99 (Brasil)
	cc.mcpServers["99"] = &MCPServer{
		Name:        "99",
		Description: "App de corridas 99 (Brasil)",
		Command:     "npx",
		Args:        []string{"-y", "99-mcp-server"},
		IsEnabled:   false,
	}

	// iFood
	cc.mcpServers["ifood"] = &MCPServer{
		Name:        "iFood",
		Description: "Delivery de comida iFood",
		Command:     "npx",
		Args:        []string{"-y", "ifood-mcp-server"},
		IsEnabled:   false,
	}

	// Rappi
	cc.mcpServers["rappi"] = &MCPServer{
		Name:        "Rappi",
		Description: "Super app Rappi",
		Command:     "npx",
		Args:        []string{"-y", "rappi-mcp-server"},
		IsEnabled:   false,
	}
}

// RegisterShoppingMCP registra servidores de Compras
func (cc *ClaudeCode) RegisterShoppingMCP() {
	// Amazon
	cc.mcpServers["amazon"] = &MCPServer{
		Name:        "Amazon",
		Description: "Compras e pedidos Amazon",
		Command:     "npx",
		Args:        []string{"-y", "amazon-mcp-server"},
		Env:         map[string]string{"AMAZON_ACCESS_KEY": "", "AMAZON_SECRET_KEY": ""},
		IsEnabled:   false,
	}

	// eBay
	cc.mcpServers["ebay"] = &MCPServer{
		Name:        "eBay",
		Description: "Compras e leilões eBay",
		Command:     "npx",
		Args:        []string{"-y", "ebay-mcp-server"},
		Env:         map[string]string{"EBAY_APP_ID": ""},
		IsEnabled:   false,
	}

	// Shopify
	cc.mcpServers["shopify"] = &MCPServer{
		Name:        "Shopify",
		Description: "Gerenciamento de loja Shopify",
		Command:     "npx",
		Args:        []string{"-y", "shopify-mcp-server"},
		Env:         map[string]string{"SHOPIFY_STORE": "", "SHOPIFY_ACCESS_TOKEN": ""},
		IsEnabled:   false,
	}

	// WooCommerce
	cc.mcpServers["woocommerce"] = &MCPServer{
		Name:        "WooCommerce",
		Description: "E-commerce WordPress",
		Command:     "npx",
		Args:        []string{"-y", "woocommerce-mcp-server"},
		Env:         map[string]string{"WC_URL": "", "WC_KEY": "", "WC_SECRET": ""},
		IsEnabled:   false,
	}

	// Mercado Livre
	cc.mcpServers["mercadolivre"] = &MCPServer{
		Name:        "Mercado Livre",
		Description: "Marketplace Mercado Livre",
		Command:     "npx",
		Args:        []string{"-y", "mercadolivre-mcp-server"},
		Env:         map[string]string{"MELI_CLIENT_ID": "", "MELI_CLIENT_SECRET": ""},
		IsEnabled:   false,
	}

	// AliExpress
	cc.mcpServers["aliexpress"] = &MCPServer{
		Name:        "AliExpress",
		Description: "Compras internacionais AliExpress",
		Command:     "npx",
		Args:        []string{"-y", "aliexpress-mcp-server"},
		IsEnabled:   false,
	}

	// Walmart
	cc.mcpServers["walmart"] = &MCPServer{
		Name:        "Walmart",
		Description: "Compras Walmart",
		Command:     "npx",
		Args:        []string{"-y", "walmart-mcp-server"},
		Env:         map[string]string{"WALMART_CLIENT_ID": "", "WALMART_CLIENT_SECRET": ""},
		IsEnabled:   false,
	}

	// Target
	cc.mcpServers["target"] = &MCPServer{
		Name:        "Target",
		Description: "Compras Target",
		Command:     "npx",
		Args:        []string{"-y", "target-mcp-server"},
		IsEnabled:   false,
	}

	// Best Buy
	cc.mcpServers["bestbuy"] = &MCPServer{
		Name:        "Best Buy",
		Description: "Eletrônicos Best Buy",
		Command:     "npx",
		Args:        []string{"-y", "bestbuy-mcp-server"},
		Env:         map[string]string{"BESTBUY_API_KEY": ""},
		IsEnabled:   false,
	}

	// Etsy
	cc.mcpServers["etsy"] = &MCPServer{
		Name:        "Etsy",
		Description: "Produtos artesanais Etsy",
		Command:     "npx",
		Args:        []string{"-y", "etsy-mcp-server"},
		Env:         map[string]string{"ETSY_API_KEY": ""},
		IsEnabled:   false,
	}

	// Wish
	cc.mcpServers["wish"] = &MCPServer{
		Name:        "Wish",
		Description: "Marketplace Wish",
		Command:     "npx",
		Args:        []string{"-y", "wish-mcp-server"},
		IsEnabled:   false,
	}

	// Shein
	cc.mcpServers["shein"] = &MCPServer{
		Name:        "Shein",
		Description: "Moda Shein",
		Command:     "npx",
		Args:        []string{"-y", "shein-mcp-server"},
		IsEnabled:   false,
	}

	// Magazine Luiza
	cc.mcpServers["magalu"] = &MCPServer{
		Name:        "Magazine Luiza",
		Description: "Varejo Magalu",
		Command:     "npx",
		Args:        []string{"-y", "magalu-mcp-server"},
		IsEnabled:   false,
	}

	// Americanas
	cc.mcpServers["americanas"] = &MCPServer{
		Name:        "Americanas",
		Description: "Lojas Americanas",
		Command:     "npx",
		Args:        []string{"-y", "americanas-mcp-server"},
		IsEnabled:   false,
	}
}
