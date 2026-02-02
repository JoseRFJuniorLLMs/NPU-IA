package productivity

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ==================== DESPERTADOR ====================

// AlarmClock despertador inteligente
type AlarmClock struct {
	alarms   map[string]*Alarm
	mu       sync.RWMutex
	tts      TTSInterface
	onAlarm  func(alarm *Alarm)
	stopChan chan struct{}
}

// Alarm alarme individual
type Alarm struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Time        time.Time     `json:"time"`
	Repeat      []time.Weekday `json:"repeat"` // Dias da semana para repetir
	Enabled     bool          `json:"enabled"`
	Sound       string        `json:"sound"`
	Message     string        `json:"message"`
	Snooze      time.Duration `json:"snooze"`
	SnoozeCount int           `json:"snooze_count"`
}

// TTSInterface interface para Text-to-Speech
type TTSInterface interface {
	Speak(text string) error
}

// NewAlarmClock cria novo despertador
func NewAlarmClock(tts TTSInterface) *AlarmClock {
	ac := &AlarmClock{
		alarms:   make(map[string]*Alarm),
		tts:      tts,
		stopChan: make(chan struct{}),
	}
	go ac.run()
	return ac
}

// SetAlarm configura alarme
func (ac *AlarmClock) SetAlarm(name string, t time.Time, message string) *Alarm {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	id := fmt.Sprintf("alarm_%d", time.Now().UnixNano())
	alarm := &Alarm{
		ID:      id,
		Name:    name,
		Time:    t,
		Enabled: true,
		Message: message,
		Snooze:  5 * time.Minute,
	}

	ac.alarms[id] = alarm
	return alarm
}

// SetAlarmFromText configura alarme por comando de voz
func (ac *AlarmClock) SetAlarmFromText(command string) (*Alarm, error) {
	// Exemplos:
	// "me acorda às 7 da manhã"
	// "alarme para daqui 30 minutos"
	// "despertador às 6:30"

	now := time.Now()
	var alarmTime time.Time
	var name string

	// TODO: Usar LLM para parsear comando natural
	// Por enquanto, parse simples
	alarmTime = now.Add(30 * time.Minute)
	name = "Alarme"

	return ac.SetAlarm(name, alarmTime, "Hora de acordar!"), nil
}

// SetRepeatingAlarm alarme que repete
func (ac *AlarmClock) SetRepeatingAlarm(name string, hour, minute int, days []time.Weekday, message string) *Alarm {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	now := time.Now()
	alarmTime := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())

	// Se já passou hoje, agenda para amanhã
	if alarmTime.Before(now) {
		alarmTime = alarmTime.Add(24 * time.Hour)
	}

	id := fmt.Sprintf("alarm_%d", time.Now().UnixNano())
	alarm := &Alarm{
		ID:      id,
		Name:    name,
		Time:    alarmTime,
		Repeat:  days,
		Enabled: true,
		Message: message,
		Snooze:  5 * time.Minute,
	}

	ac.alarms[id] = alarm
	return alarm
}

// Snooze adia alarme
func (ac *AlarmClock) Snooze(alarmID string) error {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	alarm, ok := ac.alarms[alarmID]
	if !ok {
		return fmt.Errorf("alarme não encontrado")
	}

	alarm.Time = time.Now().Add(alarm.Snooze)
	alarm.SnoozeCount++
	alarm.Enabled = true

	return nil
}

// Dismiss desliga alarme
func (ac *AlarmClock) Dismiss(alarmID string) {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	if alarm, ok := ac.alarms[alarmID]; ok {
		if len(alarm.Repeat) == 0 {
			delete(ac.alarms, alarmID)
		} else {
			// Reagenda para próximo dia
			alarm.Time = ac.nextOccurrence(alarm)
			alarm.Enabled = true
		}
	}
}

// nextOccurrence calcula próxima ocorrência
func (ac *AlarmClock) nextOccurrence(alarm *Alarm) time.Time {
	now := time.Now()
	next := alarm.Time.Add(24 * time.Hour)

	for i := 0; i < 7; i++ {
		weekday := next.Weekday()
		for _, day := range alarm.Repeat {
			if weekday == day {
				return next
			}
		}
		next = next.Add(24 * time.Hour)
	}

	return next
}

// ListAlarms lista alarmes
func (ac *AlarmClock) ListAlarms() []*Alarm {
	ac.mu.RLock()
	defer ac.mu.RUnlock()

	alarms := make([]*Alarm, 0, len(ac.alarms))
	for _, alarm := range ac.alarms {
		alarms = append(alarms, alarm)
	}
	return alarms
}

// run loop principal do despertador
func (ac *AlarmClock) run() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ac.stopChan:
			return
		case now := <-ticker.C:
			ac.checkAlarms(now)
		}
	}
}

// checkAlarms verifica alarmes
func (ac *AlarmClock) checkAlarms(now time.Time) {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	for _, alarm := range ac.alarms {
		if !alarm.Enabled {
			continue
		}

		// Verifica se é hora do alarme (com margem de 1 segundo)
		diff := alarm.Time.Sub(now)
		if diff >= 0 && diff < time.Second {
			alarm.Enabled = false
			go ac.triggerAlarm(alarm)
		}
	}
}

// triggerAlarm dispara alarme
func (ac *AlarmClock) triggerAlarm(alarm *Alarm) {
	message := alarm.Message
	if message == "" {
		message = fmt.Sprintf("Alarme: %s", alarm.Name)
	}

	// Fala o alarme
	if ac.tts != nil {
		ac.tts.Speak(message)
	}

	// Callback
	if ac.onAlarm != nil {
		ac.onAlarm(alarm)
	}
}

// Stop para o despertador
func (ac *AlarmClock) Stop() {
	close(ac.stopChan)
}

// ==================== POMODORO ====================

// PomodoroTimer timer Pomodoro
type PomodoroTimer struct {
	config      PomodoroConfig
	state       PomodoroState
	currentTask string
	sessions    int
	startTime   time.Time
	tts         TTSInterface
	onStateChange func(state PomodoroState, remaining time.Duration)
	mu          sync.RWMutex
	stopChan    chan struct{}
	pauseChan   chan struct{}
	resumeChan  chan struct{}
}

// PomodoroConfig configuração do Pomodoro
type PomodoroConfig struct {
	WorkDuration      time.Duration `json:"work_duration"`       // Padrão: 25 min
	ShortBreak        time.Duration `json:"short_break"`         // Padrão: 5 min
	LongBreak         time.Duration `json:"long_break"`          // Padrão: 15 min
	SessionsUntilLong int           `json:"sessions_until_long"` // Padrão: 4
	AutoStart         bool          `json:"auto_start"`
}

// PomodoroState estado do Pomodoro
type PomodoroState string

const (
	StateIdle       PomodoroState = "idle"
	StateWorking    PomodoroState = "working"
	StateShortBreak PomodoroState = "short_break"
	StateLongBreak  PomodoroState = "long_break"
	StatePaused     PomodoroState = "paused"
)

// PomodoroStats estatísticas
type PomodoroStats struct {
	TotalSessions    int           `json:"total_sessions"`
	TotalWorkTime    time.Duration `json:"total_work_time"`
	TotalBreakTime   time.Duration `json:"total_break_time"`
	CurrentStreak    int           `json:"current_streak"`
	TasksCompleted   []string      `json:"tasks_completed"`
}

// DefaultPomodoroConfig configuração padrão
func DefaultPomodoroConfig() PomodoroConfig {
	return PomodoroConfig{
		WorkDuration:      25 * time.Minute,
		ShortBreak:        5 * time.Minute,
		LongBreak:         15 * time.Minute,
		SessionsUntilLong: 4,
		AutoStart:         false,
	}
}

// NewPomodoroTimer cria timer Pomodoro
func NewPomodoroTimer(config PomodoroConfig, tts TTSInterface) *PomodoroTimer {
	return &PomodoroTimer{
		config:     config,
		state:      StateIdle,
		tts:        tts,
		stopChan:   make(chan struct{}),
		pauseChan:  make(chan struct{}),
		resumeChan: make(chan struct{}),
	}
}

// Start inicia sessão de trabalho
func (p *PomodoroTimer) Start(task string) {
	p.mu.Lock()
	p.currentTask = task
	p.state = StateWorking
	p.startTime = time.Now()
	p.mu.Unlock()

	if p.tts != nil {
		p.tts.Speak(fmt.Sprintf("Iniciando Pomodoro. Foco em: %s. Você tem 25 minutos.", task))
	}

	go p.run(p.config.WorkDuration)
}

// Pause pausa timer
func (p *PomodoroTimer) Pause() {
	p.mu.Lock()
	p.state = StatePaused
	p.mu.Unlock()
	p.pauseChan <- struct{}{}
}

// Resume continua timer
func (p *PomodoroTimer) Resume() {
	p.resumeChan <- struct{}{}
}

// Skip pula para próxima fase
func (p *PomodoroTimer) Skip() {
	p.stopChan <- struct{}{}
}

// run executa o timer
func (p *PomodoroTimer) run(duration time.Duration) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	remaining := duration
	paused := false

	for remaining > 0 {
		select {
		case <-p.stopChan:
			return
		case <-p.pauseChan:
			paused = true
		case <-p.resumeChan:
			paused = false
			p.mu.Lock()
			p.state = StateWorking
			p.mu.Unlock()
		case <-ticker.C:
			if !paused {
				remaining -= time.Second

				// Callback de progresso
				if p.onStateChange != nil {
					p.onStateChange(p.state, remaining)
				}

				// Avisos
				if remaining == 5*time.Minute {
					if p.tts != nil {
						p.tts.Speak("Faltam 5 minutos.")
					}
				} else if remaining == 1*time.Minute {
					if p.tts != nil {
						p.tts.Speak("Faltam 1 minuto.")
					}
				}
			}
		}
	}

	// Timer terminou
	p.onTimerComplete()
}

// onTimerComplete chamado quando timer termina
func (p *PomodoroTimer) onTimerComplete() {
	p.mu.Lock()
	defer p.mu.Unlock()

	switch p.state {
	case StateWorking:
		p.sessions++

		var breakDuration time.Duration
		var message string

		if p.sessions%p.config.SessionsUntilLong == 0 {
			p.state = StateLongBreak
			breakDuration = p.config.LongBreak
			message = fmt.Sprintf("Excelente! %d sessões completas. Pausa longa de 15 minutos.", p.sessions)
		} else {
			p.state = StateShortBreak
			breakDuration = p.config.ShortBreak
			message = fmt.Sprintf("Sessão %d completa! Pausa de 5 minutos.", p.sessions)
		}

		if p.tts != nil {
			p.tts.Speak(message)
		}

		if p.config.AutoStart {
			go p.run(breakDuration)
		}

	case StateShortBreak, StateLongBreak:
		p.state = StateIdle
		if p.tts != nil {
			p.tts.Speak("Pausa terminada. Pronto para mais uma sessão?")
		}
	}
}

// GetStatus retorna status atual
func (p *PomodoroTimer) GetStatus() (PomodoroState, time.Duration, int) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var remaining time.Duration
	if p.state == StateWorking || p.state == StateShortBreak || p.state == StateLongBreak {
		elapsed := time.Since(p.startTime)
		switch p.state {
		case StateWorking:
			remaining = p.config.WorkDuration - elapsed
		case StateShortBreak:
			remaining = p.config.ShortBreak - elapsed
		case StateLongBreak:
			remaining = p.config.LongBreak - elapsed
		}
	}

	return p.state, remaining, p.sessions
}

// GetStats retorna estatísticas
func (p *PomodoroTimer) GetStats() PomodoroStats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return PomodoroStats{
		TotalSessions:  p.sessions,
		TotalWorkTime:  time.Duration(p.sessions) * p.config.WorkDuration,
		CurrentStreak:  p.sessions,
	}
}
