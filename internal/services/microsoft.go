package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// MicrosoftServices integração com Microsoft Graph API
type MicrosoftServices struct {
	accessToken string
	client      *http.Client
	baseURL     string
}

// Outlook Email
type OutlookEmail struct {
	ID      string `json:"id"`
	Subject string `json:"subject"`
	From    struct {
		EmailAddress struct {
			Name    string `json:"name"`
			Address string `json:"address"`
		} `json:"emailAddress"`
	} `json:"from"`
	Body struct {
		Content string `json:"content"`
	} `json:"body"`
	ReceivedDateTime string `json:"receivedDateTime"`
	IsRead           bool   `json:"isRead"`
}

// OneDriveItem item do OneDrive
type OneDriveItem struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Size   int64  `json:"size"`
	Folder *struct {
		ChildCount int `json:"childCount"`
	} `json:"folder,omitempty"`
	File *struct {
		MimeType string `json:"mimeType"`
	} `json:"file,omitempty"`
}

// TeamsMessage mensagem do Teams
type TeamsMessage struct {
	ID      string `json:"id"`
	Content string `json:"body>content"`
	From    struct {
		User struct {
			DisplayName string `json:"displayName"`
		} `json:"user"`
	} `json:"from"`
}

// TodoTask tarefa do Microsoft To Do
type TodoTask struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Status    string `json:"status"`
	DueDate   string `json:"dueDateTime>dateTime,omitempty"`
	Completed bool   `json:"isComplete"`
}

// NewMicrosoftServices cria cliente Microsoft
func NewMicrosoftServices(clientID, clientSecret, tenantID string) (*MicrosoftServices, error) {
	ms := &MicrosoftServices{
		client:  &http.Client{Timeout: 30 * time.Second},
		baseURL: "https://graph.microsoft.com/v1.0",
	}

	// Obtém token OAuth2
	token, err := ms.getAccessToken(clientID, clientSecret, tenantID)
	if err != nil {
		return nil, err
	}
	ms.accessToken = token

	return ms, nil
}

// getAccessToken obtém token de acesso
func (ms *MicrosoftServices) getAccessToken(clientID, clientSecret, tenantID string) (string, error) {
	tokenURL := fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", tenantID)

	data := url.Values{}
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("scope", "https://graph.microsoft.com/.default")
	data.Set("grant_type", "client_credentials")

	resp, err := http.PostForm(tokenURL, data)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken string `json:"access_token"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	return result.AccessToken, nil
}

// request faz requisição à API
func (ms *MicrosoftServices) request(method, endpoint string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, ms.baseURL+endpoint, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+ms.accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := ms.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// ==================== OUTLOOK ====================

// ListEmails lista emails do Outlook
func (ms *MicrosoftServices) ListEmails(folder string, top int) ([]*OutlookEmail, error) {
	if folder == "" {
		folder = "inbox"
	}
	if top == 0 {
		top = 10
	}

	endpoint := fmt.Sprintf("/me/mailFolders/%s/messages?$top=%d&$orderby=receivedDateTime desc", folder, top)
	data, err := ms.request("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Value []*OutlookEmail `json:"value"`
	}
	json.Unmarshal(data, &result)

	return result.Value, nil
}

// SendEmail envia email pelo Outlook
func (ms *MicrosoftServices) SendEmail(to, subject, body string) error {
	message := map[string]interface{}{
		"message": map[string]interface{}{
			"subject": subject,
			"body": map[string]string{
				"contentType": "Text",
				"content":     body,
			},
			"toRecipients": []map[string]interface{}{
				{"emailAddress": map[string]string{"address": to}},
			},
		},
	}

	_, err := ms.request("POST", "/me/sendMail", message)
	return err
}

// ==================== ONEDRIVE ====================

// ListOneDriveFiles lista arquivos
func (ms *MicrosoftServices) ListOneDriveFiles(folderPath string) ([]*OneDriveItem, error) {
	endpoint := "/me/drive/root/children"
	if folderPath != "" {
		endpoint = fmt.Sprintf("/me/drive/root:/%s:/children", folderPath)
	}

	data, err := ms.request("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Value []*OneDriveItem `json:"value"`
	}
	json.Unmarshal(data, &result)

	return result.Value, nil
}

// SearchOneDrive busca arquivos
func (ms *MicrosoftServices) SearchOneDrive(query string) ([]*OneDriveItem, error) {
	endpoint := fmt.Sprintf("/me/drive/root/search(q='%s')", url.QueryEscape(query))
	data, err := ms.request("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Value []*OneDriveItem `json:"value"`
	}
	json.Unmarshal(data, &result)

	return result.Value, nil
}

// DownloadOneDriveFile baixa arquivo
func (ms *MicrosoftServices) DownloadOneDriveFile(itemID string) ([]byte, error) {
	endpoint := fmt.Sprintf("/me/drive/items/%s/content", itemID)
	return ms.request("GET", endpoint, nil)
}

// ==================== TEAMS ====================

// ListTeams lista times
func (ms *MicrosoftServices) ListTeams() ([]map[string]interface{}, error) {
	data, err := ms.request("GET", "/me/joinedTeams", nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Value []map[string]interface{} `json:"value"`
	}
	json.Unmarshal(data, &result)

	return result.Value, nil
}

// SendTeamsMessage envia mensagem no Teams
func (ms *MicrosoftServices) SendTeamsMessage(teamID, channelID, message string) error {
	endpoint := fmt.Sprintf("/teams/%s/channels/%s/messages", teamID, channelID)
	body := map[string]interface{}{
		"body": map[string]string{
			"content": message,
		},
	}
	_, err := ms.request("POST", endpoint, body)
	return err
}

// ==================== TO DO ====================

// ListTodoTasks lista tarefas
func (ms *MicrosoftServices) ListTodoTasks() ([]*TodoTask, error) {
	// Primeiro pega as listas
	listsData, err := ms.request("GET", "/me/todo/lists", nil)
	if err != nil {
		return nil, err
	}

	var lists struct {
		Value []struct {
			ID string `json:"id"`
		} `json:"value"`
	}
	json.Unmarshal(listsData, &lists)

	if len(lists.Value) == 0 {
		return nil, nil
	}

	// Pega tarefas da primeira lista
	endpoint := fmt.Sprintf("/me/todo/lists/%s/tasks", lists.Value[0].ID)
	data, err := ms.request("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Value []*TodoTask `json:"value"`
	}
	json.Unmarshal(data, &result)

	return result.Value, nil
}

// CreateTodoTask cria tarefa
func (ms *MicrosoftServices) CreateTodoTask(title string, dueDate time.Time) (*TodoTask, error) {
	listsData, _ := ms.request("GET", "/me/todo/lists", nil)
	var lists struct {
		Value []struct {
			ID string `json:"id"`
		} `json:"value"`
	}
	json.Unmarshal(listsData, &lists)

	if len(lists.Value) == 0 {
		return nil, fmt.Errorf("nenhuma lista encontrada")
	}

	endpoint := fmt.Sprintf("/me/todo/lists/%s/tasks", lists.Value[0].ID)
	body := map[string]interface{}{
		"title": title,
	}
	if !dueDate.IsZero() {
		body["dueDateTime"] = map[string]string{
			"dateTime": dueDate.Format(time.RFC3339),
			"timeZone": "UTC",
		}
	}

	data, err := ms.request("POST", endpoint, body)
	if err != nil {
		return nil, err
	}

	var task TodoTask
	json.Unmarshal(data, &task)

	return &task, nil
}

// ==================== CALENDAR ====================

// GetCalendarEvents eventos do calendário
func (ms *MicrosoftServices) GetCalendarEvents(start, end time.Time) ([]map[string]interface{}, error) {
	endpoint := fmt.Sprintf("/me/calendarview?startDateTime=%s&endDateTime=%s",
		start.Format(time.RFC3339),
		end.Format(time.RFC3339))

	data, err := ms.request("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Value []map[string]interface{} `json:"value"`
	}
	json.Unmarshal(data, &result)

	return result.Value, nil
}

// CreateCalendarEvent cria evento
func (ms *MicrosoftServices) CreateCalendarEvent(subject string, start, end time.Time, body string) error {
	event := map[string]interface{}{
		"subject": subject,
		"body": map[string]string{
			"contentType": "Text",
			"content":     body,
		},
		"start": map[string]string{
			"dateTime": start.Format(time.RFC3339),
			"timeZone": "UTC",
		},
		"end": map[string]string{
			"dateTime": end.Format(time.RFC3339),
			"timeZone": "UTC",
		},
	}

	_, err := ms.request("POST", "/me/events", event)
	return err
}

// Placeholder para warning
var _ = strings.TrimSpace
