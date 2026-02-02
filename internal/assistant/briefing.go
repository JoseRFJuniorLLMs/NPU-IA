package assistant

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// ==================== DAILY BRIEFING ====================

// DailyBriefing gera resumo diário
type DailyBriefing struct {
	memory    *Memory
	calendar  CalendarInterface
	email     EmailInterface
	tasks     TaskInterface
	weather   WeatherInterface
	news      NewsInterface
	crypto    CryptoInterface
	habits    *HabitTracker
	tts       TTSInterface
}

// CalendarInterface interface para calendário
type CalendarInterface interface {
	GetTodayEvents() ([]Event, error)
}

// EmailInterface interface para email
type EmailInterface interface {
	GetUnreadCount() (int, error)
	GetImportantEmails(limit int) ([]Email, error)
}

// TaskInterface interface para tarefas
type TaskInterface interface {
	GetPendingTasks() ([]Task, error)
	GetOverdueTasks() ([]Task, error)
}

// WeatherInterface interface para clima
type WeatherInterface interface {
	GetCurrentWeather(city string) (*Weather, error)
}

// NewsInterface interface para notícias
type NewsInterface interface {
	GetTopHeadlines(category string, limit int) ([]News, error)
}

// CryptoInterface interface para cripto
type CryptoInterface interface {
	GetPrice(symbol string) (*CryptoPrice, error)
}

// TTSInterface interface para TTS
type TTSInterface interface {
	Speak(text string) error
}

// Event evento do calendário
type Event struct {
	Title     string    `json:"title"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Location  string    `json:"location"`
}

// Email email
type Email struct {
	From    string `json:"from"`
	Subject string `json:"subject"`
	IsUrgent bool   `json:"is_urgent"`
}

// Task tarefa
type Task struct {
	Title    string    `json:"title"`
	DueDate  time.Time `json:"due_date"`
	Priority int       `json:"priority"`
	Project  string    `json:"project"`
}

// Weather clima
type Weather struct {
	Temperature float64 `json:"temperature"`
	Condition   string  `json:"condition"`
	Humidity    int     `json:"humidity"`
	City        string  `json:"city"`
}

// News notícia
type News struct {
	Title  string `json:"title"`
	Source string `json:"source"`
}

// CryptoPrice preço de criptomoeda
type CryptoPrice struct {
	Symbol    string  `json:"symbol"`
	Price     float64 `json:"price"`
	Change24h float64 `json:"change_24h"`
}

// Briefing resultado do briefing
type Briefing struct {
	Greeting       string         `json:"greeting"`
	Date           string         `json:"date"`
	Weather        *Weather       `json:"weather"`
	Events         []Event        `json:"events"`
	UnreadEmails   int            `json:"unread_emails"`
	ImportantEmails []Email       `json:"important_emails"`
	Tasks          []Task         `json:"tasks"`
	OverdueTasks   []Task         `json:"overdue_tasks"`
	HabitSummary   *HabitSummary  `json:"habit_summary"`
	CryptoPrices   []CryptoPrice  `json:"crypto_prices"`
	News           []News         `json:"news"`
	Quote          string         `json:"quote"`
	GeneratedAt    time.Time      `json:"generated_at"`
}

// HabitSummary resumo de hábitos
type HabitSummary struct {
	TodayComplete   int `json:"today_complete"`
	TodayPending    int `json:"today_pending"`
	CurrentStreak   int `json:"current_streak"`
	WeeklyProgress  int `json:"weekly_progress"` // percentual
}

// NewDailyBriefing cria briefing diário
func NewDailyBriefing(memory *Memory, habits *HabitTracker) *DailyBriefing {
	return &DailyBriefing{
		memory: memory,
		habits: habits,
	}
}

// SetCalendar configura calendário
func (db *DailyBriefing) SetCalendar(cal CalendarInterface) {
	db.calendar = cal
}

// SetEmail configura email
func (db *DailyBriefing) SetEmail(email EmailInterface) {
	db.email = email
}

// SetTasks configura tarefas
func (db *DailyBriefing) SetTasks(tasks TaskInterface) {
	db.tasks = tasks
}

// SetWeather configura clima
func (db *DailyBriefing) SetWeather(weather WeatherInterface) {
	db.weather = weather
}

// SetTTS configura TTS
func (db *DailyBriefing) SetTTS(tts TTSInterface) {
	db.tts = tts
}

// Generate gera o briefing completo
func (db *DailyBriefing) Generate(ctx context.Context) (*Briefing, error) {
	briefing := &Briefing{
		GeneratedAt: time.Now(),
		Date:        time.Now().Format("Monday, 02 de January de 2006"),
	}

	// Saudação personalizada
	briefing.Greeting = db.generateGreeting()

	// Clima
	if db.weather != nil {
		weather, err := db.weather.GetCurrentWeather("São Paulo")
		if err == nil {
			briefing.Weather = weather
		}
	}

	// Eventos do dia
	if db.calendar != nil {
		events, err := db.calendar.GetTodayEvents()
		if err == nil {
			briefing.Events = events
		}
	}

	// Emails
	if db.email != nil {
		unread, _ := db.email.GetUnreadCount()
		briefing.UnreadEmails = unread

		important, _ := db.email.GetImportantEmails(5)
		briefing.ImportantEmails = important
	}

	// Tarefas
	if db.tasks != nil {
		tasks, _ := db.tasks.GetPendingTasks()
		briefing.Tasks = tasks

		overdue, _ := db.tasks.GetOverdueTasks()
		briefing.OverdueTasks = overdue
	}

	// Hábitos
	if db.habits != nil {
		briefing.HabitSummary = db.getHabitSummary()
	}

	// Quote motivacional
	briefing.Quote = db.getDailyQuote()

	return briefing, nil
}

// generateGreeting gera saudação personalizada
func (db *DailyBriefing) generateGreeting() string {
	hour := time.Now().Hour()
	name := "você"

	if db.memory != nil {
		name = db.memory.GetUserName()
	}

	var greeting string
	switch {
	case hour < 6:
		greeting = fmt.Sprintf("Boa madrugada, %s! Acordou cedo hoje.", name)
	case hour < 12:
		greeting = fmt.Sprintf("Bom dia, %s!", name)
	case hour < 18:
		greeting = fmt.Sprintf("Boa tarde, %s!", name)
	default:
		greeting = fmt.Sprintf("Boa noite, %s!", name)
	}

	return greeting
}

// getHabitSummary obtém resumo de hábitos
func (db *DailyBriefing) getHabitSummary() *HabitSummary {
	if db.habits == nil {
		return nil
	}

	habits := db.habits.GetAllHabits()
	todayComplete := 0
	todayPending := 0
	maxStreak := 0

	for _, h := range habits {
		if db.habits.IsCompletedToday(h.ID) {
			todayComplete++
		} else {
			todayPending++
		}
		if h.CurrentStreak > maxStreak {
			maxStreak = h.CurrentStreak
		}
	}

	weeklyProgress := 0
	if len(habits) > 0 {
		weeklyProgress = (todayComplete * 100) / len(habits)
	}

	return &HabitSummary{
		TodayComplete:  todayComplete,
		TodayPending:   todayPending,
		CurrentStreak:  maxStreak,
		WeeklyProgress: weeklyProgress,
	}
}

// getDailyQuote retorna quote do dia
func (db *DailyBriefing) getDailyQuote() string {
	quotes := []string{
		"O segredo de ir em frente é começar. - Mark Twain",
		"A única maneira de fazer um ótimo trabalho é amar o que você faz. - Steve Jobs",
		"Foco é dizer não a centenas de boas ideias. - Steve Jobs",
		"Disciplina é a ponte entre metas e realizações. - Jim Rohn",
		"O tempo que você gosta de desperdiçar não é tempo desperdiçado. - Bertrand Russell",
		"Produtividade não é sobre estar ocupado, é sobre ser eficaz. - Tim Ferriss",
		"A respiração é a ponte que conecta a vida à consciência. - Thich Nhat Hanh",
		"Quanto mais você sua em treinamento, menos sangra em batalha. - Richard Marcinko",
		"O cold não é seu inimigo, é seu professor. - Wim Hof",
		"Mantenha-se com fome, mantenha-se tolo. - Steve Jobs",
	}

	// Usa o dia do ano para selecionar quote
	dayOfYear := time.Now().YearDay()
	return quotes[dayOfYear%len(quotes)]
}

// Speak fala o briefing
func (db *DailyBriefing) Speak(briefing *Briefing) error {
	if db.tts == nil {
		return nil
	}

	var script strings.Builder

	// Saudação
	script.WriteString(briefing.Greeting + ". ")

	// Data
	script.WriteString(fmt.Sprintf("Hoje é %s. ", briefing.Date))

	// Clima
	if briefing.Weather != nil {
		script.WriteString(fmt.Sprintf("A temperatura está em %.0f graus, %s. ",
			briefing.Weather.Temperature, briefing.Weather.Condition))
	}

	// Agenda
	if len(briefing.Events) > 0 {
		script.WriteString(fmt.Sprintf("Você tem %d eventos hoje. ", len(briefing.Events)))
		for i, event := range briefing.Events {
			if i < 3 { // Fala só os 3 primeiros
				script.WriteString(fmt.Sprintf("%s às %s. ",
					event.Title, event.StartTime.Format("15:04")))
			}
		}
	} else {
		script.WriteString("Sua agenda está livre hoje. ")
	}

	// Emails
	if briefing.UnreadEmails > 0 {
		script.WriteString(fmt.Sprintf("Você tem %d emails não lidos. ", briefing.UnreadEmails))

		urgentCount := 0
		for _, email := range briefing.ImportantEmails {
			if email.IsUrgent {
				urgentCount++
			}
		}
		if urgentCount > 0 {
			script.WriteString(fmt.Sprintf("%d são urgentes. ", urgentCount))
		}
	}

	// Tarefas atrasadas
	if len(briefing.OverdueTasks) > 0 {
		script.WriteString(fmt.Sprintf("Atenção: você tem %d tarefas atrasadas. ", len(briefing.OverdueTasks)))
	}

	// Tarefas do dia
	if len(briefing.Tasks) > 0 {
		script.WriteString(fmt.Sprintf("Você tem %d tarefas pendentes para hoje. ", len(briefing.Tasks)))
	}

	// Hábitos
	if briefing.HabitSummary != nil {
		if briefing.HabitSummary.CurrentStreak > 5 {
			script.WriteString(fmt.Sprintf("Parabéns pelo streak de %d dias! ",
				briefing.HabitSummary.CurrentStreak))
		}
	}

	// Quote
	script.WriteString("Pensamento do dia: " + briefing.Quote)

	return db.tts.Speak(script.String())
}

// ToText converte briefing para texto formatado
func (db *DailyBriefing) ToText(briefing *Briefing) string {
	var text strings.Builder

	text.WriteString("╔══════════════════════════════════════════╗\n")
	text.WriteString("║           ☀️  BRIEFING DIÁRIO             ║\n")
	text.WriteString("╚══════════════════════════════════════════╝\n\n")

	text.WriteString(fmt.Sprintf("%s\n", briefing.Greeting))
	text.WriteString(fmt.Sprintf("📅 %s\n\n", briefing.Date))

	// Clima
	if briefing.Weather != nil {
		text.WriteString(fmt.Sprintf("🌤️ CLIMA: %.0f°C, %s\n\n",
			briefing.Weather.Temperature, briefing.Weather.Condition))
	}

	// Agenda
	text.WriteString("📅 AGENDA HOJE:\n")
	if len(briefing.Events) == 0 {
		text.WriteString("   Nenhum evento agendado\n")
	} else {
		for _, event := range briefing.Events {
			text.WriteString(fmt.Sprintf("   • %s - %s\n",
				event.StartTime.Format("15:04"), event.Title))
		}
	}
	text.WriteString("\n")

	// Emails
	text.WriteString("📧 EMAILS:\n")
	text.WriteString(fmt.Sprintf("   • %d não lidos\n", briefing.UnreadEmails))
	if len(briefing.ImportantEmails) > 0 {
		text.WriteString("   Importantes:\n")
		for _, email := range briefing.ImportantEmails[:min(3, len(briefing.ImportantEmails))] {
			urgent := ""
			if email.IsUrgent {
				urgent = "🔴 "
			}
			text.WriteString(fmt.Sprintf("   %s• %s: %s\n", urgent, email.From, email.Subject))
		}
	}
	text.WriteString("\n")

	// Tarefas
	text.WriteString("✅ TAREFAS:\n")
	if len(briefing.OverdueTasks) > 0 {
		text.WriteString(fmt.Sprintf("   ⚠️ %d atrasadas!\n", len(briefing.OverdueTasks)))
	}
	if len(briefing.Tasks) == 0 {
		text.WriteString("   Nenhuma tarefa pendente\n")
	} else {
		for _, task := range briefing.Tasks[:min(5, len(briefing.Tasks))] {
			priority := ""
			if task.Priority >= 4 {
				priority = "🔴 "
			}
			text.WriteString(fmt.Sprintf("   %s• %s\n", priority, task.Title))
		}
	}
	text.WriteString("\n")

	// Hábitos
	if briefing.HabitSummary != nil {
		text.WriteString("🎯 HÁBITOS:\n")
		text.WriteString(fmt.Sprintf("   • Completos: %d | Pendentes: %d\n",
			briefing.HabitSummary.TodayComplete, briefing.HabitSummary.TodayPending))
		if briefing.HabitSummary.CurrentStreak > 0 {
			text.WriteString(fmt.Sprintf("   🔥 Streak: %d dias\n", briefing.HabitSummary.CurrentStreak))
		}
		text.WriteString("\n")
	}

	// Crypto
	if len(briefing.CryptoPrices) > 0 {
		text.WriteString("💰 CRYPTO:\n")
		for _, crypto := range briefing.CryptoPrices {
			change := "📈"
			if crypto.Change24h < 0 {
				change = "📉"
			}
			text.WriteString(fmt.Sprintf("   %s %s: $%.2f (%+.2f%%)\n",
				change, crypto.Symbol, crypto.Price, crypto.Change24h))
		}
		text.WriteString("\n")
	}

	// Quote
	text.WriteString("💭 PENSAMENTO DO DIA:\n")
	text.WriteString(fmt.Sprintf("   \"%s\"\n\n", briefing.Quote))

	text.WriteString("══════════════════════════════════════════\n")

	return text.String()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
