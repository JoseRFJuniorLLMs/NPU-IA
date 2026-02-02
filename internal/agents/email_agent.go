package agents

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// EmailAgent agente inteligente de e-mail
type EmailAgent struct {
	llm          LLMInterface
	emailService EmailServiceInterface
	userStyle    string // Estilo de escrita do usuário
}

// EmailSummary resumo de thread
type EmailSummary struct {
	ThreadID   string   `json:"thread_id"`
	Subject    string   `json:"subject"`
	KeyPoints  []string `json:"key_points"`
	ActionItems []string `json:"action_items"`
	Deadlines  []Deadline `json:"deadlines"`
}

// Deadline prazo extraído
type Deadline struct {
	Description string    `json:"description"`
	Date        time.Time `json:"date"`
	EmailID     string    `json:"email_id"`
}

// PriorityEmail email priorizado
type PriorityEmail struct {
	ID       string `json:"id"`
	Subject  string `json:"subject"`
	From     string `json:"from"`
	Priority int    `json:"priority"` // 1-5, sendo 5 mais urgente
	Reason   string `json:"reason"`
}

// LLMInterface interface para o modelo de linguagem
type LLMInterface interface {
	Generate(ctx context.Context, prompt string) (string, error)
}

// EmailServiceInterface interface para serviço de email
type EmailServiceInterface interface {
	ListEmails(query string, max int64) ([]map[string]string, error)
	GetEmailContent(id string) (string, error)
	SendEmail(to, subject, body string) error
	SearchEmails(query string) ([]map[string]string, error)
}

// NewEmailAgent cria agente de email
func NewEmailAgent(llm LLMInterface, emailService EmailServiceInterface) *EmailAgent {
	return &EmailAgent{
		llm:          llm,
		emailService: emailService,
	}
}

// ==================== 1. RESUMO DE THREADS ====================

// SummarizeThread resume uma thread de e-mail
func (e *EmailAgent) SummarizeThread(ctx context.Context, threadID string) (*EmailSummary, error) {
	// Obtém todos os e-mails da thread
	emails, err := e.emailService.ListEmails(fmt.Sprintf("thread:%s", threadID), 50)
	if err != nil {
		return nil, err
	}

	// Monta contexto
	var content strings.Builder
	for _, email := range emails {
		content.WriteString(fmt.Sprintf("De: %s\nAssunto: %s\n%s\n---\n",
			email["from"], email["subject"], email["body"]))
	}

	prompt := fmt.Sprintf(`Analise esta thread de e-mails e forneça:
1. 3-4 pontos principais da discussão
2. Itens de ação pendentes
3. Prazos mencionados

Thread:
%s

Responda em JSON:
{
  "key_points": ["ponto1", "ponto2"],
  "action_items": ["ação1", "ação2"],
  "deadlines": [{"description": "...", "date": "YYYY-MM-DD"}]
}`, content.String())

	response, err := e.llm.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	// Parse response
	summary := &EmailSummary{ThreadID: threadID}
	// TODO: Parse JSON response
	_ = response

	return summary, nil
}

// ==================== 2. TRIAGEM POR PRIORIDADE ====================

// PrioritizeEmails prioriza e-mails por urgência
func (e *EmailAgent) PrioritizeEmails(ctx context.Context) ([]*PriorityEmail, error) {
	// Obtém e-mails não lidos
	emails, err := e.emailService.ListEmails("is:unread", 50)
	if err != nil {
		return nil, err
	}

	var emailList strings.Builder
	for i, email := range emails {
		emailList.WriteString(fmt.Sprintf("%d. De: %s | Assunto: %s\n",
			i+1, email["from"], email["subject"]))
	}

	prompt := fmt.Sprintf(`Analise estes e-mails e classifique por prioridade (1-5, sendo 5 mais urgente).
Considere urgente:
- Cobranças e prazos
- Clientes insatisfeitos
- Problemas de produção
- Respostas esperadas por superiores

E-mails:
%s

Responda em JSON:
[{"index": 1, "priority": 5, "reason": "Cobrança com prazo hoje"}]`, emailList.String())

	response, err := e.llm.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	// Parse e ordena por prioridade
	prioritized := make([]*PriorityEmail, 0)
	// TODO: Parse JSON response
	_ = response

	return prioritized, nil
}

// ==================== 3. RASCUNHO COM "SUA VOZ" ====================

// DraftReply cria rascunho imitando estilo do usuário
func (e *EmailAgent) DraftReply(ctx context.Context, emailID, instruction string) (string, error) {
	// Obtém e-mail original
	content, err := e.emailService.GetEmailContent(emailID)
	if err != nil {
		return "", err
	}

	// Se não tem estilo definido, usa padrão
	style := e.userStyle
	if style == "" {
		style = "profissional mas amigável, direto ao ponto"
	}

	prompt := fmt.Sprintf(`Você é um assistente que escreve e-mails no estilo do usuário.

Estilo do usuário: %s

E-mail recebido:
%s

Instrução do usuário: %s

Escreva uma resposta completa que:
1. Mantenha o estilo de escrita indicado
2. Seja profissional
3. Responda ao conteúdo do e-mail
4. Seja conciso

Resposta:`, style, content, instruction)

	return e.llm.Generate(ctx, prompt)
}

// LearnUserStyle aprende estilo do usuário
func (e *EmailAgent) LearnUserStyle(ctx context.Context) error {
	// Obtém e-mails enviados pelo usuário
	emails, err := e.emailService.ListEmails("from:me", 20)
	if err != nil {
		return err
	}

	var samples strings.Builder
	for _, email := range emails[:min(10, len(emails))] {
		samples.WriteString(email["body"] + "\n---\n")
	}

	prompt := fmt.Sprintf(`Analise estes e-mails escritos pelo usuário e identifique seu estilo de escrita.
Considere:
- Tom (formal/informal)
- Uso de saudações
- Comprimento típico
- Expressões frequentes

E-mails:
%s

Descreva o estilo em uma frase:`, samples.String())

	style, err := e.llm.Generate(ctx, prompt)
	if err != nil {
		return err
	}

	e.userStyle = style
	return nil
}

// ==================== 4. BUSCA SEMÂNTICA ====================

// SemanticSearch busca por contexto
func (e *EmailAgent) SemanticSearch(ctx context.Context, query string) ([]map[string]string, error) {
	// Primeiro, expande a query para termos relacionados
	prompt := fmt.Sprintf(`O usuário quer encontrar um e-mail.
Descrição: "%s"

Gere 5 termos de busca que podem ajudar a encontrar este e-mail:`, query)

	terms, err := e.llm.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	// Busca usando termos expandidos
	allResults := make([]map[string]string, 0)
	searchTerms := strings.Split(terms, "\n")
	for _, term := range searchTerms {
		term = strings.TrimSpace(term)
		if term == "" {
			continue
		}
		results, _ := e.emailService.SearchEmails(term)
		allResults = append(allResults, results...)
	}

	// Remove duplicatas e rankeia por relevância
	// TODO: Implementar deduplicação e ranking

	return allResults, nil
}

// ==================== 5. EXTRAÇÃO DE PRAZOS ====================

// ExtractDeadlines extrai prazos de e-mails
func (e *EmailAgent) ExtractDeadlines(ctx context.Context) ([]Deadline, error) {
	// Obtém e-mails recentes
	emails, err := e.emailService.ListEmails("newer_than:7d", 100)
	if err != nil {
		return nil, err
	}

	deadlines := make([]Deadline, 0)

	for _, email := range emails {
		content, _ := e.emailService.GetEmailContent(email["id"])

		// Regex para datas comuns
		datePatterns := []string{
			`\d{1,2}/\d{1,2}/\d{4}`,
			`\d{1,2} de [a-zA-Z]+ de \d{4}`,
			`até dia \d{1,2}`,
			`prazo[:\s]+[^.]+`,
			`deadline[:\s]+[^.]+`,
			`vencimento[:\s]+[^.]+`,
		}

		for _, pattern := range datePatterns {
			re := regexp.MustCompile(pattern)
			matches := re.FindAllString(content, -1)
			for _, match := range matches {
				deadlines = append(deadlines, Deadline{
					Description: match,
					EmailID:     email["id"],
				})
			}
		}
	}

	return deadlines, nil
}

// ==================== 6. LIMPEZA DE SPAM ====================

// IdentifyUnsubscribable identifica newsletters para cancelar
func (e *EmailAgent) IdentifyUnsubscribable(ctx context.Context) ([]map[string]string, error) {
	// Busca newsletters e promoções
	emails, err := e.emailService.ListEmails("category:promotions OR unsubscribe", 100)
	if err != nil {
		return nil, err
	}

	// Agrupa por remetente e conta frequência
	senderCount := make(map[string]int)
	senderEmails := make(map[string]string)

	for _, email := range emails {
		from := email["from"]
		senderCount[from]++
		senderEmails[from] = email["id"]
	}

	// Identifica os que nunca foram abertos
	unsubscribable := make([]map[string]string, 0)
	for sender, count := range senderCount {
		if count > 5 { // Recebeu mais de 5 e-mails
			unsubscribable = append(unsubscribable, map[string]string{
				"sender":   sender,
				"count":    fmt.Sprintf("%d", count),
				"email_id": senderEmails[sender],
			})
		}
	}

	return unsubscribable, nil
}

// ==================== 7. BUSCA EM ANEXOS ====================

// SearchAttachments busca em anexos
func (e *EmailAgent) SearchAttachments(ctx context.Context, query string) ([]map[string]string, error) {
	// Busca e-mails com anexos
	emails, err := e.emailService.ListEmails("has:attachment", 50)
	if err != nil {
		return nil, err
	}

	// TODO: Para cada anexo:
	// 1. Baixar anexo
	// 2. Extrair texto (OCR para imagens, parsing para PDFs)
	// 3. Buscar query no texto

	return emails, nil
}

// ==================== 8. TRADUÇÃO INSTANTÂNEA ====================

// TranslateEmail traduz e-mail
func (e *EmailAgent) TranslateEmail(ctx context.Context, emailID, targetLang string) (string, error) {
	content, err := e.emailService.GetEmailContent(emailID)
	if err != nil {
		return "", err
	}

	prompt := fmt.Sprintf(`Traduza este e-mail para %s, mantendo o tom e formatação:

%s

Tradução:`, targetLang, content)

	return e.llm.Generate(ctx, prompt)
}

// TranslateAndReply traduz, entende e prepara resposta
func (e *EmailAgent) TranslateAndReply(ctx context.Context, emailID, replyInstruction, originalLang string) (string, error) {
	content, err := e.emailService.GetEmailContent(emailID)
	if err != nil {
		return "", err
	}

	prompt := fmt.Sprintf(`Este e-mail está em outro idioma.
1. Traduza para português para entender
2. Escreva uma resposta seguindo a instrução
3. Traduza a resposta de volta para %s

E-mail:
%s

Instrução: %s

Resposta em %s:`, originalLang, content, replyInstruction, originalLang)

	return e.llm.Generate(ctx, prompt)
}

// ==================== 9. FOLLOW-UP AUTOMÁTICO ====================

// PendingFollowUps identifica e-mails aguardando resposta
func (e *EmailAgent) PendingFollowUps(ctx context.Context, days int) ([]map[string]string, error) {
	// Busca e-mails enviados sem resposta
	query := fmt.Sprintf("from:me older_than:%dd", days)
	sentEmails, err := e.emailService.ListEmails(query, 50)
	if err != nil {
		return nil, err
	}

	pending := make([]map[string]string, 0)

	for _, email := range sentEmails {
		// Verifica se houve resposta
		threadQuery := fmt.Sprintf("to:me subject:%s", email["subject"])
		responses, _ := e.emailService.SearchEmails(threadQuery)

		if len(responses) == 0 {
			pending = append(pending, map[string]string{
				"id":      email["id"],
				"to":      email["to"],
				"subject": email["subject"],
				"sent":    email["date"],
			})
		}
	}

	return pending, nil
}

// GenerateFollowUp gera e-mail de follow-up
func (e *EmailAgent) GenerateFollowUp(ctx context.Context, originalEmailID string) (string, error) {
	content, err := e.emailService.GetEmailContent(originalEmailID)
	if err != nil {
		return "", err
	}

	prompt := fmt.Sprintf(`Escreva um e-mail de follow-up educado para este e-mail que não foi respondido:

E-mail original:
%s

Follow-up (seja breve e cordial):`, content)

	return e.llm.Generate(ctx, prompt)
}

// ==================== 10. CATEGORIZAÇÃO AUTOMÁTICA ====================

// CategorizeEmails categoriza e-mails automaticamente
func (e *EmailAgent) CategorizeEmails(ctx context.Context) (map[string][]string, error) {
	emails, err := e.emailService.ListEmails("is:unread", 50)
	if err != nil {
		return nil, err
	}

	categories := map[string][]string{
		"trabalho":  {},
		"pessoal":   {},
		"financas":  {},
		"viagens":   {},
		"compras":   {},
		"urgente":   {},
		"newsletter": {},
	}

	for _, email := range emails {
		prompt := fmt.Sprintf(`Categorize este e-mail em UMA das categorias:
- trabalho
- pessoal
- financas
- viagens
- compras
- urgente
- newsletter

De: %s
Assunto: %s

Categoria:`, email["from"], email["subject"])

		category, err := e.llm.Generate(ctx, prompt)
		if err != nil {
			continue
		}

		category = strings.ToLower(strings.TrimSpace(category))
		if _, ok := categories[category]; ok {
			categories[category] = append(categories[category], email["id"])
		}
	}

	return categories, nil
}

// Helper
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
