package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// ==================== LINKEDIN ====================

// LinkedInServices integração com LinkedIn
type LinkedInServices struct {
	accessToken string
	client      *http.Client
	baseURL     string
}

// LinkedInPost post do LinkedIn
type LinkedInPost struct {
	ID      string `json:"id"`
	Text    string `json:"text"`
	Created int64  `json:"created>time"`
}

// LinkedInMessage mensagem
type LinkedInMessage struct {
	ID      string `json:"id"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

// NewLinkedInServices cria cliente LinkedIn
func NewLinkedInServices(accessToken string) *LinkedInServices {
	return &LinkedInServices{
		accessToken: accessToken,
		client:      &http.Client{Timeout: 30 * time.Second},
		baseURL:     "https://api.linkedin.com/v2",
	}
}

// GetProfile obtém perfil
func (li *LinkedInServices) GetProfile() (map[string]interface{}, error) {
	req, _ := http.NewRequest("GET", li.baseURL+"/me", nil)
	req.Header.Set("Authorization", "Bearer "+li.accessToken)

	resp, err := li.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var profile map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&profile)

	return profile, nil
}

// SharePost publica post
func (li *LinkedInServices) SharePost(text string) error {
	// Primeiro obtém o URN do usuário
	profile, err := li.GetProfile()
	if err != nil {
		return err
	}
	userURN := fmt.Sprintf("urn:li:person:%s", profile["id"])

	body := map[string]interface{}{
		"author":         userURN,
		"lifecycleState": "PUBLISHED",
		"specificContent": map[string]interface{}{
			"com.linkedin.ugc.ShareContent": map[string]interface{}{
				"shareCommentary": map[string]string{
					"text": text,
				},
				"shareMediaCategory": "NONE",
			},
		},
		"visibility": map[string]string{
			"com.linkedin.ugc.MemberNetworkVisibility": "PUBLIC",
		},
	}

	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", li.baseURL+"/ugcPosts", bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", "Bearer "+li.accessToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Restli-Protocol-Version", "2.0.0")

	resp, err := li.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("erro ao postar: %s", string(body))
	}

	return nil
}

// GetConnections lista conexões
func (li *LinkedInServices) GetConnections() ([]map[string]interface{}, error) {
	req, _ := http.NewRequest("GET", li.baseURL+"/connections?q=viewer&start=0&count=50", nil)
	req.Header.Set("Authorization", "Bearer "+li.accessToken)

	resp, err := li.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Elements []map[string]interface{} `json:"elements"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	return result.Elements, nil
}

// ==================== X (TWITTER) ====================

// XServices integração com X (Twitter)
type XServices struct {
	bearerToken string
	client      *http.Client
	baseURL     string
}

// Tweet tweet
type Tweet struct {
	ID        string `json:"id"`
	Text      string `json:"text"`
	AuthorID  string `json:"author_id"`
	CreatedAt string `json:"created_at"`
}

// NewXServices cria cliente X
func NewXServices(bearerToken string) *XServices {
	return &XServices{
		bearerToken: bearerToken,
		client:      &http.Client{Timeout: 30 * time.Second},
		baseURL:     "https://api.twitter.com/2",
	}
}

// request faz requisição
func (x *XServices) request(method, endpoint string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, _ := http.NewRequest(method, x.baseURL+endpoint, reqBody)
	req.Header.Set("Authorization", "Bearer "+x.bearerToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := x.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// GetTimeline obtém timeline
func (x *XServices) GetTimeline(userID string, max int) ([]*Tweet, error) {
	if max == 0 {
		max = 10
	}
	endpoint := fmt.Sprintf("/users/%s/tweets?max_results=%d&tweet.fields=created_at", userID, max)
	data, err := x.request("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Data []*Tweet `json:"data"`
	}
	json.Unmarshal(data, &result)

	return result.Data, nil
}

// PostTweet publica tweet
func (x *XServices) PostTweet(text string) (*Tweet, error) {
	body := map[string]string{"text": text}
	data, err := x.request("POST", "/tweets", body)
	if err != nil {
		return nil, err
	}

	var result struct {
		Data *Tweet `json:"data"`
	}
	json.Unmarshal(data, &result)

	return result.Data, nil
}

// SearchTweets busca tweets
func (x *XServices) SearchTweets(query string, max int) ([]*Tweet, error) {
	if max == 0 {
		max = 10
	}
	endpoint := fmt.Sprintf("/tweets/search/recent?query=%s&max_results=%d", url.QueryEscape(query), max)
	data, err := x.request("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Data []*Tweet `json:"data"`
	}
	json.Unmarshal(data, &result)

	return result.Data, nil
}

// GetMe obtém perfil autenticado
func (x *XServices) GetMe() (map[string]interface{}, error) {
	data, err := x.request("GET", "/users/me", nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Data map[string]interface{} `json:"data"`
	}
	json.Unmarshal(data, &result)

	return result.Data, nil
}

// ==================== DISCORD ====================

// DiscordServices integração com Discord
type DiscordServices struct {
	botToken string
	client   *http.Client
	baseURL  string
}

// DiscordMessage mensagem
type DiscordMessage struct {
	ID        string `json:"id"`
	Content   string `json:"content"`
	ChannelID string `json:"channel_id"`
	Author    struct {
		Username string `json:"username"`
	} `json:"author"`
}

// NewDiscordServices cria cliente Discord
func NewDiscordServices(botToken string) *DiscordServices {
	return &DiscordServices{
		botToken: botToken,
		client:   &http.Client{Timeout: 30 * time.Second},
		baseURL:  "https://discord.com/api/v10",
	}
}

// request faz requisição
func (d *DiscordServices) request(method, endpoint string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, _ := http.NewRequest(method, d.baseURL+endpoint, reqBody)
	req.Header.Set("Authorization", "Bot "+d.botToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// SendMessage envia mensagem
func (d *DiscordServices) SendMessage(channelID, content string) (*DiscordMessage, error) {
	endpoint := fmt.Sprintf("/channels/%s/messages", channelID)
	body := map[string]string{"content": content}
	data, err := d.request("POST", endpoint, body)
	if err != nil {
		return nil, err
	}

	var msg DiscordMessage
	json.Unmarshal(data, &msg)

	return &msg, nil
}

// GetMessages obtém mensagens
func (d *DiscordServices) GetMessages(channelID string, limit int) ([]*DiscordMessage, error) {
	if limit == 0 {
		limit = 50
	}
	endpoint := fmt.Sprintf("/channels/%s/messages?limit=%d", channelID, limit)
	data, err := d.request("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var messages []*DiscordMessage
	json.Unmarshal(data, &messages)

	return messages, nil
}

// GetGuilds lista servidores
func (d *DiscordServices) GetGuilds() ([]map[string]interface{}, error) {
	data, err := d.request("GET", "/users/@me/guilds", nil)
	if err != nil {
		return nil, err
	}

	var guilds []map[string]interface{}
	json.Unmarshal(data, &guilds)

	return guilds, nil
}

// ==================== SLACK ====================

// SlackServices integração com Slack
type SlackServices struct {
	botToken string
	client   *http.Client
	baseURL  string
}

// SlackMessage mensagem
type SlackMessage struct {
	Channel string `json:"channel"`
	Text    string `json:"text"`
	TS      string `json:"ts"`
}

// NewSlackServices cria cliente Slack
func NewSlackServices(botToken string) *SlackServices {
	return &SlackServices{
		botToken: botToken,
		client:   &http.Client{Timeout: 30 * time.Second},
		baseURL:  "https://slack.com/api",
	}
}

// SendMessage envia mensagem
func (s *SlackServices) SendMessage(channel, text string) error {
	body := map[string]string{
		"channel": channel,
		"text":    text,
	}

	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", s.baseURL+"/chat.postMessage", bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", "Bearer "+s.botToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// ListChannels lista canais
func (s *SlackServices) ListChannels() ([]map[string]interface{}, error) {
	req, _ := http.NewRequest("GET", s.baseURL+"/conversations.list", nil)
	req.Header.Set("Authorization", "Bearer "+s.botToken)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Channels []map[string]interface{} `json:"channels"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	return result.Channels, nil
}

// ==================== TELEGRAM ====================

// TelegramServices integração com Telegram
type TelegramServices struct {
	botToken string
	client   *http.Client
	baseURL  string
}

// NewTelegramServices cria cliente Telegram
func NewTelegramServices(botToken string) *TelegramServices {
	return &TelegramServices{
		botToken: botToken,
		client:   &http.Client{Timeout: 30 * time.Second},
		baseURL:  fmt.Sprintf("https://api.telegram.org/bot%s", botToken),
	}
}

// SendMessage envia mensagem
func (t *TelegramServices) SendMessage(chatID, text string) error {
	endpoint := fmt.Sprintf("%s/sendMessage?chat_id=%s&text=%s", t.baseURL, chatID, url.QueryEscape(text))
	_, err := t.client.Get(endpoint)
	return err
}

// GetUpdates obtém atualizações
func (t *TelegramServices) GetUpdates() ([]map[string]interface{}, error) {
	resp, err := t.client.Get(t.baseURL + "/getUpdates")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Result []map[string]interface{} `json:"result"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	return result.Result, nil
}
