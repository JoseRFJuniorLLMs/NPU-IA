package productivity

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

// CoreFlow fluxo principal de produtividade
// 45 min trabalho → Wim Hof → 45 min trabalho → ...
type CoreFlow struct {
	pomodoro   *PomodoroTimer
	wimhof     *WimHofBreathing
	zettel     *Zettelkasten
	alarm      *AlarmClock
	tts        TTSInterface
	audio      *AudioPlayer
	config     CoreFlowConfig
	state      FlowState
	session    *FlowSession
	mu         sync.RWMutex
	stopChan   chan struct{}
	onUpdate   func(update FlowUpdate)
}

// CoreFlowConfig configuração do fluxo
type CoreFlowConfig struct {
	// Pomodoro
	WorkDuration      time.Duration `json:"work_duration"`       // 45 min
	ShortBreakType    BreakType     `json:"short_break_type"`    // wimhof
	LongBreakDuration time.Duration `json:"long_break_duration"` // 15 min
	SessionsUntilLong int           `json:"sessions_until_long"` // 4

	// Wim Hof
	WimHofRounds int    `json:"wimhof_rounds"`     // 3 rounds
	WimHofAudio  string `json:"wimhof_audio"`      // Arquivo de áudio (winhof.mp3)

	// Áudio
	AudioPath          string `json:"audio_path"`           // Pasta de áudios
	WorkStartSound     string `json:"work_start_sound"`     // Som ao iniciar trabalho
	BreakStartSound    string `json:"break_start_sound"`    // Som ao iniciar pausa
	SessionCompleteSound string `json:"session_complete_sound"` // Som ao completar sessão

	// Extras
	AutoStartNext     bool `json:"auto_start_next"`
	QuickNoteEnabled  bool `json:"quick_note_enabled"` // Captura nota rápida antes da pausa
	DailyGoalSessions int  `json:"daily_goal_sessions"`
}

// BreakType tipo de pausa
type BreakType string

const (
	BreakTypeWimHof    BreakType = "wimhof"
	BreakTypeBoxBreath BreakType = "box_breathing"
	BreakType478       BreakType = "478_breathing"
	BreakTypeRest      BreakType = "rest"
)

// FlowState estado do fluxo
type FlowState string

const (
	FlowStateIdle      FlowState = "idle"
	FlowStateWorking   FlowState = "working"
	FlowStateBreathing FlowState = "breathing"
	FlowStateLongBreak FlowState = "long_break"
	FlowStatePaused    FlowState = "paused"
	FlowStateComplete  FlowState = "complete"
)

// FlowSession sessão do dia
type FlowSession struct {
	Date             time.Time        `json:"date"`
	SessionsCompleted int             `json:"sessions_completed"`
	TotalWorkTime    time.Duration    `json:"total_work_time"`
	TotalBreakTime   time.Duration    `json:"total_break_time"`
	WimHofSessions   []WimHofSession  `json:"wimhof_sessions"`
	Notes            []string         `json:"notes"` // IDs das notas criadas
	StartTime        time.Time        `json:"start_time"`
	EndTime          time.Time        `json:"end_time"`
}

// FlowUpdate atualização do estado
type FlowUpdate struct {
	State          FlowState     `json:"state"`
	Session        int           `json:"session"`
	TimeRemaining  time.Duration `json:"time_remaining"`
	Message        string        `json:"message"`
	CurrentTask    string        `json:"current_task"`
	Progress       float64       `json:"progress"` // 0-100
}

// DefaultCoreFlowConfig configuração padrão
func DefaultCoreFlowConfig() CoreFlowConfig {
	return CoreFlowConfig{
		WorkDuration:         45 * time.Minute,
		ShortBreakType:       BreakTypeWimHof,
		LongBreakDuration:    15 * time.Minute,
		SessionsUntilLong:    4,
		WimHofRounds:         3,
		WimHofAudio:          "winhof.mp3", // Áudio guiado do Wim Hof
		AudioPath:            "audio",
		WorkStartSound:       "work_start.mp3",
		BreakStartSound:      "break_start.mp3",
		SessionCompleteSound: "complete.mp3",
		AutoStartNext:        false,
		QuickNoteEnabled:     true,
		DailyGoalSessions:    8, // 8 sessões = 6 horas de foco
	}
}

// NewCoreFlow cria novo fluxo
func NewCoreFlow(config CoreFlowConfig, tts TTSInterface, zettel *Zettelkasten) *CoreFlow {
	cf := &CoreFlow{
		config:   config,
		tts:      tts,
		zettel:   zettel,
		state:    FlowStateIdle,
		stopChan: make(chan struct{}),
	}

	// Configura Pomodoro
	pomodoroConfig := PomodoroConfig{
		WorkDuration:      config.WorkDuration,
		ShortBreak:        10 * time.Minute, // Tempo do Wim Hof
		LongBreak:         config.LongBreakDuration,
		SessionsUntilLong: config.SessionsUntilLong,
		AutoStart:         false,
	}
	cf.pomodoro = NewPomodoroTimer(pomodoroConfig, tts)

	// Configura Wim Hof
	wimhofConfig := WimHofConfig{
		Rounds:          config.WimHofRounds,
		BreathsPerRound: 30,
		BreathInTime:    2 * time.Second,
		BreathOutTime:   2 * time.Second,
		RetentionTime:   90 * time.Second,
		RecoveryTime:    15 * time.Second,
		VoiceGuided:     true,
	}
	cf.wimhof = NewWimHofBreathing(wimhofConfig, tts)

	// Configura Alarme
	cf.alarm = NewAlarmClock(tts)

	// Configura Audio Player
	cf.audio = NewAudioPlayer(config.AudioPath)

	return cf
}

// Start inicia o fluxo de produtividade
func (cf *CoreFlow) Start(ctx context.Context, task string) error {
	cf.mu.Lock()
	cf.session = &FlowSession{
		Date:      time.Now(),
		StartTime: time.Now(),
		Notes:     make([]string, 0),
	}
	cf.state = FlowStateWorking
	cf.mu.Unlock()

	cf.speak("Iniciando modo de produtividade. 45 minutos de foco.")
	cf.speak(fmt.Sprintf("Tarefa: %s", task))
	time.Sleep(2 * time.Second)
	cf.speak("Que comece o deep work!")

	return cf.runLoop(ctx, task)
}

// runLoop loop principal do fluxo
func (cf *CoreFlow) runLoop(ctx context.Context, task string) error {
	sessionNum := 0

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-cf.stopChan:
			cf.complete()
			return nil
		default:
		}

		sessionNum++

		// ========== FASE DE TRABALHO ==========
		cf.mu.Lock()
		cf.state = FlowStateWorking
		cf.mu.Unlock()

		cf.notifyUpdate(FlowUpdate{
			State:         FlowStateWorking,
			Session:       sessionNum,
			TimeRemaining: cf.config.WorkDuration,
			CurrentTask:   task,
			Message:       fmt.Sprintf("Sessão %d - FOCO!", sessionNum),
		})

		// Timer de trabalho
		workComplete := cf.workPhase(ctx, sessionNum, task)
		if !workComplete {
			return nil
		}

		cf.mu.Lock()
		cf.session.SessionsCompleted++
		cf.session.TotalWorkTime += cf.config.WorkDuration
		cf.mu.Unlock()

		// Verifica se atingiu meta diária
		if cf.session.SessionsCompleted >= cf.config.DailyGoalSessions {
			cf.speak(fmt.Sprintf("Parabéns! Você completou %d sessões hoje. Meta diária atingida!",
				cf.session.SessionsCompleted))
		}

		// ========== CAPTURA RÁPIDA (OPCIONAL) ==========
		if cf.config.QuickNoteEnabled {
			cf.speak("Alguma nota rápida antes da pausa?")
			// TODO: Aguardar input de voz para nota
		}

		// ========== FASE DE PAUSA ==========
		isLongBreak := sessionNum%cf.config.SessionsUntilLong == 0

		if isLongBreak {
			cf.mu.Lock()
			cf.state = FlowStateLongBreak
			cf.mu.Unlock()

			cf.speak(fmt.Sprintf("Excelente! %d sessões completas. Pausa longa de %d minutos.",
				sessionNum, int(cf.config.LongBreakDuration.Minutes())))

			cf.notifyUpdate(FlowUpdate{
				State:         FlowStateLongBreak,
				Session:       sessionNum,
				TimeRemaining: cf.config.LongBreakDuration,
				Message:       "Pausa longa - Descanse!",
			})

			time.Sleep(cf.config.LongBreakDuration)
		} else {
			cf.mu.Lock()
			cf.state = FlowStateBreathing
			cf.mu.Unlock()

			cf.notifyUpdate(FlowUpdate{
				State:   FlowStateBreathing,
				Session: sessionNum,
				Message: "Respiração Wim Hof",
			})

			// Toca áudio do Wim Hof se configurado
			if cf.config.WimHofAudio != "" && cf.audio != nil {
				cf.speak("Iniciando sessão de respiração Wim Hof com áudio guiado.")
				cf.audio.PlayMP3(cf.config.WimHofAudio)

				// Aguarda o áudio terminar (aproximadamente 11 minutos para 3 rounds)
				wimhofDuration := time.Duration(cf.config.WimHofRounds) * 4 * time.Minute
				time.Sleep(wimhofDuration)
				cf.audio.Stop()

				cf.mu.Lock()
				cf.session.WimHofSessions = append(cf.session.WimHofSessions, WimHofSession{
					StartTime: time.Now().Add(-wimhofDuration),
					EndTime:   time.Now(),
					Rounds:    cf.config.WimHofRounds,
					Completed: true,
				})
				cf.session.TotalBreakTime += wimhofDuration
				cf.mu.Unlock()
			} else {
				// Wim Hof guiado por TTS
				wimhofSession, err := cf.wimhof.Start(ctx)
				if err != nil {
					return err
				}

				cf.mu.Lock()
				cf.session.WimHofSessions = append(cf.session.WimHofSessions, *wimhofSession)
				cf.session.TotalBreakTime += time.Since(wimhofSession.StartTime)
				cf.mu.Unlock()
			}
		}

		// ========== PRÓXIMA SESSÃO ==========
		if cf.config.AutoStartNext {
			cf.speak("Iniciando próxima sessão em 10 segundos.")
			time.Sleep(10 * time.Second)
		} else {
			cf.speak("Pronto para a próxima sessão? Diga 'iniciar' quando estiver pronto.")
			// TODO: Aguardar comando de voz
			time.Sleep(5 * time.Second)
		}
	}
}

// workPhase fase de trabalho
func (cf *CoreFlow) workPhase(ctx context.Context, session int, task string) bool {
	duration := cf.config.WorkDuration
	startTime := time.Now()
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	// Avisos em marcos específicos
	milestones := map[time.Duration]string{
		30 * time.Minute: "30 minutos. Você está indo bem!",
		15 * time.Minute: "15 minutos restantes.",
		5 * time.Minute:  "5 minutos restantes. Finalize o que está fazendo.",
		1 * time.Minute:  "1 minuto restante.",
	}

	announcedMilestones := make(map[time.Duration]bool)

	for {
		select {
		case <-ctx.Done():
			return false
		case <-cf.stopChan:
			return false
		case now := <-ticker.C:
			elapsed := now.Sub(startTime)
			remaining := duration - elapsed

			if remaining <= 0 {
				cf.speak("Tempo! Sessão de trabalho completa.")
				return true
			}

			// Verifica milestones
			for milestone, message := range milestones {
				if remaining <= milestone && !announcedMilestones[milestone] {
					cf.speak(message)
					announcedMilestones[milestone] = true
				}
			}

			// Atualiza progresso
			progress := float64(elapsed) / float64(duration) * 100
			cf.notifyUpdate(FlowUpdate{
				State:         FlowStateWorking,
				Session:       session,
				TimeRemaining: remaining,
				CurrentTask:   task,
				Progress:      progress,
			})
		}
	}
}

// Stop para o fluxo
func (cf *CoreFlow) Stop() {
	select {
	case cf.stopChan <- struct{}{}:
	default:
	}
}

// Pause pausa o fluxo
func (cf *CoreFlow) Pause() {
	cf.mu.Lock()
	cf.state = FlowStatePaused
	cf.mu.Unlock()
	cf.speak("Fluxo pausado.")
}

// Resume continua o fluxo
func (cf *CoreFlow) Resume() {
	cf.mu.Lock()
	cf.state = FlowStateWorking
	cf.mu.Unlock()
	cf.speak("Continuando...")
}

// complete finaliza a sessão
func (cf *CoreFlow) complete() {
	cf.mu.Lock()
	cf.session.EndTime = time.Now()
	cf.state = FlowStateComplete
	cf.mu.Unlock()

	totalTime := cf.session.EndTime.Sub(cf.session.StartTime)

	cf.speak(fmt.Sprintf(
		"Sessão de produtividade finalizada. "+
		"Você completou %d sessões de foco, totalizando %d minutos de trabalho. "+
		"Excelente trabalho!",
		cf.session.SessionsCompleted,
		int(cf.session.TotalWorkTime.Minutes()),
	))

	// Salva resumo como nota se Zettelkasten disponível
	if cf.zettel != nil {
		summary := fmt.Sprintf(`# Resumo da Sessão - %s

## Estatísticas
- Sessões de foco: %d
- Tempo de trabalho: %s
- Tempo de pausa: %s
- Duração total: %s

## Respirações Wim Hof
%d sessões realizadas

## Notas capturadas
%d notas

---
Gerado automaticamente pelo NPU-IA`,
			cf.session.Date.Format("02/01/2006"),
			cf.session.SessionsCompleted,
			cf.session.TotalWorkTime.Round(time.Minute),
			cf.session.TotalBreakTime.Round(time.Minute),
			totalTime.Round(time.Minute),
			len(cf.session.WimHofSessions),
			len(cf.session.Notes),
		)

		cf.zettel.CreateNote(
			fmt.Sprintf("Produtividade: %s", cf.session.Date.Format("02/01/2006")),
			summary,
			[]string{"produtividade", "resumo-diário"},
			NoteTypeIndex,
		)
	}
}

// AddNote adiciona nota rápida à sessão
func (cf *CoreFlow) AddNote(content string) error {
	if cf.zettel == nil {
		return fmt.Errorf("Zettelkasten não configurado")
	}

	note, err := cf.zettel.QuickCapture(content, "sessão-trabalho")
	if err != nil {
		return err
	}

	cf.mu.Lock()
	cf.session.Notes = append(cf.session.Notes, note.ID)
	cf.mu.Unlock()

	cf.speak("Nota capturada.")
	return nil
}

// GetStatus retorna status atual
func (cf *CoreFlow) GetStatus() FlowUpdate {
	cf.mu.RLock()
	defer cf.mu.RUnlock()

	update := FlowUpdate{
		State:   cf.state,
		Session: 0,
	}

	if cf.session != nil {
		update.Session = cf.session.SessionsCompleted
	}

	return update
}

// GetSession retorna sessão atual
func (cf *CoreFlow) GetSession() *FlowSession {
	cf.mu.RLock()
	defer cf.mu.RUnlock()
	return cf.session
}

// SetOnUpdate define callback de atualização
func (cf *CoreFlow) SetOnUpdate(fn func(FlowUpdate)) {
	cf.onUpdate = fn
}

// notifyUpdate notifica atualização
func (cf *CoreFlow) notifyUpdate(update FlowUpdate) {
	if cf.onUpdate != nil {
		cf.onUpdate(update)
	}
}

// speak fala texto
func (cf *CoreFlow) speak(text string) {
	if cf.tts != nil {
		cf.tts.Speak(text)
	}
}

// ==================== COMANDOS DE VOZ ====================

// HandleVoiceCommand processa comando de voz
func (cf *CoreFlow) HandleVoiceCommand(command string) string {
	command = strings.ToLower(strings.TrimSpace(command))

	switch {
	case contains(command, "iniciar", "começar", "start"):
		go cf.Start(context.Background(), "Tarefa geral")
		return "Iniciando sessão de produtividade."

	case contains(command, "parar", "stop", "encerrar"):
		cf.Stop()
		return "Encerrando sessão."

	case contains(command, "pausar", "pause"):
		cf.Pause()
		return "Pausado."

	case contains(command, "continuar", "resume"):
		cf.Resume()
		return "Continuando..."

	case contains(command, "status", "quanto"):
		status := cf.GetStatus()
		return fmt.Sprintf("Estado: %s. Sessão %d. Tempo restante: %s.",
			status.State, status.Session, status.TimeRemaining)

	case contains(command, "nota", "anotar"):
		// Extrai conteúdo da nota
		content := strings.TrimPrefix(command, "nota ")
		content = strings.TrimPrefix(content, "anotar ")
		if content != "" {
			cf.AddNote(content)
			return "Nota salva."
		}
		return "Diga o que deseja anotar."

	default:
		return "Comando não reconhecido."
	}
}

// contains verifica se texto contém alguma das palavras
func contains(text string, words ...string) bool {
	for _, word := range words {
		if strings.Contains(text, word) {
			return true
		}
	}
	return false
}

// TTSInterface interface para TTS (evita import cíclico)
type TTSInterface interface {
	Speak(text string) error
}
