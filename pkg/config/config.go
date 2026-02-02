package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config configuração principal
type Config struct {
	Audio   AudioConfig   `yaml:"audio"`
	STT     STTConfig     `yaml:"stt"`
	TTS     TTSConfig     `yaml:"tts"`
	Models  ModelsConfig  `yaml:"models"`
	Memory  MemoryConfig  `yaml:"memory"`
	Actions ActionsConfig `yaml:"actions"`
}

// AudioConfig configuração de áudio
type AudioConfig struct {
	SampleRate    int     `yaml:"sample_rate"`
	VADThreshold  float32 `yaml:"vad_threshold"`
	SilenceMs     int     `yaml:"silence_ms"`
	MaxDurationMs int     `yaml:"max_duration_ms"`
}

// STTConfig configuração do Speech-to-Text
type STTConfig struct {
	ModelPath string `yaml:"model_path"`
	Language  string `yaml:"language"`
	ModelSize string `yaml:"model_size"` // tiny, base, small, medium, large
}

// TTSConfig configuração do Text-to-Speech
type TTSConfig struct {
	PiperPath  string `yaml:"piper_path"`
	VoicePath  string `yaml:"voice_path"`
	VoiceName  string `yaml:"voice_name"`
	SpeakRate  float32 `yaml:"speak_rate"`
}

// ModelsConfig configuração dos modelos LLM
type ModelsConfig struct {
	LoadAll bool        `yaml:"load_all"` // Carrega todos na inicialização
	Phi     ModelConfig `yaml:"phi"`
	Llama   ModelConfig `yaml:"llama"`
	Qwen    ModelConfig `yaml:"qwen"`
	Vision  ModelConfig `yaml:"vision"`
	Coder   ModelConfig `yaml:"coder"`
}

// ModelConfig configuração de um modelo específico
type ModelConfig struct {
	Name          string `yaml:"name"`
	Path          string `yaml:"path"`
	TokenizerPath string `yaml:"tokenizer_path"`
	MaxTokens     int    `yaml:"max_tokens"`
	Temperature   float32 `yaml:"temperature"`
	SystemPrompt  string `yaml:"system_prompt"`
}

// MemoryConfig configuração de gerenciamento de memória
type MemoryConfig struct {
	UnloadAfter time.Duration `yaml:"unload_after"` // Tempo para descarregar modelo inativo
	Persistent  []string      `yaml:"persistent"`   // Modelos que nunca descarrega
}

// ActionsConfig configuração de ações
type ActionsConfig struct {
	AllowedCommands []string `yaml:"allowed_commands"`
	EmailEnabled    bool     `yaml:"email_enabled"`
	BrowserEnabled  bool     `yaml:"browser_enabled"`
}

// Load carrega configuração de um arquivo YAML
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// Aplica defaults
	cfg.applyDefaults()

	return &cfg, nil
}

// Default retorna configuração padrão
func Default() *Config {
	cfg := &Config{}
	cfg.applyDefaults()
	return cfg
}

// applyDefaults aplica valores padrão
func (c *Config) applyDefaults() {
	// Audio
	if c.Audio.SampleRate == 0 {
		c.Audio.SampleRate = 16000
	}
	if c.Audio.VADThreshold == 0 {
		c.Audio.VADThreshold = 0.01
	}
	if c.Audio.SilenceMs == 0 {
		c.Audio.SilenceMs = 1000 // 1 segundo
	}
	if c.Audio.MaxDurationMs == 0 {
		c.Audio.MaxDurationMs = 30000 // 30 segundos
	}

	// STT
	if c.STT.Language == "" {
		c.STT.Language = "pt"
	}
	if c.STT.ModelSize == "" {
		c.STT.ModelSize = "medium"
	}
	if c.STT.ModelPath == "" {
		c.STT.ModelPath = "models/whisper-medium.onnx"
	}

	// TTS
	if c.TTS.VoiceName == "" {
		c.TTS.VoiceName = "pt_BR-faber-medium"
	}
	if c.TTS.SpeakRate == 0 {
		c.TTS.SpeakRate = 1.0
	}

	// Models
	if c.Models.Phi.Name == "" {
		c.Models.Phi.Name = "phi-3.5-mini"
		c.Models.Phi.Path = "models/phi-3.5-mini.onnx"
		c.Models.Phi.MaxTokens = 512
		c.Models.Phi.Temperature = 0.7
	}
	if c.Models.Llama.Name == "" {
		c.Models.Llama.Name = "llama-3.2-3b"
		c.Models.Llama.Path = "models/llama-3.2-3b.onnx"
		c.Models.Llama.MaxTokens = 1024
		c.Models.Llama.Temperature = 0.7
	}
	if c.Models.Qwen.Name == "" {
		c.Models.Qwen.Name = "qwen-2.5-3b"
		c.Models.Qwen.Path = "models/qwen-2.5-3b.onnx"
		c.Models.Qwen.MaxTokens = 512
		c.Models.Qwen.Temperature = 0.3 // Mais determinístico para ações
	}
	if c.Models.Vision.Name == "" {
		c.Models.Vision.Name = "minicpm-v"
		c.Models.Vision.Path = "models/minicpm-v.onnx"
		c.Models.Vision.MaxTokens = 256
	}
	if c.Models.Coder.Name == "" {
		c.Models.Coder.Name = "qwen-coder-3b"
		c.Models.Coder.Path = "models/qwen-coder-3b.onnx"
		c.Models.Coder.MaxTokens = 1024
		c.Models.Coder.Temperature = 0.2 // Bem determinístico para código
	}

	// Memory
	if c.Memory.UnloadAfter == 0 {
		c.Memory.UnloadAfter = 5 * time.Minute
	}
	if len(c.Memory.Persistent) == 0 {
		c.Memory.Persistent = []string{"whisper", "phi"} // Sempre na memória
	}

	// Actions
	if len(c.Actions.AllowedCommands) == 0 {
		c.Actions.AllowedCommands = []string{"dir", "echo", "date", "time", "hostname"}
	}
	c.Actions.EmailEnabled = true
	c.Actions.BrowserEnabled = true
}

// Save salva configuração em arquivo YAML
func (c *Config) Save(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
