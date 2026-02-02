package actions

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

// GmailClient cliente para acesso ao Gmail
type GmailClient struct {
	service *gmail.Service
	userID  string
}

// Email representa um email
type Email struct {
	ID      string   `json:"id"`
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	Body    string   `json:"body"`
	Date    string   `json:"date"`
	Unread  bool     `json:"unread"`
}

// NewGmailClient cria um novo cliente Gmail
func NewGmailClient(credentialsPath string) (*GmailClient, error) {
	ctx := context.Background()

	// Lê credenciais
	credentials, err := os.ReadFile(credentialsPath)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler credenciais: %w", err)
	}

	// Configura OAuth2
	config, err := google.ConfigFromJSON(credentials,
		gmail.GmailReadonlyScope,
		gmail.GmailSendScope,
		gmail.GmailModifyScope,
	)
	if err != nil {
		return nil, fmt.Errorf("erro ao configurar OAuth: %w", err)
	}

	// Obtém token
	token, err := getToken(config)
	if err != nil {
		return nil, fmt.Errorf("erro ao obter token: %w", err)
	}

	// Cria cliente
	client := config.Client(ctx, token)

	// Cria serviço Gmail
	service, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("erro ao criar serviço Gmail: %w", err)
	}

	return &GmailClient{
		service: service,
		userID:  "me",
	}, nil
}

// getToken obtém ou renova token OAuth
func getToken(config *oauth2.Config) (*oauth2.Token, error) {
	tokenPath := "configs/gmail_token.json"

	// Tenta carregar token existente
	token, err := loadToken(tokenPath)
	if err == nil {
		return token, nil
	}

	// Se não existe, inicia fluxo OAuth
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("\n🔐 Acesse este link para autorizar o Gmail:\n%s\n\n", authURL)
	fmt.Print("Cole o código de autorização: ")

	var code string
	fmt.Scan(&code)

	token, err = config.Exchange(context.Background(), code)
	if err != nil {
		return nil, err
	}

	// Salva token
	saveToken(tokenPath, token)

	return token, nil
}

// loadToken carrega token de arquivo
func loadToken(path string) (*oauth2.Token, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var token oauth2.Token
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, err
	}

	return &token, nil
}

// saveToken salva token em arquivo
func saveToken(path string, token *oauth2.Token) error {
	data, err := json.Marshal(token)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

// ListUnread lista emails não lidos
func (g *GmailClient) ListUnread(maxResults int64) ([]*Email, error) {
	if maxResults == 0 {
		maxResults = 10
	}

	// Lista mensagens não lidas
	result, err := g.service.Users.Messages.List(g.userID).
		Q("is:unread").
		MaxResults(maxResults).
		Do()
	if err != nil {
		return nil, err
	}

	emails := make([]*Email, 0, len(result.Messages))
	for _, msg := range result.Messages {
		email, err := g.GetMessage(msg.Id)
		if err != nil {
			continue
		}
		emails = append(emails, email)
	}

	return emails, nil
}

// GetMessage obtém detalhes de um email
func (g *GmailClient) GetMessage(messageID string) (*Email, error) {
	msg, err := g.service.Users.Messages.Get(g.userID, messageID).
		Format("full").
		Do()
	if err != nil {
		return nil, err
	}

	email := &Email{
		ID:     msg.Id,
		Unread: containsLabel(msg.LabelIds, "UNREAD"),
	}

	// Extrai headers
	for _, header := range msg.Payload.Headers {
		switch header.Name {
		case "From":
			email.From = header.Value
		case "To":
			email.To = strings.Split(header.Value, ", ")
		case "Subject":
			email.Subject = header.Value
		case "Date":
			email.Date = header.Value
		}
	}

	// Extrai corpo
	email.Body = extractBody(msg.Payload)

	return email, nil
}

// SendEmail envia um email
func (g *GmailClient) SendEmail(to, subject, body string) error {
	// Monta mensagem
	message := fmt.Sprintf(
		"To: %s\r\n"+
			"Subject: %s\r\n"+
			"Content-Type: text/plain; charset=UTF-8\r\n"+
			"\r\n"+
			"%s",
		to, subject, body,
	)

	// Codifica em base64
	raw := base64.URLEncoding.EncodeToString([]byte(message))

	// Envia
	_, err := g.service.Users.Messages.Send(g.userID, &gmail.Message{
		Raw: raw,
	}).Do()

	return err
}

// MarkAsRead marca email como lido
func (g *GmailClient) MarkAsRead(messageID string) error {
	_, err := g.service.Users.Messages.Modify(g.userID, messageID, &gmail.ModifyMessageRequest{
		RemoveLabelIds: []string{"UNREAD"},
	}).Do()
	return err
}

// Search busca emails
func (g *GmailClient) Search(query string, maxResults int64) ([]*Email, error) {
	if maxResults == 0 {
		maxResults = 10
	}

	result, err := g.service.Users.Messages.List(g.userID).
		Q(query).
		MaxResults(maxResults).
		Do()
	if err != nil {
		return nil, err
	}

	emails := make([]*Email, 0, len(result.Messages))
	for _, msg := range result.Messages {
		email, err := g.GetMessage(msg.Id)
		if err != nil {
			continue
		}
		emails = append(emails, email)
	}

	return emails, nil
}

// Summarize cria um resumo dos emails não lidos
func (g *GmailClient) Summarize() (string, error) {
	emails, err := g.ListUnread(5)
	if err != nil {
		return "", err
	}

	if len(emails) == 0 {
		return "Você não tem emails não lidos.", nil
	}

	var summary strings.Builder
	summary.WriteString(fmt.Sprintf("Você tem %d emails não lidos:\n\n", len(emails)))

	for i, email := range emails {
		summary.WriteString(fmt.Sprintf("%d. De: %s\n   Assunto: %s\n\n",
			i+1, email.From, email.Subject))
	}

	return summary.String(), nil
}

// containsLabel verifica se label está na lista
func containsLabel(labels []string, target string) bool {
	for _, label := range labels {
		if label == target {
			return true
		}
	}
	return false
}

// extractBody extrai corpo do email
func extractBody(payload *gmail.MessagePart) string {
	if payload.Body.Data != "" {
		data, err := base64.URLEncoding.DecodeString(payload.Body.Data)
		if err == nil {
			return string(data)
		}
	}

	for _, part := range payload.Parts {
		if body := extractBody(part); body != "" {
			return body
		}
	}

	return ""
}

// Placeholder para remover warning
var _ = http.StatusOK
