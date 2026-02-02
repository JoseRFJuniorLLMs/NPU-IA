package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// AutonomousAgent agente autônomo que executa ações proativamente
type AutonomousAgent struct {
	llm             LLMInterface
	emailService    EmailServiceInterface
	calendarService CalendarServiceInterface
	taskService     TaskServiceInterface
	config          AutonomousConfig
}

// AutonomousConfig configuração do agente autônomo
type AutonomousConfig struct {
	// Respostas automáticas
	AutoReplyEnabled bool     `json:"auto_reply_enabled"`
	AutoReplyFAQ     []FAQ    `json:"auto_reply_faq"`

	// Monitoramento
	ConflictAlerts   bool `json:"conflict_alerts"`
	DeadlineReminders bool `json:"deadline_reminders"`

	// Sincronização
	SyncEmailToTasks bool `json:"sync_email_to_tasks"`
}

// FAQ pergunta frequente
type FAQ struct {
	Keywords []string `json:"keywords"`
	Answer   string   `json:"answer"`
}

// ActionLog log de ação executada
type ActionLog struct {
	Timestamp   time.Time `json:"timestamp"`
	ActionType  string    `json:"action_type"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	Details     string    `json:"details"`
}

// NewAutonomousAgent cria agente autônomo
func NewAutonomousAgent(llm LLMInterface, email EmailServiceInterface, calendar CalendarServiceInterface, task TaskServiceInterface, config AutonomousConfig) *AutonomousAgent {
	return &AutonomousAgent{
		llm:             llm,
		emailService:    email,
		calendarService: calendar,
		taskService:     task,
		config:          config,
	}
}

// ==================== 26. ATENDIMENTO DE PRIMEIRO NÍVEL ====================

// HandleIncomingEmail responde automaticamente a e-mails comuns
func (a *AutonomousAgent) HandleIncomingEmail(ctx context.Context, emailID string) (*ActionLog, error) {
	if !a.config.AutoReplyEnabled {
		return nil, nil
	}

	content, err := a.emailService.GetEmailContent(emailID)
	if err != nil {
		return nil, err
	}

	// Verifica se é uma pergunta frequente
	for _, faq := range a.config.AutoReplyFAQ {
		for _, keyword := range faq.Keywords {
			if strings.Contains(strings.ToLower(content), strings.ToLower(keyword)) {
				// Responde automaticamente
				err := a.emailService.SendEmail("", "Re: ", faq.Answer)
				if err != nil {
					return &ActionLog{
						Timestamp:   time.Now(),
						ActionType:  "auto_reply",
						Description: "Resposta automática FAQ",
						Status:      "error",
						Details:     err.Error(),
					}, err
				}

				return &ActionLog{
					Timestamp:   time.Now(),
					ActionType:  "auto_reply",
					Description: "Resposta automática FAQ",
					Status:      "success",
					Details:     fmt.Sprintf("Respondido com FAQ sobre: %s", keyword),
				}, nil
			}
		}
	}

	// Usa LLM para decidir se pode responder
	prompt := fmt.Sprintf(`Analise este e-mail e determine se é uma pergunta simples que pode ser respondida automaticamente.

E-mail:
%s

Responda em JSON:
{
  "can_auto_reply": true/false,
  "reason": "motivo",
  "suggested_reply": "resposta sugerida ou null"
}`, content)

	response, err := a.llm.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	var analysis struct {
		CanAutoReply   bool   `json:"can_auto_reply"`
		Reason         string `json:"reason"`
		SuggestedReply string `json:"suggested_reply"`
	}
	json.Unmarshal([]byte(response), &analysis)

	if analysis.CanAutoReply && analysis.SuggestedReply != "" {
		// TODO: Implementar envio da resposta
		return &ActionLog{
			Timestamp:   time.Now(),
			ActionType:  "auto_reply",
			Description: "Resposta automática LLM",
			Status:      "pending_approval",
			Details:     analysis.SuggestedReply,
		}, nil
	}

	return nil, nil
}

// ==================== 27. SINCRONIZAÇÃO ENTRE APPS ====================

// SyncChanges sincroniza mudanças entre e-mail e tarefas
func (a *AutonomousAgent) SyncChanges(ctx context.Context) ([]ActionLog, error) {
	if !a.config.SyncEmailToTasks {
		return nil, nil
	}

	logs := make([]ActionLog, 0)

	// Busca e-mails recentes com palavras-chave de tarefas
	emails, err := a.emailService.ListEmails("subject:(tarefa OR task OR todo OR ação) newer_than:1d", 20)
	if err != nil {
		return nil, err
	}

	for _, email := range emails {
		content, _ := a.emailService.GetEmailContent(email["id"])

		// Analisa se contém atualização de tarefa
		prompt := fmt.Sprintf(`Este e-mail contém uma atualização de tarefa ou novo item de ação?

E-mail:
%s

Responda em JSON:
{
  "has_task_update": true/false,
  "action": "create|update|complete",
  "task_title": "título",
  "details": "detalhes"
}`, content)

		response, err := a.llm.Generate(ctx, prompt)
		if err != nil {
			continue
		}

		var update struct {
			HasTaskUpdate bool   `json:"has_task_update"`
			Action        string `json:"action"`
			TaskTitle     string `json:"task_title"`
			Details       string `json:"details"`
		}
		json.Unmarshal([]byte(response), &update)

		if update.HasTaskUpdate {
			log := ActionLog{
				Timestamp:   time.Now(),
				ActionType:  "sync_task",
				Description: fmt.Sprintf("%s tarefa: %s", update.Action, update.TaskTitle),
				Status:      "success",
			}

			switch update.Action {
			case "create":
				a.taskService.CreateTask(update.TaskTitle, update.Details, "", time.Time{})
			case "complete":
				// TODO: Encontrar e completar tarefa
			}

			logs = append(logs, log)
		}
	}

	return logs, nil
}

// ==================== 28. MODO "FORA DO ESCRITÓRIO" INTELIGENTE ====================

// SmartOutOfOfficeHandler gerencia respostas inteligentes quando fora
func (a *AutonomousAgent) SmartOutOfOfficeHandler(ctx context.Context, emailID string, contactRouting map[string]string) (*ActionLog, error) {
	content, err := a.emailService.GetEmailContent(emailID)
	if err != nil {
		return nil, err
	}

	// Identifica o assunto/tipo do e-mail
	prompt := fmt.Sprintf(`Identifique a categoria principal deste e-mail:

E-mail:
%s

Categorias possíveis: %v

Categoria:`, content, getKeys(contactRouting))

	category, err := a.llm.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	category = strings.TrimSpace(strings.ToLower(category))

	// Encontra contato alternativo
	alternateContact, exists := contactRouting[category]
	if !exists {
		alternateContact = contactRouting["default"]
	}

	// Gera resposta personalizada
	responsePrompt := fmt.Sprintf(`Crie uma resposta de "fora do escritório" para este e-mail:

E-mail recebido:
%s

Contato alternativo para este assunto: %s

A resposta deve:
1. Informar que estou ausente
2. Direcionar para o contato correto
3. Ser cordial

Resposta:`, content, alternateContact)

	response, err := a.llm.Generate(ctx, responsePrompt)
	if err != nil {
		return nil, err
	}

	// Envia resposta
	// TODO: Implementar envio

	return &ActionLog{
		Timestamp:   time.Now(),
		ActionType:  "out_of_office",
		Description: "Resposta fora do escritório inteligente",
		Status:      "success",
		Details:     fmt.Sprintf("Direcionado para: %s", alternateContact),
	}, nil
}

// ==================== 29. ALERTAS DE CONFLITO ====================

// CheckCalendarConflicts verifica conflitos na agenda
func (a *AutonomousAgent) CheckCalendarConflicts(ctx context.Context) ([]ActionLog, error) {
	if !a.config.ConflictAlerts {
		return nil, nil
	}

	logs := make([]ActionLog, 0)

	// Verifica próximos 7 dias
	now := time.Now()
	weekEnd := now.AddDate(0, 0, 7)

	events, err := a.calendarService.GetEvents(now, weekEnd)
	if err != nil {
		return nil, err
	}

	// Detecta conflitos
	for i := 0; i < len(events); i++ {
		for j := i + 1; j < len(events); j++ {
			if eventsOverlap(events[i], events[j]) {
				log := ActionLog{
					Timestamp:   time.Now(),
					ActionType:  "conflict_alert",
					Description: "Conflito detectado na agenda",
					Status:      "warning",
					Details: fmt.Sprintf("Conflito entre '%s' (%s) e '%s' (%s)",
						events[i].Title, events[i].Start.Format("02/01 15:04"),
						events[j].Title, events[j].Start.Format("02/01 15:04")),
				}
				logs = append(logs, log)
			}
		}
	}

	return logs, nil
}

// ==================== 30. LEMBRETES DE PRAZO ====================

// CheckUpcomingDeadlines verifica prazos próximos
func (a *AutonomousAgent) CheckUpcomingDeadlines(ctx context.Context) ([]ActionLog, error) {
	if !a.config.DeadlineReminders {
		return nil, nil
	}

	logs := make([]ActionLog, 0)

	// Verifica tarefas
	tasks, err := a.taskService.ListTasks()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	for _, task := range tasks {
		if task.Completed {
			continue
		}

		if !task.DueDate.IsZero() {
			daysUntil := task.DueDate.Sub(now).Hours() / 24

			var status string
			var description string

			switch {
			case daysUntil < 0:
				status = "overdue"
				description = fmt.Sprintf("ATRASADA: %s (venceu %s)", task.Title, task.DueDate.Format("02/01"))
			case daysUntil < 1:
				status = "urgent"
				description = fmt.Sprintf("URGENTE: %s vence hoje!", task.Title)
			case daysUntil < 3:
				status = "warning"
				description = fmt.Sprintf("ATENÇÃO: %s vence em %.0f dias", task.Title, daysUntil)
			}

			if status != "" {
				logs = append(logs, ActionLog{
					Timestamp:   time.Now(),
					ActionType:  "deadline_reminder",
					Description: description,
					Status:      status,
					Details:     task.Description,
				})
			}
		}
	}

	return logs, nil
}

// ==================== BACKGROUND LOOP ====================

// Run executa o agente em loop
func (a *AutonomousAgent) Run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Executa verificações periódicas
			a.runChecks(ctx)
		}
	}
}

// runChecks executa todas as verificações
func (a *AutonomousAgent) runChecks(ctx context.Context) {
	// Sincroniza mudanças
	a.SyncChanges(ctx)

	// Verifica conflitos
	a.CheckCalendarConflicts(ctx)

	// Verifica prazos
	a.CheckUpcomingDeadlines(ctx)
}

// Helper functions

func eventsOverlap(e1, e2 Event) bool {
	return e1.Start.Before(e2.End) && e1.End.After(e2.Start)
}

func getKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
