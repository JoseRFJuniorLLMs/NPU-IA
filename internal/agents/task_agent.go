package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// TaskAgent agente de automação de tarefas
type TaskAgent struct {
	llm          LLMInterface
	taskService  TaskServiceInterface
	emailService EmailServiceInterface
}

// TaskServiceInterface interface para serviço de tarefas
type TaskServiceInterface interface {
	CreateTask(title, description, project string, dueDate time.Time) (string, error)
	ListTasks() ([]Task, error)
	CompleteTask(taskID string) error
	GetProjects() ([]Project, error)
}

// Task tarefa
type Task struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Project     string    `json:"project"`
	DueDate     time.Time `json:"due_date"`
	Completed   bool      `json:"completed"`
}

// Project projeto
type Project struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// TravelItinerary itinerário de viagem
type TravelItinerary struct {
	Destination string       `json:"destination"`
	StartDate   time.Time    `json:"start_date"`
	EndDate     time.Time    `json:"end_date"`
	Flights     []FlightInfo `json:"flights"`
	Hotels      []HotelInfo  `json:"hotels"`
	Events      []Event      `json:"events"`
}

// FlightInfo informações de voo
type FlightInfo struct {
	Airline     string    `json:"airline"`
	FlightNo    string    `json:"flight_no"`
	Departure   time.Time `json:"departure"`
	Arrival     time.Time `json:"arrival"`
	Origin      string    `json:"origin"`
	Destination string    `json:"destination"`
}

// HotelInfo informações de hotel
type HotelInfo struct {
	Name     string    `json:"name"`
	CheckIn  time.Time `json:"check_in"`
	CheckOut time.Time `json:"check_out"`
	Address  string    `json:"address"`
}

// NewTaskAgent cria agente de tarefas
func NewTaskAgent(llm LLMInterface, taskService TaskServiceInterface, emailService EmailServiceInterface) *TaskAgent {
	return &TaskAgent{
		llm:          llm,
		taskService:  taskService,
		emailService: emailService,
	}
}

// ==================== 16. CRIAÇÃO DE TAREFAS ====================

// CreateTaskFromEmail cria tarefa a partir de e-mail
func (t *TaskAgent) CreateTaskFromEmail(ctx context.Context, emailID string) (*Task, error) {
	content, err := t.emailService.GetEmailContent(emailID)
	if err != nil {
		return nil, err
	}

	prompt := fmt.Sprintf(`Analise este e-mail e extraia uma tarefa:

%s

Responda em JSON:
{
  "title": "título curto da tarefa",
  "description": "detalhes relevantes",
  "due_date": "YYYY-MM-DD ou null",
  "project": "nome do projeto sugerido"
}`, content)

	response, err := t.llm.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	var taskInfo struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		DueDate     string `json:"due_date"`
		Project     string `json:"project"`
	}
	json.Unmarshal([]byte(response), &taskInfo)

	var dueDate time.Time
	if taskInfo.DueDate != "" {
		dueDate, _ = time.Parse("2006-01-02", taskInfo.DueDate)
	}

	taskID, err := t.taskService.CreateTask(taskInfo.Title, taskInfo.Description, taskInfo.Project, dueDate)
	if err != nil {
		return nil, err
	}

	return &Task{
		ID:          taskID,
		Title:       taskInfo.Title,
		Description: taskInfo.Description,
		DueDate:     dueDate,
	}, nil
}

// ==================== 17. PREENCHIMENTO DE FORMULÁRIOS ====================

// ExtractFormData extrai dados para preencher formulário
func (t *TaskAgent) ExtractFormData(ctx context.Context, emailContent string, formFields []string) (map[string]string, error) {
	fieldsStr := strings.Join(formFields, ", ")

	prompt := fmt.Sprintf(`Extraia os seguintes campos deste e-mail:
Campos: %s

E-mail:
%s

Responda em JSON com os campos preenchidos:`, fieldsStr, emailContent)

	response, err := t.llm.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	var data map[string]string
	json.Unmarshal([]byte(response), &data)

	return data, nil
}

// ==================== 18. PESQUISA PROFUNDA ====================

// DeepResearch faz pesquisa detalhada
func (t *TaskAgent) DeepResearch(ctx context.Context, topic string) (string, error) {
	prompt := fmt.Sprintf(`Você é um pesquisador especialista.

Faça uma pesquisa detalhada sobre: %s

Estruture sua resposta com:
1. Resumo Executivo
2. Contexto e Background
3. Principais Descobertas
4. Dados e Estatísticas
5. Conclusões e Recomendações
6. Fontes e Referências

Pesquisa:`, topic)

	research, err := t.llm.Generate(ctx, prompt)
	if err != nil {
		return "", err
	}

	return research, nil
}

// ==================== 19. ORGANIZAÇÃO DE VIAGENS ====================

// OrganizeTravel organiza viagem a partir de e-mails
func (t *TaskAgent) OrganizeTravel(ctx context.Context) (*TravelItinerary, error) {
	// Busca e-mails de viagem
	emails, err := t.emailService.ListEmails("subject:(confirmação OR confirmation OR booking OR reserva) newer_than:30d", 20)
	if err != nil {
		return nil, err
	}

	itinerary := &TravelItinerary{}

	for _, email := range emails {
		content, _ := t.emailService.GetEmailContent(email["id"])

		// Detecta tipo de confirmação
		prompt := fmt.Sprintf(`Analise esta confirmação de viagem e extraia as informações:

%s

Responda em JSON:
{
  "type": "flight|hotel|car|event",
  "details": {
    // campos específicos do tipo
  }
}`, content)

		response, err := t.llm.Generate(ctx, prompt)
		if err != nil {
			continue
		}

		var info struct {
			Type    string                 `json:"type"`
			Details map[string]interface{} `json:"details"`
		}
		json.Unmarshal([]byte(response), &info)

		// TODO: Parse e adiciona ao itinerário baseado no tipo
		_ = info
	}

	return itinerary, nil
}

// ==================== 20. MONITORAMENTO DE PREÇOS ====================

// SetPriceAlert configura alerta de preço
func (t *TaskAgent) SetPriceAlert(ctx context.Context, product, url string, targetPrice float64) error {
	// Cria tarefa de monitoramento
	description := fmt.Sprintf(`Monitorar preço de: %s
URL: %s
Preço alvo: R$ %.2f

Verificar periodicamente e alertar quando atingir o preço.`, product, url, targetPrice)

	_, err := t.taskService.CreateTask(
		fmt.Sprintf("Alerta de preço: %s", product),
		description,
		"Monitoramento",
		time.Time{}, // Sem data específica
	)

	return err
}

// ==================== SYNC DE APPS ====================

// SyncEmailToTask sincroniza e-mail com tarefa
func (t *TaskAgent) SyncEmailToTask(ctx context.Context, emailID, taskID string) error {
	content, err := t.emailService.GetEmailContent(emailID)
	if err != nil {
		return err
	}

	// Extrai atualizações
	prompt := fmt.Sprintf(`Este e-mail contém atualizações sobre uma tarefa existente?
Extraia qualquer mudança de prazo, status ou detalhes adicionais.

E-mail:
%s

Atualizações em JSON:
{
  "new_deadline": "YYYY-MM-DD ou null",
  "status_update": "descrição ou null",
  "additional_info": "info ou null"
}`, content)

	response, err := t.llm.Generate(ctx, prompt)
	if err != nil {
		return err
	}

	// TODO: Aplicar atualizações à tarefa
	_ = response

	return nil
}

// ==================== FORA DO ESCRITÓRIO ====================

// SmartOutOfOffice configura resposta automática inteligente
func (t *TaskAgent) SmartOutOfOffice(ctx context.Context, startDate, endDate time.Time, contacts map[string]string) (string, error) {
	// contacts: map[assunto]contato_alternativo

	contactList := ""
	for subject, contact := range contacts {
		contactList += fmt.Sprintf("- %s: %s\n", subject, contact)
	}

	prompt := fmt.Sprintf(`Crie uma mensagem de "fora do escritório" profissional:

Período: %s a %s
Contatos alternativos:
%s

A mensagem deve:
1. Informar o período de ausência
2. Direcionar para o contato correto baseado no assunto
3. Ser cordial e profissional

Mensagem:`, startDate.Format("02/01"), endDate.Format("02/01"), contactList)

	return t.llm.Generate(ctx, prompt)
}

// ==================== ANÁLISE DE CONTRATOS ====================

// AnalyzeContract analisa contrato
func (t *TaskAgent) AnalyzeContract(ctx context.Context, contractText string) (map[string]interface{}, error) {
	prompt := fmt.Sprintf(`Você é um especialista em análise de contratos.

Analise este contrato e identifique:
1. Partes envolvidas
2. Objeto do contrato
3. Valor e forma de pagamento
4. Prazo de vigência
5. Cláusulas de rescisão
6. Multas e penalidades
7. Datas importantes (renovação, vencimento)
8. Cláusulas de risco ou atenção

Contrato:
%s

Análise em JSON:
{
  "parties": [],
  "object": "",
  "value": "",
  "payment_terms": "",
  "validity": "",
  "termination_clauses": [],
  "penalties": [],
  "important_dates": [],
  "risk_clauses": [],
  "recommendations": []
}`, contractText)

	response, err := t.llm.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	var analysis map[string]interface{}
	json.Unmarshal([]byte(response), &analysis)

	return analysis, nil
}

// ==================== LOCALIZAÇÃO DE INFORMAÇÕES ====================

// FindPersonalInfo busca informação pessoal em e-mails
func (t *TaskAgent) FindPersonalInfo(ctx context.Context, query string) (string, error) {
	// Detecta tipo de informação sendo buscada
	infoTypes := map[string]string{
		"matrícula":     "número de matrícula",
		"cpf":           "número de CPF",
		"telefone":      "número de telefone",
		"endereço":      "endereço",
		"senha":         "ATENÇÃO: não posso buscar senhas",
		"conta":         "número de conta",
		"reserva":       "código de reserva",
		"confirmação":   "número de confirmação",
	}

	searchQuery := query
	for keyword, description := range infoTypes {
		if strings.Contains(strings.ToLower(query), keyword) {
			searchQuery = description
			break
		}
	}

	// Busca em e-mails
	emails, err := t.emailService.SearchEmails(searchQuery)
	if err != nil {
		return "", err
	}

	if len(emails) == 0 {
		return "Não encontrei essa informação nos seus e-mails.", nil
	}

	// Analisa resultados
	var contents strings.Builder
	for _, email := range emails[:min(5, len(emails))] {
		content, _ := t.emailService.GetEmailContent(email["id"])
		contents.WriteString(content + "\n---\n")
	}

	prompt := fmt.Sprintf(`O usuário está procurando: %s

Analise estes e-mails e extraia a informação:
%s

Resposta:`, query, contents.String())

	return t.llm.Generate(ctx, prompt)
}

// ==================== MELHORIA DE ESCRITA ====================

// ImproveWriting melhora texto
func (t *TaskAgent) ImproveWriting(ctx context.Context, text, style string) (string, error) {
	styleInstructions := map[string]string{
		"profissional": "Torne o texto mais formal e profissional, mantendo a clareza.",
		"curto":        "Reduza o texto ao essencial, mantendo a mensagem principal.",
		"amigável":     "Torne o texto mais acolhedor e amigável, mantendo profissionalismo.",
		"assertivo":    "Torne o texto mais direto e assertivo.",
		"detalhado":    "Adicione mais detalhes e contexto ao texto.",
	}

	instruction := styleInstructions[style]
	if instruction == "" {
		instruction = styleInstructions["profissional"]
	}

	prompt := fmt.Sprintf(`%s

Texto original:
%s

Texto melhorado:`, instruction, text)

	return t.llm.Generate(ctx, prompt)
}

// ==================== HELPER FUNCTIONS ====================

// ExtractDatesFromText extrai datas de um texto
func ExtractDatesFromText(text string) []string {
	patterns := []string{
		`\d{1,2}/\d{1,2}/\d{4}`,
		`\d{1,2}-\d{1,2}-\d{4}`,
		`\d{1,2} de [a-zA-Záéíóú]+ de \d{4}`,
	}

	dates := make([]string, 0)
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllString(text, -1)
		dates = append(dates, matches...)
	}

	return dates
}
