package assistant

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

// ==================== FOCUS MUSIC ====================

// FocusMusic player de música para foco
type FocusMusic struct {
	basePath     string
	playlists    map[string]*Playlist
	currentTrack *Track
	isPlaying    bool
	volume       int // 0-100
	mode         PlayMode
	queue        []*Track
	currentCmd   *exec.Cmd
	mu           sync.RWMutex
	stopChan     chan struct{}

	// Integração com Spotify (opcional)
	spotifyToken string
}

// Playlist playlist de músicas
type Playlist struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Tracks      []*Track `json:"tracks"`
	Category    string   `json:"category"` // focus, relax, ambient, nature
	Duration    time.Duration `json:"duration"`
}

// Track faixa de música
type Track struct {
	ID       string        `json:"id"`
	Name     string        `json:"name"`
	Artist   string        `json:"artist"`
	Path     string        `json:"path"`     // Caminho local
	URL      string        `json:"url"`      // URL de streaming
	Duration time.Duration `json:"duration"`
	BPM      int           `json:"bpm"`      // Batidas por minuto
	Category string        `json:"category"`
}

// PlayMode modo de reprodução
type PlayMode string

const (
	PlayModeSequential PlayMode = "sequential"
	PlayModeShuffle    PlayMode = "shuffle"
	PlayModeRepeat     PlayMode = "repeat"
	PlayModeRepeatOne  PlayMode = "repeat_one"
)

// NewFocusMusic cria player de foco
func NewFocusMusic(basePath string) *FocusMusic {
	fm := &FocusMusic{
		basePath:  basePath,
		playlists: make(map[string]*Playlist),
		volume:    30, // Volume baixo por padrão
		mode:      PlayModeShuffle,
		stopChan:  make(chan struct{}),
	}

	// Cria diretório
	os.MkdirAll(basePath, 0755)

	// Cria playlists padrão
	fm.createDefaultPlaylists()

	return fm
}

// createDefaultPlaylists cria playlists pré-definidas
func (fm *FocusMusic) createDefaultPlaylists() {
	fm.playlists["lofi"] = &Playlist{
		ID:          "lofi",
		Name:        "Lo-Fi Focus",
		Description: "Beats relaxantes para concentração",
		Category:    "focus",
		Tracks:      make([]*Track, 0),
	}

	fm.playlists["ambient"] = &Playlist{
		ID:          "ambient",
		Name:        "Ambient",
		Description: "Música ambiente suave",
		Category:    "ambient",
		Tracks:      make([]*Track, 0),
	}

	fm.playlists["nature"] = &Playlist{
		ID:          "nature",
		Name:        "Nature Sounds",
		Description: "Sons da natureza: chuva, floresta, oceano",
		Category:    "nature",
		Tracks:      make([]*Track, 0),
	}

	fm.playlists["classical"] = &Playlist{
		ID:          "classical",
		Name:        "Classical Focus",
		Description: "Música clássica para concentração",
		Category:    "focus",
		Tracks:      make([]*Track, 0),
	}

	fm.playlists["binaural"] = &Playlist{
		ID:          "binaural",
		Name:        "Binaural Beats",
		Description: "Frequências binaurais para foco profundo",
		Category:    "focus",
		Tracks:      make([]*Track, 0),
	}
}

// ==================== CONTROLES ====================

// Play inicia reprodução
func (fm *FocusMusic) Play(playlistID string) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	playlist, ok := fm.playlists[playlistID]
	if !ok {
		return fmt.Errorf("playlist não encontrada: %s", playlistID)
	}

	if len(playlist.Tracks) == 0 {
		// Tenta reproduzir arquivos locais da categoria
		return fm.playLocalFiles(playlistID)
	}

	// Prepara queue
	fm.queue = make([]*Track, len(playlist.Tracks))
	copy(fm.queue, playlist.Tracks)

	if fm.mode == PlayModeShuffle {
		fm.shuffleQueue()
	}

	fm.isPlaying = true
	go fm.playLoop()

	return nil
}

// PlayFile reproduz arquivo específico
func (fm *FocusMusic) PlayFile(filePath string) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	if fm.currentCmd != nil {
		fm.currentCmd.Process.Kill()
	}

	// Usa ffplay para reproduzir
	fm.currentCmd = exec.Command("ffplay", "-nodisp", "-autoexit", "-volume", fmt.Sprintf("%d", fm.volume), filePath)
	fm.isPlaying = true

	go func() {
		fm.currentCmd.Run()
		fm.mu.Lock()
		fm.isPlaying = false
		fm.mu.Unlock()
	}()

	return nil
}

// playLocalFiles reproduz arquivos locais de uma categoria
func (fm *FocusMusic) playLocalFiles(category string) error {
	// Busca arquivos MP3 na pasta da categoria
	categoryPath := filepath.Join(fm.basePath, category)
	files, err := filepath.Glob(filepath.Join(categoryPath, "*.mp3"))
	if err != nil || len(files) == 0 {
		// Tenta pasta principal
		files, _ = filepath.Glob(filepath.Join(fm.basePath, "*.mp3"))
	}

	if len(files) == 0 {
		return fmt.Errorf("nenhum arquivo de música encontrado")
	}

	// Cria queue com arquivos locais
	fm.queue = make([]*Track, 0)
	for _, file := range files {
		fm.queue = append(fm.queue, &Track{
			ID:   filepath.Base(file),
			Name: filepath.Base(file),
			Path: file,
		})
	}

	if fm.mode == PlayModeShuffle {
		fm.shuffleQueue()
	}

	fm.isPlaying = true
	go fm.playLoop()

	return nil
}

// Pause pausa reprodução
func (fm *FocusMusic) Pause() {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	if fm.currentCmd != nil {
		// Envia sinal de pausa (não disponível no ffplay, então para)
		fm.currentCmd.Process.Kill()
	}
	fm.isPlaying = false
}

// Resume continua reprodução
func (fm *FocusMusic) Resume() {
	if !fm.isPlaying && len(fm.queue) > 0 {
		fm.isPlaying = true
		go fm.playLoop()
	}
}

// Stop para reprodução
func (fm *FocusMusic) Stop() {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	if fm.currentCmd != nil {
		fm.currentCmd.Process.Kill()
	}
	fm.isPlaying = false
	fm.queue = nil
	fm.currentTrack = nil

	select {
	case fm.stopChan <- struct{}{}:
	default:
	}
}

// Next próxima faixa
func (fm *FocusMusic) Next() {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	if fm.currentCmd != nil {
		fm.currentCmd.Process.Kill()
	}
}

// SetVolume ajusta volume
func (fm *FocusMusic) SetVolume(volume int) {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	if volume < 0 {
		volume = 0
	}
	if volume > 100 {
		volume = 100
	}
	fm.volume = volume

	// Ajusta volume do sistema também
	fm.setSystemVolume(volume)
}

// setSystemVolume ajusta volume do sistema
func (fm *FocusMusic) setSystemVolume(percent int) {
	script := fmt.Sprintf(`
		$obj = New-Object -ComObject WScript.Shell
		1..50 | ForEach-Object { $obj.SendKeys([char]174) }
		1..%d | ForEach-Object { $obj.SendKeys([char]175) }
	`, percent/2)

	exec.Command("powershell", "-c", script).Run()
}

// LowerVolume abaixa volume (para quando falar)
func (fm *FocusMusic) LowerVolume() {
	fm.mu.Lock()
	originalVolume := fm.volume
	fm.volume = fm.volume / 3 // Reduz para 1/3
	fm.mu.Unlock()

	fm.setSystemVolume(fm.volume)

	// Restaura após 5 segundos
	go func() {
		time.Sleep(5 * time.Second)
		fm.mu.Lock()
		fm.volume = originalVolume
		fm.mu.Unlock()
		fm.setSystemVolume(fm.volume)
	}()
}

// SetMode define modo de reprodução
func (fm *FocusMusic) SetMode(mode PlayMode) {
	fm.mu.Lock()
	defer fm.mu.Unlock()
	fm.mode = mode

	if mode == PlayModeShuffle && len(fm.queue) > 0 {
		fm.shuffleQueue()
	}
}

// ==================== PLAYBACK LOOP ====================

// playLoop loop de reprodução
func (fm *FocusMusic) playLoop() {
	for {
		fm.mu.RLock()
		if !fm.isPlaying || len(fm.queue) == 0 {
			fm.mu.RUnlock()
			return
		}

		track := fm.queue[0]
		fm.currentTrack = track
		fm.mu.RUnlock()

		// Remove da queue (ou move para o final)
		fm.mu.Lock()
		if fm.mode == PlayModeRepeat || fm.mode == PlayModeShuffle {
			fm.queue = append(fm.queue[1:], track)
		} else if fm.mode == PlayModeRepeatOne {
			// Mantém na frente
		} else {
			fm.queue = fm.queue[1:]
		}
		fm.mu.Unlock()

		// Reproduz
		if track.Path != "" {
			fm.playTrack(track)
		} else if track.URL != "" {
			fm.playURL(track.URL)
		}

		// Verifica se deve parar
		select {
		case <-fm.stopChan:
			return
		default:
		}
	}
}

// playTrack reproduz uma faixa
func (fm *FocusMusic) playTrack(track *Track) {
	fm.mu.Lock()
	fm.currentCmd = exec.Command("ffplay", "-nodisp", "-autoexit",
		"-volume", fmt.Sprintf("%d", fm.volume),
		track.Path)
	fm.mu.Unlock()

	fm.currentCmd.Run()
}

// playURL reproduz URL de streaming
func (fm *FocusMusic) playURL(url string) {
	fm.mu.Lock()
	fm.currentCmd = exec.Command("ffplay", "-nodisp", "-autoexit",
		"-volume", fmt.Sprintf("%d", fm.volume),
		url)
	fm.mu.Unlock()

	fm.currentCmd.Run()
}

// shuffleQueue embaralha a queue
func (fm *FocusMusic) shuffleQueue() {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(fm.queue), func(i, j int) {
		fm.queue[i], fm.queue[j] = fm.queue[j], fm.queue[i]
	})
}

// ==================== PLAYLISTS ====================

// AddTrack adiciona faixa a uma playlist
func (fm *FocusMusic) AddTrack(playlistID string, track *Track) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	playlist, ok := fm.playlists[playlistID]
	if !ok {
		return fmt.Errorf("playlist não encontrada")
	}

	playlist.Tracks = append(playlist.Tracks, track)
	return nil
}

// CreatePlaylist cria nova playlist
func (fm *FocusMusic) CreatePlaylist(name, description, category string) *Playlist {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	id := fmt.Sprintf("playlist_%d", time.Now().UnixNano())
	playlist := &Playlist{
		ID:          id,
		Name:        name,
		Description: description,
		Category:    category,
		Tracks:      make([]*Track, 0),
	}

	fm.playlists[id] = playlist
	return playlist
}

// GetPlaylists retorna todas as playlists
func (fm *FocusMusic) GetPlaylists() []*Playlist {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	playlists := make([]*Playlist, 0, len(fm.playlists))
	for _, p := range fm.playlists {
		playlists = append(playlists, p)
	}
	return playlists
}

// ScanLocalMusic escaneia músicas locais
func (fm *FocusMusic) ScanLocalMusic() int {
	count := 0

	// Escaneia pastas
	categories := []string{"lofi", "ambient", "nature", "classical", "binaural"}

	for _, category := range categories {
		categoryPath := filepath.Join(fm.basePath, category)
		files, _ := filepath.Glob(filepath.Join(categoryPath, "*.mp3"))

		for _, file := range files {
			track := &Track{
				ID:       filepath.Base(file),
				Name:     filepath.Base(file),
				Path:     file,
				Category: category,
			}
			fm.AddTrack(category, track)
			count++
		}
	}

	// Escaneia pasta principal
	files, _ := filepath.Glob(filepath.Join(fm.basePath, "*.mp3"))
	for _, file := range files {
		track := &Track{
			ID:   filepath.Base(file),
			Name: filepath.Base(file),
			Path: file,
		}
		fm.AddTrack("lofi", track) // Adiciona ao lofi por padrão
		count++
	}

	return count
}

// ==================== STATUS ====================

// GetStatus retorna status atual
func (fm *FocusMusic) GetStatus() map[string]interface{} {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	status := map[string]interface{}{
		"is_playing": fm.isPlaying,
		"volume":     fm.volume,
		"mode":       fm.mode,
		"queue_size": len(fm.queue),
	}

	if fm.currentTrack != nil {
		status["current_track"] = fm.currentTrack.Name
	}

	return status
}

// IsPlaying verifica se está tocando
func (fm *FocusMusic) IsPlaying() bool {
	fm.mu.RLock()
	defer fm.mu.RUnlock()
	return fm.isPlaying
}

// GetCurrentTrack retorna faixa atual
func (fm *FocusMusic) GetCurrentTrack() *Track {
	fm.mu.RLock()
	defer fm.mu.RUnlock()
	return fm.currentTrack
}

// ==================== INTEGRAÇÃO COM CORE FLOW ====================

// StartFocusSession inicia música para sessão de foco
func (fm *FocusMusic) StartFocusSession(duration time.Duration) {
	// Escolhe playlist baseada na duração
	var playlistID string
	if duration >= 45*time.Minute {
		playlistID = "lofi" // Sessões longas: lo-fi
	} else if duration >= 25*time.Minute {
		playlistID = "ambient" // Sessões médias: ambient
	} else {
		playlistID = "binaural" // Sessões curtas: binaural
	}

	fm.SetVolume(25) // Volume baixo
	fm.SetMode(PlayModeShuffle)
	fm.Play(playlistID)
}

// StartBreakSession inicia música para pausa
func (fm *FocusMusic) StartBreakSession() {
	fm.SetVolume(35) // Volume um pouco mais alto
	fm.Play("nature")
}

// FadeOut diminui volume gradualmente e para
func (fm *FocusMusic) FadeOut(duration time.Duration) {
	fm.mu.RLock()
	currentVolume := fm.volume
	fm.mu.RUnlock()

	steps := 10
	stepDuration := duration / time.Duration(steps)
	volumeStep := currentVolume / steps

	for i := 0; i < steps; i++ {
		time.Sleep(stepDuration)
		fm.SetVolume(currentVolume - (volumeStep * (i + 1)))
	}

	fm.Stop()
}

// Integration com contexto
func (fm *FocusMusic) PlayWithContext(ctx context.Context, playlistID string) error {
	if err := fm.Play(playlistID); err != nil {
		return err
	}

	go func() {
		<-ctx.Done()
		fm.Stop()
	}()

	return nil
}
