package productivity

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

// AudioPlayer reproduz áudios
type AudioPlayer struct {
	basePath  string
	current   *exec.Cmd
	isPlaying bool
	mu        sync.Mutex
}

// AudioFile arquivo de áudio disponível
type AudioFile struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	Duration time.Duration `json:"duration"`
}

// NewAudioPlayer cria player de áudio
func NewAudioPlayer(basePath string) *AudioPlayer {
	return &AudioPlayer{
		basePath: basePath,
	}
}

// Play reproduz arquivo de áudio
func (ap *AudioPlayer) Play(filename string) error {
	ap.mu.Lock()
	defer ap.mu.Unlock()

	// Para qualquer áudio em reprodução
	if ap.current != nil && ap.isPlaying {
		ap.current.Process.Kill()
	}

	// Resolve caminho
	path := filename
	if !filepath.IsAbs(filename) {
		path = filepath.Join(ap.basePath, filename)
	}

	// Verifica se arquivo existe
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("arquivo não encontrado: %s", path)
	}

	// Usa PowerShell para reproduzir no Windows
	ap.current = exec.Command("powershell", "-c",
		fmt.Sprintf(`(New-Object Media.SoundPlayer '%s').PlaySync()`, path))

	ap.isPlaying = true

	go func() {
		ap.current.Run()
		ap.mu.Lock()
		ap.isPlaying = false
		ap.mu.Unlock()
	}()

	return nil
}

// PlayMP3 reproduz arquivo MP3 usando ffplay ou Windows Media Player
func (ap *AudioPlayer) PlayMP3(filename string) error {
	ap.mu.Lock()
	defer ap.mu.Unlock()

	if ap.current != nil && ap.isPlaying {
		ap.current.Process.Kill()
	}

	path := filename
	if !filepath.IsAbs(filename) {
		path = filepath.Join(ap.basePath, filename)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("arquivo não encontrado: %s", path)
	}

	// Tenta usar ffplay (se FFmpeg instalado)
	ap.current = exec.Command("ffplay", "-nodisp", "-autoexit", path)
	err := ap.current.Start()

	if err != nil {
		// Fallback: usa Windows Media Player via PowerShell
		script := fmt.Sprintf(`
			Add-Type -AssemblyName presentationCore
			$player = New-Object System.Windows.Media.MediaPlayer
			$player.Open('%s')
			$player.Play()
			Start-Sleep -Seconds 600
		`, path)
		ap.current = exec.Command("powershell", "-c", script)
		err = ap.current.Start()
	}

	if err != nil {
		// Último fallback: abre com player padrão
		ap.current = exec.Command("cmd", "/c", "start", path)
		err = ap.current.Start()
	}

	if err == nil {
		ap.isPlaying = true
	}

	return err
}

// PlayAsync reproduz em background
func (ap *AudioPlayer) PlayAsync(filename string) error {
	go ap.PlayMP3(filename)
	return nil
}

// Stop para reprodução
func (ap *AudioPlayer) Stop() {
	ap.mu.Lock()
	defer ap.mu.Unlock()

	if ap.current != nil && ap.isPlaying {
		ap.current.Process.Kill()
		ap.isPlaying = false
	}
}

// IsPlaying verifica se está tocando
func (ap *AudioPlayer) IsPlaying() bool {
	ap.mu.Lock()
	defer ap.mu.Unlock()
	return ap.isPlaying
}

// FadeOut reduz volume gradualmente (precisa de controle de volume)
func (ap *AudioPlayer) FadeOut(duration time.Duration) {
	// TODO: Implementar fade out
	time.Sleep(duration)
	ap.Stop()
}

// SetVolume ajusta volume do sistema (Windows)
func (ap *AudioPlayer) SetVolume(percent int) error {
	// Usa nircmd ou PowerShell
	script := fmt.Sprintf(`
		$wshShell = New-Object -ComObject WScript.Shell
		1..50 | ForEach-Object { $wshShell.SendKeys([char]174) }
		1..%d | ForEach-Object { $wshShell.SendKeys([char]175) }
	`, percent/2)

	cmd := exec.Command("powershell", "-c", script)
	return cmd.Run()
}

// ListAudioFiles lista arquivos de áudio disponíveis
func (ap *AudioPlayer) ListAudioFiles() ([]AudioFile, error) {
	files := make([]AudioFile, 0)

	patterns := []string{"*.mp3", "*.wav", "*.ogg", "*.flac"}

	for _, pattern := range patterns {
		matches, err := filepath.Glob(filepath.Join(ap.basePath, pattern))
		if err != nil {
			continue
		}

		for _, match := range matches {
			files = append(files, AudioFile{
				Name: filepath.Base(match),
				Path: match,
			})
		}
	}

	return files, nil
}
