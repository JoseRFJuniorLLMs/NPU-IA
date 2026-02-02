package coder

import (
	"context"
	"fmt"
	"strings"

	ort "github.com/yalue/onnxruntime_go"
	"github.com/JoseRFJuniorLLMs/NPU-IA/pkg/config"
)

// Model representa o modelo de código Qwen-Coder
type Model struct {
	session *ort.DynamicAdvancedSession
	config  config.ModelConfig
}

// New cria um novo modelo de código
func New(cfg config.ModelConfig) (*Model, error) {
	options, err := ort.NewSessionOptions()
	if err != nil {
		return nil, err
	}

	// Usa DirectML para NPU AMD
	err = options.AppendExecutionProviderDirectML(0)
	if err != nil {
		fmt.Println("DirectML não disponível para Coder, usando CPU")
	}

	// Carrega modelo Qwen-Coder ONNX
	session, err := ort.NewDynamicAdvancedSession(
		cfg.Path,
		[]string{"input_ids", "attention_mask"},
		[]string{"logits"},
		options,
	)
	if err != nil {
		return nil, fmt.Errorf("erro ao carregar modelo de código: %w", err)
	}

	return &Model{
		session: session,
		config:  cfg,
	}, nil
}

// Generate gera código ou analisa código existente
func (m *Model) Generate(ctx context.Context, prompt string) (string, error) {
	// Detecta tipo de tarefa
	taskType := m.detectTask(prompt)

	var systemPrompt string
	switch taskType {
	case "generate":
		systemPrompt = `Você é um assistente especializado em programação.
Gere código limpo, bem documentado e funcional.
Responda APENAS com o código, sem explicações adicionais.`

	case "fix":
		systemPrompt = `Você é um assistente especializado em debugging.
Analise o código, identifique o problema e corrija.
Mostre o código corrigido e explique brevemente o que estava errado.`

	case "explain":
		systemPrompt = `Você é um professor de programação.
Explique o código de forma clara e didática em português.
Use exemplos quando apropriado.`

	case "review":
		systemPrompt = `Você é um revisor de código experiente.
Analise o código e forneça feedback sobre:
- Possíveis bugs
- Performance
- Boas práticas
- Sugestões de melhoria`

	default:
		systemPrompt = `Você é um assistente de programação.
Ajude com qualquer tarefa relacionada a código.`
	}

	fullPrompt := fmt.Sprintf(`<|system|>
%s
<|end|>
<|user|>
%s
<|end|>
<|assistant|>
`, systemPrompt, prompt)

	// TODO: Implementar inferência real
	// Por enquanto, retorna placeholder
	return m.mockGenerate(fullPrompt), nil
}

// detectTask detecta o tipo de tarefa de código
func (m *Model) detectTask(prompt string) string {
	lower := strings.ToLower(prompt)

	if strings.Contains(lower, "cria") || strings.Contains(lower, "gera") ||
		strings.Contains(lower, "escreve") || strings.Contains(lower, "faz") {
		return "generate"
	}

	if strings.Contains(lower, "corrige") || strings.Contains(lower, "fix") ||
		strings.Contains(lower, "erro") || strings.Contains(lower, "bug") {
		return "fix"
	}

	if strings.Contains(lower, "explica") || strings.Contains(lower, "como funciona") ||
		strings.Contains(lower, "o que") {
		return "explain"
	}

	if strings.Contains(lower, "revisa") || strings.Contains(lower, "review") ||
		strings.Contains(lower, "analisa") {
		return "review"
	}

	return "general"
}

// mockGenerate simula geração (placeholder)
func (m *Model) mockGenerate(prompt string) string {
	return "// Código gerado pelo NPU-IA\n// TODO: Implementar inferência real"
}

// GenerateFromScreenshot gera/analisa código de uma imagem
func (m *Model) GenerateFromScreenshot(ctx context.Context, imageData []byte, prompt string) (string, error) {
	// TODO: Integrar com modelo de visão para OCR/análise de código na tela
	return "", fmt.Errorf("não implementado ainda")
}

// Close libera recursos
func (m *Model) Close() error {
	if m.session != nil {
		return m.session.Destroy()
	}
	return nil
}
