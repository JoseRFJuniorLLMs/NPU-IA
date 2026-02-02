package actions

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

// BrowserClient controle avançado do navegador
type BrowserClient struct {
	ctx    context.Context
	cancel context.CancelFunc
}

// NewBrowserClient cria cliente do navegador
func NewBrowserClient() (*BrowserClient, error) {
	// Cria contexto do Chrome DevTools Protocol
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
		chromedp.Flag("disable-gpu", false),
		chromedp.Flag("start-maximized", true),
	)

	allocCtx, _ := chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, cancel := chromedp.NewContext(allocCtx)

	return &BrowserClient{
		ctx:    ctx,
		cancel: cancel,
	}, nil
}

// Navigate navega para uma URL
func (b *BrowserClient) Navigate(url string) error {
	if !strings.HasPrefix(url, "http") {
		url = "https://" + url
	}

	return chromedp.Run(b.ctx, chromedp.Navigate(url))
}

// Click clica em um elemento
func (b *BrowserClient) Click(selector string) error {
	return chromedp.Run(b.ctx,
		chromedp.WaitVisible(selector, chromedp.ByQuery),
		chromedp.Click(selector, chromedp.ByQuery),
	)
}

// Type digita em um campo
func (b *BrowserClient) Type(selector, text string) error {
	return chromedp.Run(b.ctx,
		chromedp.WaitVisible(selector, chromedp.ByQuery),
		chromedp.SendKeys(selector, text, chromedp.ByQuery),
	)
}

// GetText obtém texto de um elemento
func (b *BrowserClient) GetText(selector string) (string, error) {
	var text string
	err := chromedp.Run(b.ctx,
		chromedp.WaitVisible(selector, chromedp.ByQuery),
		chromedp.Text(selector, &text, chromedp.ByQuery),
	)
	return text, err
}

// Screenshot captura screenshot
func (b *BrowserClient) Screenshot() ([]byte, error) {
	var buf []byte
	err := chromedp.Run(b.ctx, chromedp.CaptureScreenshot(&buf))
	return buf, err
}

// GetPageContent obtém conteúdo da página
func (b *BrowserClient) GetPageContent() (string, error) {
	var html string
	err := chromedp.Run(b.ctx, chromedp.OuterHTML("html", &html))
	return html, err
}

// Search faz busca no Google
func (b *BrowserClient) Search(query string) ([]string, error) {
	url := fmt.Sprintf("https://www.google.com/search?q=%s", strings.ReplaceAll(query, " ", "+"))

	var results []string
	err := chromedp.Run(b.ctx,
		chromedp.Navigate(url),
		chromedp.Sleep(2*time.Second),
		chromedp.Evaluate(`
			Array.from(document.querySelectorAll('h3')).map(h => h.textContent).slice(0, 5)
		`, &results),
	)

	return results, err
}

// OpenYouTube abre YouTube e busca
func (b *BrowserClient) OpenYouTube(query string) error {
	url := fmt.Sprintf("https://www.youtube.com/results?search_query=%s", strings.ReplaceAll(query, " ", "+"))
	return b.Navigate(url)
}

// OpenSpotify abre Spotify (app ou web)
func (b *BrowserClient) OpenSpotify(query string) error {
	// Tenta abrir app
	cmd := exec.Command("cmd", "/c", "start", "spotify:")
	if err := cmd.Run(); err != nil {
		// Fallback para web
		url := fmt.Sprintf("https://open.spotify.com/search/%s", strings.ReplaceAll(query, " ", "%20"))
		return b.Navigate(url)
	}
	return nil
}

// OpenWhatsApp abre WhatsApp Web
func (b *BrowserClient) OpenWhatsApp() error {
	return b.Navigate("https://web.whatsapp.com")
}

// SendWhatsApp envia mensagem no WhatsApp (precisa estar logado)
func (b *BrowserClient) SendWhatsApp(contact, message string) error {
	// Busca contato
	err := chromedp.Run(b.ctx,
		chromedp.Navigate("https://web.whatsapp.com"),
		chromedp.Sleep(5*time.Second), // Aguarda carregar
		chromedp.WaitVisible(`div[data-testid="chat-list-search"]`, chromedp.ByQuery),
		chromedp.Click(`div[data-testid="chat-list-search"]`, chromedp.ByQuery),
		chromedp.SendKeys(`div[data-testid="chat-list-search"] input`, contact, chromedp.ByQuery),
		chromedp.Sleep(2*time.Second),
	)
	if err != nil {
		return err
	}

	// Clica no contato e envia mensagem
	return chromedp.Run(b.ctx,
		chromedp.Click(`span[title="`+contact+`"]`, chromedp.ByQuery),
		chromedp.Sleep(1*time.Second),
		chromedp.SendKeys(`div[data-testid="conversation-compose-box-input"]`, message, chromedp.ByQuery),
		chromedp.Click(`button[data-testid="send"]`, chromedp.ByQuery),
	)
}

// Close fecha o navegador
func (b *BrowserClient) Close() error {
	b.cancel()
	return nil
}
