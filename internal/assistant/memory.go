package assistant

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// ==================== MEMÓRIA PERSISTENTE ====================

// Memory sistema de memória de longo prazo
type Memory struct {
	basePath    string
	preferences *UserPreferences
	facts       map[string]*Fact
	conversations []ConversationSummary
	patterns    map[string]*Pattern
	llm         LLMInterface
	mu          sync.RWMutex
}

// LLMInterface interface para o modelo
type LLMInterface interface {
	Generate(ctx context.Context, prompt string) (string, error)
}

// UserPreferences preferências do usuário
type UserPreferences struct {
	Name              string            `json:"name"`
	Nickname          string            `json:"nickname"`
	Language          string            `json:"language"`
	Timezone          string            `json:"timezone"`
	WakeUpTime        string            `json:"wake_up_time"`
	SleepTime         string            `json:"sleep_time"`
	WorkStartTime     string            `json:"work_start_time"`
	WorkEndTime       string            `json:"work_end_time"`
	PreferredLLM      string            `json:"preferred_llm"`
	VoiceSpeed        float64           `json:"voice_speed"`
	FormalityLevel    string            `json:"formality_level"` // formal, casual, mixed
	Interests         []string          `json:"interests"`
	Skills            []string          `json:"skills"`
	Goals             []string          `json:"goals"`
	DietaryPrefs      []string          `json:"dietary_prefs"`
	CustomPreferences map[string]string `json:"custom_preferences"`
	UpdatedAt         time.Time         `json:"updated_at"`
}

// Fact fato sobre o usuário ou contexto
type Fact struct {
	ID         string    `json:"id"`
	Category   string    `json:"category"` // personal, work, preference, relationship, etc
	Subject    string    `json:"subject"`
	Content    string    `json:"content"`
	Source     string    `json:"source"` // Como aprendeu: conversa, email, inferência
	Confidence float64   `json:"confidence"` // 0-1
	CreatedAt  time.Time `json:"created_at"`
	LastUsed   time.Time `json:"last_used"`
	UseCount   int       `json:"use_count"`
}

// ConversationSummary resumo de conversa
type ConversationSummary struct {
	ID        string    `json:"id"`
	Date      time.Time `json:"date"`
	Topics    []string  `json:"topics"`
	KeyPoints []string  `json:"key_points"`
	Decisions []string  `json:"decisions"`
	FollowUps []string  `json:"follow_ups"`
	Sentiment string    `json:"sentiment"`
}

// Pattern padrão detectado
type Pattern struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"` // time, behavior, preference
	Description string    `json:"description"`
	Occurrences int       `json:"occurrences"`
	LastSeen    time.Time `json:"last_seen"`
	Confidence  float64   `json:"confidence"`
}

// NewMemory cria sistema de memória
func NewMemory(basePath string, llm LLMInterface) (*Memory, error) {
	m := &Memory{
		basePath:    basePath,
		facts:       make(map[string]*Fact),
		patterns:    make(map[string]*Pattern),
		llm:         llm,
		preferences: &UserPreferences{
			Language:          "pt-BR",
			Timezone:          "America/Sao_Paulo",
			FormalityLevel:    "casual",
			CustomPreferences: make(map[string]string),
		},
	}

	// Cria diretório
	os.MkdirAll(basePath, 0755)

	// Carrega memória existente
	m.load()

	return m, nil
}

// ==================== PREFERÊNCIAS ====================

// SetPreference define preferência
func (m *Memory) SetPreference(key, value string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	switch key {
	case "name":
		m.preferences.Name = value
	case "nickname":
		m.preferences.Nickname = value
	case "wake_time":
		m.preferences.WakeUpTime = value
	case "sleep_time":
		m.preferences.SleepTime = value
	default:
		m.preferences.CustomPreferences[key] = value
	}
	m.preferences.UpdatedAt = time.Now()
	m.save()
}

// GetPreference obtém preferência
func (m *Memory) GetPreference(key string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	switch key {
	case "name":
		return m.preferences.Name
	case "nickname":
		return m.preferences.Nickname
	default:
		return m.preferences.CustomPreferences[key]
	}
}

// GetUserName retorna nome do usuário
func (m *Memory) GetUserName() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.preferences.Nickname != "" {
		return m.preferences.Nickname
	}
	if m.preferences.Name != "" {
		return m.preferences.Name
	}
	return "você"
}

// ==================== FATOS ====================

// LearnFact aprende um novo fato
func (m *Memory) LearnFact(category, subject, content, source string, confidence float64) *Fact {
	m.mu.Lock()
	defer m.mu.Unlock()

	id := fmt.Sprintf("fact_%d", time.Now().UnixNano())
	fact := &Fact{
		ID:         id,
		Category:   category,
		Subject:    subject,
		Content:    content,
		Source:     source,
		Confidence: confidence,
		CreatedAt:  time.Now(),
		LastUsed:   time.Now(),
		UseCount:   0,
	}

	m.facts[id] = fact
	m.save()

	return fact
}

// RecallFacts busca fatos relevantes
func (m *Memory) RecallFacts(query string, limit int) []*Fact {
	m.mu.RLock()
	defer m.mu.RUnlock()

	queryLower := strings.ToLower(query)
	relevant := make([]*Fact, 0)

	for _, fact := range m.facts {
		// Busca em subject e content
		if strings.Contains(strings.ToLower(fact.Subject), queryLower) ||
			strings.Contains(strings.ToLower(fact.Content), queryLower) ||
			strings.Contains(strings.ToLower(fact.Category), queryLower) {
			relevant = append(relevant, fact)
		}
	}

	// Ordena por relevância (confidence * recency)
	sort.Slice(relevant, func(i, j int) bool {
		scoreI := relevant[i].Confidence * float64(relevant[i].UseCount+1)
		scoreJ := relevant[j].Confidence * float64(relevant[j].UseCount+1)
		return scoreI > scoreJ
	})

	if limit > 0 && len(relevant) > limit {
		return relevant[:limit]
	}
	return relevant
}

// GetFactsByCategory busca por categoria
func (m *Memory) GetFactsByCategory(category string) []*Fact {
	m.mu.RLock()
	defer m.mu.RUnlock()

	facts := make([]*Fact, 0)
	for _, fact := range m.facts {
		if fact.Category == category {
			facts = append(facts, fact)
		}
	}
	return facts
}

// UseFact marca fato como usado
func (m *Memory) UseFact(factID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if fact, ok := m.facts[factID]; ok {
		fact.LastUsed = time.Now()
		fact.UseCount++
	}
}

// ==================== CONVERSAS ====================

// SummarizeConversation resume e armazena conversa
func (m *Memory) SummarizeConversation(ctx context.Context, messages []string) error {
	// Junta mensagens
	conversation := strings.Join(messages, "\n")

	prompt := fmt.Sprintf(`Analise esta conversa e extraia:
1. Tópicos discutidos
2. Pontos-chave
3. Decisões tomadas
4. Itens de follow-up
5. Novos fatos sobre o usuário

Conversa:
%s

Responda em JSON:
{
  "topics": [],
  "key_points": [],
  "decisions": [],
  "follow_ups": [],
  "new_facts": [{"category": "", "subject": "", "content": ""}],
  "sentiment": "positive/neutral/negative"
}`, conversation)

	response, err := m.llm.Generate(ctx, prompt)
	if err != nil {
		return err
	}

	var summary struct {
		Topics    []string `json:"topics"`
		KeyPoints []string `json:"key_points"`
		Decisions []string `json:"decisions"`
		FollowUps []string `json:"follow_ups"`
		NewFacts  []struct {
			Category string `json:"category"`
			Subject  string `json:"subject"`
			Content  string `json:"content"`
		} `json:"new_facts"`
		Sentiment string `json:"sentiment"`
	}
	json.Unmarshal([]byte(response), &summary)

	// Salva resumo
	m.mu.Lock()
	m.conversations = append(m.conversations, ConversationSummary{
		ID:        fmt.Sprintf("conv_%d", time.Now().UnixNano()),
		Date:      time.Now(),
		Topics:    summary.Topics,
		KeyPoints: summary.KeyPoints,
		Decisions: summary.Decisions,
		FollowUps: summary.FollowUps,
		Sentiment: summary.Sentiment,
	})
	m.mu.Unlock()

	// Aprende novos fatos
	for _, fact := range summary.NewFacts {
		m.LearnFact(fact.Category, fact.Subject, fact.Content, "conversa", 0.8)
	}

	m.save()
	return nil
}

// ==================== PADRÕES ====================

// DetectPattern detecta um padrão
func (m *Memory) DetectPattern(patternType, description string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Verifica se padrão já existe
	for _, p := range m.patterns {
		if p.Type == patternType && p.Description == description {
			p.Occurrences++
			p.LastSeen = time.Now()
			p.Confidence = min(1.0, p.Confidence+0.1)
			m.save()
			return
		}
	}

	// Novo padrão
	id := fmt.Sprintf("pattern_%d", time.Now().UnixNano())
	m.patterns[id] = &Pattern{
		ID:          id,
		Type:        patternType,
		Description: description,
		Occurrences: 1,
		LastSeen:    time.Now(),
		Confidence:  0.3,
	}
	m.save()
}

// GetPatterns retorna padrões
func (m *Memory) GetPatterns(patternType string) []*Pattern {
	m.mu.RLock()
	defer m.mu.RUnlock()

	patterns := make([]*Pattern, 0)
	for _, p := range m.patterns {
		if patternType == "" || p.Type == patternType {
			patterns = append(patterns, p)
		}
	}
	return patterns
}

// ==================== CONTEXTO ====================

// GetContext gera contexto para o LLM
func (m *Memory) GetContext(query string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var context strings.Builder

	// Adiciona informações do usuário
	context.WriteString("## Sobre o usuário:\n")
	if m.preferences.Name != "" {
		context.WriteString(fmt.Sprintf("- Nome: %s\n", m.preferences.Name))
	}
	if m.preferences.WakeUpTime != "" {
		context.WriteString(fmt.Sprintf("- Acorda às: %s\n", m.preferences.WakeUpTime))
	}
	if len(m.preferences.Interests) > 0 {
		context.WriteString(fmt.Sprintf("- Interesses: %s\n", strings.Join(m.preferences.Interests, ", ")))
	}

	// Adiciona fatos relevantes
	facts := m.RecallFacts(query, 5)
	if len(facts) > 0 {
		context.WriteString("\n## Fatos relevantes:\n")
		for _, fact := range facts {
			context.WriteString(fmt.Sprintf("- %s: %s\n", fact.Subject, fact.Content))
		}
	}

	// Adiciona padrões
	patterns := m.GetPatterns("")
	highConfPatterns := make([]*Pattern, 0)
	for _, p := range patterns {
		if p.Confidence > 0.7 {
			highConfPatterns = append(highConfPatterns, p)
		}
	}
	if len(highConfPatterns) > 0 {
		context.WriteString("\n## Padrões observados:\n")
		for _, p := range highConfPatterns {
			context.WriteString(fmt.Sprintf("- %s\n", p.Description))
		}
	}

	return context.String()
}

// ==================== PERSISTÊNCIA ====================

// save salva memória em disco
func (m *Memory) save() error {
	data := struct {
		Preferences   *UserPreferences       `json:"preferences"`
		Facts         map[string]*Fact       `json:"facts"`
		Conversations []ConversationSummary  `json:"conversations"`
		Patterns      map[string]*Pattern    `json:"patterns"`
	}{
		Preferences:   m.preferences,
		Facts:         m.facts,
		Conversations: m.conversations,
		Patterns:      m.patterns,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(m.basePath, "memory.json"), jsonData, 0644)
}

// load carrega memória do disco
func (m *Memory) load() error {
	data, err := os.ReadFile(filepath.Join(m.basePath, "memory.json"))
	if err != nil {
		return nil // Arquivo não existe ainda
	}

	var loaded struct {
		Preferences   *UserPreferences       `json:"preferences"`
		Facts         map[string]*Fact       `json:"facts"`
		Conversations []ConversationSummary  `json:"conversations"`
		Patterns      map[string]*Pattern    `json:"patterns"`
	}

	if err := json.Unmarshal(data, &loaded); err != nil {
		return err
	}

	if loaded.Preferences != nil {
		m.preferences = loaded.Preferences
	}
	if loaded.Facts != nil {
		m.facts = loaded.Facts
	}
	if loaded.Conversations != nil {
		m.conversations = loaded.Conversations
	}
	if loaded.Patterns != nil {
		m.patterns = loaded.Patterns
	}

	return nil
}

// GetStats retorna estatísticas da memória
func (m *Memory) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]interface{}{
		"total_facts":         len(m.facts),
		"total_conversations": len(m.conversations),
		"total_patterns":      len(m.patterns),
		"preferences_set":     m.preferences.Name != "",
	}
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
