package llm

import (
	"context"
	"fmt"
	"strings"

	ort "github.com/yalue/onnxruntime_go"
	"github.com/JoseRFJuniorLLMs/NPU-IA/pkg/config"
)

// Model representa um modelo de linguagem
type Model struct {
	session    *ort.DynamicAdvancedSession
	config     config.ModelConfig
	tokenizer  *Tokenizer
	systemPrompt string
}

// New cria um novo modelo LLM
func New(cfg config.ModelConfig) (*Model, error) {
	// Configura opções ONNX
	options, err := ort.NewSessionOptions()
	if err != nil {
		return nil, err
	}

	// Usa DirectML para NPU AMD
	err = options.AppendExecutionProviderDirectML(0)
	if err != nil {
		fmt.Printf("DirectML não disponível para %s, usando CPU\n", cfg.Name)
	}

	// Carrega modelo
	session, err := ort.NewDynamicAdvancedSession(
		cfg.Path,
		[]string{"input_ids", "attention_mask"},
		[]string{"logits"},
		options,
	)
	if err != nil {
		return nil, fmt.Errorf("erro ao carregar %s: %w", cfg.Name, err)
	}

	// Carrega tokenizer
	tokenizer, err := NewTokenizer(cfg.TokenizerPath)
	if err != nil {
		return nil, fmt.Errorf("erro ao carregar tokenizer: %w", err)
	}

	return &Model{
		session:      session,
		config:       cfg,
		tokenizer:    tokenizer,
		systemPrompt: cfg.SystemPrompt,
	}, nil
}

// Generate gera texto a partir do prompt
func (m *Model) Generate(ctx context.Context, prompt string) (string, error) {
	// Monta prompt completo
	fullPrompt := m.buildPrompt(prompt)

	// Tokeniza
	inputIDs, attentionMask := m.tokenizer.Encode(fullPrompt)

	// Prepara tensores
	inputShape := ort.NewShape(1, int64(len(inputIDs)))

	inputTensor, err := ort.NewTensor(inputShape, inputIDs)
	if err != nil {
		return "", err
	}
	defer inputTensor.Destroy()

	maskTensor, err := ort.NewTensor(inputShape, attentionMask)
	if err != nil {
		return "", err
	}
	defer maskTensor.Destroy()

	// Geração autoregressiva
	var generatedIDs []int64
	maxTokens := m.config.MaxTokens

	for i := 0; i < maxTokens; i++ {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		// Executa inferência
		outputs, err := m.session.Run(map[string]*ort.Tensor[int64]{
			"input_ids":      inputTensor,
			"attention_mask": maskTensor,
		})
		if err != nil {
			return "", err
		}

		// Pega próximo token (sampling)
		nextToken := m.sampleNextToken(outputs["logits"])

		// Verifica EOS
		if nextToken == m.tokenizer.EOSToken() {
			break
		}

		generatedIDs = append(generatedIDs, nextToken)

		// Atualiza input para próxima iteração
		inputIDs = append(inputIDs, nextToken)
		attentionMask = append(attentionMask, 1)
	}

	// Decodifica
	result := m.tokenizer.Decode(generatedIDs)

	return strings.TrimSpace(result), nil
}

// GenerateAction gera uma ação estruturada
func (m *Model) GenerateAction(ctx context.Context, prompt string) (string, error) {
	actionPrompt := fmt.Sprintf(`Você é um assistente que executa ações no computador.
Dado o comando do usuário, retorne APENAS o JSON da ação, sem explicações.

Formato:
{"action": "tipo_acao", "params": {"param1": "valor1"}}

Ações disponíveis:
- open_app: abre aplicativo {"app": "nome"}
- open_url: abre URL {"url": "endereco"}
- type_text: digita texto {"text": "texto"}
- read_email: lê emails {}
- send_email: envia email {"to": "email", "subject": "assunto", "body": "corpo"}
- volume: ajusta volume {"level": 50}
- screenshot: captura tela {}

Comando: %s

JSON:`, prompt)

	return m.Generate(ctx, actionPrompt)
}

// buildPrompt monta o prompt completo com system message
func (m *Model) buildPrompt(userPrompt string) string {
	if m.systemPrompt == "" {
		m.systemPrompt = "Você é um assistente IA útil que responde em português brasileiro de forma concisa."
	}

	return fmt.Sprintf(`<|system|>
%s
<|end|>
<|user|>
%s
<|end|>
<|assistant|>
`, m.systemPrompt, userPrompt)
}

// sampleNextToken faz sampling do próximo token
func (m *Model) sampleNextToken(logits *ort.Tensor[float32]) int64 {
	// TODO: Implementar sampling com temperature
	// Por enquanto, greedy (argmax)
	data := logits.GetData()
	maxIdx := int64(0)
	maxVal := data[0]

	for i, v := range data {
		if v > maxVal {
			maxVal = v
			maxIdx = int64(i)
		}
	}

	return maxIdx
}

// Close libera recursos
func (m *Model) Close() error {
	if m.session != nil {
		return m.session.Destroy()
	}
	return nil
}
