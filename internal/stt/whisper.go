package stt

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	ort "github.com/yalue/onnxruntime_go"
	"github.com/JoseRFJuniorLLMs/NPU-IA/pkg/config"
)

// Whisper implementa Speech-to-Text usando Faster-Whisper ONNX
type Whisper struct {
	session   *ort.DynamicAdvancedSession
	config    config.STTConfig
	language  string
	tokenizer *WhisperTokenizer
}

// WhisperTokenizer tokenizer específico para Whisper
type WhisperTokenizer struct {
	vocab       map[int64]string
	specialIDs  map[string]int64
	timestamps  bool
}

// NewWhisperTokenizer cria tokenizer para Whisper
func NewWhisperTokenizer(modelPath string) (*WhisperTokenizer, error) {
	t := &WhisperTokenizer{
		vocab:      make(map[int64]string),
		specialIDs: make(map[string]int64),
		timestamps: true,
	}

	// Tenta carregar vocab.json do diretório do modelo
	vocabPath := filepath.Join(filepath.Dir(modelPath), "vocab.json")
	if data, err := os.ReadFile(vocabPath); err == nil {
		var vocab map[string]int64
		if err := json.Unmarshal(data, &vocab); err == nil {
			for token, id := range vocab {
				t.vocab[id] = token
			}
		}
	}

	// IDs especiais do Whisper
	t.specialIDs = map[string]int64{
		"<|endoftext|>":      50257,
		"<|startoftranscript|>": 50258,
		"<|translate|>":      50358,
		"<|transcribe|>":     50359,
		"<|startoflm|>":      50360,
		"<|startofprev|>":    50361,
		"<|nospeech|>":       50362,
		"<|notimestamps|>":   50363,
		"<|pt|>":             50316, // Português
		"<|en|>":             50259, // Inglês
	}

	return t, nil
}

// NewWhisper cria uma nova instância do Whisper
func NewWhisper(cfg config.STTConfig) (*Whisper, error) {
	// Inicializa ONNX Runtime com DirectML (NPU AMD)
	err := ort.InitializeEnvironment()
	if err != nil {
		return nil, fmt.Errorf("erro ao inicializar ONNX: %w", err)
	}

	// Configura para usar DirectML (NPU)
	options, err := ort.NewSessionOptions()
	if err != nil {
		return nil, err
	}

	// Adiciona DirectML como provider (usa NPU AMD)
	err = options.AppendExecutionProviderDirectML(0)
	if err != nil {
		// Fallback para CPU se DirectML não disponível
		fmt.Println("DirectML não disponível, usando CPU")
	}

	// Carrega o modelo Whisper ONNX
	// Nomes de entrada/saída variam por versão do modelo
	inputNames := []string{"audio_pcm", "min_length", "max_length", "num_beams", "num_return_sequences", "length_penalty", "repetition_penalty"}
	outputNames := []string{"str"}

	// Tenta carregar com nomes padrão do whisper ONNX
	session, err := ort.NewDynamicAdvancedSession(
		cfg.ModelPath,
		inputNames,
		outputNames,
		options,
	)
	if err != nil {
		// Tenta com nomes alternativos
		session, err = ort.NewDynamicAdvancedSession(
			cfg.ModelPath,
			[]string{"audio_input"},
			[]string{"logits"},
			options,
		)
		if err != nil {
			return nil, fmt.Errorf("erro ao carregar modelo Whisper: %w", err)
		}
	}

	// Carrega tokenizer
	tokenizer, err := NewWhisperTokenizer(cfg.ModelPath)
	if err != nil {
		fmt.Printf("Aviso: não foi possível carregar tokenizer: %v\n", err)
	}

	return &Whisper{
		session:   session,
		config:    cfg,
		language:  cfg.Language,
		tokenizer: tokenizer,
	}, nil
}

// Transcribe converte áudio em texto
func (w *Whisper) Transcribe(audioData []float32) (string, error) {
	if len(audioData) == 0 {
		return "", nil
	}

	// Prepara input tensor
	inputShape := ort.NewShape(1, int64(len(audioData)))
	inputTensor, err := ort.NewTensor(inputShape, audioData)
	if err != nil {
		return "", fmt.Errorf("erro ao criar tensor: %w", err)
	}
	defer inputTensor.Destroy()

	// Executa inferência
	outputs, err := w.session.Run(map[string]*ort.Tensor[float32]{
		"audio_input": inputTensor,
	})
	if err != nil {
		return "", fmt.Errorf("erro na inferência: %w", err)
	}

	// Processa output
	transcription := w.decodeOutput(outputs)

	return transcription, nil
}

// decodeOutput decodifica o output do modelo
func (w *Whisper) decodeOutput(outputs map[string]*ort.Tensor[float32]) string {
	// Obtém tensor de logits
	logits, ok := outputs["logits"]
	if !ok {
		// Tenta nome alternativo
		for name, tensor := range outputs {
			if strings.Contains(name, "logit") || strings.Contains(name, "output") {
				logits = tensor
				break
			}
		}
	}

	if logits == nil {
		return ""
	}

	data := logits.GetData()
	if len(data) == 0 {
		return ""
	}

	// Decodifica tokens usando greedy search
	tokens := w.greedyDecode(data)

	// Converte tokens em texto
	return w.tokensToText(tokens)
}

// greedyDecode faz decodificação greedy dos logits
func (w *Whisper) greedyDecode(logits []float32) []int64 {
	// Whisper vocab size é ~51865
	vocabSize := 51865

	if len(logits) < vocabSize {
		vocabSize = len(logits)
	}

	var tokens []int64
	endOfText := int64(50257) // <|endoftext|>

	// Processa em blocos de vocabSize
	for i := 0; i < len(logits); i += vocabSize {
		end := i + vocabSize
		if end > len(logits) {
			break
		}

		// Encontra argmax
		maxIdx := 0
		maxVal := logits[i]
		for j := 1; j < vocabSize && i+j < len(logits); j++ {
			if logits[i+j] > maxVal {
				maxVal = logits[i + j]
				maxIdx = j
			}
		}

		token := int64(maxIdx)

		// Para se encontrar end of text
		if token == endOfText {
			break
		}

		// Ignora tokens especiais de timestamp (<|0.00|> a <|30.00|>)
		if token >= 50364 && token <= 50864 {
			continue
		}

		// Ignora outros tokens especiais
		if token >= 50257 && token < 50364 {
			continue
		}

		tokens = append(tokens, token)
	}

	return tokens
}

// tokensToText converte tokens em texto
func (w *Whisper) tokensToText(tokens []int64) string {
	var result strings.Builder

	for _, token := range tokens {
		// Tenta decodificar do vocabulário
		if w.tokenizer != nil && w.tokenizer.vocab != nil {
			if text, ok := w.tokenizer.vocab[token]; ok {
				// Whisper usa Ġ para espaço
				text = strings.ReplaceAll(text, "Ġ", " ")
				// Remove caracteres especiais de byte
				if !strings.HasPrefix(text, "<|") {
					result.WriteString(text)
				}
				continue
			}
		}

		// Fallback: decodifica usando byte-level BPE básico
		if token < 256 {
			result.WriteByte(byte(token))
		}
	}

	text := result.String()

	// Limpa o texto
	text = strings.TrimSpace(text)

	// Remove múltiplos espaços
	for strings.Contains(text, "  ") {
		text = strings.ReplaceAll(text, "  ", " ")
	}

	return text
}

// Close libera recursos
func (w *Whisper) Close() error {
	if w.session != nil {
		return w.session.Destroy()
	}
	return nil
}

// SetLanguage define o idioma para transcrição
func (w *Whisper) SetLanguage(lang string) {
	w.language = lang
}
