package services

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// ==================== EVA-IA INTEGRATION ====================
// Integração com EVA (eva-ia.org) via WebSocket
// Voice-to-Voice usando Google Gemini

// EVAClient cliente para EVA-IA
type EVAClient struct {
	wsURL         string
	apiKey        string
	geminiKey     string
	conn          *websocket.Conn
	isConnected   bool
	reconnectWait time.Duration

	// Audio
	audioCapture  *AudioCapture
	audioPlayer   *AudioPlayer
	sampleRate    int
	channels      int

	// Gemini config
	geminiModel   string
	voiceName     string
	language      string

	// State
	isListening   bool
	isSpeaking    bool
	sessionID     string
	context       []Message
	mu            sync.RWMutex

	// Callbacks
	onTranscript  func(text string)
	onResponse    func(text string)
	onAudioStart  func()
	onAudioEnd    func()
	onError       func(err error)
}

// EVAConfig configuração EVA
type EVAConfig struct {
	WebSocketURL string // ws://eva-ia.org/ws ou wss://
	APIKey       string
	GeminiAPIKey string
	GeminiModel  string // gemini-2.0-flash, gemini-pro, etc
	VoiceName    string // pt-BR voices
	Language     string // pt-BR, en-US
	SampleRate   int    // 16000, 24000, 48000
}

// EVAMessage mensagem WebSocket
type EVAMessage struct {
	Type      string          `json:"type"`
	SessionID string          `json:"session_id,omitempty"`
	Data      json.RawMessage `json:"data,omitempty"`
	Timestamp int64           `json:"timestamp"`
}

// EVATranscript transcrição de áudio
type EVATranscript struct {
	Text       string  `json:"text"`
	IsFinal    bool    `json:"is_final"`
	Confidence float64 `json:"confidence"`
	Language   string  `json:"language"`
}

// EVAResponse resposta do modelo
type EVAResponse struct {
	Text     string `json:"text"`
	AudioURL string `json:"audio_url,omitempty"`
	Audio    []byte `json:"audio,omitempty"` // Base64 decoded
	Emotion  string `json:"emotion,omitempty"`
}

// GeminiRequest request para Gemini
type GeminiRequest struct {
	Contents         []GeminiContent        `json:"contents"`
	GenerationConfig *GeminiGenerationConfig `json:"generationConfig,omitempty"`
	SystemInstruction *GeminiContent         `json:"systemInstruction,omitempty"`
}

// GeminiContent conteúdo Gemini
type GeminiContent struct {
	Role  string       `json:"role"`
	Parts []GeminiPart `json:"parts"`
}

// GeminiPart parte do conteúdo
type GeminiPart struct {
	Text       string          `json:"text,omitempty"`
	InlineData *GeminiInlineData `json:"inlineData,omitempty"`
}

// GeminiInlineData dados inline (áudio/imagem)
type GeminiInlineData struct {
	MimeType string `json:"mimeType"`
	Data     string `json:"data"` // Base64
}

// GeminiGenerationConfig configuração de geração
type GeminiGenerationConfig struct {
	Temperature     float64  `json:"temperature,omitempty"`
	TopP            float64  `json:"topP,omitempty"`
	TopK            int      `json:"topK,omitempty"`
	MaxOutputTokens int      `json:"maxOutputTokens,omitempty"`
	StopSequences   []string `json:"stopSequences,omitempty"`
	ResponseMimeType string  `json:"responseMimeType,omitempty"`
}

// GeminiResponse resposta Gemini
type GeminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
		FinishReason string `json:"finishReason"`
	} `json:"candidates"`
}

// AudioCapture captura de áudio
type AudioCapture struct {
	cmd       *exec.Cmd
	isRunning bool
	mu        sync.Mutex
}

// AudioPlayer player de áudio
type AudioPlayer struct {
	cmd       *exec.Cmd
	isPlaying bool
	mu        sync.Mutex
}

// NewEVAClient cria cliente EVA
func NewEVAClient(config EVAConfig) *EVAClient {
	if config.WebSocketURL == "" {
		config.WebSocketURL = "wss://eva-ia.org/ws"
	}
	if config.GeminiModel == "" {
		config.GeminiModel = "gemini-2.0-flash-exp"
	}
	if config.Language == "" {
		config.Language = "pt-BR"
	}
	if config.SampleRate == 0 {
		config.SampleRate = 16000
	}

	return &EVAClient{
		wsURL:         config.WebSocketURL,
		apiKey:        config.APIKey,
		geminiKey:     config.GeminiAPIKey,
		geminiModel:   config.GeminiModel,
		voiceName:     config.VoiceName,
		language:      config.Language,
		sampleRate:    config.SampleRate,
		channels:      1,
		reconnectWait: 5 * time.Second,
		context:       make([]Message, 0),
		audioCapture:  &AudioCapture{},
		audioPlayer:   &AudioPlayer{},
	}
}

// ==================== CONEXÃO ====================

// Connect conecta ao WebSocket EVA
func (eva *EVAClient) Connect(ctx context.Context) error {
	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	headers := http.Header{}
	if eva.apiKey != "" {
		headers.Set("Authorization", "Bearer "+eva.apiKey)
	}

	conn, _, err := dialer.DialContext(ctx, eva.wsURL, headers)
	if err != nil {
		return fmt.Errorf("erro ao conectar WebSocket: %w", err)
	}

	eva.mu.Lock()
	eva.conn = conn
	eva.isConnected = true
	eva.sessionID = fmt.Sprintf("session_%d", time.Now().UnixNano())
	eva.mu.Unlock()

	// Inicia receiver
	go eva.receiveLoop()

	// Envia handshake
	return eva.sendHandshake()
}

// Disconnect desconecta
func (eva *EVAClient) Disconnect() error {
	eva.mu.Lock()
	defer eva.mu.Unlock()

	if eva.conn != nil {
		eva.isConnected = false
		return eva.conn.Close()
	}
	return nil
}

// sendHandshake envia handshake inicial
func (eva *EVAClient) sendHandshake() error {
	handshake := map[string]interface{}{
		"type":        "handshake",
		"session_id":  eva.sessionID,
		"language":    eva.language,
		"sample_rate": eva.sampleRate,
		"channels":    eva.channels,
		"model":       eva.geminiModel,
	}

	return eva.sendJSON(handshake)
}

// sendJSON envia mensagem JSON
func (eva *EVAClient) sendJSON(v interface{}) error {
	eva.mu.RLock()
	conn := eva.conn
	eva.mu.RUnlock()

	if conn == nil {
		return fmt.Errorf("não conectado")
	}

	return conn.WriteJSON(v)
}

// receiveLoop loop de recebimento
func (eva *EVAClient) receiveLoop() {
	for {
		eva.mu.RLock()
		conn := eva.conn
		connected := eva.isConnected
		eva.mu.RUnlock()

		if !connected || conn == nil {
			return
		}

		_, message, err := conn.ReadMessage()
		if err != nil {
			if eva.onError != nil {
				eva.onError(err)
			}
			eva.handleDisconnect()
			return
		}

		eva.handleMessage(message)
	}
}

// handleMessage processa mensagem recebida
func (eva *EVAClient) handleMessage(data []byte) {
	var msg EVAMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return
	}

	switch msg.Type {
	case "transcript":
		var transcript EVATranscript
		json.Unmarshal(msg.Data, &transcript)
		if eva.onTranscript != nil {
			eva.onTranscript(transcript.Text)
		}
		if transcript.IsFinal {
			go eva.processWithGemini(transcript.Text)
		}

	case "response":
		var response EVAResponse
		json.Unmarshal(msg.Data, &response)
		if eva.onResponse != nil {
			eva.onResponse(response.Text)
		}
		if len(response.Audio) > 0 {
			go eva.playAudio(response.Audio)
		}

	case "audio_start":
		eva.mu.Lock()
		eva.isSpeaking = true
		eva.mu.Unlock()
		if eva.onAudioStart != nil {
			eva.onAudioStart()
		}

	case "audio_end":
		eva.mu.Lock()
		eva.isSpeaking = false
		eva.mu.Unlock()
		if eva.onAudioEnd != nil {
			eva.onAudioEnd()
		}

	case "error":
		var errMsg struct {
			Message string `json:"message"`
		}
		json.Unmarshal(msg.Data, &errMsg)
		if eva.onError != nil {
			eva.onError(fmt.Errorf(errMsg.Message))
		}
	}
}

// handleDisconnect trata desconexão
func (eva *EVAClient) handleDisconnect() {
	eva.mu.Lock()
	eva.isConnected = false
	eva.mu.Unlock()

	// Tenta reconectar
	go func() {
		time.Sleep(eva.reconnectWait)
		eva.Connect(context.Background())
	}()
}

// ==================== VOICE TO VOICE ====================

// StartListening inicia captura de áudio
func (eva *EVAClient) StartListening() error {
	eva.mu.Lock()
	if eva.isListening {
		eva.mu.Unlock()
		return nil
	}
	eva.isListening = true
	eva.mu.Unlock()

	// Inicia captura com ffmpeg
	go eva.captureAudioLoop()

	return nil
}

// StopListening para captura
func (eva *EVAClient) StopListening() {
	eva.mu.Lock()
	eva.isListening = false
	eva.mu.Unlock()

	eva.audioCapture.mu.Lock()
	if eva.audioCapture.cmd != nil && eva.audioCapture.isRunning {
		eva.audioCapture.cmd.Process.Kill()
		eva.audioCapture.isRunning = false
	}
	eva.audioCapture.mu.Unlock()
}

// captureAudioLoop loop de captura de áudio
func (eva *EVAClient) captureAudioLoop() {
	eva.audioCapture.mu.Lock()
	// Windows: usa dshow para captura
	eva.audioCapture.cmd = exec.Command("ffmpeg",
		"-f", "dshow",
		"-i", "audio=Microphone",
		"-ar", fmt.Sprintf("%d", eva.sampleRate),
		"-ac", fmt.Sprintf("%d", eva.channels),
		"-f", "wav",
		"-")
	eva.audioCapture.isRunning = true
	eva.audioCapture.mu.Unlock()

	stdout, err := eva.audioCapture.cmd.StdoutPipe()
	if err != nil {
		return
	}

	if err := eva.audioCapture.cmd.Start(); err != nil {
		return
	}

	buffer := make([]byte, 4096)
	for {
		eva.mu.RLock()
		listening := eva.isListening
		eva.mu.RUnlock()

		if !listening {
			break
		}

		n, err := stdout.Read(buffer)
		if err != nil {
			break
		}

		// Envia áudio para o servidor
		eva.sendAudio(buffer[:n])
	}

	eva.audioCapture.cmd.Wait()
}

// sendAudio envia chunk de áudio
func (eva *EVAClient) sendAudio(audio []byte) error {
	msg := map[string]interface{}{
		"type":       "audio",
		"session_id": eva.sessionID,
		"data":       base64.StdEncoding.EncodeToString(audio),
		"timestamp":  time.Now().UnixMilli(),
	}

	return eva.sendJSON(msg)
}

// playAudio reproduz áudio de resposta
func (eva *EVAClient) playAudio(audio []byte) {
	eva.audioPlayer.mu.Lock()
	defer eva.audioPlayer.mu.Unlock()

	// Salva em arquivo temporário
	tmpFile, err := os.CreateTemp("", "eva_audio_*.wav")
	if err != nil {
		return
	}
	defer os.Remove(tmpFile.Name())

	tmpFile.Write(audio)
	tmpFile.Close()

	// Reproduz com ffplay
	eva.audioPlayer.cmd = exec.Command("ffplay", "-nodisp", "-autoexit", tmpFile.Name())
	eva.audioPlayer.isPlaying = true
	eva.audioPlayer.cmd.Run()
	eva.audioPlayer.isPlaying = false
}

// ==================== GEMINI INTEGRATION ====================

// processWithGemini processa texto com Gemini
func (eva *EVAClient) processWithGemini(text string) {
	ctx := context.Background()

	// Adiciona ao contexto
	eva.context = append(eva.context, Message{
		Role:    "user",
		Content: text,
	})

	// Monta request
	contents := make([]GeminiContent, 0)

	// System instruction
	systemPrompt := `Você é EVA, uma assistente de IA avançada e amigável.
Responda de forma natural, concisa e útil.
Você pode conversar em português brasileiro.
Mantenha suas respostas curtas para uma conversa fluida por voz.`

	// Adiciona histórico
	for _, msg := range eva.context {
		role := "user"
		if msg.Role == "assistant" {
			role = "model"
		}
		contents = append(contents, GeminiContent{
			Role: role,
			Parts: []GeminiPart{{Text: msg.Content}},
		})
	}

	request := GeminiRequest{
		Contents: contents,
		SystemInstruction: &GeminiContent{
			Role:  "user",
			Parts: []GeminiPart{{Text: systemPrompt}},
		},
		GenerationConfig: &GeminiGenerationConfig{
			Temperature:     0.7,
			MaxOutputTokens: 256,
		},
	}

	// Chama Gemini
	response, err := eva.callGemini(ctx, request)
	if err != nil {
		if eva.onError != nil {
			eva.onError(err)
		}
		return
	}

	// Adiciona resposta ao contexto
	eva.context = append(eva.context, Message{
		Role:    "assistant",
		Content: response,
	})

	// Callback
	if eva.onResponse != nil {
		eva.onResponse(response)
	}

	// Converte para áudio e envia
	audio, err := eva.textToSpeech(ctx, response)
	if err == nil && len(audio) > 0 {
		eva.playAudio(audio)
	}

	// Envia resposta pelo WebSocket
	eva.sendJSON(map[string]interface{}{
		"type":       "response",
		"session_id": eva.sessionID,
		"data": map[string]string{
			"text": response,
		},
	})
}

// callGemini chama API Gemini
func (eva *EVAClient) callGemini(ctx context.Context, request GeminiRequest) (string, error) {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
		eva.geminiModel, eva.geminiKey)

	body, err := json.Marshal(request)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("erro Gemini (%d): %s", resp.StatusCode, string(body))
	}

	var geminiResp GeminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		return "", err
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("resposta vazia do Gemini")
	}

	return geminiResp.Candidates[0].Content.Parts[0].Text, nil
}

// textToSpeech converte texto em áudio
func (eva *EVAClient) textToSpeech(ctx context.Context, text string) ([]byte, error) {
	// Usa Google Cloud TTS ou alternativa local
	// Por enquanto, usa Piper local se disponível

	// Verifica se Piper está disponível
	piperPath := "piper"
	if _, err := exec.LookPath(piperPath); err != nil {
		// Tenta usar edge-tts como alternativa
		return eva.edgeTTS(ctx, text)
	}

	// Usa Piper
	cmd := exec.CommandContext(ctx, piperPath,
		"--model", "pt_BR-faber-medium",
		"--output_raw")

	cmd.Stdin = bytes.NewReader([]byte(text))
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return output, nil
}

// edgeTTS usa Microsoft Edge TTS (gratuito)
func (eva *EVAClient) edgeTTS(ctx context.Context, text string) ([]byte, error) {
	// edge-tts é uma ferramenta Python que usa o TTS do Edge
	tmpFile, err := os.CreateTemp("", "tts_*.mp3")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	voice := "pt-BR-FranciscaNeural" // Voz feminina brasileira
	if eva.voiceName != "" {
		voice = eva.voiceName
	}

	cmd := exec.CommandContext(ctx, "edge-tts",
		"--voice", voice,
		"--text", text,
		"--write-media", tmpFile.Name())

	if err := cmd.Run(); err != nil {
		return nil, err
	}

	return os.ReadFile(tmpFile.Name())
}

// ==================== LIVE VOICE (STREAMING) ====================

// StartLiveVoice inicia modo voz ao vivo bidirecional
func (eva *EVAClient) StartLiveVoice() error {
	if err := eva.Connect(context.Background()); err != nil {
		return err
	}

	// Registra callbacks padrão se não definidos
	if eva.onTranscript == nil {
		eva.onTranscript = func(text string) {
			fmt.Printf("🎤 Você: %s\n", text)
		}
	}

	if eva.onResponse == nil {
		eva.onResponse = func(text string) {
			fmt.Printf("🤖 EVA: %s\n", text)
		}
	}

	// Inicia escuta
	return eva.StartListening()
}

// StopLiveVoice para modo voz ao vivo
func (eva *EVAClient) StopLiveVoice() {
	eva.StopListening()
	eva.Disconnect()
}

// ==================== CALLBACKS ====================

// OnTranscript registra callback para transcrições
func (eva *EVAClient) OnTranscript(callback func(text string)) {
	eva.onTranscript = callback
}

// OnResponse registra callback para respostas
func (eva *EVAClient) OnResponse(callback func(text string)) {
	eva.onResponse = callback
}

// OnAudioStart registra callback para início de áudio
func (eva *EVAClient) OnAudioStart(callback func()) {
	eva.onAudioStart = callback
}

// OnAudioEnd registra callback para fim de áudio
func (eva *EVAClient) OnAudioEnd(callback func()) {
	eva.onAudioEnd = callback
}

// OnError registra callback para erros
func (eva *EVAClient) OnError(callback func(err error)) {
	eva.onError = callback
}

// ==================== CONFIGURAÇÕES ====================

// SetGeminiModel define modelo Gemini
func (eva *EVAClient) SetGeminiModel(model string) {
	eva.geminiModel = model
}

// SetVoice define voz para TTS
func (eva *EVAClient) SetVoice(voice string) {
	eva.voiceName = voice
}

// SetLanguage define idioma
func (eva *EVAClient) SetLanguage(lang string) {
	eva.language = lang
}

// ClearContext limpa contexto da conversa
func (eva *EVAClient) ClearContext() {
	eva.mu.Lock()
	eva.context = make([]Message, 0)
	eva.mu.Unlock()
}

// GetContext retorna contexto atual
func (eva *EVAClient) GetContext() []Message {
	eva.mu.RLock()
	defer eva.mu.RUnlock()
	return eva.context
}

// ==================== STATUS ====================

// GetStatus retorna status
func (eva *EVAClient) GetStatus() map[string]interface{} {
	eva.mu.RLock()
	defer eva.mu.RUnlock()

	return map[string]interface{}{
		"connected":    eva.isConnected,
		"listening":    eva.isListening,
		"speaking":     eva.isSpeaking,
		"session_id":   eva.sessionID,
		"model":        eva.geminiModel,
		"language":     eva.language,
		"context_size": len(eva.context),
	}
}

// IsConnected verifica conexão
func (eva *EVAClient) IsConnected() bool {
	eva.mu.RLock()
	defer eva.mu.RUnlock()
	return eva.isConnected
}

// IsListening verifica se está ouvindo
func (eva *EVAClient) IsListening() bool {
	eva.mu.RLock()
	defer eva.mu.RUnlock()
	return eva.isListening
}

// IsSpeaking verifica se está falando
func (eva *EVAClient) IsSpeaking() bool {
	eva.mu.RLock()
	defer eva.mu.RUnlock()
	return eva.isSpeaking
}

// ==================== INTEGRAÇÃO COM NPU-IA ====================

// EVABridge ponte entre NPU-IA e EVA
type EVABridge struct {
	eva       *EVAClient
	npuRouter interface{}
	useLocal  bool
}

// NewEVABridge cria ponte
func NewEVABridge(eva *EVAClient) *EVABridge {
	return &EVABridge{
		eva:      eva,
		useLocal: false,
	}
}

// ProcessVoice processa comando de voz
func (eb *EVABridge) ProcessVoice(ctx context.Context, audio []byte) (string, error) {
	// Envia áudio para EVA
	eb.eva.sendAudio(audio)

	// Aguarda resposta (simplificado - em produção usar channels)
	time.Sleep(2 * time.Second)

	// Retorna última resposta do contexto
	context := eb.eva.GetContext()
	if len(context) > 0 {
		return context[len(context)-1].Content, nil
	}

	return "", fmt.Errorf("sem resposta")
}

// StartAssistant inicia assistente de voz
func (eb *EVABridge) StartAssistant() error {
	return eb.eva.StartLiveVoice()
}

// StopAssistant para assistente
func (eb *EVABridge) StopAssistant() {
	eb.eva.StopLiveVoice()
}
