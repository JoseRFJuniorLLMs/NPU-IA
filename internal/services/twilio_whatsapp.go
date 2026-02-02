package services

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ==================== TWILIO WHATSAPP BUSINESS API ====================
// Integração completa com WhatsApp usando Twilio
// Permite usar seu próprio número de telefone

// TwilioWhatsApp cliente WhatsApp via Twilio
type TwilioWhatsApp struct {
	accountSID     string
	authToken      string
	phoneNumber    string // Seu número WhatsApp Business (formato: +5511999999999)
	messagingSID   string // Messaging Service SID (opcional)
	webhookURL     string
	statusCallback string

	// Cache e estado
	conversations map[string]*WhatsAppConversation
	templates     map[string]*MessageTemplate
	contacts      map[string]*WhatsAppContact
	mediaCache    map[string]string

	// Configurações
	defaultCountry string
	autoReply      bool
	typingDelay    time.Duration

	// HTTP client
	client *http.Client
	mu     sync.RWMutex

	// Callbacks
	onMessage       func(msg *WhatsAppMessage)
	onStatusUpdate  func(status *MessageStatus)
	onMediaReceived func(media *WhatsAppMedia)
}

// WhatsAppConversation conversa WhatsApp
type WhatsAppConversation struct {
	ID           string             `json:"id"`
	Contact      *WhatsAppContact   `json:"contact"`
	Messages     []*WhatsAppMessage `json:"messages"`
	LastActivity time.Time          `json:"last_activity"`
	IsOpen       bool               `json:"is_open"`
	Context      map[string]string  `json:"context"`
}

// WhatsAppContact contato WhatsApp
type WhatsAppContact struct {
	PhoneNumber string `json:"phone_number"`
	Name        string `json:"name"`
	ProfileName string `json:"profile_name"`
	WaID        string `json:"wa_id"`
	IsVerified  bool   `json:"is_verified"`
}

// WhatsAppMessage mensagem WhatsApp
type WhatsAppMessage struct {
	SID         string            `json:"sid"`
	From        string            `json:"from"`
	To          string            `json:"to"`
	Body        string            `json:"body"`
	MediaURL    string            `json:"media_url,omitempty"`
	MediaType   string            `json:"media_type,omitempty"`
	Status      string            `json:"status"`
	Direction   string            `json:"direction"` // inbound, outbound
	Timestamp   time.Time         `json:"timestamp"`
	IsRead      bool              `json:"is_read"`
	ReplyTo     string            `json:"reply_to,omitempty"`
	Buttons     []MessageButton   `json:"buttons,omitempty"`
	Location    *MessageLocation  `json:"location,omitempty"`
	Contacts    []MessageContact  `json:"contacts,omitempty"`
	Interactive *InteractiveMsg   `json:"interactive,omitempty"`
}

// MessageButton botão de mensagem
type MessageButton struct {
	Type    string `json:"type"` // reply, url, call
	Title   string `json:"title"`
	ID      string `json:"id,omitempty"`
	URL     string `json:"url,omitempty"`
	Phone   string `json:"phone,omitempty"`
}

// MessageLocation localização
type MessageLocation struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Name      string  `json:"name,omitempty"`
	Address   string  `json:"address,omitempty"`
}

// MessageContact contato compartilhado
type MessageContact struct {
	Name   string   `json:"name"`
	Phones []string `json:"phones"`
	Emails []string `json:"emails,omitempty"`
}

// InteractiveMsg mensagem interativa
type InteractiveMsg struct {
	Type    string           `json:"type"` // list, button, product
	Header  string           `json:"header,omitempty"`
	Body    string           `json:"body"`
	Footer  string           `json:"footer,omitempty"`
	Buttons []MessageButton  `json:"buttons,omitempty"`
	List    *InteractiveList `json:"list,omitempty"`
}

// InteractiveList lista interativa
type InteractiveList struct {
	ButtonText string          `json:"button_text"`
	Sections   []ListSection   `json:"sections"`
}

// ListSection seção da lista
type ListSection struct {
	Title string     `json:"title"`
	Rows  []ListRow  `json:"rows"`
}

// ListRow item da lista
type ListRow struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
}

// MessageTemplate template de mensagem
type MessageTemplate struct {
	Name       string            `json:"name"`
	Language   string            `json:"language"`
	Category   string            `json:"category"` // marketing, utility, authentication
	Components []TemplateComponent `json:"components"`
	Status     string            `json:"status"`
}

// TemplateComponent componente do template
type TemplateComponent struct {
	Type       string               `json:"type"` // header, body, footer, buttons
	Format     string               `json:"format,omitempty"`
	Text       string               `json:"text,omitempty"`
	Parameters []TemplateParameter  `json:"parameters,omitempty"`
}

// TemplateParameter parâmetro do template
type TemplateParameter struct {
	Type string `json:"type"` // text, image, document, video
	Text string `json:"text,omitempty"`
	URL  string `json:"url,omitempty"`
}

// MessageStatus status da mensagem
type MessageStatus struct {
	MessageSID string    `json:"message_sid"`
	Status     string    `json:"status"` // queued, sent, delivered, read, failed
	Timestamp  time.Time `json:"timestamp"`
	ErrorCode  string    `json:"error_code,omitempty"`
	ErrorMsg   string    `json:"error_message,omitempty"`
}

// WhatsAppMedia mídia WhatsApp
type WhatsAppMedia struct {
	SID         string `json:"sid"`
	ContentType string `json:"content_type"`
	URL         string `json:"url"`
	Filename    string `json:"filename,omitempty"`
	Size        int64  `json:"size"`
}

// TwilioConfig configuração Twilio
type TwilioConfig struct {
	AccountSID     string
	AuthToken      string
	PhoneNumber    string // Formato: +5511999999999
	MessagingSID   string // Opcional
	WebhookURL     string
	StatusCallback string
}

// NewTwilioWhatsApp cria cliente WhatsApp
func NewTwilioWhatsApp(config TwilioConfig) *TwilioWhatsApp {
	tw := &TwilioWhatsApp{
		accountSID:     config.AccountSID,
		authToken:      config.AuthToken,
		phoneNumber:    config.PhoneNumber,
		messagingSID:   config.MessagingSID,
		webhookURL:     config.WebhookURL,
		statusCallback: config.StatusCallback,
		conversations:  make(map[string]*WhatsAppConversation),
		templates:      make(map[string]*MessageTemplate),
		contacts:       make(map[string]*WhatsAppContact),
		mediaCache:     make(map[string]string),
		defaultCountry: "BR",
		autoReply:      false,
		typingDelay:    1 * time.Second,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	return tw
}

// ==================== ENVIO DE MENSAGENS ====================

// SendMessage envia mensagem de texto
func (tw *TwilioWhatsApp) SendMessage(ctx context.Context, to, message string) (*WhatsAppMessage, error) {
	to = tw.formatPhoneNumber(to)

	data := url.Values{}
	data.Set("To", "whatsapp:"+to)
	data.Set("From", "whatsapp:"+tw.phoneNumber)
	data.Set("Body", message)

	if tw.statusCallback != "" {
		data.Set("StatusCallback", tw.statusCallback)
	}

	return tw.sendRequest(ctx, data)
}

// SendMediaMessage envia mensagem com mídia
func (tw *TwilioWhatsApp) SendMediaMessage(ctx context.Context, to, mediaURL, caption string) (*WhatsAppMessage, error) {
	to = tw.formatPhoneNumber(to)

	data := url.Values{}
	data.Set("To", "whatsapp:"+to)
	data.Set("From", "whatsapp:"+tw.phoneNumber)
	data.Set("MediaUrl", mediaURL)
	if caption != "" {
		data.Set("Body", caption)
	}

	if tw.statusCallback != "" {
		data.Set("StatusCallback", tw.statusCallback)
	}

	return tw.sendRequest(ctx, data)
}

// SendImage envia imagem
func (tw *TwilioWhatsApp) SendImage(ctx context.Context, to, imageURL, caption string) (*WhatsAppMessage, error) {
	return tw.SendMediaMessage(ctx, to, imageURL, caption)
}

// SendDocument envia documento
func (tw *TwilioWhatsApp) SendDocument(ctx context.Context, to, documentURL, filename string) (*WhatsAppMessage, error) {
	return tw.SendMediaMessage(ctx, to, documentURL, filename)
}

// SendAudio envia áudio
func (tw *TwilioWhatsApp) SendAudio(ctx context.Context, to, audioURL string) (*WhatsAppMessage, error) {
	return tw.SendMediaMessage(ctx, to, audioURL, "")
}

// SendVideo envia vídeo
func (tw *TwilioWhatsApp) SendVideo(ctx context.Context, to, videoURL, caption string) (*WhatsAppMessage, error) {
	return tw.SendMediaMessage(ctx, to, videoURL, caption)
}

// SendLocation envia localização
func (tw *TwilioWhatsApp) SendLocation(ctx context.Context, to string, location *MessageLocation) (*WhatsAppMessage, error) {
	to = tw.formatPhoneNumber(to)

	// WhatsApp via Twilio não suporta location diretamente
	// Enviamos como link do Google Maps
	mapURL := fmt.Sprintf("https://www.google.com/maps?q=%f,%f", location.Latitude, location.Longitude)

	message := fmt.Sprintf("📍 %s\n%s\n%s", location.Name, location.Address, mapURL)

	return tw.SendMessage(ctx, to, message)
}

// SendContact envia contato
func (tw *TwilioWhatsApp) SendContact(ctx context.Context, to string, contact *MessageContact) (*WhatsAppMessage, error) {
	// Formata como vCard simplificado
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("👤 *%s*\n", contact.Name))
	for _, phone := range contact.Phones {
		sb.WriteString(fmt.Sprintf("📞 %s\n", phone))
	}
	for _, email := range contact.Emails {
		sb.WriteString(fmt.Sprintf("📧 %s\n", email))
	}

	return tw.SendMessage(ctx, to, sb.String())
}

// SendTemplate envia mensagem usando template aprovado
func (tw *TwilioWhatsApp) SendTemplate(ctx context.Context, to, templateName string, params map[string]string) (*WhatsAppMessage, error) {
	to = tw.formatPhoneNumber(to)

	// Monta Content SID para templates
	data := url.Values{}
	data.Set("To", "whatsapp:"+to)
	data.Set("From", "whatsapp:"+tw.phoneNumber)

	// Para templates, usa ContentSid ou ContentVariables
	if tw.templates[templateName] != nil {
		// Se tiver template registrado, usa ContentSid
		// Isso requer pré-registro no Twilio Console
		contentVars := make(map[string]string)
		i := 1
		for key, value := range params {
			contentVars[fmt.Sprintf("%d", i)] = value
			i++
			_ = key // key não usado no Twilio, só posição
		}
		varsJSON, _ := json.Marshal(contentVars)
		data.Set("ContentVariables", string(varsJSON))
	}

	if tw.statusCallback != "" {
		data.Set("StatusCallback", tw.statusCallback)
	}

	return tw.sendRequest(ctx, data)
}

// SendInteractiveButtons envia mensagem com botões
func (tw *TwilioWhatsApp) SendInteractiveButtons(ctx context.Context, to, body string, buttons []MessageButton) (*WhatsAppMessage, error) {
	// Twilio não suporta botões interativos diretamente via API básica
	// Formata como texto com opções numeradas
	var sb strings.Builder
	sb.WriteString(body)
	sb.WriteString("\n\n")
	for i, btn := range buttons {
		sb.WriteString(fmt.Sprintf("*%d.* %s\n", i+1, btn.Title))
	}
	sb.WriteString("\n_Responda com o número da opção_")

	return tw.SendMessage(ctx, to, sb.String())
}

// SendInteractiveList envia lista interativa
func (tw *TwilioWhatsApp) SendInteractiveList(ctx context.Context, to string, interactive *InteractiveMsg) (*WhatsAppMessage, error) {
	// Formata como texto estruturado
	var sb strings.Builder

	if interactive.Header != "" {
		sb.WriteString(fmt.Sprintf("*%s*\n\n", interactive.Header))
	}
	sb.WriteString(interactive.Body)
	sb.WriteString("\n\n")

	if interactive.List != nil {
		for _, section := range interactive.List.Sections {
			sb.WriteString(fmt.Sprintf("📋 *%s*\n", section.Title))
			for _, row := range section.Rows {
				sb.WriteString(fmt.Sprintf("  • %s", row.Title))
				if row.Description != "" {
					sb.WriteString(fmt.Sprintf(" - %s", row.Description))
				}
				sb.WriteString("\n")
			}
			sb.WriteString("\n")
		}
	}

	if interactive.Footer != "" {
		sb.WriteString(fmt.Sprintf("_%s_", interactive.Footer))
	}

	return tw.SendMessage(ctx, to, sb.String())
}

// ReplyToMessage responde a uma mensagem específica
func (tw *TwilioWhatsApp) ReplyToMessage(ctx context.Context, originalMsgSID, to, message string) (*WhatsAppMessage, error) {
	// Twilio não suporta reply direto, envia como mensagem normal
	return tw.SendMessage(ctx, to, message)
}

// ==================== UPLOAD DE MÍDIA ====================

// UploadMedia faz upload de arquivo local
func (tw *TwilioWhatsApp) UploadMedia(ctx context.Context, filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("MediaUrl", filepath.Base(filePath))
	if err != nil {
		return "", err
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return "", err
	}

	writer.Close()

	url := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", tw.accountSID)
	req, err := http.NewRequestWithContext(ctx, "POST", url, body)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.SetBasicAuth(tw.accountSID, tw.authToken)

	resp, err := tw.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		MediaUrl string `json:"media_url"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	return result.MediaUrl, nil
}

// ==================== RECEBIMENTO (WEBHOOK) ====================

// HandleWebhook processa webhook do Twilio
func (tw *TwilioWhatsApp) HandleWebhook(r *http.Request) (*WhatsAppMessage, error) {
	if err := r.ParseForm(); err != nil {
		return nil, err
	}

	msg := &WhatsAppMessage{
		SID:       r.FormValue("MessageSid"),
		From:      strings.TrimPrefix(r.FormValue("From"), "whatsapp:"),
		To:        strings.TrimPrefix(r.FormValue("To"), "whatsapp:"),
		Body:      r.FormValue("Body"),
		Status:    r.FormValue("SmsStatus"),
		Direction: "inbound",
		Timestamp: time.Now(),
	}

	// Mídia anexada
	numMedia := r.FormValue("NumMedia")
	if numMedia != "" && numMedia != "0" {
		msg.MediaURL = r.FormValue("MediaUrl0")
		msg.MediaType = r.FormValue("MediaContentType0")
	}

	// Localização (se enviada)
	if lat := r.FormValue("Latitude"); lat != "" {
		var latitude, longitude float64
		fmt.Sscanf(lat, "%f", &latitude)
		fmt.Sscanf(r.FormValue("Longitude"), "%f", &longitude)
		msg.Location = &MessageLocation{
			Latitude:  latitude,
			Longitude: longitude,
		}
	}

	// Armazena na conversa
	tw.mu.Lock()
	conv, ok := tw.conversations[msg.From]
	if !ok {
		conv = &WhatsAppConversation{
			ID: msg.From,
			Contact: &WhatsAppContact{
				PhoneNumber: msg.From,
				ProfileName: r.FormValue("ProfileName"),
			},
			Messages: make([]*WhatsAppMessage, 0),
			IsOpen:   true,
			Context:  make(map[string]string),
		}
		tw.conversations[msg.From] = conv
	}
	conv.Messages = append(conv.Messages, msg)
	conv.LastActivity = time.Now()
	tw.mu.Unlock()

	// Callback
	if tw.onMessage != nil {
		tw.onMessage(msg)
	}

	return msg, nil
}

// HandleStatusCallback processa callback de status
func (tw *TwilioWhatsApp) HandleStatusCallback(r *http.Request) (*MessageStatus, error) {
	if err := r.ParseForm(); err != nil {
		return nil, err
	}

	status := &MessageStatus{
		MessageSID: r.FormValue("MessageSid"),
		Status:     r.FormValue("MessageStatus"),
		Timestamp:  time.Now(),
		ErrorCode:  r.FormValue("ErrorCode"),
		ErrorMsg:   r.FormValue("ErrorMessage"),
	}

	if tw.onStatusUpdate != nil {
		tw.onStatusUpdate(status)
	}

	return status, nil
}

// ==================== CONVERSAS ====================

// GetConversation obtém conversa
func (tw *TwilioWhatsApp) GetConversation(phoneNumber string) *WhatsAppConversation {
	tw.mu.RLock()
	defer tw.mu.RUnlock()

	phoneNumber = tw.formatPhoneNumber(phoneNumber)
	return tw.conversations[phoneNumber]
}

// GetAllConversations retorna todas as conversas
func (tw *TwilioWhatsApp) GetAllConversations() []*WhatsAppConversation {
	tw.mu.RLock()
	defer tw.mu.RUnlock()

	convs := make([]*WhatsAppConversation, 0, len(tw.conversations))
	for _, c := range tw.conversations {
		convs = append(convs, c)
	}
	return convs
}

// GetUnreadConversations retorna conversas não lidas
func (tw *TwilioWhatsApp) GetUnreadConversations() []*WhatsAppConversation {
	tw.mu.RLock()
	defer tw.mu.RUnlock()

	unread := make([]*WhatsAppConversation, 0)
	for _, c := range tw.conversations {
		hasUnread := false
		for _, m := range c.Messages {
			if m.Direction == "inbound" && !m.IsRead {
				hasUnread = true
				break
			}
		}
		if hasUnread {
			unread = append(unread, c)
		}
	}
	return unread
}

// MarkAsRead marca mensagens como lidas
func (tw *TwilioWhatsApp) MarkAsRead(phoneNumber string) {
	tw.mu.Lock()
	defer tw.mu.Unlock()

	phoneNumber = tw.formatPhoneNumber(phoneNumber)
	if conv, ok := tw.conversations[phoneNumber]; ok {
		for _, m := range conv.Messages {
			m.IsRead = true
		}
	}
}

// ==================== HELPERS ====================

// formatPhoneNumber formata número de telefone
func (tw *TwilioWhatsApp) formatPhoneNumber(phone string) string {
	// Remove caracteres não numéricos
	phone = strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' || r == '+' {
			return r
		}
		return -1
	}, phone)

	// Adiciona código do país se não tiver
	if !strings.HasPrefix(phone, "+") {
		switch tw.defaultCountry {
		case "BR":
			phone = "+55" + phone
		case "US":
			phone = "+1" + phone
		default:
			phone = "+" + phone
		}
	}

	return phone
}

// sendRequest envia request para API Twilio
func (tw *TwilioWhatsApp) sendRequest(ctx context.Context, data url.Values) (*WhatsAppMessage, error) {
	apiURL := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", tw.accountSID)

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(tw.accountSID, tw.authToken)

	resp, err := tw.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("erro Twilio (%d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		SID        string `json:"sid"`
		Status     string `json:"status"`
		To         string `json:"to"`
		From       string `json:"from"`
		Body       string `json:"body"`
		DateCreated string `json:"date_created"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	msg := &WhatsAppMessage{
		SID:       result.SID,
		From:      strings.TrimPrefix(result.From, "whatsapp:"),
		To:        strings.TrimPrefix(result.To, "whatsapp:"),
		Body:      result.Body,
		Status:    result.Status,
		Direction: "outbound",
		Timestamp: time.Now(),
	}

	// Armazena na conversa
	tw.mu.Lock()
	to := msg.To
	conv, ok := tw.conversations[to]
	if !ok {
		conv = &WhatsAppConversation{
			ID:       to,
			Contact:  &WhatsAppContact{PhoneNumber: to},
			Messages: make([]*WhatsAppMessage, 0),
			IsOpen:   true,
			Context:  make(map[string]string),
		}
		tw.conversations[to] = conv
	}
	conv.Messages = append(conv.Messages, msg)
	conv.LastActivity = time.Now()
	tw.mu.Unlock()

	return msg, nil
}

// ==================== CALLBACKS ====================

// OnMessage registra callback para mensagens recebidas
func (tw *TwilioWhatsApp) OnMessage(callback func(msg *WhatsAppMessage)) {
	tw.onMessage = callback
}

// OnStatusUpdate registra callback para atualizações de status
func (tw *TwilioWhatsApp) OnStatusUpdate(callback func(status *MessageStatus)) {
	tw.onStatusUpdate = callback
}

// OnMediaReceived registra callback para mídia recebida
func (tw *TwilioWhatsApp) OnMediaReceived(callback func(media *WhatsAppMedia)) {
	tw.onMediaReceived = callback
}

// ==================== CONFIGURAÇÕES ====================

// SetDefaultCountry define país padrão
func (tw *TwilioWhatsApp) SetDefaultCountry(country string) {
	tw.defaultCountry = country
}

// SetAutoReply habilita auto-resposta
func (tw *TwilioWhatsApp) SetAutoReply(enabled bool) {
	tw.autoReply = enabled
}

// SetTypingDelay define delay de digitação simulado
func (tw *TwilioWhatsApp) SetTypingDelay(delay time.Duration) {
	tw.typingDelay = delay
}

// ==================== UTILIDADES ====================

// GetMessageStatus obtém status de uma mensagem
func (tw *TwilioWhatsApp) GetMessageStatus(ctx context.Context, messageSID string) (*MessageStatus, error) {
	apiURL := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages/%s.json",
		tw.accountSID, messageSID)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(tw.accountSID, tw.authToken)

	resp, err := tw.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Status       string `json:"status"`
		ErrorCode    string `json:"error_code"`
		ErrorMessage string `json:"error_message"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &MessageStatus{
		MessageSID: messageSID,
		Status:     result.Status,
		Timestamp:  time.Now(),
		ErrorCode:  result.ErrorCode,
		ErrorMsg:   result.ErrorMessage,
	}, nil
}

// DeleteMessage deleta mensagem (se suportado)
func (tw *TwilioWhatsApp) DeleteMessage(ctx context.Context, messageSID string) error {
	apiURL := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages/%s.json",
		tw.accountSID, messageSID)

	req, err := http.NewRequestWithContext(ctx, "DELETE", apiURL, nil)
	if err != nil {
		return err
	}

	req.SetBasicAuth(tw.accountSID, tw.authToken)

	resp, err := tw.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("erro ao deletar mensagem: %d", resp.StatusCode)
	}

	return nil
}

// GetStatus retorna status do serviço
func (tw *TwilioWhatsApp) GetStatus() map[string]interface{} {
	tw.mu.RLock()
	defer tw.mu.RUnlock()

	totalMessages := 0
	for _, c := range tw.conversations {
		totalMessages += len(c.Messages)
	}

	return map[string]interface{}{
		"phone_number":       tw.phoneNumber,
		"total_conversations": len(tw.conversations),
		"total_messages":     totalMessages,
		"auto_reply":         tw.autoReply,
		"default_country":    tw.defaultCountry,
	}
}

// ==================== INTEGRAÇÃO COM NPU-IA ====================

// WhatsAppAssistant assistente WhatsApp
type WhatsAppAssistant struct {
	whatsapp    *TwilioWhatsApp
	llm         LLMInterface
	memory      *Memory
	autoRespond bool
}

// NewWhatsAppAssistant cria assistente
func NewWhatsAppAssistant(whatsapp *TwilioWhatsApp, llm LLMInterface, memory *Memory) *WhatsAppAssistant {
	wa := &WhatsAppAssistant{
		whatsapp:    whatsapp,
		llm:         llm,
		memory:      memory,
		autoRespond: true,
	}

	// Registra handler de mensagens
	whatsapp.OnMessage(func(msg *WhatsAppMessage) {
		if wa.autoRespond {
			go wa.handleMessage(msg)
		}
	})

	return wa
}

// handleMessage processa mensagem recebida
func (wa *WhatsAppAssistant) handleMessage(msg *WhatsAppMessage) {
	ctx := context.Background()

	// Obtém contexto da memória
	context := ""
	if wa.memory != nil {
		context = wa.memory.GetContext(msg.Body)
	}

	// Gera resposta
	prompt := fmt.Sprintf(`Você é um assistente pessoal via WhatsApp. Responda de forma concisa e amigável.

Contexto do usuário:
%s

Mensagem recebida de %s:
%s

Responda de forma natural e útil:`, context, msg.From, msg.Body)

	response, err := wa.llm.Generate(ctx, prompt)
	if err != nil {
		response = "Desculpe, não consegui processar sua mensagem. Tente novamente."
	}

	// Envia resposta
	wa.whatsapp.SendMessage(ctx, msg.From, response)
}

// SendBriefing envia briefing diário via WhatsApp
func (wa *WhatsAppAssistant) SendBriefing(ctx context.Context, to string, briefing string) error {
	_, err := wa.whatsapp.SendMessage(ctx, to, briefing)
	return err
}

// SendReminder envia lembrete
func (wa *WhatsAppAssistant) SendReminder(ctx context.Context, to, reminder string) error {
	message := fmt.Sprintf("⏰ *Lembrete*\n\n%s", reminder)
	_, err := wa.whatsapp.SendMessage(ctx, to, message)
	return err
}

// SendTaskUpdate envia atualização de tarefa
func (wa *WhatsAppAssistant) SendTaskUpdate(ctx context.Context, to, task, status string) error {
	emoji := "📋"
	if status == "completed" {
		emoji = "✅"
	} else if status == "overdue" {
		emoji = "⚠️"
	}

	message := fmt.Sprintf("%s *Tarefa*: %s\n*Status*: %s", emoji, task, status)
	_, err := wa.whatsapp.SendMessage(ctx, to, message)
	return err
}

// ==================== HELPER PARA AUTENTICAÇÃO ====================

// BasicAuth gera header de autenticação
func (tw *TwilioWhatsApp) BasicAuth() string {
	auth := tw.accountSID + ":" + tw.authToken
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
}
