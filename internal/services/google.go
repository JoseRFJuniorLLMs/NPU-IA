package services

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/docs/v1"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
	"google.golang.org/api/tasks/v1"
)

// GoogleServices integração completa com Google
type GoogleServices struct {
	client   *oauth2.Token
	gmail    *gmail.Service
	calendar *calendar.Service
	drive    *drive.Service
	docs     *docs.Service
	sheets   *sheets.Service
	tasks    *tasks.Service
}

// NewGoogleServices inicializa todos os serviços Google
func NewGoogleServices(credentialsPath string) (*GoogleServices, error) {
	ctx := context.Background()

	credentials, err := os.ReadFile(credentialsPath)
	if err != nil {
		return nil, err
	}

	// Escopos para todos os serviços
	scopes := []string{
		gmail.GmailModifyScope,
		calendar.CalendarScope,
		drive.DriveScope,
		docs.DocumentsScope,
		sheets.SpreadsheetsScope,
		tasks.TasksScope,
	}

	config, err := google.ConfigFromJSON(credentials, scopes...)
	if err != nil {
		return nil, err
	}

	token, err := getGoogleToken(config)
	if err != nil {
		return nil, err
	}

	httpClient := config.Client(ctx, token)

	gs := &GoogleServices{}

	// Inicializa todos os serviços
	gs.gmail, _ = gmail.NewService(ctx, option.WithHTTPClient(httpClient))
	gs.calendar, _ = calendar.NewService(ctx, option.WithHTTPClient(httpClient))
	gs.drive, _ = drive.NewService(ctx, option.WithHTTPClient(httpClient))
	gs.docs, _ = docs.NewService(ctx, option.WithHTTPClient(httpClient))
	gs.sheets, _ = sheets.NewService(ctx, option.WithHTTPClient(httpClient))
	gs.tasks, _ = tasks.NewService(ctx, option.WithHTTPClient(httpClient))

	return gs, nil
}

// ==================== GMAIL ====================

// ListEmails lista emails
func (g *GoogleServices) ListEmails(query string, max int64) ([]map[string]string, error) {
	if max == 0 {
		max = 10
	}

	result, err := g.gmail.Users.Messages.List("me").Q(query).MaxResults(max).Do()
	if err != nil {
		return nil, err
	}

	emails := make([]map[string]string, 0)
	for _, msg := range result.Messages {
		full, _ := g.gmail.Users.Messages.Get("me", msg.Id).Format("metadata").Do()
		email := map[string]string{"id": msg.Id}
		for _, h := range full.Payload.Headers {
			if h.Name == "From" || h.Name == "Subject" || h.Name == "Date" {
				email[strings.ToLower(h.Name)] = h.Value
			}
		}
		emails = append(emails, email)
	}

	return emails, nil
}

// SendEmail envia email
func (g *GoogleServices) SendEmail(to, subject, body string) error {
	msg := fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s", to, subject, body)
	_, err := g.gmail.Users.Messages.Send("me", &gmail.Message{
		Raw: base64Encode(msg),
	}).Do()
	return err
}

// ==================== DRIVE ====================

// ListDriveFiles lista arquivos no Drive
func (g *GoogleServices) ListDriveFiles(query string, max int64) ([]*drive.File, error) {
	if max == 0 {
		max = 20
	}

	result, err := g.drive.Files.List().Q(query).PageSize(max).Fields("files(id, name, mimeType, size)").Do()
	if err != nil {
		return nil, err
	}

	return result.Files, nil
}

// SearchDrive busca arquivos
func (g *GoogleServices) SearchDrive(name string) ([]*drive.File, error) {
	query := fmt.Sprintf("name contains '%s'", name)
	return g.ListDriveFiles(query, 10)
}

// DownloadFile baixa arquivo do Drive
func (g *GoogleServices) DownloadFile(fileID, destPath string) error {
	resp, err := g.drive.Files.Get(fileID).Download()
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// UploadFile faz upload de arquivo
func (g *GoogleServices) UploadFile(filePath string, parentFolderID string) (*drive.File, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	driveFile := &drive.File{
		Name: filepath.Base(filePath),
	}
	if parentFolderID != "" {
		driveFile.Parents = []string{parentFolderID}
	}

	return g.drive.Files.Create(driveFile).Media(file).Do()
}

// ==================== DOCS ====================

// CreateDoc cria documento
func (g *GoogleServices) CreateDoc(title, content string) (*docs.Document, error) {
	doc := &docs.Document{Title: title}
	created, err := g.docs.Documents.Create(doc).Do()
	if err != nil {
		return nil, err
	}

	// Adiciona conteúdo
	if content != "" {
		requests := []*docs.Request{
			{
				InsertText: &docs.InsertTextRequest{
					Location: &docs.Location{Index: 1},
					Text:     content,
				},
			},
		}
		g.docs.Documents.BatchUpdate(created.DocumentId, &docs.BatchUpdateDocumentRequest{
			Requests: requests,
		}).Do()
	}

	return created, nil
}

// ReadDoc lê documento
func (g *GoogleServices) ReadDoc(docID string) (string, error) {
	doc, err := g.docs.Documents.Get(docID).Do()
	if err != nil {
		return "", err
	}

	var content strings.Builder
	for _, element := range doc.Body.Content {
		if element.Paragraph != nil {
			for _, elem := range element.Paragraph.Elements {
				if elem.TextRun != nil {
					content.WriteString(elem.TextRun.Content)
				}
			}
		}
	}

	return content.String(), nil
}

// ==================== SHEETS ====================

// ReadSheet lê planilha
func (g *GoogleServices) ReadSheet(sheetID, range_ string) ([][]interface{}, error) {
	result, err := g.sheets.Spreadsheets.Values.Get(sheetID, range_).Do()
	if err != nil {
		return nil, err
	}
	return result.Values, nil
}

// WriteSheet escreve na planilha
func (g *GoogleServices) WriteSheet(sheetID, range_ string, values [][]interface{}) error {
	valueRange := &sheets.ValueRange{Values: values}
	_, err := g.sheets.Spreadsheets.Values.Update(sheetID, range_, valueRange).
		ValueInputOption("USER_ENTERED").Do()
	return err
}

// CreateSheet cria planilha
func (g *GoogleServices) CreateSheet(title string) (*sheets.Spreadsheet, error) {
	sheet := &sheets.Spreadsheet{
		Properties: &sheets.SpreadsheetProperties{Title: title},
	}
	return g.sheets.Spreadsheets.Create(sheet).Do()
}

// ==================== CALENDAR ====================

// GetTodayEvents eventos de hoje
func (g *GoogleServices) GetTodayEvents() ([]*calendar.Event, error) {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	end := start.Add(24 * time.Hour)

	result, err := g.calendar.Events.List("primary").
		TimeMin(start.Format(time.RFC3339)).
		TimeMax(end.Format(time.RFC3339)).
		SingleEvents(true).
		OrderBy("startTime").
		Do()
	if err != nil {
		return nil, err
	}

	return result.Items, nil
}

// CreateEvent cria evento
func (g *GoogleServices) CreateEvent(title string, start, end time.Time, description string) (*calendar.Event, error) {
	event := &calendar.Event{
		Summary:     title,
		Description: description,
		Start:       &calendar.EventDateTime{DateTime: start.Format(time.RFC3339)},
		End:         &calendar.EventDateTime{DateTime: end.Format(time.RFC3339)},
	}
	return g.calendar.Events.Insert("primary", event).Do()
}

// ==================== TASKS ====================

// ListTasks lista tarefas
func (g *GoogleServices) ListTasks() ([]*tasks.Task, error) {
	// Primeiro pega a lista padrão
	lists, err := g.tasks.Tasklists.List().Do()
	if err != nil || len(lists.Items) == 0 {
		return nil, err
	}

	result, err := g.tasks.Tasks.List(lists.Items[0].Id).Do()
	if err != nil {
		return nil, err
	}

	return result.Items, nil
}

// CreateTask cria tarefa
func (g *GoogleServices) CreateTask(title, notes string, due time.Time) (*tasks.Task, error) {
	lists, _ := g.tasks.Tasklists.List().Do()
	if len(lists.Items) == 0 {
		return nil, fmt.Errorf("nenhuma lista de tarefas encontrada")
	}

	task := &tasks.Task{
		Title: title,
		Notes: notes,
	}
	if !due.IsZero() {
		task.Due = due.Format(time.RFC3339)
	}

	return g.tasks.Tasks.Insert(lists.Items[0].Id, task).Do()
}

// CompleteTask marca tarefa como concluída
func (g *GoogleServices) CompleteTask(taskID string) error {
	lists, _ := g.tasks.Tasklists.List().Do()
	if len(lists.Items) == 0 {
		return fmt.Errorf("nenhuma lista encontrada")
	}

	task, err := g.tasks.Tasks.Get(lists.Items[0].Id, taskID).Do()
	if err != nil {
		return err
	}

	task.Status = "completed"
	_, err = g.tasks.Tasks.Update(lists.Items[0].Id, taskID, task).Do()
	return err
}

// Helper functions
func base64Encode(s string) string {
	return strings.TrimRight(
		strings.ReplaceAll(
			strings.ReplaceAll(
				string([]byte(s)),
				"+", "-"),
			"/", "_"),
		"=")
}

func getGoogleToken(config *oauth2.Config) (*oauth2.Token, error) {
	tokenPath := "configs/google_token.json"
	// Implementação simplificada - ver gmail.go para versão completa
	return loadTokenFromFile(tokenPath)
}

func loadTokenFromFile(path string) (*oauth2.Token, error) {
	// Placeholder
	return nil, fmt.Errorf("token não encontrado")
}
