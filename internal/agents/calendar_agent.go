package agents

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// CalendarAgent agente de calendário inteligente
type CalendarAgent struct {
	llm             LLMInterface
	calendarService CalendarServiceInterface
	emailService    EmailServiceInterface
}

// CalendarServiceInterface interface para serviço de calendário
type CalendarServiceInterface interface {
	GetEvents(start, end time.Time) ([]Event, error)
	CreateEvent(title string, start, end time.Time, description string) error
	DeleteEvent(eventID string) error
	GetFreeSlots(start, end time.Time, duration time.Duration) ([]TimeSlot, error)
}

// Event evento do calendário
type Event struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Start       time.Time `json:"start"`
	End         time.Time `json:"end"`
	Description string    `json:"description"`
	Location    string    `json:"location"`
	Attendees   []string  `json:"attendees"`
}

// TimeSlot horário disponível
type TimeSlot struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// MeetingSuggestion sugestão de reunião
type MeetingSuggestion struct {
	Slot        TimeSlot `json:"slot"`
	LocalTime   string   `json:"local_time"`
	RemoteTime  string   `json:"remote_time"`
	RemoteTZ    string   `json:"remote_timezone"`
}

// NewCalendarAgent cria agente de calendário
func NewCalendarAgent(llm LLMInterface, calendarService CalendarServiceInterface, emailService EmailServiceInterface) *CalendarAgent {
	return &CalendarAgent{
		llm:             llm,
		calendarService: calendarService,
		emailService:    emailService,
	}
}

// ==================== 11. AGENDAMENTO POR TEXTO ====================

// ScheduleByText agenda reunião por comando de texto
func (c *CalendarAgent) ScheduleByText(ctx context.Context, command string) (*Event, error) {
	// Usa LLM para extrair informações do comando
	prompt := fmt.Sprintf(`Extraia as informações de agendamento deste comando:
"%s"

Responda em JSON:
{
  "participant": "nome da pessoa",
  "period": "manhã/tarde/noite",
  "day": "hoje/amanhã/próxima semana/YYYY-MM-DD",
  "duration_minutes": 30,
  "subject": "assunto da reunião"
}`, command)

	response, err := c.llm.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	// TODO: Parse JSON response
	_ = response

	// Encontra horário disponível
	now := time.Now()
	tomorrow := now.AddDate(0, 0, 1)
	slots, err := c.calendarService.GetFreeSlots(now, tomorrow, 30*time.Minute)
	if err != nil {
		return nil, err
	}

	if len(slots) == 0 {
		return nil, fmt.Errorf("nenhum horário disponível")
	}

	// Cria evento
	event := &Event{
		Title: "Reunião",
		Start: slots[0].Start,
		End:   slots[0].End,
	}

	err = c.calendarService.CreateEvent(event.Title, event.Start, event.End, "")
	return event, err
}

// ==================== 12. ATA DE REUNIÃO ====================

// GenerateMeetingNotes gera ata de reunião
func (c *CalendarAgent) GenerateMeetingNotes(ctx context.Context, transcript string, attendees []string) (string, error) {
	prompt := fmt.Sprintf(`Você é um assistente que cria atas de reunião.

Transcrição da reunião:
%s

Participantes: %s

Crie uma ata estruturada com:
1. Resumo (2-3 frases)
2. Decisões tomadas (bullet points)
3. Itens de ação (quem, o quê, prazo)
4. Próximos passos

Ata:`, transcript, strings.Join(attendees, ", "))

	return c.llm.Generate(ctx, prompt)
}

// SendMeetingNotes envia ata por e-mail
func (c *CalendarAgent) SendMeetingNotes(ctx context.Context, transcript string, attendees []string) error {
	notes, err := c.GenerateMeetingNotes(ctx, transcript, attendees)
	if err != nil {
		return err
	}

	for _, attendee := range attendees {
		c.emailService.SendEmail(
			attendee,
			"Ata da Reunião",
			notes,
		)
	}

	return nil
}

// ==================== 13. BLOQUEIO DE FOCO ====================

// BlockFocusTime bloqueia tempo para foco
func (c *CalendarAgent) BlockFocusTime(ctx context.Context, duration time.Duration, taskDescription string) ([]TimeSlot, error) {
	// Encontra buracos na agenda
	now := time.Now()
	weekEnd := now.AddDate(0, 0, 7)

	slots, err := c.calendarService.GetFreeSlots(now, weekEnd, duration)
	if err != nil {
		return nil, err
	}

	// Filtra slots ideais para foco (manhã ou final da tarde)
	idealSlots := make([]TimeSlot, 0)
	for _, slot := range slots {
		hour := slot.Start.Hour()
		// Manhã (8-12) ou final da tarde (15-18)
		if (hour >= 8 && hour <= 12) || (hour >= 15 && hour <= 18) {
			idealSlots = append(idealSlots, slot)
		}
	}

	// Cria bloqueios de foco
	for i, slot := range idealSlots[:min(3, len(idealSlots))] {
		title := fmt.Sprintf("🎯 Foco: %s", taskDescription)
		if taskDescription == "" {
			title = "🎯 Tempo de Foco"
		}
		c.calendarService.CreateEvent(title, slot.Start, slot.End, "Tempo bloqueado para trabalho focado")
		idealSlots[i] = slot
	}

	return idealSlots, nil
}

// ==================== 14. GESTÃO DE FUSO HORÁRIO ====================

// SuggestMeetingTime sugere horários considerando fusos
func (c *CalendarAgent) SuggestMeetingTime(ctx context.Context, participants []string, timezones []string, duration time.Duration) ([]MeetingSuggestion, error) {
	now := time.Now()
	weekEnd := now.AddDate(0, 0, 7)

	// Obtém slots livres
	slots, err := c.calendarService.GetFreeSlots(now, weekEnd, duration)
	if err != nil {
		return nil, err
	}

	suggestions := make([]MeetingSuggestion, 0)

	for _, slot := range slots {
		// Verifica se é horário comercial em todos os fusos
		isGoodForAll := true
		for _, tz := range timezones {
			loc, err := time.LoadLocation(tz)
			if err != nil {
				continue
			}
			remoteTime := slot.Start.In(loc)
			hour := remoteTime.Hour()
			// Entre 9h e 18h
			if hour < 9 || hour > 18 {
				isGoodForAll = false
				break
			}
		}

		if isGoodForAll {
			suggestion := MeetingSuggestion{
				Slot:      slot,
				LocalTime: slot.Start.Format("15:04"),
			}
			if len(timezones) > 0 {
				loc, _ := time.LoadLocation(timezones[0])
				suggestion.RemoteTime = slot.Start.In(loc).Format("15:04")
				suggestion.RemoteTZ = timezones[0]
			}
			suggestions = append(suggestions, suggestion)
		}
	}

	return suggestions[:min(5, len(suggestions))], nil
}

// ==================== 15. LINK DE CONVITE INTELIGENTE ====================

// CreateMeetingWithLink cria reunião com link
func (c *CalendarAgent) CreateMeetingWithLink(ctx context.Context, title string, start, end time.Time, attendees []string, platform string) (string, error) {
	var meetingLink string

	// Gera link baseado na plataforma
	switch strings.ToLower(platform) {
	case "meet", "google meet":
		// TODO: Usar API do Google Meet
		meetingLink = "https://meet.google.com/xxx-xxxx-xxx"
	case "zoom":
		// TODO: Usar API do Zoom
		meetingLink = "https://zoom.us/j/123456789"
	case "teams":
		// TODO: Usar API do Teams
		meetingLink = "https://teams.microsoft.com/l/meetup-join/..."
	default:
		meetingLink = "https://meet.google.com/xxx-xxxx-xxx"
	}

	description := fmt.Sprintf("Link da reunião: %s\n\nParticipantes: %s",
		meetingLink, strings.Join(attendees, ", "))

	err := c.calendarService.CreateEvent(title, start, end, description)
	if err != nil {
		return "", err
	}

	// Envia convites por e-mail
	for _, attendee := range attendees {
		emailBody := fmt.Sprintf(`Você foi convidado para: %s

📅 Data: %s
⏰ Horário: %s - %s
🔗 Link: %s

Aceitar | Recusar | Talvez`,
			title,
			start.Format("02/01/2006"),
			start.Format("15:04"),
			end.Format("15:04"),
			meetingLink)

		c.emailService.SendEmail(attendee, fmt.Sprintf("Convite: %s", title), emailBody)
	}

	return meetingLink, nil
}

// ==================== CONFLITOS ====================

// CheckConflicts verifica conflitos na agenda
func (c *CalendarAgent) CheckConflicts(ctx context.Context, start, end time.Time) ([]Event, error) {
	events, err := c.calendarService.GetEvents(start, end)
	if err != nil {
		return nil, err
	}

	conflicts := make([]Event, 0)
	for _, event := range events {
		if event.Start.Before(end) && event.End.After(start) {
			conflicts = append(conflicts, event)
		}
	}

	return conflicts, nil
}

// SuggestReschedule sugere novo horário para evento conflitante
func (c *CalendarAgent) SuggestReschedule(ctx context.Context, conflictingEvent Event) ([]TimeSlot, error) {
	duration := conflictingEvent.End.Sub(conflictingEvent.Start)
	now := time.Now()
	weekEnd := now.AddDate(0, 0, 7)

	return c.calendarService.GetFreeSlots(now, weekEnd, duration)
}
