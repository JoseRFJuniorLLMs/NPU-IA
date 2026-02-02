package productivity

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// WimHofBreathing exercício de respiração Wim Hof
type WimHofBreathing struct {
	config    WimHofConfig
	state     WimHofState
	round     int
	breath    int
	tts       TTSInterface
	onUpdate  func(state WimHofState, info string)
	mu        sync.RWMutex
	stopChan  chan struct{}
	pauseChan chan struct{}
}

// WimHofConfig configuração do exercício
type WimHofConfig struct {
	Rounds          int           `json:"rounds"`           // Número de rounds (padrão: 3)
	BreathsPerRound int           `json:"breaths_per_round"` // Respirações por round (padrão: 30)
	BreathInTime    time.Duration `json:"breath_in_time"`   // Tempo para inspirar
	BreathOutTime   time.Duration `json:"breath_out_time"`  // Tempo para expirar
	RetentionTime   time.Duration `json:"retention_time"`   // Tempo de retenção (padrão: progressivo)
	RecoveryTime    time.Duration `json:"recovery_time"`    // Tempo de recuperação (15s)
	VoiceGuided     bool          `json:"voice_guided"`
}

// WimHofState estado do exercício
type WimHofState string

const (
	WHStateIdle       WimHofState = "idle"
	WHStateBreathIn   WimHofState = "breath_in"
	WHStateBreathOut  WimHofState = "breath_out"
	WHStateRetention  WimHofState = "retention"
	WHStateRecovery   WimHofState = "recovery"
	WHStateComplete   WimHofState = "complete"
	WHStatePaused     WimHofState = "paused"
)

// WimHofSession sessão de exercício
type WimHofSession struct {
	StartTime      time.Time       `json:"start_time"`
	EndTime        time.Time       `json:"end_time"`
	Rounds         int             `json:"rounds"`
	RetentionTimes []time.Duration `json:"retention_times"`
	Completed      bool            `json:"completed"`
}

// DefaultWimHofConfig configuração padrão
func DefaultWimHofConfig() WimHofConfig {
	return WimHofConfig{
		Rounds:          3,
		BreathsPerRound: 30,
		BreathInTime:    2 * time.Second,
		BreathOutTime:   2 * time.Second,
		RetentionTime:   90 * time.Second, // Base, aumenta por round
		RecoveryTime:    15 * time.Second,
		VoiceGuided:     true,
	}
}

// NewWimHofBreathing cria novo exercício
func NewWimHofBreathing(config WimHofConfig, tts TTSInterface) *WimHofBreathing {
	return &WimHofBreathing{
		config:    config,
		state:     WHStateIdle,
		tts:       tts,
		stopChan:  make(chan struct{}),
		pauseChan: make(chan struct{}),
	}
}

// Start inicia sessão
func (w *WimHofBreathing) Start(ctx context.Context) (*WimHofSession, error) {
	session := &WimHofSession{
		StartTime:      time.Now(),
		RetentionTimes: make([]time.Duration, 0),
	}

	w.speak("Iniciando respiração Wim Hof. Encontre uma posição confortável.")
	time.Sleep(3 * time.Second)

	w.speak(fmt.Sprintf("Faremos %d rounds de %d respirações profundas.",
		w.config.Rounds, w.config.BreathsPerRound))
	time.Sleep(2 * time.Second)

	// Executa rounds
	for round := 1; round <= w.config.Rounds; round++ {
		w.mu.Lock()
		w.round = round
		w.mu.Unlock()

		select {
		case <-ctx.Done():
			return session, ctx.Err()
		case <-w.stopChan:
			return session, nil
		default:
		}

		w.speak(fmt.Sprintf("Round %d. Vamos começar.", round))
		time.Sleep(2 * time.Second)

		// Fase de respiração
		retentionTime := w.breathingPhase(ctx, round)
		if retentionTime > 0 {
			session.RetentionTimes = append(session.RetentionTimes, retentionTime)
		}

		if round < w.config.Rounds {
			w.speak("Preparando para o próximo round.")
			time.Sleep(3 * time.Second)
		}
	}

	w.speak("Sessão completa! Excelente trabalho.")

	session.EndTime = time.Now()
	session.Rounds = w.config.Rounds
	session.Completed = true

	w.mu.Lock()
	w.state = WHStateComplete
	w.mu.Unlock()

	return session, nil
}

// breathingPhase fase de respiração
func (w *WimHofBreathing) breathingPhase(ctx context.Context, round int) time.Duration {
	// 30 respirações profundas
	for breath := 1; breath <= w.config.BreathsPerRound; breath++ {
		w.mu.Lock()
		w.breath = breath
		w.state = WHStateBreathIn
		w.mu.Unlock()

		select {
		case <-ctx.Done():
			return 0
		case <-w.stopChan:
			return 0
		default:
		}

		// Inspirar
		if w.config.VoiceGuided {
			if breath%5 == 1 {
				w.speak(fmt.Sprintf("%d", breath))
			}
		}

		if w.onUpdate != nil {
			w.onUpdate(WHStateBreathIn, fmt.Sprintf("Inspire - %d/%d", breath, w.config.BreathsPerRound))
		}
		time.Sleep(w.config.BreathInTime)

		// Expirar
		w.mu.Lock()
		w.state = WHStateBreathOut
		w.mu.Unlock()

		if w.onUpdate != nil {
			w.onUpdate(WHStateBreathOut, fmt.Sprintf("Expire - %d/%d", breath, w.config.BreathsPerRound))
		}
		time.Sleep(w.config.BreathOutTime)
	}

	// Retenção
	w.speak("Expire totalmente e segure.")

	w.mu.Lock()
	w.state = WHStateRetention
	w.mu.Unlock()

	// Tempo de retenção aumenta por round
	retentionTime := w.config.RetentionTime + time.Duration(round-1)*30*time.Second
	retentionStart := time.Now()

	// Timer de retenção com contagem
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	retentionTimer := time.NewTimer(retentionTime)
	defer retentionTimer.Stop()

	elapsed := time.Duration(0)
	for {
		select {
		case <-ctx.Done():
			return time.Since(retentionStart)
		case <-w.stopChan:
			return time.Since(retentionStart)
		case <-ticker.C:
			elapsed += 30 * time.Second
			if w.onUpdate != nil {
				w.onUpdate(WHStateRetention, fmt.Sprintf("Retenção: %v", elapsed))
			}
			w.speak(fmt.Sprintf("%d segundos", int(elapsed.Seconds())))
		case <-retentionTimer.C:
			actualRetention := time.Since(retentionStart)
			w.speak(fmt.Sprintf("Inspire e segure por 15 segundos. Retenção: %d segundos.",
				int(actualRetention.Seconds())))

			// Recuperação
			w.mu.Lock()
			w.state = WHStateRecovery
			w.mu.Unlock()

			if w.onUpdate != nil {
				w.onUpdate(WHStateRecovery, "Recuperação - 15 segundos")
			}
			time.Sleep(w.config.RecoveryTime)

			w.speak("Solte o ar.")
			return actualRetention
		}
	}
}

// Stop para exercício
func (w *WimHofBreathing) Stop() {
	select {
	case w.stopChan <- struct{}{}:
	default:
	}
}

// GetStatus retorna status atual
func (w *WimHofBreathing) GetStatus() (WimHofState, int, int) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.state, w.round, w.breath
}

// speak fala texto se TTS disponível
func (w *WimHofBreathing) speak(text string) {
	if w.tts != nil && w.config.VoiceGuided {
		w.tts.Speak(text)
	}
}

// ==================== BOX BREATHING ====================

// BoxBreathing respiração em caixa (4-4-4-4)
type BoxBreathing struct {
	duration time.Duration // Duração de cada fase
	cycles   int           // Número de ciclos
	tts      TTSInterface
}

// NewBoxBreathing cria respiração em caixa
func NewBoxBreathing(phaseDuration time.Duration, cycles int, tts TTSInterface) *BoxBreathing {
	return &BoxBreathing{
		duration: phaseDuration,
		cycles:   cycles,
		tts:      tts,
	}
}

// Start inicia exercício
func (b *BoxBreathing) Start(ctx context.Context) error {
	if b.tts != nil {
		b.tts.Speak(fmt.Sprintf("Iniciando respiração em caixa. %d segundos para cada fase, %d ciclos.",
			int(b.duration.Seconds()), b.cycles))
	}
	time.Sleep(2 * time.Second)

	for cycle := 1; cycle <= b.cycles; cycle++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if b.tts != nil {
			b.tts.Speak(fmt.Sprintf("Ciclo %d", cycle))
		}

		// Inspirar
		if b.tts != nil {
			b.tts.Speak("Inspire")
		}
		time.Sleep(b.duration)

		// Segurar
		if b.tts != nil {
			b.tts.Speak("Segure")
		}
		time.Sleep(b.duration)

		// Expirar
		if b.tts != nil {
			b.tts.Speak("Expire")
		}
		time.Sleep(b.duration)

		// Segurar vazio
		if b.tts != nil {
			b.tts.Speak("Segure vazio")
		}
		time.Sleep(b.duration)
	}

	if b.tts != nil {
		b.tts.Speak("Exercício completo. Bem feito!")
	}

	return nil
}

// ==================== 4-7-8 BREATHING ====================

// Breathing478 técnica de respiração 4-7-8
type Breathing478 struct {
	cycles int
	tts    TTSInterface
}

// NewBreathing478 cria exercício 4-7-8
func NewBreathing478(cycles int, tts TTSInterface) *Breathing478 {
	return &Breathing478{
		cycles: cycles,
		tts:    tts,
	}
}

// Start inicia exercício
func (b *Breathing478) Start(ctx context.Context) error {
	if b.tts != nil {
		b.tts.Speak("Iniciando respiração 4-7-8. Esta técnica ajuda a relaxar e dormir.")
	}
	time.Sleep(2 * time.Second)

	for cycle := 1; cycle <= b.cycles; cycle++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Inspirar por 4 segundos
		if b.tts != nil {
			b.tts.Speak("Inspire pelo nariz")
		}
		time.Sleep(4 * time.Second)

		// Segurar por 7 segundos
		if b.tts != nil {
			b.tts.Speak("Segure")
		}
		time.Sleep(7 * time.Second)

		// Expirar por 8 segundos
		if b.tts != nil {
			b.tts.Speak("Expire pela boca lentamente")
		}
		time.Sleep(8 * time.Second)
	}

	if b.tts != nil {
		b.tts.Speak("Exercício completo. Sinta-se relaxado.")
	}

	return nil
}
