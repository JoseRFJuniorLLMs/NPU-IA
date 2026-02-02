package stt

import (
	"fmt"

	ort "github.com/yalue/onnxruntime_go"
	"github.com/JoseRFJuniorLLMs/NPU-IA/pkg/config"
)

// Whisper implementa Speech-to-Text usando Faster-Whisper ONNX
type Whisper struct {
	session  *ort.DynamicAdvancedSession
	config   config.STTConfig
	language string
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
	session, err := ort.NewDynamicAdvancedSession(
		cfg.ModelPath,
		[]string{"audio_input"},
		[]string{"transcription"},
		options,
	)
	if err != nil {
		return nil, fmt.Errorf("erro ao carregar modelo Whisper: %w", err)
	}

	return &Whisper{
		session:  session,
		config:   cfg,
		language: cfg.Language,
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
	// TODO: Implementar decodificação do tokenizer
	// Por enquanto retorna placeholder
	return ""
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
