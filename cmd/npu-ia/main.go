package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/JoseRFJuniorLLMs/NPU-IA/internal/assistant"
	"github.com/JoseRFJuniorLLMs/NPU-IA/internal/audio"
	"github.com/JoseRFJuniorLLMs/NPU-IA/internal/npu"
	"github.com/JoseRFJuniorLLMs/NPU-IA/internal/productivity"
	"github.com/JoseRFJuniorLLMs/NPU-IA/internal/router"
	"github.com/JoseRFJuniorLLMs/NPU-IA/internal/tts"
	"github.com/JoseRFJuniorLLMs/NPU-IA/pkg/config"
)

const banner = `
╔═══════════════════════════════════════════════════════════════╗
║                                                               ║
║    ███╗   ██╗██████╗ ██╗   ██╗      ██╗ █████╗               ║
║    ████╗  ██║██╔══██╗██║   ██║      ██║██╔══██╗              ║
║    ██╔██╗ ██║██████╔╝██║   ██║█████╗██║███████║              ║
║    ██║╚██╗██║██╔═══╝ ██║   ██║╚════╝██║██╔══██║              ║
║    ██║ ╚████║██║     ╚██████╔╝      ██║██║  ██║              ║
║    ╚═╝  ╚═══╝╚═╝      ╚═════╝       ╚═╝╚═╝  ╚═╝              ║
║                                                               ║
║    Assistente IA 100%% Local - Powered by AMD NPU             ║
║    55 TOPS | 32GB RAM | Zero Cloud                           ║
║                                                               ║
╚═══════════════════════════════════════════════════════════════╝
`

// Application estrutura principal da aplicação
type Application struct {
	ctx    context.Context
	cancel context.CancelFunc
	cfg    *config.Config

	// Core
	router  *router.Router
	speaker *tts.Piper
	mic     *audio.Capture
	dm      *npu.DirectML

	// Assistant
	memory   *assistant.Memory
	briefing *assistant.DailyBriefing
	habits   *assistant.HabitTracker
	music    *assistant.FocusMusic
	ebook    *assistant.EbookReader

	// Productivity
	coreFlow    *productivity.CoreFlow
	zettel      *productivity.Zettelkasten
	audioPlayer *productivity.AudioPlayer

	// State
	conversationHistory []string
	lastInteraction     time.Time
}

func main() {
	fmt.Println(banner)
	log.Println("Iniciando NPU-IA...")

	// Carrega configurações
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		log.Printf("Usando configurações padrão: %v", err)
		cfg = config.Default()
	}

	// Cria aplicação
	app, err := NewApplication(cfg)
	if err != nil {
		log.Fatalf("Erro ao inicializar aplicação: %v", err)
	}
	defer app.Close()

	// Executa
	app.Run()
}

// NewApplication cria nova instância da aplicação
func NewApplication(cfg *config.Config) (*Application, error) {
	ctx, cancel := context.WithCancel(context.Background())

	app := &Application{
		ctx:                 ctx,
		cancel:              cancel,
		cfg:                 cfg,
		conversationHistory: make([]string, 0),
		lastInteraction:     time.Now(),
	}

	// Inicializa NPU/DirectML
	log.Println("Detectando NPU...")
	dm, err := npu.NewDirectML()
	if err != nil {
		log.Printf("Aviso: NPU não disponível, usando CPU: %v", err)
	} else {
		dm.PrintDeviceInfo()
		app.dm = dm
	}

	// Inicializa Router (carrega modelos)
	log.Println("Carregando modelos na NPU...")
	r, err := router.New(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("erro ao inicializar router: %w", err)
	}
	app.router = r

	// Inicializa TTS
	log.Println("Inicializando TTS...")
	speaker, err := tts.New(cfg.TTS)
	if err != nil {
		return nil, fmt.Errorf("erro ao inicializar TTS: %w", err)
	}
	app.speaker = speaker

	// Inicializa captura de áudio
	log.Println("Inicializando microfone...")
	mic, err := audio.NewCapture(cfg.Audio)
	if err != nil {
		return nil, fmt.Errorf("erro ao inicializar microfone: %w", err)
	}
	app.mic = mic

	// Cria diretório de dados
	dataDir := filepath.Join(getHomeDir(), ".npu-ia")
	os.MkdirAll(dataDir, 0755)

	// Inicializa Memory
	log.Println("Carregando memória...")
	memory, err := assistant.NewMemory(filepath.Join(dataDir, "memory"), nil)
	if err != nil {
		log.Printf("Aviso: não foi possível carregar memória: %v", err)
	}
	app.memory = memory

	// Inicializa Habit Tracker
	log.Println("Carregando hábitos...")
	habits, err := assistant.NewHabitTracker(filepath.Join(dataDir, "habits"), &ttsWrapper{speaker})
	if err != nil {
		log.Printf("Aviso: não foi possível carregar hábitos: %v", err)
	}
	app.habits = habits

	// Inicializa Daily Briefing
	app.briefing = assistant.NewDailyBriefing(memory, habits)
	app.briefing.SetTTS(&ttsWrapper{speaker})

	// Inicializa Focus Music
	musicDir := filepath.Join(dataDir, "music")
	os.MkdirAll(musicDir, 0755)
	app.music = assistant.NewFocusMusic(musicDir)

	// Inicializa Zettelkasten
	zettelDir := filepath.Join(dataDir, "notes")
	app.zettel = productivity.NewZettelkasten(zettelDir)

	// Inicializa Audio Player
	audioDir := "audio"
	app.audioPlayer = productivity.NewAudioPlayer(audioDir)

	// Inicializa Core Flow (Pomodoro + Wim Hof)
	coreFlowConfig := productivity.DefaultCoreFlowConfig()
	coreFlowConfig.WimHofAudio = "winhof.mp3"
	app.coreFlow = productivity.NewCoreFlow(coreFlowConfig, &ttsWrapper{speaker}, app.zettel)

	// Callback de atualização do Core Flow
	app.coreFlow.SetOnUpdate(func(update productivity.FlowUpdate) {
		log.Printf("Flow: %s | Sessão %d | %s", update.State, update.Session, update.Message)
	})

	// Inicializa eBook Reader
	ebookDir := filepath.Join(dataDir, "books")
	os.MkdirAll(ebookDir, 0755)
	app.ebook = assistant.NewEbookReader(ebookDir, &ttsWrapper{speaker})

	log.Println("✓ Todos os módulos inicializados!")
	return app, nil
}

// Run executa o loop principal
func (app *Application) Run() {
	// Sinal de pronto
	log.Println("✓ NPU-IA pronto! Ouvindo...")

	// Daily Briefing ao iniciar (se for horário apropriado)
	hour := time.Now().Hour()
	if hour >= 6 && hour <= 10 {
		app.runDailyBriefing()
	} else {
		app.speaker.Speak("Olá! Estou pronto para ajudar.")
	}

	// Mostra dashboard de hábitos
	if app.habits != nil {
		fmt.Println(app.habits.RenderDashboard())
	}

	// Loop principal em goroutine
	go app.mainLoop()

	// Aguarda sinal de shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("\nDesligando NPU-IA...")

	// Salva contexto da conversa
	if app.memory != nil && len(app.conversationHistory) > 0 {
		app.memory.SummarizeConversation(app.ctx, app.conversationHistory)
	}

	app.speaker.Speak("Até logo!")
}

// mainLoop loop principal de escuta e processamento
func (app *Application) mainLoop() {
	for {
		select {
		case <-app.ctx.Done():
			return
		default:
			// Captura áudio
			audioData, err := app.mic.Listen()
			if err != nil {
				log.Printf("Erro ao capturar áudio: %v", err)
				continue
			}

			if len(audioData) == 0 {
				continue
			}

			// Processa comando
			response, err := app.processCommand(audioData)
			if err != nil {
				log.Printf("Erro ao processar: %v", err)
				app.speaker.Speak("Desculpe, não entendi.")
				continue
			}

			// Responde
			if response != "" {
				app.speaker.Speak(response)
			}

			app.lastInteraction = time.Now()
		}
	}
}

// processCommand processa comando de áudio
func (app *Application) processCommand(audioData []float32) (string, error) {
	// Processa com o router
	response, err := app.router.Process(app.ctx, audioData)
	if err != nil {
		return "", err
	}

	if response.Text == "" {
		return "", nil
	}

	// Guarda no histórico
	app.conversationHistory = append(app.conversationHistory, "Usuário: "+response.Text)

	// Verifica comandos especiais primeiro
	text := strings.ToLower(response.Text)

	// Comandos de produtividade
	if specialResponse := app.handleSpecialCommands(text); specialResponse != "" {
		app.conversationHistory = append(app.conversationHistory, "Assistente: "+specialResponse)
		return specialResponse, nil
	}

	// Resposta normal do LLM
	app.conversationHistory = append(app.conversationHistory, "Assistente: "+response.Text)
	return response.Text, nil
}

// handleSpecialCommands processa comandos especiais
func (app *Application) handleSpecialCommands(text string) string {
	// === PRODUTIVIDADE ===
	if contains(text, "iniciar foco", "começar foco", "modo foco", "pomodoro") {
		task := extractAfter(text, "foco", "sobre")
		if task == "" {
			task = "Tarefa geral"
		}
		go app.coreFlow.Start(app.ctx, task)
		return "Iniciando modo de produtividade. 45 minutos de foco profundo."
	}

	if contains(text, "parar foco", "encerrar foco", "parar pomodoro") {
		app.coreFlow.Stop()
		return "Sessão de foco encerrada."
	}

	if contains(text, "pausar", "pause") {
		app.coreFlow.Pause()
		return "Pausado."
	}

	if contains(text, "continuar", "resume", "retomar") {
		app.coreFlow.Resume()
		return "Continuando..."
	}

	// === RESPIRAÇÃO WIM HOF ===
	if contains(text, "wim hof", "respiração", "breathing") {
		if app.audioPlayer != nil {
			go app.audioPlayer.PlayMP3("winhof.mp3")
			return "Iniciando sessão de respiração Wim Hof. Siga as instruções do áudio."
		}
		return "Áudio do Wim Hof não encontrado."
	}

	// === HÁBITOS ===
	if contains(text, "meus hábitos", "mostrar hábitos", "listar hábitos") {
		if app.habits != nil {
			fmt.Println(app.habits.RenderDashboard())
			completed, total, _ := app.habits.GetDailyProgress()
			return fmt.Sprintf("Você tem %d de %d hábitos completos hoje.", completed, total)
		}
		return "Tracker de hábitos não disponível."
	}

	if contains(text, "completar hábito", "marcar hábito", "fiz") {
		// Tenta identificar qual hábito
		habits := app.habits.GetAllHabits()
		for _, h := range habits {
			if contains(text, strings.ToLower(h.Name)) {
				if err := app.habits.Complete(h.ID, "", 0); err != nil {
					return fmt.Sprintf("%s já foi completado hoje.", h.Name)
				}
				return fmt.Sprintf("%s marcado como completo! %s", h.Icon, h.Name)
			}
		}
		return "Qual hábito você completou? Diga o nome do hábito."
	}

	if contains(text, "relatório semanal", "resumo semanal") {
		if app.habits != nil {
			fmt.Println(app.habits.GetWeeklyReport())
			return "Relatório semanal exibido no console."
		}
		return "Tracker de hábitos não disponível."
	}

	// === MÚSICA ===
	if contains(text, "tocar música", "música para foco", "lofi", "ambient") {
		if app.music != nil {
			playlist := "lofi"
			if contains(text, "ambient") {
				playlist = "ambient"
			} else if contains(text, "nature", "natureza") {
				playlist = "nature"
			} else if contains(text, "classical", "clássica") {
				playlist = "classical"
			}
			app.music.StartPlaylist(app.ctx, playlist)
			return fmt.Sprintf("Tocando playlist %s para foco.", playlist)
		}
		return "Player de música não disponível."
	}

	if contains(text, "parar música", "silêncio") {
		if app.music != nil {
			app.music.Stop()
			return "Música parada."
		}
		return "Player de música não disponível."
	}

	// === NOTAS (ZETTELKASTEN) ===
	if contains(text, "anotar", "criar nota", "lembrar") {
		content := extractAfter(text, "anotar", "nota", "lembrar")
		if content != "" && app.zettel != nil {
			note, err := app.zettel.QuickCapture(content, "voz")
			if err != nil {
				return "Erro ao salvar nota."
			}
			return fmt.Sprintf("Nota salva: %s", note.ID)
		}
		return "O que você quer anotar?"
	}

	if contains(text, "minhas notas", "listar notas") {
		if app.zettel != nil {
			notes := app.zettel.GetAllNotes()
			if len(notes) == 0 {
				return "Você não tem notas salvas."
			}
			return fmt.Sprintf("Você tem %d notas. As mais recentes são exibidas no console.", len(notes))
		}
		return "Sistema de notas não disponível."
	}

	// === BRIEFING ===
	if contains(text, "briefing", "resumo do dia", "meu dia") {
		app.runDailyBriefing()
		return "" // O briefing já fala
	}

	// === EBOOKS ===
	if contains(text, "ler livro", "abrir livro", "meus livros") {
		if app.ebook != nil {
			books := app.ebook.ListBooks()
			if len(books) == 0 {
				return "Nenhum livro encontrado na biblioteca."
			}
			return fmt.Sprintf("Você tem %d livros. Diga qual deseja ler.", len(books))
		}
		return "Leitor de livros não disponível."
	}

	// === INFORMAÇÕES ===
	if contains(text, "que horas", "hora atual") {
		return fmt.Sprintf("São %s.", time.Now().Format("15:04"))
	}

	if contains(text, "que dia", "data de hoje") {
		return fmt.Sprintf("Hoje é %s.", time.Now().Format("02 de January de 2006"))
	}

	// === STATUS ===
	if contains(text, "status", "como está") {
		status := app.coreFlow.GetStatus()
		if status.State != productivity.FlowStateIdle {
			return fmt.Sprintf("Estado: %s. Sessão %d.", status.State, status.Session)
		}

		memStats := app.memory.GetStats()
		return fmt.Sprintf("Sistema operacional. %d fatos na memória. Pronto para ajudar.",
			memStats["total_facts"])
	}

	return "" // Não é comando especial, usar resposta do LLM
}

// runDailyBriefing executa o briefing diário
func (app *Application) runDailyBriefing() {
	if app.briefing == nil {
		app.speaker.Speak("Briefing não disponível.")
		return
	}

	briefing, err := app.briefing.Generate(app.ctx)
	if err != nil {
		log.Printf("Erro ao gerar briefing: %v", err)
		app.speaker.Speak("Não foi possível gerar o briefing.")
		return
	}

	// Exibe no console
	fmt.Println(app.briefing.ToText(briefing))

	// Fala o briefing
	app.briefing.Speak(briefing)
}

// Close fecha todos os recursos
func (app *Application) Close() error {
	app.cancel()

	if app.router != nil {
		app.router.Close()
	}
	if app.speaker != nil {
		app.speaker.Close()
	}
	if app.mic != nil {
		app.mic.Close()
	}
	if app.dm != nil {
		app.dm.Close()
	}
	if app.music != nil {
		app.music.Stop()
	}
	if app.audioPlayer != nil {
		app.audioPlayer.Stop()
	}

	return nil
}

// === HELPERS ===

// ttsWrapper wrapper para implementar interface TTSInterface
type ttsWrapper struct {
	piper *tts.Piper
}

func (w *ttsWrapper) Speak(text string) error {
	if w.piper != nil {
		w.piper.Speak(text)
	}
	return nil
}

// contains verifica se texto contém alguma das palavras
func contains(text string, words ...string) bool {
	for _, word := range words {
		if strings.Contains(text, word) {
			return true
		}
	}
	return false
}

// extractAfter extrai texto após uma das palavras-chave
func extractAfter(text string, keywords ...string) string {
	for _, kw := range keywords {
		if idx := strings.Index(text, kw); idx != -1 {
			after := strings.TrimSpace(text[idx+len(kw):])
			// Remove preposições comuns
			after = strings.TrimPrefix(after, "sobre ")
			after = strings.TrimPrefix(after, "que ")
			after = strings.TrimPrefix(after, "de ")
			return after
		}
	}
	return ""
}

// getHomeDir retorna diretório home do usuário
func getHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	return home
}
