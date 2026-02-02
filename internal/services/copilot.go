package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

// ==================== WINDOWS COPILOT INTEGRATION ====================

var (
	user32                  = syscall.NewLazyDLL("user32.dll")
	procKeyboardEvent       = user32.NewProc("keybd_event")
	procFindWindow          = user32.NewProc("FindWindowW")
	procSetForegroundWindow = user32.NewProc("SetForegroundWindow")
	procSendMessage         = user32.NewProc("SendMessageW")
	procGetClipboardData    = user32.NewProc("GetClipboardData")
	procOpenClipboard       = user32.NewProc("OpenClipboard")
	procCloseClipboard      = user32.NewProc("CloseClipboard")
	procSetClipboardData    = user32.NewProc("SetClipboardData")
	procEmptyClipboard      = user32.NewProc("EmptyClipboard")
	procGlobalAlloc         = syscall.NewLazyDLL("kernel32.dll").NewProc("GlobalAlloc")
	procGlobalLock          = syscall.NewLazyDLL("kernel32.dll").NewProc("GlobalLock")
	procGlobalUnlock        = syscall.NewLazyDLL("kernel32.dll").NewProc("GlobalUnlock")
)

const (
	VK_LWIN     = 0x5B
	VK_C        = 0x43
	VK_RETURN   = 0x0D
	VK_CONTROL  = 0x11
	VK_V        = 0x56
	KEYEVENTF_KEYUP = 0x0002
	CF_UNICODETEXT  = 13
	GMEM_MOVEABLE   = 0x0002
)

// WindowsCopilot integração com Windows Copilot
type WindowsCopilot struct {
	isEnabled     bool
	isAvailable   bool
	lastQuery     string
	lastResponse  string
	queryHistory  []CopilotQuery
	plugins       map[string]*CopilotPlugin
	callbacks     map[string]CopilotCallback
	mu            sync.RWMutex

	// Configurações
	autoOpen      bool
	timeout       time.Duration
	responseWait  time.Duration
}

// CopilotQuery histórico de queries
type CopilotQuery struct {
	ID        string    `json:"id"`
	Query     string    `json:"query"`
	Response  string    `json:"response"`
	Timestamp time.Time `json:"timestamp"`
	Duration  time.Duration `json:"duration"`
	Success   bool      `json:"success"`
}

// CopilotPlugin plugin do Copilot
type CopilotPlugin struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Commands    []string `json:"commands"`
	IsEnabled   bool     `json:"is_enabled"`
}

// CopilotCallback callback para respostas
type CopilotCallback func(response string) error

// CopilotCapability capacidade do Copilot
type CopilotCapability struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Available   bool   `json:"available"`
}

// NewWindowsCopilot cria integração com Copilot
func NewWindowsCopilot() *WindowsCopilot {
	wc := &WindowsCopilot{
		isEnabled:    true,
		plugins:      make(map[string]*CopilotPlugin),
		callbacks:    make(map[string]CopilotCallback),
		queryHistory: make([]CopilotQuery, 0),
		autoOpen:     true,
		timeout:      30 * time.Second,
		responseWait: 3 * time.Second,
	}

	// Verifica disponibilidade
	wc.checkAvailability()

	// Registra plugins padrão
	wc.registerDefaultPlugins()

	return wc
}

// checkAvailability verifica se Copilot está disponível
func (wc *WindowsCopilot) checkAvailability() {
	// Verifica versão do Windows (Copilot requer Windows 11 23H2+)
	cmd := exec.Command("powershell", "-c", "[System.Environment]::OSVersion.Version")
	output, err := cmd.Output()
	if err != nil {
		wc.isAvailable = false
		return
	}

	version := strings.TrimSpace(string(output))
	// Windows 11 tem build >= 22000
	wc.isAvailable = strings.Contains(version, "10.0.22") || strings.Contains(version, "10.0.23")

	// Verifica se Copilot está instalado
	copilotPath := filepath.Join(os.Getenv("LOCALAPPDATA"), "Microsoft", "WindowsApps", "Microsoft.Copilot_8wekyb3d8bbwe")
	if _, err := os.Stat(copilotPath); err == nil {
		wc.isAvailable = true
	}
}

// registerDefaultPlugins registra plugins padrão
func (wc *WindowsCopilot) registerDefaultPlugins() {
	wc.plugins["system"] = &CopilotPlugin{
		ID:          "system",
		Name:        "System Control",
		Description: "Controle de configurações do Windows",
		Commands:    []string{"open settings", "change theme", "adjust volume", "toggle wifi"},
		IsEnabled:   true,
	}

	wc.plugins["search"] = &CopilotPlugin{
		ID:          "search",
		Name:        "Web Search",
		Description: "Busca na web via Bing",
		Commands:    []string{"search", "find", "look up"},
		IsEnabled:   true,
	}

	wc.plugins["apps"] = &CopilotPlugin{
		ID:          "apps",
		Name:        "App Control",
		Description: "Abrir e controlar aplicativos",
		Commands:    []string{"open", "launch", "start", "close"},
		IsEnabled:   true,
	}

	wc.plugins["files"] = &CopilotPlugin{
		ID:          "files",
		Name:        "File Operations",
		Description: "Operações com arquivos",
		Commands:    []string{"find file", "open folder", "recent files"},
		IsEnabled:   true,
	}

	wc.plugins["productivity"] = &CopilotPlugin{
		ID:          "productivity",
		Name:        "Productivity",
		Description: "Integração com Office e Teams",
		Commands:    []string{"create document", "schedule meeting", "send email"},
		IsEnabled:   true,
	}
}

// ==================== CONTROLES PRINCIPAIS ====================

// Open abre o Windows Copilot
func (wc *WindowsCopilot) Open() error {
	if !wc.isAvailable {
		return fmt.Errorf("Windows Copilot não está disponível")
	}

	// Win + C abre o Copilot
	keyPress(VK_LWIN, 0)
	keyPress(VK_C, 0)
	keyRelease(VK_C)
	keyRelease(VK_LWIN)

	time.Sleep(500 * time.Millisecond)
	return nil
}

// Close fecha o Windows Copilot
func (wc *WindowsCopilot) Close() error {
	// Win + C também fecha se já estiver aberto
	return wc.Open()
}

// Toggle alterna Copilot
func (wc *WindowsCopilot) Toggle() error {
	return wc.Open()
}

// SendQuery envia query para o Copilot
func (wc *WindowsCopilot) SendQuery(ctx context.Context, query string) (string, error) {
	wc.mu.Lock()
	defer wc.mu.Unlock()

	if !wc.isAvailable {
		return "", fmt.Errorf("Windows Copilot não está disponível")
	}

	startTime := time.Now()

	// Abre Copilot se necessário
	if wc.autoOpen {
		wc.Open()
		time.Sleep(1 * time.Second)
	}

	// Copia query para clipboard
	if err := setClipboard(query); err != nil {
		return "", fmt.Errorf("erro ao copiar query: %w", err)
	}

	// Cola no Copilot (Ctrl+V)
	keyPress(VK_CONTROL, 0)
	keyPress(VK_V, 0)
	keyRelease(VK_V)
	keyRelease(VK_CONTROL)

	time.Sleep(200 * time.Millisecond)

	// Enter para enviar
	keyPress(VK_RETURN, 0)
	keyRelease(VK_RETURN)

	// Aguarda resposta
	time.Sleep(wc.responseWait)

	// Registra query
	queryRecord := CopilotQuery{
		ID:        fmt.Sprintf("query_%d", time.Now().UnixNano()),
		Query:     query,
		Timestamp: time.Now(),
		Duration:  time.Since(startTime),
		Success:   true,
	}
	wc.queryHistory = append(wc.queryHistory, queryRecord)
	wc.lastQuery = query

	return "Query enviada ao Copilot", nil
}

// SendCommand envia comando específico
func (wc *WindowsCopilot) SendCommand(ctx context.Context, command string, args map[string]string) error {
	// Formata comando
	query := command
	for key, value := range args {
		query = strings.ReplaceAll(query, "{"+key+"}", value)
	}

	_, err := wc.SendQuery(ctx, query)
	return err
}

// ==================== COMANDOS ESPECÍFICOS ====================

// OpenApp abre aplicativo via Copilot
func (wc *WindowsCopilot) OpenApp(appName string) error {
	ctx := context.Background()
	_, err := wc.SendQuery(ctx, fmt.Sprintf("Open %s", appName))
	return err
}

// SearchWeb busca na web
func (wc *WindowsCopilot) SearchWeb(query string) error {
	ctx := context.Background()
	_, err := wc.SendQuery(ctx, fmt.Sprintf("Search for %s", query))
	return err
}

// OpenSettings abre configurações
func (wc *WindowsCopilot) OpenSettings(setting string) error {
	ctx := context.Background()
	_, err := wc.SendQuery(ctx, fmt.Sprintf("Open %s settings", setting))
	return err
}

// ChangeTheme muda tema do Windows
func (wc *WindowsCopilot) ChangeTheme(theme string) error {
	ctx := context.Background()
	_, err := wc.SendQuery(ctx, fmt.Sprintf("Change to %s theme", theme))
	return err
}

// SetVolume ajusta volume
func (wc *WindowsCopilot) SetVolume(level int) error {
	ctx := context.Background()
	_, err := wc.SendQuery(ctx, fmt.Sprintf("Set volume to %d%%", level))
	return err
}

// ToggleWifi liga/desliga WiFi
func (wc *WindowsCopilot) ToggleWifi(enabled bool) error {
	ctx := context.Background()
	action := "off"
	if enabled {
		action = "on"
	}
	_, err := wc.SendQuery(ctx, fmt.Sprintf("Turn WiFi %s", action))
	return err
}

// ToggleBluetooth liga/desliga Bluetooth
func (wc *WindowsCopilot) ToggleBluetooth(enabled bool) error {
	ctx := context.Background()
	action := "off"
	if enabled {
		action = "on"
	}
	_, err := wc.SendQuery(ctx, fmt.Sprintf("Turn Bluetooth %s", action))
	return err
}

// ToggleDarkMode alterna modo escuro
func (wc *WindowsCopilot) ToggleDarkMode(enabled bool) error {
	ctx := context.Background()
	mode := "light"
	if enabled {
		mode = "dark"
	}
	_, err := wc.SendQuery(ctx, fmt.Sprintf("Switch to %s mode", mode))
	return err
}

// CreateReminder cria lembrete
func (wc *WindowsCopilot) CreateReminder(title string, when time.Time) error {
	ctx := context.Background()
	_, err := wc.SendQuery(ctx, fmt.Sprintf("Remind me to %s at %s", title, when.Format("3:04 PM")))
	return err
}

// ScheduleMeeting agenda reunião
func (wc *WindowsCopilot) ScheduleMeeting(title string, when time.Time, attendees []string) error {
	ctx := context.Background()
	query := fmt.Sprintf("Schedule a meeting called %s for %s", title, when.Format("Monday at 3:04 PM"))
	if len(attendees) > 0 {
		query += fmt.Sprintf(" with %s", strings.Join(attendees, ", "))
	}
	_, err := wc.SendQuery(ctx, query)
	return err
}

// SendEmail compõe email
func (wc *WindowsCopilot) SendEmail(to, subject, body string) error {
	ctx := context.Background()
	_, err := wc.SendQuery(ctx, fmt.Sprintf("Write an email to %s about %s saying %s", to, subject, body))
	return err
}

// SummarizeContent resume conteúdo
func (wc *WindowsCopilot) SummarizeContent(content string) error {
	ctx := context.Background()
	_, err := wc.SendQuery(ctx, fmt.Sprintf("Summarize this: %s", content))
	return err
}

// TranslateText traduz texto
func (wc *WindowsCopilot) TranslateText(text, targetLang string) error {
	ctx := context.Background()
	_, err := wc.SendQuery(ctx, fmt.Sprintf("Translate to %s: %s", targetLang, text))
	return err
}

// ExplainCode explica código
func (wc *WindowsCopilot) ExplainCode(code string) error {
	ctx := context.Background()
	_, err := wc.SendQuery(ctx, fmt.Sprintf("Explain this code: %s", code))
	return err
}

// GenerateImage gera imagem
func (wc *WindowsCopilot) GenerateImage(prompt string) error {
	ctx := context.Background()
	_, err := wc.SendQuery(ctx, fmt.Sprintf("Create an image of %s", prompt))
	return err
}

// ==================== INTEGRAÇÃO COM NPU-IA ====================

// CopilotBridge ponte entre NPU-IA e Copilot
type CopilotBridge struct {
	copilot    *WindowsCopilot
	npuRouter  interface{} // Router do NPU-IA
	useLocal   bool        // Preferir modelo local
	hybridMode bool        // Modo híbrido
}

// NewCopilotBridge cria ponte
func NewCopilotBridge(copilot *WindowsCopilot) *CopilotBridge {
	return &CopilotBridge{
		copilot:    copilot,
		useLocal:   true,
		hybridMode: true,
	}
}

// Route roteia query para Copilot ou modelo local
func (cb *CopilotBridge) Route(ctx context.Context, query string) (string, error) {
	// Detecta se deve usar Copilot
	useCopilot := cb.shouldUseCopilot(query)

	if useCopilot && cb.copilot.isAvailable {
		return cb.copilot.SendQuery(ctx, query)
	}

	// Usa modelo local (implementação no router)
	return "", fmt.Errorf("modelo local não configurado")
}

// shouldUseCopilot determina se deve usar Copilot
func (cb *CopilotBridge) shouldUseCopilot(query string) bool {
	queryLower := strings.ToLower(query)

	// Comandos que Copilot executa melhor
	copilotKeywords := []string{
		"open", "abrir",
		"search", "buscar", "pesquisar",
		"settings", "configurações",
		"create", "criar",
		"schedule", "agendar",
		"remind", "lembrar",
		"email", "e-mail",
		"meeting", "reunião",
		"theme", "tema",
		"volume",
		"wifi", "bluetooth",
		"translate", "traduzir",
		"image", "imagem", "generate",
	}

	for _, keyword := range copilotKeywords {
		if strings.Contains(queryLower, keyword) {
			return true
		}
	}

	return false
}

// ==================== STATUS E HISTÓRICO ====================

// GetStatus retorna status do Copilot
func (wc *WindowsCopilot) GetStatus() map[string]interface{} {
	wc.mu.RLock()
	defer wc.mu.RUnlock()

	return map[string]interface{}{
		"is_available":  wc.isAvailable,
		"is_enabled":    wc.isEnabled,
		"total_queries": len(wc.queryHistory),
		"last_query":    wc.lastQuery,
		"plugins":       len(wc.plugins),
		"auto_open":     wc.autoOpen,
	}
}

// GetCapabilities retorna capacidades
func (wc *WindowsCopilot) GetCapabilities() []CopilotCapability {
	return []CopilotCapability{
		{Name: "Web Search", Description: "Busca na web via Bing", Available: true},
		{Name: "App Launch", Description: "Abrir aplicativos", Available: true},
		{Name: "System Control", Description: "Controle de configurações", Available: true},
		{Name: "Productivity", Description: "Office e Teams", Available: true},
		{Name: "Image Generation", Description: "Geração de imagens DALL-E", Available: true},
		{Name: "Code Explanation", Description: "Explicação de código", Available: true},
		{Name: "Translation", Description: "Tradução de textos", Available: true},
		{Name: "Summarization", Description: "Resumo de conteúdo", Available: true},
		{Name: "Email Composition", Description: "Composição de emails", Available: true},
		{Name: "Calendar", Description: "Agendamento de eventos", Available: true},
	}
}

// GetQueryHistory retorna histórico
func (wc *WindowsCopilot) GetQueryHistory(limit int) []CopilotQuery {
	wc.mu.RLock()
	defer wc.mu.RUnlock()

	if limit <= 0 || limit > len(wc.queryHistory) {
		return wc.queryHistory
	}

	start := len(wc.queryHistory) - limit
	return wc.queryHistory[start:]
}

// GetPlugins retorna plugins
func (wc *WindowsCopilot) GetPlugins() []*CopilotPlugin {
	wc.mu.RLock()
	defer wc.mu.RUnlock()

	plugins := make([]*CopilotPlugin, 0, len(wc.plugins))
	for _, p := range wc.plugins {
		plugins = append(plugins, p)
	}
	return plugins
}

// EnablePlugin habilita plugin
func (wc *WindowsCopilot) EnablePlugin(pluginID string) error {
	wc.mu.Lock()
	defer wc.mu.Unlock()

	if plugin, ok := wc.plugins[pluginID]; ok {
		plugin.IsEnabled = true
		return nil
	}
	return fmt.Errorf("plugin não encontrado: %s", pluginID)
}

// DisablePlugin desabilita plugin
func (wc *WindowsCopilot) DisablePlugin(pluginID string) error {
	wc.mu.Lock()
	defer wc.mu.Unlock()

	if plugin, ok := wc.plugins[pluginID]; ok {
		plugin.IsEnabled = false
		return nil
	}
	return fmt.Errorf("plugin não encontrado: %s", pluginID)
}

// RegisterCallback registra callback
func (wc *WindowsCopilot) RegisterCallback(event string, callback CopilotCallback) {
	wc.mu.Lock()
	defer wc.mu.Unlock()
	wc.callbacks[event] = callback
}

// SetAutoOpen configura abertura automática
func (wc *WindowsCopilot) SetAutoOpen(enabled bool) {
	wc.mu.Lock()
	defer wc.mu.Unlock()
	wc.autoOpen = enabled
}

// SetTimeout configura timeout
func (wc *WindowsCopilot) SetTimeout(timeout time.Duration) {
	wc.mu.Lock()
	defer wc.mu.Unlock()
	wc.timeout = timeout
}

// IsAvailable verifica disponibilidade
func (wc *WindowsCopilot) IsAvailable() bool {
	return wc.isAvailable
}

// ==================== HELPERS ====================

// keyPress simula pressionamento de tecla
func keyPress(vk byte, scan byte) {
	procKeyboardEvent.Call(uintptr(vk), uintptr(scan), 0, 0)
}

// keyRelease simula liberação de tecla
func keyRelease(vk byte) {
	procKeyboardEvent.Call(uintptr(vk), 0, KEYEVENTF_KEYUP, 0)
}

// setClipboard define conteúdo do clipboard
func setClipboard(text string) error {
	r, _, err := procOpenClipboard.Call(0)
	if r == 0 {
		return fmt.Errorf("erro ao abrir clipboard: %v", err)
	}
	defer procCloseClipboard.Call()

	procEmptyClipboard.Call()

	// Converte para UTF-16
	utf16 := syscall.StringToUTF16(text)
	size := len(utf16) * 2

	// Aloca memória global
	hMem, _, _ := procGlobalAlloc.Call(GMEM_MOVEABLE, uintptr(size))
	if hMem == 0 {
		return fmt.Errorf("erro ao alocar memória")
	}

	// Lock e copia
	ptr, _, _ := procGlobalLock.Call(hMem)
	if ptr == 0 {
		return fmt.Errorf("erro ao fazer lock da memória")
	}

	for i, v := range utf16 {
		*(*uint16)(unsafe.Pointer(ptr + uintptr(i*2))) = v
	}

	procGlobalUnlock.Call(hMem)

	// Define no clipboard
	procSetClipboardData.Call(CF_UNICODETEXT, hMem)

	return nil
}

// ==================== EXPORTS JSON ====================

// ExportHistory exporta histórico para JSON
func (wc *WindowsCopilot) ExportHistory(filepath string) error {
	wc.mu.RLock()
	defer wc.mu.RUnlock()

	data, err := json.MarshalIndent(wc.queryHistory, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath, data, 0644)
}

// ExportPlugins exporta plugins para JSON
func (wc *WindowsCopilot) ExportPlugins(filepath string) error {
	wc.mu.RLock()
	defer wc.mu.RUnlock()

	data, err := json.MarshalIndent(wc.plugins, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath, data, 0644)
}
