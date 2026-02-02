package audio

import (
	"fmt"
	"sync"
	"time"

	"github.com/gordonklaus/portaudio"
	"github.com/JoseRFJuniorLLMs/NPU-IA/pkg/config"
)

// Capture gerencia captura de áudio do microfone
type Capture struct {
	stream     *portaudio.Stream
	buffer     []float32
	config     config.AudioConfig
	mu         sync.Mutex
	isListening bool

	// Voice Activity Detection
	vadThreshold float32
	silenceTime  time.Duration
	maxDuration  time.Duration
}

// NewCapture cria uma nova instância de captura
func NewCapture(cfg config.AudioConfig) (*Capture, error) {
	// Inicializa PortAudio
	if err := portaudio.Initialize(); err != nil {
		return nil, fmt.Errorf("erro ao inicializar PortAudio: %w", err)
	}

	c := &Capture{
		config:       cfg,
		vadThreshold: cfg.VADThreshold,
		silenceTime:  time.Duration(cfg.SilenceMs) * time.Millisecond,
		maxDuration:  time.Duration(cfg.MaxDurationMs) * time.Millisecond,
	}

	// Configura stream
	inputChannels := 1
	sampleRate := cfg.SampleRate
	if sampleRate == 0 {
		sampleRate = 16000
	}
	framesPerBuffer := 1024

	stream, err := portaudio.OpenDefaultStream(
		inputChannels,    // input channels
		0,                // output channels
		float64(sampleRate),
		framesPerBuffer,
		c.processAudio,
	)
	if err != nil {
		portaudio.Terminate()
		return nil, fmt.Errorf("erro ao abrir stream: %w", err)
	}

	c.stream = stream

	return c, nil
}

// processAudio callback do PortAudio
func (c *Capture) processAudio(in []float32) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isListening {
		c.buffer = append(c.buffer, in...)
	}
}

// Listen aguarda fala e retorna áudio capturado
func (c *Capture) Listen() ([]float32, error) {
	c.mu.Lock()
	c.buffer = make([]float32, 0, c.config.SampleRate*10) // 10 segundos max
	c.isListening = true
	c.mu.Unlock()

	// Inicia stream
	if err := c.stream.Start(); err != nil {
		return nil, err
	}
	defer c.stream.Stop()

	// Aguarda atividade de voz
	startTime := time.Now()
	lastActivity := time.Now()
	speechDetected := false

	for {
		time.Sleep(50 * time.Millisecond)

		// Verifica timeout máximo
		if time.Since(startTime) > c.maxDuration {
			break
		}

		// Analisa energia do áudio
		c.mu.Lock()
		energy := c.calculateEnergy(c.buffer[max(0, len(c.buffer)-1600):]) // últimos 100ms
		c.mu.Unlock()

		if energy > c.vadThreshold {
			speechDetected = true
			lastActivity = time.Now()
		} else if speechDetected && time.Since(lastActivity) > c.silenceTime {
			// Silêncio após fala detectada
			break
		}

		// Se não detectou fala por muito tempo, reseta
		if !speechDetected && time.Since(startTime) > 5*time.Second {
			return nil, nil
		}
	}

	// Para de escutar
	c.mu.Lock()
	c.isListening = false
	result := make([]float32, len(c.buffer))
	copy(result, c.buffer)
	c.mu.Unlock()

	if len(result) < c.config.SampleRate/2 { // menos de 0.5s
		return nil, nil
	}

	return result, nil
}

// calculateEnergy calcula energia RMS do áudio
func (c *Capture) calculateEnergy(samples []float32) float32 {
	if len(samples) == 0 {
		return 0
	}

	var sum float32
	for _, s := range samples {
		sum += s * s
	}

	return sum / float32(len(samples))
}

// SetVADThreshold ajusta sensibilidade do VAD
func (c *Capture) SetVADThreshold(threshold float32) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.vadThreshold = threshold
}

// Close libera recursos
func (c *Capture) Close() error {
	if c.stream != nil {
		c.stream.Close()
	}
	portaudio.Terminate()
	return nil
}

// helper function
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
