package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// GitHubServices integração com GitHub
type GitHubServices struct {
	token   string
	client  *http.Client
	baseURL string
	user    string
}

// Repository repositório
type Repository struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	Description string `json:"description"`
	Private     bool   `json:"private"`
	HTMLURL     string `json:"html_url"`
	CloneURL    string `json:"clone_url"`
	Stars       int    `json:"stargazers_count"`
	Forks       int    `json:"forks_count"`
	Language    string `json:"language"`
}

// Issue issue
type Issue struct {
	ID        int64  `json:"id"`
	Number    int    `json:"number"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	State     string `json:"state"`
	HTMLURL   string `json:"html_url"`
	CreatedAt string `json:"created_at"`
	User      struct {
		Login string `json:"login"`
	} `json:"user"`
}

// PullRequest PR
type PullRequest struct {
	ID        int64  `json:"id"`
	Number    int    `json:"number"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	State     string `json:"state"`
	HTMLURL   string `json:"html_url"`
	Head      struct {
		Ref string `json:"ref"`
	} `json:"head"`
	Base struct {
		Ref string `json:"ref"`
	} `json:"base"`
	Mergeable bool `json:"mergeable"`
}

// Notification notificação
type Notification struct {
	ID         string `json:"id"`
	Unread     bool   `json:"unread"`
	Reason     string `json:"reason"`
	Subject    struct {
		Title string `json:"title"`
		Type  string `json:"type"`
	} `json:"subject"`
	Repository struct {
		FullName string `json:"full_name"`
	} `json:"repository"`
}

// Gist gist
type Gist struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Public      bool   `json:"public"`
	HTMLURL     string `json:"html_url"`
	Files       map[string]struct {
		Filename string `json:"filename"`
		Content  string `json:"content"`
	} `json:"files"`
}

// NewGitHubServices cria cliente GitHub
func NewGitHubServices(token string) (*GitHubServices, error) {
	gh := &GitHubServices{
		token:   token,
		client:  &http.Client{Timeout: 30 * time.Second},
		baseURL: "https://api.github.com",
	}

	// Obtém usuário autenticado
	user, err := gh.GetAuthenticatedUser()
	if err != nil {
		return nil, err
	}
	gh.user = user

	return gh, nil
}

// request faz requisição à API
func (gh *GitHubServices) request(method, endpoint string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, gh.baseURL+endpoint, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+gh.token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := gh.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// GetAuthenticatedUser retorna usuário autenticado
func (gh *GitHubServices) GetAuthenticatedUser() (string, error) {
	data, err := gh.request("GET", "/user", nil)
	if err != nil {
		return "", err
	}

	var user struct {
		Login string `json:"login"`
	}
	json.Unmarshal(data, &user)

	return user.Login, nil
}

// ==================== REPOS ====================

// ListRepos lista repositórios
func (gh *GitHubServices) ListRepos() ([]*Repository, error) {
	data, err := gh.request("GET", "/user/repos?sort=updated&per_page=30", nil)
	if err != nil {
		return nil, err
	}

	var repos []*Repository
	json.Unmarshal(data, &repos)

	return repos, nil
}

// GetRepo obtém repositório
func (gh *GitHubServices) GetRepo(owner, repo string) (*Repository, error) {
	endpoint := fmt.Sprintf("/repos/%s/%s", owner, repo)
	data, err := gh.request("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var repository Repository
	json.Unmarshal(data, &repository)

	return &repository, nil
}

// CreateRepo cria repositório
func (gh *GitHubServices) CreateRepo(name, description string, private bool) (*Repository, error) {
	body := map[string]interface{}{
		"name":        name,
		"description": description,
		"private":     private,
		"auto_init":   true,
	}

	data, err := gh.request("POST", "/user/repos", body)
	if err != nil {
		return nil, err
	}

	var repo Repository
	json.Unmarshal(data, &repo)

	return &repo, nil
}

// ==================== ISSUES ====================

// ListIssues lista issues
func (gh *GitHubServices) ListIssues(owner, repo string) ([]*Issue, error) {
	endpoint := fmt.Sprintf("/repos/%s/%s/issues?state=open", owner, repo)
	data, err := gh.request("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var issues []*Issue
	json.Unmarshal(data, &issues)

	return issues, nil
}

// CreateIssue cria issue
func (gh *GitHubServices) CreateIssue(owner, repo, title, body string, labels []string) (*Issue, error) {
	endpoint := fmt.Sprintf("/repos/%s/%s/issues", owner, repo)
	reqBody := map[string]interface{}{
		"title":  title,
		"body":   body,
		"labels": labels,
	}

	data, err := gh.request("POST", endpoint, reqBody)
	if err != nil {
		return nil, err
	}

	var issue Issue
	json.Unmarshal(data, &issue)

	return &issue, nil
}

// CloseIssue fecha issue
func (gh *GitHubServices) CloseIssue(owner, repo string, number int) error {
	endpoint := fmt.Sprintf("/repos/%s/%s/issues/%d", owner, repo, number)
	body := map[string]string{"state": "closed"}
	_, err := gh.request("PATCH", endpoint, body)
	return err
}

// ==================== PULL REQUESTS ====================

// ListPullRequests lista PRs
func (gh *GitHubServices) ListPullRequests(owner, repo string) ([]*PullRequest, error) {
	endpoint := fmt.Sprintf("/repos/%s/%s/pulls?state=open", owner, repo)
	data, err := gh.request("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var prs []*PullRequest
	json.Unmarshal(data, &prs)

	return prs, nil
}

// CreatePullRequest cria PR
func (gh *GitHubServices) CreatePullRequest(owner, repo, title, body, head, base string) (*PullRequest, error) {
	endpoint := fmt.Sprintf("/repos/%s/%s/pulls", owner, repo)
	reqBody := map[string]string{
		"title": title,
		"body":  body,
		"head":  head,
		"base":  base,
	}

	data, err := gh.request("POST", endpoint, reqBody)
	if err != nil {
		return nil, err
	}

	var pr PullRequest
	json.Unmarshal(data, &pr)

	return &pr, nil
}

// MergePullRequest faz merge do PR
func (gh *GitHubServices) MergePullRequest(owner, repo string, number int) error {
	endpoint := fmt.Sprintf("/repos/%s/%s/pulls/%d/merge", owner, repo, number)
	_, err := gh.request("PUT", endpoint, map[string]string{})
	return err
}

// ==================== NOTIFICATIONS ====================

// ListNotifications lista notificações
func (gh *GitHubServices) ListNotifications() ([]*Notification, error) {
	data, err := gh.request("GET", "/notifications", nil)
	if err != nil {
		return nil, err
	}

	var notifications []*Notification
	json.Unmarshal(data, &notifications)

	return notifications, nil
}

// MarkNotificationsRead marca todas como lidas
func (gh *GitHubServices) MarkNotificationsRead() error {
	_, err := gh.request("PUT", "/notifications", map[string]string{})
	return err
}

// ==================== GISTS ====================

// ListGists lista gists
func (gh *GitHubServices) ListGists() ([]*Gist, error) {
	data, err := gh.request("GET", "/gists", nil)
	if err != nil {
		return nil, err
	}

	var gists []*Gist
	json.Unmarshal(data, &gists)

	return gists, nil
}

// CreateGist cria gist
func (gh *GitHubServices) CreateGist(description, filename, content string, public bool) (*Gist, error) {
	body := map[string]interface{}{
		"description": description,
		"public":      public,
		"files": map[string]interface{}{
			filename: map[string]string{
				"content": content,
			},
		},
	}

	data, err := gh.request("POST", "/gists", body)
	if err != nil {
		return nil, err
	}

	var gist Gist
	json.Unmarshal(data, &gist)

	return &gist, nil
}

// ==================== SEARCH ====================

// SearchCode busca código
func (gh *GitHubServices) SearchCode(query string) ([]map[string]interface{}, error) {
	endpoint := fmt.Sprintf("/search/code?q=%s", query)
	data, err := gh.request("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Items []map[string]interface{} `json:"items"`
	}
	json.Unmarshal(data, &result)

	return result.Items, nil
}

// SearchRepos busca repositórios
func (gh *GitHubServices) SearchRepos(query string) ([]*Repository, error) {
	endpoint := fmt.Sprintf("/search/repositories?q=%s&sort=stars", query)
	data, err := gh.request("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Items []*Repository `json:"items"`
	}
	json.Unmarshal(data, &result)

	return result.Items, nil
}
