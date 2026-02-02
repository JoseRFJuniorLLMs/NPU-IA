package router

import (
	"context"
	"log"
	"strings"
	"sync"

	"github.com/JoseRFJuniorLLMs/NPU-IA/internal/actions"
	"github.com/JoseRFJuniorLLMs/NPU-IA/internal/coder"
	"github.com/JoseRFJuniorLLMs/NPU-IA/internal/llm"
	"github.com/JoseRFJuniorLLMs/NPU-IA/internal/stt"
	"github.com/JoseRFJuniorLLMs/NPU-IA/internal/vision"
	"github.com/JoseRFJuniorLLMs/NPU-IA/pkg/config"
)

// Intent representa a intenção detectada
type Intent string

const (
	IntentSimple   Intent = "simple"   // Comandos rápidos
	IntentAction   Intent = "action"   // Executar ações
	IntentContext  Intent = "context"  // Conversa longa
	IntentVision   Intent = "vision"   // Ver tela
	IntentCode     Intent = "code"     // Código
)

// Input representa a entrada do usuário
type Input struct {
	Audio    []float32
	Text     string
	HasImage bool
	Image    []byte
}

// Response representa a resposta do sistema
type Response struct {
	Text    string
	Action  *actions.Action
	Success bool
}

// Router gerencia os modelos e roteia requisições
type Router struct {
	// STT
	whisper *stt.Whisper

	// LLMs
	phi   *llm.Model // Rápido
	llama *llm.Model // Contexto
	qwen  *llm.Model // Ações

	// Especialistas
	vision *vision.Model // Ver tela
	coder  *coder.Model  // Código

	// Executor de ações
	executor *actions.Executor

	// Config
	cfg *config.Config
	mu  sync.RWMutex

	// Pool de modelos carregados
	loaded map[string]bool
}

// New cria um novo Router
func New(ctx context.Context, cfg *config.Config) (*Router, error) {
	r := &Router{
		cfg:    cfg,
		loaded: make(map[string]bool),
	}

	// Whisper sempre carrega (STT principal)
	log.Println("  → Carregando Whisper (STT)...")
	whisper, err := stt.NewWhisper(cfg.STT)
	if err != nil {
		return nil, err
	}
	r.whisper = whisper
	r.loaded["whisper"] = true

	// Carrega modelos conforme configuração
	if cfg.Models.LoadAll {
		if err := r.loadAllModels(ctx); err != nil {
			return nil, err
		}
	} else {
		// Lazy loading - carrega só o Phi (rápido) por padrão
		log.Println("  → Carregando Phi-3.5 (LLM rápido)...")
		phi, err := llm.New(cfg.Models.Phi)
		if err != nil {
			return nil, err
		}
		r.phi = phi
		r.loaded["phi"] = true
	}

	// Executor de ações
	r.executor = actions.NewExecutor()

	log.Println("  ✓ Router inicializado!")
	return r, nil
}

// loadAllModels carrega todos os modelos na memória
func (r *Router) loadAllModels(ctx context.Context) error {
	var wg sync.WaitGroup
	errChan := make(chan error, 5)

	// Phi
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Println("  → Carregando Phi-3.5...")
		phi, err := llm.New(r.cfg.Models.Phi)
		if err != nil {
			errChan <- err
			return
		}
		r.mu.Lock()
		r.phi = phi
		r.loaded["phi"] = true
		r.mu.Unlock()
	}()

	// Llama
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Println("  → Carregando Llama 3.2...")
		llama, err := llm.New(r.cfg.Models.Llama)
		if err != nil {
			errChan <- err
			return
		}
		r.mu.Lock()
		r.llama = llama
		r.loaded["llama"] = true
		r.mu.Unlock()
	}()

	// Qwen
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Println("  → Carregando Qwen 2.5...")
		qwen, err := llm.New(r.cfg.Models.Qwen)
		if err != nil {
			errChan <- err
			return
		}
		r.mu.Lock()
		r.qwen = qwen
		r.loaded["qwen"] = true
		r.mu.Unlock()
	}()

	// Vision
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Println("  → Carregando MiniCPM-V (visão)...")
		v, err := vision.New(r.cfg.Models.Vision)
		if err != nil {
			errChan <- err
			return
		}
		r.mu.Lock()
		r.vision = v
		r.loaded["vision"] = true
		r.mu.Unlock()
	}()

	// Coder
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Println("  → Carregando Qwen-Coder...")
		c, err := coder.New(r.cfg.Models.Coder)
		if err != nil {
			errChan <- err
			return
		}
		r.mu.Lock()
		r.coder = c
		r.loaded["coder"] = true
		r.mu.Unlock()
	}()

	wg.Wait()
	close(errChan)

	// Verifica erros
	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}

// Process processa a entrada e retorna resposta
func (r *Router) Process(ctx context.Context, audioData []float32) (*Response, error) {
	// 1. Transcreve áudio
	text, err := r.whisper.Transcribe(audioData)
	if err != nil {
		return nil, err
	}

	if text == "" {
		return &Response{}, nil
	}

	log.Printf("🎤 Você: %s", text)

	// 2. Detecta intenção
	intent := r.detectIntent(text)
	log.Printf("🎯 Intenção: %s", intent)

	// 3. Roteia para modelo apropriado
	var response *Response

	switch intent {
	case IntentVision:
		response, err = r.handleVision(ctx, text)
	case IntentCode:
		response, err = r.handleCode(ctx, text)
	case IntentAction:
		response, err = r.handleAction(ctx, text)
	case IntentContext:
		response, err = r.handleContext(ctx, text)
	default:
		response, err = r.handleSimple(ctx, text)
	}

	if err != nil {
		return nil, err
	}

	log.Printf("🤖 NPU-IA: %s", response.Text)
	return response, nil
}

// detectIntent detecta a intenção do usuário
func (r *Router) detectIntent(text string) Intent {
	lower := strings.ToLower(text)

	// Visão
	visionKeywords := []string{"tela", "vendo", "olha", "mostra", "screenshot", "imagem", "o que tem"}
	for _, kw := range visionKeywords {
		if strings.Contains(lower, kw) {
			return IntentVision
		}
	}

	// Código
	codeKeywords := []string{"código", "codigo", "função", "funcao", "bug", "erro", "programa", "script", "python", "go ", "javascript"}
	for _, kw := range codeKeywords {
		if strings.Contains(lower, kw) {
			return IntentCode
		}
	}

	// Ações
	actionKeywords := []string{"abre", "abra", "fecha", "feche", "envia", "manda", "lê ", "ler ", "email", "chrome", "navegador", "volume", "brilho"}
	for _, kw := range actionKeywords {
		if strings.Contains(lower, kw) {
			return IntentAction
		}
	}

	// Contexto longo
	contextKeywords := []string{"explica", "conte", "história", "como funciona", "por que", "porque"}
	for _, kw := range contextKeywords {
		if strings.Contains(lower, kw) {
			return IntentContext
		}
	}

	return IntentSimple
}

// handleSimple usa Phi para respostas rápidas
func (r *Router) handleSimple(ctx context.Context, text string) (*Response, error) {
	r.ensureLoaded("phi")
	result, err := r.phi.Generate(ctx, text)
	if err != nil {
		return nil, err
	}
	return &Response{Text: result, Success: true}, nil
}

// handleAction usa Qwen para ações
func (r *Router) handleAction(ctx context.Context, text string) (*Response, error) {
	r.ensureLoaded("qwen")

	// Gera o comando de ação
	result, err := r.qwen.GenerateAction(ctx, text)
	if err != nil {
		return nil, err
	}

	// Executa a ação
	action, err := r.executor.Execute(result)
	if err != nil {
		return &Response{
			Text:    "Desculpe, não consegui executar: " + err.Error(),
			Success: false,
		}, nil
	}

	return &Response{
		Text:    action.Response,
		Action:  action,
		Success: true,
	}, nil
}

// handleContext usa Llama para conversas longas
func (r *Router) handleContext(ctx context.Context, text string) (*Response, error) {
	r.ensureLoaded("llama")
	result, err := r.llama.Generate(ctx, text)
	if err != nil {
		return nil, err
	}
	return &Response{Text: result, Success: true}, nil
}

// handleVision usa MiniCPM-V para ver a tela
func (r *Router) handleVision(ctx context.Context, text string) (*Response, error) {
	r.ensureLoaded("vision")

	// Captura screenshot
	screenshot, err := r.vision.CaptureScreen()
	if err != nil {
		return nil, err
	}

	// Analisa com visão
	result, err := r.vision.Analyze(ctx, screenshot, text)
	if err != nil {
		return nil, err
	}

	return &Response{Text: result, Success: true}, nil
}

// handleCode usa Qwen-Coder para código
func (r *Router) handleCode(ctx context.Context, text string) (*Response, error) {
	r.ensureLoaded("coder")
	result, err := r.coder.Generate(ctx, text)
	if err != nil {
		return nil, err
	}
	return &Response{Text: result, Success: true}, nil
}

// ensureLoaded carrega modelo se necessário (lazy loading)
func (r *Router) ensureLoaded(name string) {
	r.mu.RLock()
	if r.loaded[name] {
		r.mu.RUnlock()
		return
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()

	// Double check
	if r.loaded[name] {
		return
	}

	log.Printf("  → Carregando %s sob demanda...", name)

	var err error
	switch name {
	case "phi":
		r.phi, err = llm.New(r.cfg.Models.Phi)
	case "llama":
		r.llama, err = llm.New(r.cfg.Models.Llama)
	case "qwen":
		r.qwen, err = llm.New(r.cfg.Models.Qwen)
	case "vision":
		r.vision, err = vision.New(r.cfg.Models.Vision)
	case "coder":
		r.coder, err = coder.New(r.cfg.Models.Coder)
	}

	if err != nil {
		log.Printf("Erro ao carregar %s: %v", name, err)
		return
	}

	r.loaded[name] = true
}

// Close libera recursos
func (r *Router) Close() error {
	if r.whisper != nil {
		r.whisper.Close()
	}
	if r.phi != nil {
		r.phi.Close()
	}
	if r.llama != nil {
		r.llama.Close()
	}
	if r.qwen != nil {
		r.qwen.Close()
	}
	if r.vision != nil {
		r.vision.Close()
	}
	if r.coder != nil {
		r.coder.Close()
	}
	return nil
}
