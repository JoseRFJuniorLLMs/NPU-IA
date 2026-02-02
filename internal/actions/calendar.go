package actions

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// CalendarClient cliente para Google Calendar
type CalendarClient struct {
	service *calendar.Service
}

// Event representa um evento
type Event struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Location    string    `json:"location"`
	Start       time.Time `json:"start"`
	End         time.Time `json:"end"`
	AllDay      bool      `json:"all_day"`
}

// NewCalendarClient cria cliente do Calendar
func NewCalendarClient(credentialsPath string) (*CalendarClient, error) {
	ctx := context.Background()

	credentials, err := os.ReadFile(credentialsPath)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler credenciais: %w", err)
	}

	config, err := google.ConfigFromJSON(credentials, calendar.CalendarReadonlyScope, calendar.CalendarEventsScope)
	if err != nil {
		return nil, err
	}

	token, err := getToken(config)
	if err != nil {
		return nil, err
	}

	client := config.Client(ctx, token)
	service, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}

	return &CalendarClient{service: service}, nil
}

// GetTodayEvents retorna eventos de hoje
func (c *CalendarClient) GetTodayEvents() ([]*Event, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	return c.GetEvents(startOfDay, endOfDay)
}

// GetEvents retorna eventos em um período
func (c *CalendarClient) GetEvents(start, end time.Time) ([]*Event, error) {
	events, err := c.service.Events.List("primary").
		TimeMin(start.Format(time.RFC3339)).
		TimeMax(end.Format(time.RFC3339)).
		SingleEvents(true).
		OrderBy("startTime").
		Do()
	if err != nil {
		return nil, err
	}

	result := make([]*Event, 0, len(events.Items))
	for _, item := range events.Items {
		event := &Event{
			ID:          item.Id,
			Title:       item.Summary,
			Description: item.Description,
			Location:    item.Location,
		}

		// Parse datas
		if item.Start.DateTime != "" {
			event.Start, _ = time.Parse(time.RFC3339, item.Start.DateTime)
			event.End, _ = time.Parse(time.RFC3339, item.End.DateTime)
		} else {
			event.AllDay = true
			event.Start, _ = time.Parse("2006-01-02", item.Start.Date)
			event.End, _ = time.Parse("2006-01-02", item.End.Date)
		}

		result = append(result, event)
	}

	return result, nil
}

// CreateEvent cria um novo evento
func (c *CalendarClient) CreateEvent(title string, start, end time.Time, description string) (*Event, error) {
	event := &calendar.Event{
		Summary:     title,
		Description: description,
		Start: &calendar.EventDateTime{
			DateTime: start.Format(time.RFC3339),
		},
		End: &calendar.EventDateTime{
			DateTime: end.Format(time.RFC3339),
		},
	}

	created, err := c.service.Events.Insert("primary", event).Do()
	if err != nil {
		return nil, err
	}

	return &Event{
		ID:          created.Id,
		Title:       created.Summary,
		Description: created.Description,
		Start:       start,
		End:         end,
	}, nil
}

// DeleteEvent deleta um evento
func (c *CalendarClient) DeleteEvent(eventID string) error {
	return c.service.Events.Delete("primary", eventID).Do()
}

// Summarize cria resumo da agenda
func (c *CalendarClient) Summarize() (string, error) {
	events, err := c.GetTodayEvents()
	if err != nil {
		return "", err
	}

	if len(events) == 0 {
		return "Você não tem eventos agendados para hoje.", nil
	}

	var summary strings.Builder
	summary.WriteString(fmt.Sprintf("Você tem %d eventos hoje:\n\n", len(events)))

	for i, event := range events {
		if event.AllDay {
			summary.WriteString(fmt.Sprintf("%d. %s (dia inteiro)\n", i+1, event.Title))
		} else {
			summary.WriteString(fmt.Sprintf("%d. %s às %s\n",
				i+1, event.Title, event.Start.Format("15:04")))
		}
	}

	return summary.String(), nil
}

// GetNextEvent retorna o próximo evento
func (c *CalendarClient) GetNextEvent() (*Event, error) {
	now := time.Now()
	end := now.Add(7 * 24 * time.Hour) // Próximos 7 dias

	events, err := c.GetEvents(now, end)
	if err != nil {
		return nil, err
	}

	if len(events) == 0 {
		return nil, nil
	}

	return events[0], nil
}
