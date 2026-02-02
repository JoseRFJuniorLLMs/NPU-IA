package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ==================== NOTION ====================

// NotionServices integração com Notion
type NotionServices struct {
	apiKey  string
	client  *http.Client
	baseURL string
	version string
}

// NotionPage página
type NotionPage struct {
	ID         string                 `json:"id"`
	Title      string                 `json:"title"`
	Properties map[string]interface{} `json:"properties"`
}

// NotionDatabase database
type NotionDatabase struct {
	ID    string `json:"id"`
	Title []struct {
		Text struct {
			Content string `json:"content"`
		} `json:"text"`
	} `json:"title"`
}

// NewNotionServices cria cliente Notion
func NewNotionServices(apiKey string) *NotionServices {
	return &NotionServices{
		apiKey:  apiKey,
		client:  &http.Client{Timeout: 30 * time.Second},
		baseURL: "https://api.notion.com/v1",
		version: "2022-06-28",
	}
}

// request faz requisição
func (n *NotionServices) request(method, endpoint string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, _ := http.NewRequest(method, n.baseURL+endpoint, reqBody)
	req.Header.Set("Authorization", "Bearer "+n.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Notion-Version", n.version)

	resp, err := n.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// SearchPages busca páginas
func (n *NotionServices) SearchPages(query string) ([]*NotionPage, error) {
	body := map[string]interface{}{
		"query": query,
		"filter": map[string]string{
			"property": "object",
			"value":    "page",
		},
	}

	data, err := n.request("POST", "/search", body)
	if err != nil {
		return nil, err
	}

	var result struct {
		Results []*NotionPage `json:"results"`
	}
	json.Unmarshal(data, &result)

	return result.Results, nil
}

// CreatePage cria página
func (n *NotionServices) CreatePage(parentID, title, content string) (*NotionPage, error) {
	body := map[string]interface{}{
		"parent": map[string]string{
			"page_id": parentID,
		},
		"properties": map[string]interface{}{
			"title": map[string]interface{}{
				"title": []map[string]interface{}{
					{"text": map[string]string{"content": title}},
				},
			},
		},
		"children": []map[string]interface{}{
			{
				"object": "block",
				"type":   "paragraph",
				"paragraph": map[string]interface{}{
					"rich_text": []map[string]interface{}{
						{"type": "text", "text": map[string]string{"content": content}},
					},
				},
			},
		},
	}

	data, err := n.request("POST", "/pages", body)
	if err != nil {
		return nil, err
	}

	var page NotionPage
	json.Unmarshal(data, &page)

	return &page, nil
}

// QueryDatabase consulta database
func (n *NotionServices) QueryDatabase(databaseID string, filter map[string]interface{}) ([]map[string]interface{}, error) {
	data, err := n.request("POST", "/databases/"+databaseID+"/query", filter)
	if err != nil {
		return nil, err
	}

	var result struct {
		Results []map[string]interface{} `json:"results"`
	}
	json.Unmarshal(data, &result)

	return result.Results, nil
}

// ==================== TODOIST ====================

// TodoistServices integração com Todoist
type TodoistServices struct {
	apiKey  string
	client  *http.Client
	baseURL string
}

// TodoistTask tarefa
type TodoistTask struct {
	ID          string `json:"id"`
	Content     string `json:"content"`
	Description string `json:"description"`
	ProjectID   string `json:"project_id"`
	Priority    int    `json:"priority"`
	Due         *struct {
		Date string `json:"date"`
	} `json:"due,omitempty"`
	Completed bool `json:"is_completed"`
}

// TodoistProject projeto
type TodoistProject struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// NewTodoistServices cria cliente Todoist
func NewTodoistServices(apiKey string) *TodoistServices {
	return &TodoistServices{
		apiKey:  apiKey,
		client:  &http.Client{Timeout: 30 * time.Second},
		baseURL: "https://api.todoist.com/rest/v2",
	}
}

// request faz requisição
func (t *TodoistServices) request(method, endpoint string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, _ := http.NewRequest(method, t.baseURL+endpoint, reqBody)
	req.Header.Set("Authorization", "Bearer "+t.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// GetTasks lista tarefas
func (t *TodoistServices) GetTasks() ([]*TodoistTask, error) {
	data, err := t.request("GET", "/tasks", nil)
	if err != nil {
		return nil, err
	}

	var tasks []*TodoistTask
	json.Unmarshal(data, &tasks)

	return tasks, nil
}

// CreateTask cria tarefa
func (t *TodoistServices) CreateTask(content string, projectID string, dueDate string, priority int) (*TodoistTask, error) {
	body := map[string]interface{}{
		"content":  content,
		"priority": priority,
	}
	if projectID != "" {
		body["project_id"] = projectID
	}
	if dueDate != "" {
		body["due_date"] = dueDate
	}

	data, err := t.request("POST", "/tasks", body)
	if err != nil {
		return nil, err
	}

	var task TodoistTask
	json.Unmarshal(data, &task)

	return &task, nil
}

// CompleteTask completa tarefa
func (t *TodoistServices) CompleteTask(taskID string) error {
	_, err := t.request("POST", "/tasks/"+taskID+"/close", nil)
	return err
}

// GetProjects lista projetos
func (t *TodoistServices) GetProjects() ([]*TodoistProject, error) {
	data, err := t.request("GET", "/projects", nil)
	if err != nil {
		return nil, err
	}

	var projects []*TodoistProject
	json.Unmarshal(data, &projects)

	return projects, nil
}

// ==================== SPOTIFY ====================

// SpotifyServices integração com Spotify
type SpotifyServices struct {
	accessToken string
	client      *http.Client
	baseURL     string
}

// SpotifyTrack música
type SpotifyTrack struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Artists []struct {
		Name string `json:"name"`
	} `json:"artists"`
	Album struct {
		Name string `json:"name"`
	} `json:"album"`
	URI string `json:"uri"`
}

// NewSpotifyServices cria cliente Spotify
func NewSpotifyServices(accessToken string) *SpotifyServices {
	return &SpotifyServices{
		accessToken: accessToken,
		client:      &http.Client{Timeout: 30 * time.Second},
		baseURL:     "https://api.spotify.com/v1",
	}
}

// request faz requisição
func (s *SpotifyServices) request(method, endpoint string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, _ := http.NewRequest(method, s.baseURL+endpoint, reqBody)
	req.Header.Set("Authorization", "Bearer "+s.accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// GetCurrentlyPlaying música atual
func (s *SpotifyServices) GetCurrentlyPlaying() (*SpotifyTrack, error) {
	data, err := s.request("GET", "/me/player/currently-playing", nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Item *SpotifyTrack `json:"item"`
	}
	json.Unmarshal(data, &result)

	return result.Item, nil
}

// Play inicia reprodução
func (s *SpotifyServices) Play() error {
	_, err := s.request("PUT", "/me/player/play", nil)
	return err
}

// Pause pausa reprodução
func (s *SpotifyServices) Pause() error {
	_, err := s.request("PUT", "/me/player/pause", nil)
	return err
}

// Next próxima música
func (s *SpotifyServices) Next() error {
	_, err := s.request("POST", "/me/player/next", nil)
	return err
}

// Previous música anterior
func (s *SpotifyServices) Previous() error {
	_, err := s.request("POST", "/me/player/previous", nil)
	return err
}

// Search busca músicas
func (s *SpotifyServices) Search(query string, limit int) ([]*SpotifyTrack, error) {
	if limit == 0 {
		limit = 10
	}
	endpoint := fmt.Sprintf("/search?q=%s&type=track&limit=%d", query, limit)
	data, err := s.request("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Tracks struct {
			Items []*SpotifyTrack `json:"items"`
		} `json:"tracks"`
	}
	json.Unmarshal(data, &result)

	return result.Tracks.Items, nil
}

// PlayTrack toca música específica
func (s *SpotifyServices) PlayTrack(uri string) error {
	body := map[string]interface{}{
		"uris": []string{uri},
	}
	_, err := s.request("PUT", "/me/player/play", body)
	return err
}

// SetVolume ajusta volume
func (s *SpotifyServices) SetVolume(percent int) error {
	endpoint := fmt.Sprintf("/me/player/volume?volume_percent=%d", percent)
	_, err := s.request("PUT", endpoint, nil)
	return err
}
