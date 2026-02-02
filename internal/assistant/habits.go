package assistant

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// ==================== HABIT TRACKER ====================

// HabitTracker rastreador de hábitos
type HabitTracker struct {
	basePath    string
	habits      map[string]*Habit
	completions map[string][]Completion // habitID -> completions
	mu          sync.RWMutex
	tts         TTSInterface
}

// Habit hábito
type Habit struct {
	ID            string        `json:"id"`
	Name          string        `json:"name"`
	Description   string        `json:"description"`
	Icon          string        `json:"icon"`
	Frequency     Frequency     `json:"frequency"`
	TargetDays    []time.Weekday `json:"target_days"` // Para frequência semanal
	TargetCount   int           `json:"target_count"` // Vezes por período
	CurrentStreak int           `json:"current_streak"`
	BestStreak    int           `json:"best_streak"`
	TotalComplete int           `json:"total_complete"`
	CreatedAt     time.Time     `json:"created_at"`
	Category      string        `json:"category"`
	Reminder      string        `json:"reminder"` // Horário do lembrete
	IsActive      bool          `json:"is_active"`
}

// Frequency frequência do hábito
type Frequency string

const (
	FrequencyDaily   Frequency = "daily"
	FrequencyWeekly  Frequency = "weekly"
	FrequencyMonthly Frequency = "monthly"
)

// Completion registro de conclusão
type Completion struct {
	HabitID   string    `json:"habit_id"`
	Date      time.Time `json:"date"`
	Notes     string    `json:"notes"`
	Value     float64   `json:"value"` // Para hábitos com valor (ex: minutos meditando)
}

// HabitStats estatísticas de um hábito
type HabitStats struct {
	Habit           *Habit    `json:"habit"`
	CurrentStreak   int       `json:"current_streak"`
	BestStreak      int       `json:"best_streak"`
	TotalComplete   int       `json:"total_complete"`
	CompletionRate  float64   `json:"completion_rate"` // percentual
	Last7Days       []bool    `json:"last_7_days"`
	Last30Days      []bool    `json:"last_30_days"`
}

// NewHabitTracker cria tracker de hábitos
func NewHabitTracker(basePath string, tts TTSInterface) (*HabitTracker, error) {
	ht := &HabitTracker{
		basePath:    basePath,
		habits:      make(map[string]*Habit),
		completions: make(map[string][]Completion),
		tts:         tts,
	}

	os.MkdirAll(basePath, 0755)
	ht.load()

	return ht, nil
}

// ==================== CRUD ====================

// CreateHabit cria novo hábito
func (ht *HabitTracker) CreateHabit(name, description, category, icon string, frequency Frequency) *Habit {
	ht.mu.Lock()
	defer ht.mu.Unlock()

	id := fmt.Sprintf("habit_%d", time.Now().UnixNano())
	habit := &Habit{
		ID:          id,
		Name:        name,
		Description: description,
		Category:    category,
		Icon:        icon,
		Frequency:   frequency,
		TargetCount: 1,
		CreatedAt:   time.Now(),
		IsActive:    true,
	}

	if icon == "" {
		habit.Icon = "✅"
	}

	ht.habits[id] = habit
	ht.completions[id] = make([]Completion, 0)
	ht.save()

	return habit
}

// CreateCommonHabits cria hábitos comuns pré-definidos
func (ht *HabitTracker) CreateCommonHabits() {
	commonHabits := []struct {
		name     string
		category string
		icon     string
	}{
		{"Wim Hof", "saúde", "🧊"},
		{"Exercício", "saúde", "💪"},
		{"Meditação", "mente", "🧘"},
		{"Leitura", "aprendizado", "📚"},
		{"Água 2L", "saúde", "💧"},
		{"Sono 8h", "saúde", "😴"},
		{"Pomodoro", "produtividade", "🍅"},
		{"Journaling", "mente", "📝"},
		{"Gratidão", "mente", "🙏"},
		{"Sem redes sociais", "foco", "📵"},
	}

	for _, h := range commonHabits {
		ht.CreateHabit(h.name, "", h.category, h.icon, FrequencyDaily)
	}
}

// GetHabit obtém hábito
func (ht *HabitTracker) GetHabit(id string) *Habit {
	ht.mu.RLock()
	defer ht.mu.RUnlock()
	return ht.habits[id]
}

// GetAllHabits retorna todos os hábitos ativos
func (ht *HabitTracker) GetAllHabits() []*Habit {
	ht.mu.RLock()
	defer ht.mu.RUnlock()

	habits := make([]*Habit, 0)
	for _, h := range ht.habits {
		if h.IsActive {
			habits = append(habits, h)
		}
	}

	// Ordena por categoria
	sort.Slice(habits, func(i, j int) bool {
		return habits[i].Category < habits[j].Category
	})

	return habits
}

// GetByCategory retorna hábitos de uma categoria
func (ht *HabitTracker) GetByCategory(category string) []*Habit {
	ht.mu.RLock()
	defer ht.mu.RUnlock()

	habits := make([]*Habit, 0)
	for _, h := range ht.habits {
		if h.Category == category && h.IsActive {
			habits = append(habits, h)
		}
	}
	return habits
}

// DeleteHabit deleta hábito (soft delete)
func (ht *HabitTracker) DeleteHabit(id string) {
	ht.mu.Lock()
	defer ht.mu.Unlock()

	if habit, ok := ht.habits[id]; ok {
		habit.IsActive = false
		ht.save()
	}
}

// ==================== COMPLETIONS ====================

// Complete marca hábito como completo
func (ht *HabitTracker) Complete(habitID string, notes string, value float64) error {
	ht.mu.Lock()
	defer ht.mu.Unlock()

	habit, ok := ht.habits[habitID]
	if !ok {
		return fmt.Errorf("hábito não encontrado")
	}

	// Verifica se já foi completado hoje
	today := time.Now().Truncate(24 * time.Hour)
	for _, c := range ht.completions[habitID] {
		if c.Date.Truncate(24 * time.Hour).Equal(today) {
			return fmt.Errorf("hábito já completado hoje")
		}
	}

	// Adiciona completion
	completion := Completion{
		HabitID: habitID,
		Date:    time.Now(),
		Notes:   notes,
		Value:   value,
	}
	ht.completions[habitID] = append(ht.completions[habitID], completion)

	// Atualiza streak
	ht.updateStreak(habit)

	// Atualiza totais
	habit.TotalComplete++

	ht.save()

	// Feedback de voz
	if ht.tts != nil {
		if habit.CurrentStreak > habit.BestStreak {
			ht.tts.Speak(fmt.Sprintf("%s completo! Novo recorde: %d dias de streak!",
				habit.Name, habit.CurrentStreak))
		} else if habit.CurrentStreak > 0 && habit.CurrentStreak%7 == 0 {
			ht.tts.Speak(fmt.Sprintf("%s completo! %d dias de streak! Continue assim!",
				habit.Name, habit.CurrentStreak))
		}
	}

	return nil
}

// IsCompletedToday verifica se foi completado hoje
func (ht *HabitTracker) IsCompletedToday(habitID string) bool {
	ht.mu.RLock()
	defer ht.mu.RUnlock()

	today := time.Now().Truncate(24 * time.Hour)
	for _, c := range ht.completions[habitID] {
		if c.Date.Truncate(24 * time.Hour).Equal(today) {
			return true
		}
	}
	return false
}

// IsCompletedOn verifica se foi completado em uma data
func (ht *HabitTracker) IsCompletedOn(habitID string, date time.Time) bool {
	ht.mu.RLock()
	defer ht.mu.RUnlock()

	targetDate := date.Truncate(24 * time.Hour)
	for _, c := range ht.completions[habitID] {
		if c.Date.Truncate(24 * time.Hour).Equal(targetDate) {
			return true
		}
	}
	return false
}

// updateStreak atualiza streak do hábito
func (ht *HabitTracker) updateStreak(habit *Habit) {
	completions := ht.completions[habit.ID]
	if len(completions) == 0 {
		habit.CurrentStreak = 0
		return
	}

	// Ordena por data (mais recente primeiro)
	sort.Slice(completions, func(i, j int) bool {
		return completions[i].Date.After(completions[j].Date)
	})

	streak := 1
	today := time.Now().Truncate(24 * time.Hour)
	yesterday := today.AddDate(0, 0, -1)

	// Verifica se a última completion é de hoje ou ontem
	lastDate := completions[0].Date.Truncate(24 * time.Hour)
	if !lastDate.Equal(today) && !lastDate.Equal(yesterday) {
		habit.CurrentStreak = 0
		return
	}

	// Conta streak
	for i := 1; i < len(completions); i++ {
		prevDate := completions[i-1].Date.Truncate(24 * time.Hour)
		currDate := completions[i].Date.Truncate(24 * time.Hour)

		diff := prevDate.Sub(currDate).Hours() / 24
		if diff == 1 {
			streak++
		} else if diff == 0 {
			// Mesmo dia, não conta
			continue
		} else {
			break
		}
	}

	habit.CurrentStreak = streak
	if streak > habit.BestStreak {
		habit.BestStreak = streak
	}
}

// ==================== ESTATÍSTICAS ====================

// GetStats retorna estatísticas de um hábito
func (ht *HabitTracker) GetStats(habitID string) *HabitStats {
	ht.mu.RLock()
	defer ht.mu.RUnlock()

	habit, ok := ht.habits[habitID]
	if !ok {
		return nil
	}

	stats := &HabitStats{
		Habit:         habit,
		CurrentStreak: habit.CurrentStreak,
		BestStreak:    habit.BestStreak,
		TotalComplete: habit.TotalComplete,
	}

	// Últimos 7 dias
	stats.Last7Days = make([]bool, 7)
	for i := 0; i < 7; i++ {
		date := time.Now().AddDate(0, 0, -i)
		stats.Last7Days[6-i] = ht.IsCompletedOn(habitID, date)
	}

	// Últimos 30 dias
	stats.Last30Days = make([]bool, 30)
	completedCount := 0
	for i := 0; i < 30; i++ {
		date := time.Now().AddDate(0, 0, -i)
		completed := ht.IsCompletedOn(habitID, date)
		stats.Last30Days[29-i] = completed
		if completed {
			completedCount++
		}
	}

	// Taxa de conclusão
	stats.CompletionRate = float64(completedCount) / 30.0 * 100

	return stats
}

// GetDailyProgress retorna progresso do dia
func (ht *HabitTracker) GetDailyProgress() (completed int, total int, habits []*Habit) {
	ht.mu.RLock()
	defer ht.mu.RUnlock()

	habits = make([]*Habit, 0)
	for _, h := range ht.habits {
		if !h.IsActive {
			continue
		}
		total++
		if ht.IsCompletedToday(h.ID) {
			completed++
		}
		habits = append(habits, h)
	}

	return
}

// GetWeeklyReport relatório semanal
func (ht *HabitTracker) GetWeeklyReport() string {
	habits := ht.GetAllHabits()

	var report strings.Builder
	report.WriteString("📊 RELATÓRIO SEMANAL DE HÁBITOS\n")
	report.WriteString("═══════════════════════════════════\n\n")

	for _, habit := range habits {
		stats := ht.GetStats(habit.ID)
		if stats == nil {
			continue
		}

		// Visualização dos últimos 7 dias
		weekView := ""
		for _, completed := range stats.Last7Days {
			if completed {
				weekView += "✅"
			} else {
				weekView += "⬜"
			}
		}

		report.WriteString(fmt.Sprintf("%s %s\n", habit.Icon, habit.Name))
		report.WriteString(fmt.Sprintf("   %s\n", weekView))
		report.WriteString(fmt.Sprintf("   🔥 Streak: %d dias | Taxa: %.0f%%\n\n",
			stats.CurrentStreak, stats.CompletionRate))
	}

	return report.String()
}

// ==================== VISUALIZAÇÃO ====================

// RenderCalendarView renderiza visualização de calendário
func (ht *HabitTracker) RenderCalendarView(habitID string) string {
	stats := ht.GetStats(habitID)
	if stats == nil {
		return "Hábito não encontrado"
	}

	var view strings.Builder
	view.WriteString(fmt.Sprintf("%s %s - Últimos 30 dias\n\n",
		stats.Habit.Icon, stats.Habit.Name))

	// Renderiza em linhas de 7
	for i := 0; i < 30; i++ {
		if stats.Last30Days[i] {
			view.WriteString("🟩 ")
		} else {
			view.WriteString("⬜ ")
		}
		if (i+1)%7 == 0 {
			view.WriteString("\n")
		}
	}

	view.WriteString(fmt.Sprintf("\n\n🔥 Streak atual: %d dias\n", stats.CurrentStreak))
	view.WriteString(fmt.Sprintf("🏆 Melhor streak: %d dias\n", stats.BestStreak))
	view.WriteString(fmt.Sprintf("📊 Taxa de conclusão: %.1f%%\n", stats.CompletionRate))

	return view.String()
}

// RenderDashboard renderiza dashboard completo
func (ht *HabitTracker) RenderDashboard() string {
	completed, total, habits := ht.GetDailyProgress()

	var dash strings.Builder
	dash.WriteString("╔══════════════════════════════════════════╗\n")
	dash.WriteString("║           🎯 HABIT TRACKER               ║\n")
	dash.WriteString("╚══════════════════════════════════════════╝\n\n")

	// Progresso do dia
	progress := 0
	if total > 0 {
		progress = (completed * 100) / total
	}
	progressBar := ht.renderProgressBar(progress, 20)

	dash.WriteString(fmt.Sprintf("HOJE: %s %d/%d (%d%%)\n\n", progressBar, completed, total, progress))

	// Lista de hábitos por categoria
	categories := make(map[string][]*Habit)
	for _, h := range habits {
		categories[h.Category] = append(categories[h.Category], h)
	}

	for category, categoryHabits := range categories {
		dash.WriteString(fmt.Sprintf("── %s ──\n", strings.ToUpper(category)))
		for _, habit := range categoryHabits {
			status := "⬜"
			if ht.IsCompletedToday(habit.ID) {
				status = "✅"
			}
			streak := ""
			if habit.CurrentStreak > 0 {
				streak = fmt.Sprintf(" 🔥%d", habit.CurrentStreak)
			}
			dash.WriteString(fmt.Sprintf("%s %s %s%s\n", status, habit.Icon, habit.Name, streak))
		}
		dash.WriteString("\n")
	}

	return dash.String()
}

// renderProgressBar renderiza barra de progresso
func (ht *HabitTracker) renderProgressBar(percent, width int) string {
	filled := (percent * width) / 100
	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	return bar
}

// ==================== PERSISTÊNCIA ====================

func (ht *HabitTracker) save() error {
	data := struct {
		Habits      map[string]*Habit        `json:"habits"`
		Completions map[string][]Completion  `json:"completions"`
	}{
		Habits:      ht.habits,
		Completions: ht.completions,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(ht.basePath, "habits.json"), jsonData, 0644)
}

func (ht *HabitTracker) load() error {
	data, err := os.ReadFile(filepath.Join(ht.basePath, "habits.json"))
	if err != nil {
		return nil
	}

	var loaded struct {
		Habits      map[string]*Habit        `json:"habits"`
		Completions map[string][]Completion  `json:"completions"`
	}

	if err := json.Unmarshal(data, &loaded); err != nil {
		return err
	}

	if loaded.Habits != nil {
		ht.habits = loaded.Habits
	}
	if loaded.Completions != nil {
		ht.completions = loaded.Completions
	}

	// Recalcula streaks
	for _, habit := range ht.habits {
		ht.updateStreak(habit)
	}

	return nil
}

