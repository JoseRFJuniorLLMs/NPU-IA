package actions

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// Action representa uma ação executada
type Action struct {
	Type     string                 `json:"action"`
	Params   map[string]interface{} `json:"params"`
	Response string                 `json:"response,omitempty"`
	Success  bool                   `json:"success"`
}

// Executor executa ações no sistema
type Executor struct {
	handlers map[string]ActionHandler
}

// ActionHandler função que executa uma ação
type ActionHandler func(params map[string]interface{}) (string, error)

// NewExecutor cria um novo executor
func NewExecutor() *Executor {
	e := &Executor{
		handlers: make(map[string]ActionHandler),
	}

	// Registra handlers padrão
	e.RegisterHandler("open_app", e.openApp)
	e.RegisterHandler("open_url", e.openURL)
	e.RegisterHandler("type_text", e.typeText)
	e.RegisterHandler("read_email", e.readEmail)
	e.RegisterHandler("send_email", e.sendEmail)
	e.RegisterHandler("volume", e.setVolume)
	e.RegisterHandler("screenshot", e.takeScreenshot)
	e.RegisterHandler("search", e.search)
	e.RegisterHandler("run_command", e.runCommand)

	return e
}

// RegisterHandler registra um handler de ação
func (e *Executor) RegisterHandler(actionType string, handler ActionHandler) {
	e.handlers[actionType] = handler
}

// Execute executa uma ação a partir do JSON
func (e *Executor) Execute(actionJSON string) (*Action, error) {
	var action Action
	if err := json.Unmarshal([]byte(actionJSON), &action); err != nil {
		return nil, fmt.Errorf("JSON inválido: %w", err)
	}

	handler, ok := e.handlers[action.Type]
	if !ok {
		return nil, fmt.Errorf("ação desconhecida: %s", action.Type)
	}

	response, err := handler(action.Params)
	if err != nil {
		action.Success = false
		action.Response = fmt.Sprintf("Erro: %v", err)
		return &action, err
	}

	action.Success = true
	action.Response = response
	return &action, nil
}

// === Handlers de Ações ===

// openApp abre um aplicativo
func (e *Executor) openApp(params map[string]interface{}) (string, error) {
	app, ok := params["app"].(string)
	if !ok {
		return "", fmt.Errorf("parâmetro 'app' não fornecido")
	}

	// Mapeia nomes comuns para executáveis
	appMap := map[string]string{
		"chrome":      "chrome",
		"navegador":   "chrome",
		"firefox":     "firefox",
		"edge":        "msedge",
		"notepad":     "notepad",
		"bloco":       "notepad",
		"calculadora": "calc",
		"calc":        "calc",
		"explorer":    "explorer",
		"arquivos":    "explorer",
		"terminal":    "wt",
		"cmd":         "cmd",
		"vscode":      "code",
		"code":        "code",
		"spotify":     "spotify",
		"discord":     "discord",
		"outlook":     "outlook",
		"word":        "winword",
		"excel":       "excel",
		"powerpoint":  "powerpnt",
	}

	executable := app
	if mapped, ok := appMap[strings.ToLower(app)]; ok {
		executable = mapped
	}

	cmd := exec.Command("cmd", "/c", "start", executable)
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("erro ao abrir %s: %w", app, err)
	}

	return fmt.Sprintf("Abrindo %s", app), nil
}

// openURL abre uma URL no navegador
func (e *Executor) openURL(params map[string]interface{}) (string, error) {
	url, ok := params["url"].(string)
	if !ok {
		return "", fmt.Errorf("parâmetro 'url' não fornecido")
	}

	// Adiciona https se não tiver protocolo
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "https://" + url
	}

	cmd := exec.Command("cmd", "/c", "start", url)
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("erro ao abrir URL: %w", err)
	}

	return fmt.Sprintf("Abrindo %s", url), nil
}

// typeText digita texto
func (e *Executor) typeText(params map[string]interface{}) (string, error) {
	text, ok := params["text"].(string)
	if !ok {
		return "", fmt.Errorf("parâmetro 'text' não fornecido")
	}

	// Usa PowerShell para simular digitação
	script := fmt.Sprintf(`
		Add-Type -AssemblyName System.Windows.Forms
		[System.Windows.Forms.SendKeys]::SendWait('%s')
	`, strings.ReplaceAll(text, "'", "''"))

	cmd := exec.Command("powershell", "-Command", script)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("erro ao digitar: %w", err)
	}

	return "Texto digitado", nil
}

// readEmail lê emails (Outlook)
func (e *Executor) readEmail(params map[string]interface{}) (string, error) {
	// PowerShell script para ler emails do Outlook
	script := `
		$outlook = New-Object -ComObject Outlook.Application
		$namespace = $outlook.GetNamespace("MAPI")
		$inbox = $namespace.GetDefaultFolder(6)
		$emails = $inbox.Items | Select-Object -First 5

		$result = @()
		foreach ($email in $emails) {
			$result += @{
				Subject = $email.Subject
				From = $email.SenderName
				Date = $email.ReceivedTime.ToString()
			}
		}

		$result | ConvertTo-Json
	`

	cmd := exec.Command("powershell", "-Command", script)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("erro ao ler emails: %w", err)
	}

	return string(output), nil
}

// sendEmail envia email
func (e *Executor) sendEmail(params map[string]interface{}) (string, error) {
	to, _ := params["to"].(string)
	subject, _ := params["subject"].(string)
	body, _ := params["body"].(string)

	if to == "" || subject == "" {
		return "", fmt.Errorf("parâmetros 'to' e 'subject' são obrigatórios")
	}

	script := fmt.Sprintf(`
		$outlook = New-Object -ComObject Outlook.Application
		$mail = $outlook.CreateItem(0)
		$mail.To = '%s'
		$mail.Subject = '%s'
		$mail.Body = '%s'
		$mail.Send()
	`, to, subject, body)

	cmd := exec.Command("powershell", "-Command", script)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("erro ao enviar email: %w", err)
	}

	return fmt.Sprintf("Email enviado para %s", to), nil
}

// setVolume ajusta volume do sistema
func (e *Executor) setVolume(params map[string]interface{}) (string, error) {
	level, ok := params["level"].(float64)
	if !ok {
		return "", fmt.Errorf("parâmetro 'level' não fornecido")
	}

	// Ajusta volume via nircmd ou PowerShell
	script := fmt.Sprintf(`
		$volume = [math]::Round(%f * 655.35)
		$wshShell = New-Object -ComObject WScript.Shell
		1..50 | ForEach-Object { $wshShell.SendKeys([char]174) }
		1..%d | ForEach-Object { $wshShell.SendKeys([char]175) }
	`, level, int(level/2))

	cmd := exec.Command("powershell", "-Command", script)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("erro ao ajustar volume: %w", err)
	}

	return fmt.Sprintf("Volume ajustado para %d%%", int(level)), nil
}

// takeScreenshot captura a tela
func (e *Executor) takeScreenshot(params map[string]interface{}) (string, error) {
	// Usa snippingtool ou PowerShell
	cmd := exec.Command("snippingtool", "/clip")
	if err := cmd.Run(); err != nil {
		// Fallback: PrintScreen key
		script := `
			Add-Type -AssemblyName System.Windows.Forms
			[System.Windows.Forms.SendKeys]::SendWait('{PRTSC}')
		`
		cmd = exec.Command("powershell", "-Command", script)
		cmd.Run()
	}

	return "Screenshot capturada", nil
}

// search faz uma busca na web
func (e *Executor) search(params map[string]interface{}) (string, error) {
	query, ok := params["query"].(string)
	if !ok {
		return "", fmt.Errorf("parâmetro 'query' não fornecido")
	}

	url := fmt.Sprintf("https://www.google.com/search?q=%s", strings.ReplaceAll(query, " ", "+"))
	return e.openURL(map[string]interface{}{"url": url})
}

// runCommand executa um comando
func (e *Executor) runCommand(params map[string]interface{}) (string, error) {
	command, ok := params["command"].(string)
	if !ok {
		return "", fmt.Errorf("parâmetro 'command' não fornecido")
	}

	// Por segurança, lista de comandos permitidos
	allowedCommands := []string{"dir", "echo", "date", "time", "hostname", "whoami"}
	allowed := false
	for _, c := range allowedCommands {
		if strings.HasPrefix(strings.ToLower(command), c) {
			allowed = true
			break
		}
	}

	if !allowed {
		return "", fmt.Errorf("comando não permitido por segurança")
	}

	cmd := exec.Command("cmd", "/c", command)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(output), nil
}
