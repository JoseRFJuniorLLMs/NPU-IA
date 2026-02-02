package llm

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strings"
	"time"

	ort "github.com/yalue/onnxruntime_go"
	"github.com/JoseRFJuniorLLMs/NPU-IA/pkg/config"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// SamplingParams parâmetros de sampling
type SamplingParams struct {
	Temperature     float32 // 0.0 = greedy, 1.0 = mais criativo
	TopK            int     // Top-K sampling (0 = desativado)
	TopP            float32 // Nucleus sampling (0.0 = desativado)
	RepetitionPenalty float32 // Penalidade para repetição
}

// Model representa um modelo de linguagem
type Model struct {
	session       *ort.DynamicAdvancedSession
	config        config.ModelConfig
	tokenizer     *Tokenizer
	systemPrompt  string
	sampling      SamplingParams
	generatedIDs  []int64 // Para aplicar repetition penalty
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

	// Configura sampling baseado na temperatura do config
	sampling := SamplingParams{
		Temperature:       cfg.Temperature,
		TopK:             40,   // Default razoável
		TopP:             0.95, // Nucleus sampling
		RepetitionPenalty: 1.1,
	}

	// Ajusta para tarefas específicas
	if cfg.Temperature < 0.3 {
		// Mais determinístico (código, ações)
		sampling.TopK = 10
		sampling.TopP = 0.5
	} else if cfg.Temperature > 0.8 {
		// Mais criativo
		sampling.TopK = 100
		sampling.TopP = 0.98
	}

	return &Model{
		session:      session,
		config:       cfg,
		tokenizer:    tokenizer,
		systemPrompt: cfg.SystemPrompt,
		sampling:     sampling,
	}, nil
}

// Generate gera texto a partir do prompt
func (m *Model) Generate(ctx context.Context, prompt string) (string, error) {
	// Limpa histórico de geração anterior
	m.generatedIDs = nil

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

// sampleNextToken faz sampling do próximo token com temperature, top-k e top-p
func (m *Model) sampleNextToken(logits *ort.Tensor[float32]) int64 {
	data := logits.GetData()
	if len(data) == 0 {
		return m.tokenizer.EOSToken()
	}

	// Copia para não modificar original
	logitsCopy := make([]float32, len(data))
	copy(logitsCopy, data)

	// Aplica repetition penalty
	if m.sampling.RepetitionPenalty != 1.0 && len(m.generatedIDs) > 0 {
		m.applyRepetitionPenalty(logitsCopy)
	}

	// Se temperature = 0, usa greedy
	if m.sampling.Temperature == 0 || m.sampling.Temperature < 0.01 {
		return m.greedySample(logitsCopy)
	}

	// Aplica temperature
	m.applyTemperature(logitsCopy, m.sampling.Temperature)

	// Aplica top-k filtering
	if m.sampling.TopK > 0 {
		m.applyTopK(logitsCopy, m.sampling.TopK)
	}

	// Aplica top-p (nucleus) filtering
	if m.sampling.TopP > 0 && m.sampling.TopP < 1.0 {
		m.applyTopP(logitsCopy, m.sampling.TopP)
	}

	// Converte para probabilidades (softmax)
	probs := m.softmax(logitsCopy)

	// Amostra da distribuição
	token := m.sampleFromProbs(probs)

	// Guarda para repetition penalty
	m.generatedIDs = append(m.generatedIDs, token)

	return token
}

// greedySample retorna o token com maior logit
func (m *Model) greedySample(logits []float32) int64 {
	maxIdx := int64(0)
	maxVal := logits[0]

	for i, v := range logits {
		if v > maxVal {
			maxVal = v
			maxIdx = int64(i)
		}
	}

	return maxIdx
}

// applyTemperature aplica temperature scaling aos logits
func (m *Model) applyTemperature(logits []float32, temperature float32) {
	for i := range logits {
		logits[i] = logits[i] / temperature
	}
}

// applyRepetitionPenalty penaliza tokens já gerados
func (m *Model) applyRepetitionPenalty(logits []float32) {
	seen := make(map[int64]bool)
	for _, id := range m.generatedIDs {
		if seen[id] {
			continue
		}
		seen[id] = true

		if int(id) < len(logits) {
			if logits[id] > 0 {
				logits[id] = logits[id] / m.sampling.RepetitionPenalty
			} else {
				logits[id] = logits[id] * m.sampling.RepetitionPenalty
			}
		}
	}
}

// applyTopK mantém apenas os top-k tokens com maior probabilidade
func (m *Model) applyTopK(logits []float32, k int) {
	if k >= len(logits) {
		return
	}

	// Encontra o k-ésimo maior valor
	type idxVal struct {
		idx int
		val float32
	}

	pairs := make([]idxVal, len(logits))
	for i, v := range logits {
		pairs[i] = idxVal{i, v}
	}

	// Ordena por valor decrescente
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].val > pairs[j].val
	})

	// Pega o threshold do k-ésimo elemento
	threshold := pairs[k-1].val

	// Zera todos abaixo do threshold
	negInf := float32(math.Inf(-1))
	for i := range logits {
		if logits[i] < threshold {
			logits[i] = negInf
		}
	}
}

// applyTopP aplica nucleus sampling (top-p)
func (m *Model) applyTopP(logits []float32, p float32) {
	// Primeiro aplica softmax para obter probabilidades
	probs := m.softmax(logits)

	// Ordena por probabilidade decrescente
	type idxProb struct {
		idx  int
		prob float32
	}

	pairs := make([]idxProb, len(probs))
	for i, prob := range probs {
		pairs[i] = idxProb{i, prob}
	}

	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].prob > pairs[j].prob
	})

	// Acumula probabilidades até atingir p
	cumProb := float32(0.0)
	cutoffIdx := len(pairs)

	for i, pair := range pairs {
		cumProb += pair.prob
		if cumProb >= p {
			cutoffIdx = i + 1
			break
		}
	}

	// Cria set dos índices válidos
	validSet := make(map[int]bool)
	for i := 0; i < cutoffIdx; i++ {
		validSet[pairs[i].idx] = true
	}

	// Zera probabilidades fora do nucleus
	negInf := float32(math.Inf(-1))
	for i := range logits {
		if !validSet[i] {
			logits[i] = negInf
		}
	}
}

// softmax converte logits em probabilidades
func (m *Model) softmax(logits []float32) []float32 {
	// Encontra máximo para estabilidade numérica
	maxVal := logits[0]
	for _, v := range logits {
		if v > maxVal && !math.IsInf(float64(v), 0) {
			maxVal = v
		}
	}

	// Calcula exp(x - max) e soma
	probs := make([]float32, len(logits))
	sum := float32(0.0)

	for i, v := range logits {
		if math.IsInf(float64(v), -1) {
			probs[i] = 0
		} else {
			probs[i] = float32(math.Exp(float64(v - maxVal)))
			sum += probs[i]
		}
	}

	// Normaliza
	if sum > 0 {
		for i := range probs {
			probs[i] /= sum
		}
	}

	return probs
}

// sampleFromProbs amostra um índice baseado nas probabilidades
func (m *Model) sampleFromProbs(probs []float32) int64 {
	r := rand.Float32()
	cumProb := float32(0.0)

	for i, p := range probs {
		cumProb += p
		if r < cumProb {
			return int64(i)
		}
	}

	// Fallback: retorna o último índice válido
	for i := len(probs) - 1; i >= 0; i-- {
		if probs[i] > 0 {
			return int64(i)
		}
	}

	return 0
}

// ResetGeneration limpa o histórico de tokens gerados (para nova conversa)
func (m *Model) ResetGeneration() {
	m.generatedIDs = nil
}

// SetSamplingParams permite ajustar parâmetros de sampling em runtime
func (m *Model) SetSamplingParams(params SamplingParams) {
	m.sampling = params
}

// Close libera recursos
func (m *Model) Close() error {
	if m.session != nil {
		return m.session.Destroy()
	}
	return nil
}
