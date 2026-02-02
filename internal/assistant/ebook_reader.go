package assistant

import (
	"archive/zip"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html"
)

// ==================== EBOOK READER ====================

// EbookReader leitor de ebooks com integração Calibre
type EbookReader struct {
	calibrePath   string // Caminho da biblioteca Calibre
	currentBook   *Book
	books         map[string]*Book
	readingState  map[string]*ReadingProgress
	highlights    map[string][]*Highlight
	llm           LLMInterface
	tts           TTSInterface
	mu            sync.RWMutex
	basePath      string
}

// Book livro
type Book struct {
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	Author      string            `json:"author"`
	Path        string            `json:"path"`
	Format      string            `json:"format"` // epub, pdf, mobi
	CoverPath   string            `json:"cover_path"`
	Language    string            `json:"language"`
	Publisher   string            `json:"publisher"`
	PublishDate string            `json:"publish_date"`
	Description string            `json:"description"`
	Tags        []string          `json:"tags"`
	TotalPages  int               `json:"total_pages"`
	Chapters    []*Chapter        `json:"chapters"`
	Metadata    map[string]string `json:"metadata"`
	AddedAt     time.Time         `json:"added_at"`
}

// Chapter capítulo
type Chapter struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Index   int    `json:"index"`
	Content string `json:"content"`
	Path    string `json:"path"` // Caminho interno no EPUB
}

// ReadingProgress progresso de leitura
type ReadingProgress struct {
	BookID         string    `json:"book_id"`
	CurrentChapter int       `json:"current_chapter"`
	CurrentPage    int       `json:"current_page"`
	Percentage     float64   `json:"percentage"`
	LastRead       time.Time `json:"last_read"`
	TotalReadTime  time.Duration `json:"total_read_time"`
	SessionStart   time.Time `json:"session_start"`
}

// Highlight destaque/anotação
type Highlight struct {
	ID        string    `json:"id"`
	BookID    string    `json:"book_id"`
	ChapterID string    `json:"chapter_id"`
	Text      string    `json:"text"`
	Note      string    `json:"note"`
	Color     string    `json:"color"`
	Page      int       `json:"page"`
	CreatedAt time.Time `json:"created_at"`
}

// BookSummary resumo de livro
type BookSummary struct {
	BookID     string   `json:"book_id"`
	Summary    string   `json:"summary"`
	KeyPoints  []string `json:"key_points"`
	Quotes     []string `json:"quotes"`
	Themes     []string `json:"themes"`
	GeneratedAt time.Time `json:"generated_at"`
}

// NewEbookReader cria leitor de ebooks
func NewEbookReader(basePath, calibrePath string, llm LLMInterface, tts TTSInterface) *EbookReader {
	er := &EbookReader{
		basePath:     basePath,
		calibrePath:  calibrePath,
		books:        make(map[string]*Book),
		readingState: make(map[string]*ReadingProgress),
		highlights:   make(map[string][]*Highlight),
		llm:          llm,
		tts:          tts,
	}

	os.MkdirAll(basePath, 0755)
	er.load()

	return er
}

// ==================== CALIBRE INTEGRATION ====================

// ScanCalibreLibrary escaneia biblioteca do Calibre
func (er *EbookReader) ScanCalibreLibrary() (int, error) {
	if er.calibrePath == "" {
		return 0, fmt.Errorf("caminho do Calibre não configurado")
	}

	count := 0

	// Calibre organiza em pastas: Author/Title/files
	err := filepath.Walk(er.calibrePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".epub" || ext == ".pdf" || ext == ".mobi" {
			book, err := er.parseBookFromPath(path)
			if err == nil {
				er.books[book.ID] = book
				count++
			}
		}

		return nil
	})

	if err != nil {
		return count, err
	}

	er.save()
	return count, nil
}

// parseBookFromPath extrai informações do livro do caminho
func (er *EbookReader) parseBookFromPath(path string) (*Book, error) {
	// Tenta ler metadata.opf se existir (Calibre metadata)
	dir := filepath.Dir(path)
	metadataPath := filepath.Join(dir, "metadata.opf")

	book := &Book{
		ID:       filepath.Base(dir),
		Path:     path,
		Format:   strings.TrimPrefix(filepath.Ext(path), "."),
		AddedAt:  time.Now(),
		Metadata: make(map[string]string),
	}

	// Busca capa
	coverPath := filepath.Join(dir, "cover.jpg")
	if _, err := os.Stat(coverPath); err == nil {
		book.CoverPath = coverPath
	}

	// Tenta ler metadata do Calibre
	if _, err := os.Stat(metadataPath); err == nil {
		er.parseOPFMetadata(metadataPath, book)
	} else {
		// Fallback: extrai do nome do arquivo
		book.Title = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	}

	// Se for EPUB, extrai mais informações
	if book.Format == "epub" {
		er.parseEPUB(book)
	}

	return book, nil
}

// parseOPFMetadata lê metadata.opf do Calibre
func (er *EbookReader) parseOPFMetadata(path string, book *Book) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	type OPFMetadata struct {
		Title       string `xml:"metadata>title"`
		Creator     string `xml:"metadata>creator"`
		Publisher   string `xml:"metadata>publisher"`
		Language    string `xml:"metadata>language"`
		Description string `xml:"metadata>description"`
		Date        string `xml:"metadata>date"`
	}

	var opf OPFMetadata
	if err := xml.Unmarshal(data, &opf); err != nil {
		return err
	}

	book.Title = opf.Title
	book.Author = opf.Creator
	book.Publisher = opf.Publisher
	book.Language = opf.Language
	book.Description = opf.Description
	book.PublishDate = opf.Date

	return nil
}

// ==================== EPUB PARSING ====================

// parseEPUB extrai conteúdo do EPUB
func (er *EbookReader) parseEPUB(book *Book) error {
	reader, err := zip.OpenReader(book.Path)
	if err != nil {
		return err
	}
	defer reader.Close()

	chapters := make([]*Chapter, 0)
	chapterIndex := 0

	for _, file := range reader.File {
		if strings.HasSuffix(file.Name, ".xhtml") || strings.HasSuffix(file.Name, ".html") {
			rc, err := file.Open()
			if err != nil {
				continue
			}

			content, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				continue
			}

			// Extrai texto do HTML
			text := er.extractTextFromHTML(string(content))
			if len(text) < 100 {
				continue // Pula arquivos muito pequenos
			}

			// Extrai título do capítulo
			title := er.extractChapterTitle(string(content))
			if title == "" {
				title = fmt.Sprintf("Capítulo %d", chapterIndex+1)
			}

			chapters = append(chapters, &Chapter{
				ID:      fmt.Sprintf("ch_%d", chapterIndex),
				Title:   title,
				Index:   chapterIndex,
				Content: text,
				Path:    file.Name,
			})
			chapterIndex++
		}
	}

	// Ordena capítulos
	sort.Slice(chapters, func(i, j int) bool {
		return chapters[i].Path < chapters[j].Path
	})

	// Reindica
	for i, ch := range chapters {
		ch.Index = i
	}

	book.Chapters = chapters
	book.TotalPages = len(chapters) // Aproximação

	return nil
}

// extractTextFromHTML extrai texto limpo do HTML
func (er *EbookReader) extractTextFromHTML(htmlContent string) string {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return htmlContent
	}

	var text strings.Builder
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.TextNode {
			text.WriteString(n.Data)
			text.WriteString(" ")
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	// Limpa espaços extras
	result := regexp.MustCompile(`\s+`).ReplaceAllString(text.String(), " ")
	return strings.TrimSpace(result)
}

// extractChapterTitle extrai título do capítulo do HTML
func (er *EbookReader) extractChapterTitle(htmlContent string) string {
	// Tenta encontrar h1, h2, title
	patterns := []string{
		`<h1[^>]*>([^<]+)</h1>`,
		`<h2[^>]*>([^<]+)</h2>`,
		`<title[^>]*>([^<]+)</title>`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(htmlContent)
		if len(matches) > 1 {
			return strings.TrimSpace(matches[1])
		}
	}

	return ""
}

// ==================== LEITURA ====================

// OpenBook abre livro
func (er *EbookReader) OpenBook(bookID string) (*Book, error) {
	er.mu.Lock()
	defer er.mu.Unlock()

	book, ok := er.books[bookID]
	if !ok {
		return nil, fmt.Errorf("livro não encontrado")
	}

	er.currentBook = book

	// Atualiza progresso
	if _, ok := er.readingState[bookID]; !ok {
		er.readingState[bookID] = &ReadingProgress{
			BookID:         bookID,
			CurrentChapter: 0,
			CurrentPage:    0,
		}
	}
	er.readingState[bookID].LastRead = time.Now()
	er.readingState[bookID].SessionStart = time.Now()

	return book, nil
}

// GetCurrentChapter retorna capítulo atual
func (er *EbookReader) GetCurrentChapter() (*Chapter, error) {
	er.mu.RLock()
	defer er.mu.RUnlock()

	if er.currentBook == nil {
		return nil, fmt.Errorf("nenhum livro aberto")
	}

	progress := er.readingState[er.currentBook.ID]
	if progress.CurrentChapter >= len(er.currentBook.Chapters) {
		return nil, fmt.Errorf("fim do livro")
	}

	return er.currentBook.Chapters[progress.CurrentChapter], nil
}

// NextChapter avança para próximo capítulo
func (er *EbookReader) NextChapter() (*Chapter, error) {
	er.mu.Lock()
	defer er.mu.Unlock()

	if er.currentBook == nil {
		return nil, fmt.Errorf("nenhum livro aberto")
	}

	progress := er.readingState[er.currentBook.ID]
	if progress.CurrentChapter >= len(er.currentBook.Chapters)-1 {
		return nil, fmt.Errorf("último capítulo")
	}

	progress.CurrentChapter++
	progress.Percentage = float64(progress.CurrentChapter) / float64(len(er.currentBook.Chapters)) * 100
	er.save()

	return er.currentBook.Chapters[progress.CurrentChapter], nil
}

// PreviousChapter volta para capítulo anterior
func (er *EbookReader) PreviousChapter() (*Chapter, error) {
	er.mu.Lock()
	defer er.mu.Unlock()

	if er.currentBook == nil {
		return nil, fmt.Errorf("nenhum livro aberto")
	}

	progress := er.readingState[er.currentBook.ID]
	if progress.CurrentChapter <= 0 {
		return nil, fmt.Errorf("primeiro capítulo")
	}

	progress.CurrentChapter--
	progress.Percentage = float64(progress.CurrentChapter) / float64(len(er.currentBook.Chapters)) * 100
	er.save()

	return er.currentBook.Chapters[progress.CurrentChapter], nil
}

// GoToChapter vai para capítulo específico
func (er *EbookReader) GoToChapter(index int) (*Chapter, error) {
	er.mu.Lock()
	defer er.mu.Unlock()

	if er.currentBook == nil {
		return nil, fmt.Errorf("nenhum livro aberto")
	}

	if index < 0 || index >= len(er.currentBook.Chapters) {
		return nil, fmt.Errorf("capítulo inválido")
	}

	progress := er.readingState[er.currentBook.ID]
	progress.CurrentChapter = index
	progress.Percentage = float64(index) / float64(len(er.currentBook.Chapters)) * 100
	er.save()

	return er.currentBook.Chapters[index], nil
}

// ==================== LEITURA EM VOZ ALTA ====================

// ReadAloud lê capítulo em voz alta
func (er *EbookReader) ReadAloud(ctx context.Context) error {
	chapter, err := er.GetCurrentChapter()
	if err != nil {
		return err
	}

	if er.tts == nil {
		return fmt.Errorf("TTS não configurado")
	}

	// Divide em parágrafos para pausas naturais
	paragraphs := strings.Split(chapter.Content, "\n\n")

	for _, paragraph := range paragraphs {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		paragraph = strings.TrimSpace(paragraph)
		if len(paragraph) < 10 {
			continue
		}

		er.tts.Speak(paragraph)
		time.Sleep(500 * time.Millisecond) // Pausa entre parágrafos
	}

	return nil
}

// ==================== AI FEATURES ====================

// SummarizeChapter resume capítulo
func (er *EbookReader) SummarizeChapter(ctx context.Context, chapterIndex int) (string, error) {
	if er.currentBook == nil {
		return "", fmt.Errorf("nenhum livro aberto")
	}

	if chapterIndex >= len(er.currentBook.Chapters) {
		return "", fmt.Errorf("capítulo inválido")
	}

	chapter := er.currentBook.Chapters[chapterIndex]

	prompt := fmt.Sprintf(`Resuma este capítulo do livro "%s":

%s

Forneça:
1. Resumo em 3-5 frases
2. Pontos principais
3. Citações importantes (se houver)

Resumo:`, er.currentBook.Title, chapter.Content[:min(8000, len(chapter.Content))])

	return er.llm.Generate(ctx, prompt)
}

// SummarizeBook resume livro inteiro
func (er *EbookReader) SummarizeBook(ctx context.Context, bookID string) (*BookSummary, error) {
	book, ok := er.books[bookID]
	if !ok {
		return nil, fmt.Errorf("livro não encontrado")
	}

	// Coleta resumos de cada capítulo
	var allContent strings.Builder
	for _, chapter := range book.Chapters[:min(10, len(book.Chapters))] {
		allContent.WriteString(chapter.Content[:min(1000, len(chapter.Content))])
		allContent.WriteString("\n---\n")
	}

	prompt := fmt.Sprintf(`Analise este livro "%s" de %s:

%s

Forneça em JSON:
{
  "summary": "resumo geral em 5-10 frases",
  "key_points": ["ponto 1", "ponto 2", ...],
  "quotes": ["citação importante 1", ...],
  "themes": ["tema 1", "tema 2", ...]
}`, book.Title, book.Author, allContent.String())

	response, err := er.llm.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	var summary BookSummary
	json.Unmarshal([]byte(response), &summary)
	summary.BookID = bookID
	summary.GeneratedAt = time.Now()

	return &summary, nil
}

// AskAboutBook pergunta sobre o livro
func (er *EbookReader) AskAboutBook(ctx context.Context, question string) (string, error) {
	if er.currentBook == nil {
		return "", fmt.Errorf("nenhum livro aberto")
	}

	// Coleta contexto do capítulo atual e adjacentes
	progress := er.readingState[er.currentBook.ID]
	var context strings.Builder

	start := max(0, progress.CurrentChapter-1)
	end := min(len(er.currentBook.Chapters), progress.CurrentChapter+2)

	for i := start; i < end; i++ {
		ch := er.currentBook.Chapters[i]
		context.WriteString(fmt.Sprintf("=== %s ===\n", ch.Title))
		context.WriteString(ch.Content[:min(2000, len(ch.Content))])
		context.WriteString("\n\n")
	}

	prompt := fmt.Sprintf(`Baseado neste trecho do livro "%s":

%s

Pergunta: %s

Resposta:`, er.currentBook.Title, context.String(), question)

	return er.llm.Generate(ctx, prompt)
}

// GenerateFlashcards gera flashcards do capítulo
func (er *EbookReader) GenerateFlashcards(ctx context.Context, chapterIndex int) ([]map[string]string, error) {
	if er.currentBook == nil {
		return nil, fmt.Errorf("nenhum livro aberto")
	}

	chapter := er.currentBook.Chapters[chapterIndex]

	prompt := fmt.Sprintf(`Crie 5-10 flashcards para estudo baseados neste texto:

%s

Retorne em JSON:
[
  {"front": "pergunta", "back": "resposta"},
  ...
]`, chapter.Content[:min(5000, len(chapter.Content))])

	response, err := er.llm.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	var flashcards []map[string]string
	json.Unmarshal([]byte(response), &flashcards)

	return flashcards, nil
}

// ==================== HIGHLIGHTS ====================

// AddHighlight adiciona destaque
func (er *EbookReader) AddHighlight(bookID, chapterID, text, note, color string) *Highlight {
	er.mu.Lock()
	defer er.mu.Unlock()

	highlight := &Highlight{
		ID:        fmt.Sprintf("hl_%d", time.Now().UnixNano()),
		BookID:    bookID,
		ChapterID: chapterID,
		Text:      text,
		Note:      note,
		Color:     color,
		CreatedAt: time.Now(),
	}

	if _, ok := er.highlights[bookID]; !ok {
		er.highlights[bookID] = make([]*Highlight, 0)
	}
	er.highlights[bookID] = append(er.highlights[bookID], highlight)

	er.save()
	return highlight
}

// GetHighlights retorna destaques de um livro
func (er *EbookReader) GetHighlights(bookID string) []*Highlight {
	er.mu.RLock()
	defer er.mu.RUnlock()
	return er.highlights[bookID]
}

// ExportHighlights exporta destaques para Markdown
func (er *EbookReader) ExportHighlights(bookID string) string {
	book := er.books[bookID]
	highlights := er.highlights[bookID]

	var export strings.Builder
	export.WriteString(fmt.Sprintf("# Destaques: %s\n", book.Title))
	export.WriteString(fmt.Sprintf("**Autor:** %s\n\n", book.Author))
	export.WriteString("---\n\n")

	for _, hl := range highlights {
		export.WriteString(fmt.Sprintf("> %s\n\n", hl.Text))
		if hl.Note != "" {
			export.WriteString(fmt.Sprintf("*Nota: %s*\n\n", hl.Note))
		}
		export.WriteString("---\n\n")
	}

	return export.String()
}

// ==================== BIBLIOTECA ====================

// GetAllBooks retorna todos os livros
func (er *EbookReader) GetAllBooks() []*Book {
	er.mu.RLock()
	defer er.mu.RUnlock()

	books := make([]*Book, 0, len(er.books))
	for _, b := range er.books {
		books = append(books, b)
	}

	// Ordena por título
	sort.Slice(books, func(i, j int) bool {
		return books[i].Title < books[j].Title
	})

	return books
}

// SearchBooks busca livros
func (er *EbookReader) SearchBooks(query string) []*Book {
	er.mu.RLock()
	defer er.mu.RUnlock()

	queryLower := strings.ToLower(query)
	results := make([]*Book, 0)

	for _, book := range er.books {
		if strings.Contains(strings.ToLower(book.Title), queryLower) ||
			strings.Contains(strings.ToLower(book.Author), queryLower) {
			results = append(results, book)
		}
	}

	return results
}

// GetReadingProgress retorna progresso de leitura
func (er *EbookReader) GetReadingProgress(bookID string) *ReadingProgress {
	er.mu.RLock()
	defer er.mu.RUnlock()
	return er.readingState[bookID]
}

// GetCurrentlyReading retorna livros sendo lidos
func (er *EbookReader) GetCurrentlyReading() []*Book {
	er.mu.RLock()
	defer er.mu.RUnlock()

	reading := make([]*Book, 0)
	for bookID, progress := range er.readingState {
		if progress.Percentage > 0 && progress.Percentage < 100 {
			if book, ok := er.books[bookID]; ok {
				reading = append(reading, book)
			}
		}
	}

	// Ordena por último lido
	sort.Slice(reading, func(i, j int) bool {
		return er.readingState[reading[i].ID].LastRead.After(er.readingState[reading[j].ID].LastRead)
	})

	return reading
}

// ==================== CALIBRE CLI ====================

// OpenInCalibre abre livro no Calibre
func (er *EbookReader) OpenInCalibre(bookID string) error {
	book := er.books[bookID]
	if book == nil {
		return fmt.Errorf("livro não encontrado")
	}

	// Tenta abrir com calibre
	cmd := exec.Command("ebook-viewer", book.Path)
	return cmd.Start()
}

// ConvertFormat converte formato usando Calibre
func (er *EbookReader) ConvertFormat(bookID, targetFormat string) (string, error) {
	book := er.books[bookID]
	if book == nil {
		return "", fmt.Errorf("livro não encontrado")
	}

	outputPath := strings.TrimSuffix(book.Path, filepath.Ext(book.Path)) + "." + targetFormat

	cmd := exec.Command("ebook-convert", book.Path, outputPath)
	if err := cmd.Run(); err != nil {
		return "", err
	}

	return outputPath, nil
}

// ==================== PERSISTÊNCIA ====================

func (er *EbookReader) save() error {
	data := struct {
		Books        map[string]*Book            `json:"books"`
		ReadingState map[string]*ReadingProgress `json:"reading_state"`
		Highlights   map[string][]*Highlight     `json:"highlights"`
	}{
		Books:        er.books,
		ReadingState: er.readingState,
		Highlights:   er.highlights,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(er.basePath, "ebooks.json"), jsonData, 0644)
}

func (er *EbookReader) load() error {
	data, err := os.ReadFile(filepath.Join(er.basePath, "ebooks.json"))
	if err != nil {
		return nil
	}

	var loaded struct {
		Books        map[string]*Book            `json:"books"`
		ReadingState map[string]*ReadingProgress `json:"reading_state"`
		Highlights   map[string][]*Highlight     `json:"highlights"`
	}

	if err := json.Unmarshal(data, &loaded); err != nil {
		return err
	}

	if loaded.Books != nil {
		er.books = loaded.Books
	}
	if loaded.ReadingState != nil {
		er.readingState = loaded.ReadingState
	}
	if loaded.Highlights != nil {
		er.highlights = loaded.Highlights
	}

	return nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
