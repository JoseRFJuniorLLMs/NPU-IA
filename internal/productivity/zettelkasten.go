package productivity

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

// Zettelkasten sistema de notas Zettelkasten
type Zettelkasten struct {
	basePath string
	notes    map[string]*Note
	index    *NoteIndex
	llm      LLMInterface
	mu       sync.RWMutex
}

// LLMInterface interface para o modelo de linguagem
type LLMInterface interface {
	Generate(ctx context.Context, prompt string) (string, error)
}

// Note uma nota/zettel
type Note struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	Tags        []string  `json:"tags"`
	Links       []string  `json:"links"`       // IDs de notas linkadas
	Backlinks   []string  `json:"backlinks"`   // Notas que linkam para esta
	Created     time.Time `json:"created"`
	Modified    time.Time `json:"modified"`
	Type        NoteType  `json:"type"`
	Source      string    `json:"source"`      // Fonte original (livro, artigo, etc)
	IsFleet     bool      `json:"is_fleet"`    // Nota temporária/fleeting
	IsPermanent bool      `json:"is_permanent"` // Nota permanente/elaborada
}

// NoteType tipo de nota
type NoteType string

const (
	NoteTypeFleet     NoteType = "fleet"     // Captura rápida
	NoteTypeLiterature NoteType = "literature" // Nota de leitura
	NoteTypePermanent NoteType = "permanent" // Nota elaborada
	NoteTypeIndex     NoteType = "index"     // Nota índice/MOC
	NoteTypeProject   NoteType = "project"   // Nota de projeto
)

// NoteIndex índice de notas para busca rápida
type NoteIndex struct {
	ByTag      map[string][]string // tag -> note IDs
	ByLink     map[string][]string // linked note ID -> linking note IDs
	ByDate     map[string][]string // YYYY-MM-DD -> note IDs
	ByType     map[NoteType][]string
	FullText   map[string][]string // word -> note IDs
}

// NewZettelkasten cria novo sistema Zettelkasten
func NewZettelkasten(basePath string, llm LLMInterface) (*Zettelkasten, error) {
	z := &Zettelkasten{
		basePath: basePath,
		notes:    make(map[string]*Note),
		index: &NoteIndex{
			ByTag:    make(map[string][]string),
			ByLink:   make(map[string][]string),
			ByDate:   make(map[string][]string),
			ByType:   make(map[NoteType][]string),
			FullText: make(map[string][]string),
		},
		llm: llm,
	}

	// Cria diretório se não existe
	os.MkdirAll(basePath, 0755)

	// Carrega notas existentes
	if err := z.loadNotes(); err != nil {
		return nil, err
	}

	return z, nil
}

// ==================== CRIAÇÃO DE NOTAS ====================

// QuickCapture captura rápida (fleeting note)
func (z *Zettelkasten) QuickCapture(content string, tags ...string) (*Note, error) {
	note := &Note{
		ID:       z.generateID(),
		Content:  content,
		Tags:     tags,
		Created:  time.Now(),
		Modified: time.Now(),
		Type:     NoteTypeFleet,
		IsFleet:  true,
	}

	// Extrai título das primeiras palavras
	words := strings.Fields(content)
	if len(words) > 5 {
		note.Title = strings.Join(words[:5], " ") + "..."
	} else {
		note.Title = content
	}

	return z.saveNote(note)
}

// CreateNote cria nota elaborada
func (z *Zettelkasten) CreateNote(title, content string, tags []string, noteType NoteType) (*Note, error) {
	note := &Note{
		ID:          z.generateID(),
		Title:       title,
		Content:     content,
		Tags:        tags,
		Created:     time.Now(),
		Modified:    time.Now(),
		Type:        noteType,
		IsPermanent: noteType == NoteTypePermanent,
	}

	// Extrai links [[...]] do conteúdo
	note.Links = z.extractLinks(content)

	return z.saveNote(note)
}

// CreateLiteratureNote cria nota de leitura
func (z *Zettelkasten) CreateLiteratureNote(title, content, source string, tags []string) (*Note, error) {
	note := &Note{
		ID:       z.generateID(),
		Title:    title,
		Content:  content,
		Source:   source,
		Tags:     tags,
		Created:  time.Now(),
		Modified: time.Now(),
		Type:     NoteTypeLiterature,
	}

	note.Links = z.extractLinks(content)

	return z.saveNote(note)
}

// ==================== AI-POWERED ====================

// SmartCapture captura inteligente com AI
func (z *Zettelkasten) SmartCapture(ctx context.Context, content string) (*Note, error) {
	prompt := fmt.Sprintf(`Analise este conteúdo e extraia:
1. Um título conciso
2. Tags relevantes (max 5)
3. Possíveis conexões com conceitos

Conteúdo: %s

Responda em JSON:
{
  "title": "...",
  "tags": ["tag1", "tag2"],
  "concepts": ["conceito1", "conceito2"]
}`, content)

	response, err := z.llm.Generate(ctx, prompt)
	if err != nil {
		// Fallback para captura simples
		return z.QuickCapture(content)
	}

	var analysis struct {
		Title    string   `json:"title"`
		Tags     []string `json:"tags"`
		Concepts []string `json:"concepts"`
	}
	json.Unmarshal([]byte(response), &analysis)

	note := &Note{
		ID:       z.generateID(),
		Title:    analysis.Title,
		Content:  content,
		Tags:     analysis.Tags,
		Created:  time.Now(),
		Modified: time.Now(),
		Type:     NoteTypeFleet,
		IsFleet:  true,
	}

	// Tenta linkar com notas existentes
	for _, concept := range analysis.Concepts {
		existingNotes := z.SearchByContent(concept)
		for _, existing := range existingNotes {
			if existing.ID != note.ID {
				note.Links = append(note.Links, existing.ID)
			}
		}
	}

	return z.saveNote(note)
}

// ElaborateNote transforma nota fleeting em permanente
func (z *Zettelkasten) ElaborateNote(ctx context.Context, noteID string) (*Note, error) {
	z.mu.RLock()
	note, exists := z.notes[noteID]
	z.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("nota não encontrada: %s", noteID)
	}

	prompt := fmt.Sprintf(`Você é um assistente de escrita no estilo Zettelkasten.
Elabore esta nota rápida em uma nota permanente:
- Expanda as ideias principais
- Adicione contexto e explicações
- Mantenha uma ideia por nota
- Use linguagem clara e atemporal

Nota original:
%s

Nota elaborada:`, note.Content)

	elaborated, err := z.llm.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	note.Content = elaborated
	note.Type = NoteTypePermanent
	note.IsFleet = false
	note.IsPermanent = true
	note.Modified = time.Now()

	return z.saveNote(note)
}

// SuggestConnections sugere conexões para uma nota
func (z *Zettelkasten) SuggestConnections(ctx context.Context, noteID string) ([]string, error) {
	z.mu.RLock()
	note, exists := z.notes[noteID]
	z.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("nota não encontrada")
	}

	// Coleta notas potencialmente relacionadas
	candidates := make([]*Note, 0)

	// Por tags
	for _, tag := range note.Tags {
		for _, id := range z.index.ByTag[tag] {
			if id != noteID {
				if n, ok := z.notes[id]; ok {
					candidates = append(candidates, n)
				}
			}
		}
	}

	if len(candidates) == 0 {
		return nil, nil
	}

	// Usa LLM para ranquear
	candidateTexts := ""
	for i, c := range candidates[:min(10, len(candidates))] {
		candidateTexts += fmt.Sprintf("%d. [%s] %s\n", i+1, c.ID, c.Title)
	}

	prompt := fmt.Sprintf(`Nota atual:
Título: %s
Conteúdo: %s

Notas candidatas:
%s

Quais notas têm conexão significativa com a nota atual?
Liste os números das 3 mais relevantes:`, note.Title, note.Content, candidateTexts)

	response, err := z.llm.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	// Extrai IDs sugeridos
	suggestions := make([]string, 0)
	re := regexp.MustCompile(`\d+`)
	matches := re.FindAllString(response, -1)

	for _, match := range matches {
		idx := 0
		fmt.Sscanf(match, "%d", &idx)
		if idx > 0 && idx <= len(candidates) {
			suggestions = append(suggestions, candidates[idx-1].ID)
		}
	}

	return suggestions, nil
}

// GenerateMOC gera um Map of Content
func (z *Zettelkasten) GenerateMOC(ctx context.Context, topic string) (*Note, error) {
	// Busca notas relacionadas ao tópico
	relatedNotes := z.SearchByContent(topic)
	taggedNotes := z.GetByTag(topic)

	allNotes := make(map[string]*Note)
	for _, n := range relatedNotes {
		allNotes[n.ID] = n
	}
	for _, n := range taggedNotes {
		allNotes[n.ID] = n
	}

	if len(allNotes) == 0 {
		return nil, fmt.Errorf("nenhuma nota encontrada sobre: %s", topic)
	}

	// Lista notas para LLM
	noteList := ""
	for _, n := range allNotes {
		noteList += fmt.Sprintf("- [[%s]] %s\n", n.ID, n.Title)
	}

	prompt := fmt.Sprintf(`Crie um Map of Content (índice estruturado) sobre "%s".
Organize as notas em categorias lógicas.

Notas disponíveis:
%s

Crie o MOC em Markdown com:
- Título
- Breve introdução
- Notas organizadas por subtópicos
- Links para as notas usando [[ID]]`, topic, noteList)

	mocContent, err := z.llm.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return z.CreateNote(
		fmt.Sprintf("MOC: %s", topic),
		mocContent,
		[]string{"moc", topic},
		NoteTypeIndex,
	)
}

// ==================== BUSCA ====================

// Search busca notas
func (z *Zettelkasten) Search(query string) []*Note {
	z.mu.RLock()
	defer z.mu.RUnlock()

	results := make([]*Note, 0)
	queryLower := strings.ToLower(query)

	for _, note := range z.notes {
		// Busca em título
		if strings.Contains(strings.ToLower(note.Title), queryLower) {
			results = append(results, note)
			continue
		}
		// Busca em conteúdo
		if strings.Contains(strings.ToLower(note.Content), queryLower) {
			results = append(results, note)
			continue
		}
		// Busca em tags
		for _, tag := range note.Tags {
			if strings.Contains(strings.ToLower(tag), queryLower) {
				results = append(results, note)
				break
			}
		}
	}

	// Ordena por relevância (mais recentes primeiro)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Modified.After(results[j].Modified)
	})

	return results
}

// SearchByContent busca por conteúdo
func (z *Zettelkasten) SearchByContent(query string) []*Note {
	return z.Search(query)
}

// GetByTag busca por tag
func (z *Zettelkasten) GetByTag(tag string) []*Note {
	z.mu.RLock()
	defer z.mu.RUnlock()

	ids := z.index.ByTag[strings.ToLower(tag)]
	notes := make([]*Note, 0, len(ids))
	for _, id := range ids {
		if note, ok := z.notes[id]; ok {
			notes = append(notes, note)
		}
	}
	return notes
}

// GetBacklinks retorna notas que linkam para uma nota
func (z *Zettelkasten) GetBacklinks(noteID string) []*Note {
	z.mu.RLock()
	defer z.mu.RUnlock()

	note, exists := z.notes[noteID]
	if !exists {
		return nil
	}

	notes := make([]*Note, 0, len(note.Backlinks))
	for _, id := range note.Backlinks {
		if n, ok := z.notes[id]; ok {
			notes = append(notes, n)
		}
	}
	return notes
}

// GetFleetingNotes retorna notas para processar
func (z *Zettelkasten) GetFleetingNotes() []*Note {
	z.mu.RLock()
	defer z.mu.RUnlock()

	notes := make([]*Note, 0)
	for _, note := range z.notes {
		if note.IsFleet {
			notes = append(notes, note)
		}
	}

	sort.Slice(notes, func(i, j int) bool {
		return notes[i].Created.Before(notes[j].Created)
	})

	return notes
}

// GetDailyNote retorna nota diária
func (z *Zettelkasten) GetDailyNote(date time.Time) (*Note, error) {
	dateStr := date.Format("2006-01-02")

	z.mu.RLock()
	ids := z.index.ByDate[dateStr]
	z.mu.RUnlock()

	for _, id := range ids {
		if note, ok := z.notes[id]; ok {
			if note.Type == NoteTypeIndex && strings.HasPrefix(note.Title, "Daily:") {
				return note, nil
			}
		}
	}

	// Cria nova nota diária
	return z.CreateNote(
		fmt.Sprintf("Daily: %s", dateStr),
		fmt.Sprintf("# %s\n\n## Tarefas\n\n- [ ] \n\n## Notas\n\n## Reflexões\n\n",
			date.Format("Monday, 02 January 2006")),
		[]string{"daily", dateStr},
		NoteTypeIndex,
	)
}

// ==================== UTILITÁRIOS ====================

// generateID gera ID único
func (z *Zettelkasten) generateID() string {
	return time.Now().Format("20060102150405")
}

// extractLinks extrai links [[...]] do conteúdo
func (z *Zettelkasten) extractLinks(content string) []string {
	re := regexp.MustCompile(`\[\[([^\]]+)\]\]`)
	matches := re.FindAllStringSubmatch(content, -1)

	links := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) > 1 {
			links = append(links, match[1])
		}
	}
	return links
}

// saveNote salva nota
func (z *Zettelkasten) saveNote(note *Note) (*Note, error) {
	z.mu.Lock()
	defer z.mu.Unlock()

	z.notes[note.ID] = note

	// Atualiza índices
	for _, tag := range note.Tags {
		tagLower := strings.ToLower(tag)
		z.index.ByTag[tagLower] = append(z.index.ByTag[tagLower], note.ID)
	}

	dateStr := note.Created.Format("2006-01-02")
	z.index.ByDate[dateStr] = append(z.index.ByDate[dateStr], note.ID)
	z.index.ByType[note.Type] = append(z.index.ByType[note.Type], note.ID)

	// Atualiza backlinks
	for _, linkedID := range note.Links {
		if linked, ok := z.notes[linkedID]; ok {
			linked.Backlinks = append(linked.Backlinks, note.ID)
		}
	}

	// Salva em arquivo
	return note, z.saveToFile(note)
}

// saveToFile salva nota em arquivo
func (z *Zettelkasten) saveToFile(note *Note) error {
	filename := filepath.Join(z.basePath, note.ID+".md")

	// Formato markdown com frontmatter
	content := fmt.Sprintf(`---
id: %s
title: %s
tags: [%s]
created: %s
modified: %s
type: %s
---

# %s

%s
`,
		note.ID,
		note.Title,
		strings.Join(note.Tags, ", "),
		note.Created.Format(time.RFC3339),
		note.Modified.Format(time.RFC3339),
		note.Type,
		note.Title,
		note.Content,
	)

	return os.WriteFile(filename, []byte(content), 0644)
}

// loadNotes carrega notas do disco
func (z *Zettelkasten) loadNotes() error {
	files, err := filepath.Glob(filepath.Join(z.basePath, "*.md"))
	if err != nil {
		return err
	}

	for _, file := range files {
		// TODO: Parse frontmatter e carregar notas
		_ = file
	}

	return nil
}

// GetStats retorna estatísticas
func (z *Zettelkasten) GetStats() map[string]interface{} {
	z.mu.RLock()
	defer z.mu.RUnlock()

	fleetCount := 0
	permanentCount := 0
	for _, note := range z.notes {
		if note.IsFleet {
			fleetCount++
		}
		if note.IsPermanent {
			permanentCount++
		}
	}

	return map[string]interface{}{
		"total_notes":     len(z.notes),
		"fleeting_notes":  fleetCount,
		"permanent_notes": permanentCount,
		"total_tags":      len(z.index.ByTag),
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
