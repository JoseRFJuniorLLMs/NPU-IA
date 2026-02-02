# NPU-IA - Assistente de IA Local para AMD Ryzen AI NPU

## Visão Geral

NPU-IA é um assistente de IA pessoal completo que roda 100% localmente no NPU (Neural Processing Unit) do AMD Ryzen AI. O projeto é escrito em Go para máxima performance e integra múltiplos modelos de IA, serviços externos e automações.

**Repositório**: https://github.com/JoseRFJuniorLLMs/NPU-IA

## Hardware Alvo

- **Notebook**: HP OmniBook Ultra
- **Processador**: AMD Ryzen AI 300
- **NPU**: 55 TOPS (Trilhões de Operações por Segundo)
- **RAM**: 32GB
- **Plataforma**: Windows 11

## Arquitetura

### Modelos de IA (ONNX para NPU)

| Modelo | Uso | Carregamento |
|--------|-----|--------------|
| **Whisper** | Speech-to-Text | Sempre carregado |
| **Phi-3.5 Mini** | LLM rápido para conversas | Sempre carregado |
| **Llama 3.2** | LLM mais capaz | Sob demanda |
| **Qwen2.5** | LLM alternativo | Sob demanda |
| **MiniCPM-V** | Visão/Screenshots | Sob demanda |
| **Qwen-Coder** | Análise de código | Sob demanda |
| **Piper** | Text-to-Speech | Sempre carregado |

### Router Inteligente

O sistema possui um router que:
- Detecta intenção do usuário
- Seleciona modelo apropriado automaticamente
- Gerencia memória (descarrega modelos inativos após 5 min)
- Lazy loading para modelos pesados

## Estrutura do Projeto

```
D:\dev\NPU-IA\
├── cmd/
│   └── npu-ia/
│       └── main.go              # Entry point com banner
├── internal/
│   ├── router/
│   │   ├── router.go            # Router inteligente de modelos
│   │   └── memory.go            # Gerenciador de memória
│   ├── stt/
│   │   └── whisper.go           # Whisper STT com DirectML
│   ├── tts/
│   │   └── piper.go             # Piper TTS
│   ├── llm/
│   │   ├── model.go             # Wrapper LLM ONNX
│   │   └── tokenizer.go         # Tokenizer
│   ├── vision/
│   │   └── vision.go            # MiniCPM-V para screenshots
│   ├── coder/
│   │   └── coder.go             # Qwen-Coder para código
│   ├── audio/
│   │   └── capture.go           # Captura de áudio com VAD
│   ├── npu/
│   │   └── directml.go          # DirectML NPU wrapper
│   ├── actions/
│   │   ├── executor.go          # Executor de ações do sistema
│   │   ├── gmail.go             # Gmail OAuth
│   │   ├── calendar.go          # Google Calendar
│   │   └── browser.go           # Chrome DevTools Protocol
│   ├── services/
│   │   ├── google.go            # Google APIs (Gmail, Drive, Docs, etc)
│   │   ├── microsoft.go         # Microsoft Graph (Outlook, Teams, etc)
│   │   ├── github.go            # GitHub API
│   │   ├── social.go            # LinkedIn, X, Discord, Slack, Telegram
│   │   ├── productivity.go      # Notion, Todoist, Spotify
│   │   ├── hub.go               # Central de serviços
│   │   ├── copilot.go           # Windows Copilot integration
│   │   ├── claude_code.go       # Claude Code + MCP Servers
│   │   ├── mcp_servers_extended.go   # Smart Home, Health, Travel, Shopping
│   │   ├── mcp_servers_extended2.go  # Social, Gaming, Education, News, Sports
│   │   ├── mcp_servers_extended3.go  # Finance, Legal, Real Estate, Auto, Entertainment
│   │   ├── mcp_servers_extended4.go  # Business, HR, DevOps
│   │   ├── twilio_whatsapp.go   # WhatsApp Business via Twilio
│   │   └── eva_integration.go   # EVA-IA Voice-to-Voice com Gemini
│   ├── agents/
│   │   ├── email_agent.go       # 10 features de email
│   │   ├── calendar_agent.go    # 5 features de calendário
│   │   ├── task_agent.go        # Automação de tarefas
│   │   └── autonomous_agent.go  # Agente autônomo
│   ├── productivity/
│   │   ├── timer.go             # Pomodoro + Alarmes
│   │   ├── wimhof.go            # Respiração Wim Hof
│   │   ├── zettelkasten.go      # Sistema de notas
│   │   ├── audio_player.go      # Player para winhof.mp3
│   │   └── core_flow.go         # Fluxo: 45min trabalho + Wim Hof
│   └── assistant/
│       ├── memory.go            # Memória de longo prazo
│       ├── briefing.go          # Daily Briefing
│       ├── habits.go            # Habit Tracker
│       ├── focus_music.go       # Música para foco
│       └── ebook_reader.go      # Leitor EPUB + Calibre
├── pkg/
│   └── config/
│       └── config.go            # Configuração
├── configs/
│   └── config.yaml              # YAML config
├── scripts/
│   ├── install_deps.ps1         # Instalador Windows
│   └── download_models.ps1      # Download de modelos
├── audio/
│   └── winhof.mp3               # Áudio guiado Wim Hof
├── go.mod
├── README.md
└── claude.md                    # Este arquivo
```

## Core Flow (Fluxo Principal)

O fluxo de produtividade padrão:

1. **Sessão de Trabalho**: 45 minutos (Pomodoro estendido)
   - Música lo-fi/ambient em volume baixo
   - Bloqueio de distrações
   - Timer visual

2. **Pausa Wim Hof**: 10-15 minutos
   - Reproduz `winhof.mp3` (áudio guiado)
   - 3 rounds de respiração (30 respirações cada)
   - Retenção de ar entre rounds
   - Recuperação guiada

3. **Repetição**: Ciclo contínuo de trabalho/pausa

## Features do Assistente

### 1. Daily Briefing
- Saudação personalizada por horário
- Previsão do tempo
- Agenda do dia
- Emails importantes
- Tarefas pendentes/atrasadas
- Resumo de hábitos
- Cotações crypto
- Quote motivacional

### 2. Memória de Longo Prazo
- Preferências do usuário
- Fatos aprendidos em conversas
- Padrões de comportamento
- Resumos de conversas anteriores
- Contexto personalizado para LLM

### 3. Habit Tracker
- Hábitos diários/semanais
- Streaks e estatísticas
- Lembretes automáticos
- Integração com briefing

### 4. Focus Music
- Playlists: Lo-Fi, Ambient, Nature, Classical, Binaural
- Volume automático (baixa quando fala)
- Integração com Pomodoro
- Scan de músicas locais

### 5. Leitor de eBooks
- Integração com Calibre
- Leitura de EPUBs
- Bookmarks e anotações
- Text-to-Speech para audiobooks
- Busca na biblioteca

## Integrações MCP (300+)

### Oficiais
- Filesystem, GitHub, GitLab, PostgreSQL, SQLite
- Puppeteer, Brave Search, Google Maps, Google Drive

### Produtividade
- Notion, Todoist, Linear, Asana, Trello, ClickUp
- Calendly, Cal.com, Zoom, Google Meet, Teams

### Comunicação
- Slack, Discord, Telegram, WhatsApp Business (Twilio)
- Email (IMAP/SMTP)

### Smart Home
- Home Assistant, Philips Hue, Google Home, Alexa
- SmartThings, Ring, Wyze, Tuya, Tesla, Sonos

### Saúde/Fitness
- Fitbit, Garmin, Strava, WHOOP, Oura Ring
- MyFitnessPal, Peloton, Headspace

### Transporte
- Uber, Lyft, 99, iFood, Rappi, DoorDash

### Finanças
- Stripe, PayPal, Wise, Binance, Coinbase
- QuickBooks, YNAB, Alpha Vantage

### Entretenimento
- Spotify, Apple Music, Netflix, Plex
- YouTube, Twitch, Steam, PlayStation

### Design
- Figma, Adobe CC, Canva, Unsplash

### DevOps
- Docker, Kubernetes, Terraform, Jenkins
- Prometheus, Grafana, Datadog, Sentry

## Integrações Especiais

### Windows Copilot
- Abre/fecha via Win+C
- Envia queries e comandos
- Controle de sistema (tema, volume, WiFi)
- Agenda reuniões, cria lembretes

### Claude Code
- Integração com CLI Claude Code
- Gerenciamento de MCP servers
- Sessões interativas

### WhatsApp via Twilio
- Seu próprio número WhatsApp Business
- Envio de texto, mídia, localização
- Templates aprovados
- Webhook para receber mensagens
- Assistente automático com LLM

### EVA-IA (eva-ia.org)
- WebSocket para comunicação
- Voice-to-Voice bidirecional
- Google Gemini como modelo
- Captura de áudio em tempo real
- Text-to-Speech (Edge TTS / Piper)

## Tecnologias

- **Linguagem**: Go (Golang)
- **NPU**: AMD Ryzen AI via DirectML/ONNX Runtime
- **Modelos**: ONNX format otimizado para NPU
- **Audio**: FFmpeg, FFplay, Piper TTS
- **APIs**: OAuth2, REST, WebSocket
- **Armazenamento**: JSON local, SQLite

## Comandos de Voz Suportados

- "Olá" / "Oi" - Inicia conversa
- "Que horas são" - Hora atual
- "Como está o tempo" - Previsão
- "Leia meus emails" - Resumo de emails
- "O que tenho hoje" - Agenda
- "Iniciar foco" - Começa Pomodoro
- "Pausa" / "Continuar" - Controle de timer
- "Tocar música" - Focus music
- "Abrir [app]" - Abre aplicativo
- "Pesquisar [termo]" - Busca web
- "Ler [livro]" - Abre eBook

## Configuração

### Variáveis de Ambiente Necessárias

```env
# Google
GOOGLE_CLIENT_ID=
GOOGLE_CLIENT_SECRET=

# Microsoft
AZURE_CLIENT_ID=
AZURE_CLIENT_SECRET=

# GitHub
GITHUB_TOKEN=

# Twilio (WhatsApp)
TWILIO_ACCOUNT_SID=
TWILIO_AUTH_TOKEN=
TWILIO_PHONE_NUMBER=+55...

# Google Gemini (EVA)
GEMINI_API_KEY=

# Outros conforme necessário
```

### Arquivo config.yaml

```yaml
npu:
  device: "AMD Ryzen AI"
  memory_limit: 8GB

models:
  whisper: "models/whisper-base-onnx"
  phi: "models/phi-3.5-mini-onnx"

audio:
  sample_rate: 16000
  vad_threshold: 0.5

pomodoro:
  work_duration: 45m
  break_duration: 15m

wimhof:
  audio_file: "audio/winhof.mp3"
  rounds: 3
  breaths_per_round: 30
```

## Execução

```bash
# Build
go build -o npu-ia.exe ./cmd/npu-ia

# Run
./npu-ia.exe

# Com hot-reload (desenvolvimento)
air
```

## Próximos Passos Sugeridos

1. [ ] Implementar wake word ("Hey NPU" ou similar)
2. [ ] Dashboard web para visualização
3. [ ] Integração com calendário nativo Windows
4. [ ] Modo offline completo
5. [ ] Backup automático de memória para cloud
6. [ ] Plugin system para extensões
7. [ ] Mobile companion app

## Notas Importantes

- Todos os modelos rodam localmente no NPU (privacidade total)
- Whisper e Phi-3.5 sempre carregados para resposta rápida
- Modelos pesados carregados sob demanda e descarregados após 5 min
- Áudio winhof.mp3 deve estar em `audio/winhof.mp3`
- Calibre deve estar instalado para integração com eBooks

---

*Última atualização: Fevereiro 2026*
*Modelo: Claude Opus 4.5*
