package tts

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/JoseRFJuniorLLMs/NPU-IA/pkg/config"
)

// Piper implementa Text-to-Speech usando Piper
type Piper struct {
	config    config.TTSConfig
	voicePath string
	piperPath string
}

// New cria uma nova instância do TTS
func New(cfg config.TTSConfig) (*Piper, error) {
	// Verifica se Piper está instalado
	piperPath := cfg.PiperPath
	if piperPath == "" {
		piperPath = "piper" // Assume que está no PATH
	}

	// Verifica se o modelo de voz existe
	if cfg.VoicePath != "" {
		if _, err := os.Stat(cfg.VoicePath); os.IsNotExist(err) {
			return nil, fmt.Errorf("modelo de voz não encontrado: %s", cfg.VoicePath)
		}
	}

	return &Piper{
		config:    cfg,
		voicePath: cfg.VoicePath,
		piperPath: piperPath,
	}, nil
}

// Speak converte texto em fala e reproduz
func (p *Piper) Speak(text string) error {
	if text == "" {
		return nil
	}

	// Cria arquivo temporário para output
	tmpDir := os.TempDir()
	wavFile := filepath.Join(tmpDir, "npu_ia_speech.wav")

	// Executa Piper para gerar WAV
	cmd := exec.Command(p.piperPath,
		"--model", p.voicePath,
		"--output_file", wavFile,
	)

	// Passa texto via stdin
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("erro ao iniciar Piper: %w", err)
	}

	stdin.Write([]byte(text))
	stdin.Close()

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("erro no Piper: %w", err)
	}

	// Reproduz o áudio usando PowerShell (Windows)
	playCmd := exec.Command("powershell", "-c",
		fmt.Sprintf(`(New-Object Media.SoundPlayer '%s').PlaySync()`, wavFile))

	if err := playCmd.Run(); err != nil {
		return fmt.Errorf("erro ao reproduzir áudio: %w", err)
	}

	// Remove arquivo temporário
	os.Remove(wavFile)

	return nil
}

// SpeakAsync reproduz áudio de forma assíncrona
func (p *Piper) SpeakAsync(text string) error {
	go func() {
		if err := p.Speak(text); err != nil {
			fmt.Printf("Erro no TTS: %v\n", err)
		}
	}()
	return nil
}

// SetVoice troca o modelo de voz
func (p *Piper) SetVoice(voicePath string) error {
	if _, err := os.Stat(voicePath); os.IsNotExist(err) {
		return fmt.Errorf("modelo de voz não encontrado: %s", voicePath)
	}
	p.voicePath = voicePath
	return nil
}

// GetAvailableVoices lista vozes disponíveis
func (p *Piper) GetAvailableVoices() ([]string, error) {
	voicesDir := filepath.Dir(p.voicePath)
	entries, err := os.ReadDir(voicesDir)
	if err != nil {
		return nil, err
	}

	voices := []string{}
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) == ".onnx" {
			voices = append(voices, entry.Name())
		}
	}

	return voices, nil
}

// Close libera recursos
func (p *Piper) Close() error {
	return nil
}
