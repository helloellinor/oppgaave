package calendar

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

// RecurrenceType defines the type of recurring event
type RecurrenceType string

const (
	RecurrenceNone      RecurrenceType = "none"
	RecurrenceDaily     RecurrenceType = "daily"
	RecurrenceWeekly    RecurrenceType = "weekly"
	RecurrenceMonthly   RecurrenceType = "monthly"
	RecurrenceYearly    RecurrenceType = "yearly"
	RecurrenceCustom    RecurrenceType = "custom"
)

// Event represents a calendar event
type Event struct {
	ID          string         `json:"id"`
	Title       string         `json:"title"`
	Description string         `json:"description,omitempty"`
	Location    string         `json:"location,omitempty"`
	StartTime   time.Time      `json:"start_time"`
	EndTime     time.Time      `json:"end_time"`
	AllDay      bool           `json:"all_day"`
	Recurrence  RecurrenceRule `json:"recurrence,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	Tags        []string       `json:"tags,omitempty"`
}

// RecurrenceRule defines how an event recurs
type RecurrenceRule struct {
	Type      RecurrenceType `json:"type"`
	Interval  int            `json:"interval"`          // Every N days/weeks/months/years
	EndDate   *time.Time     `json:"end_date,omitempty"` // When recurrence ends
	Count     int            `json:"count,omitempty"`    // Number of occurrences
	WeekDays  []time.Weekday `json:"week_days,omitempty"` // For weekly recurrence
	MonthDay  int            `json:"month_day,omitempty"` // Day of month for monthly recurrence
}

// Calendar manages a collection of events
type Calendar struct {
	events map[string]*Event
}

// NewCalendar creates a new calendar instance
func NewCalendar() *Calendar {
	return &Calendar{
		events: make(map[string]*Event),
	}
}

// AddEvent adds a new event to the calendar
func (c *Calendar) AddEvent(event *Event) error {
	if event == nil {
		return fmt.Errorf("event cannot be nil")
	}

	// Generate ID if not provided
	if event.ID == "" {
		event.ID = uuid.New().String()
	}

	// Validate event
	if err := c.validateEvent(event); err != nil {
		return fmt.Errorf("invalid event: %w", err)
	}

	// Set timestamps
	now := time.Now()
	if event.CreatedAt.IsZero() {
		event.CreatedAt = now
	}
	event.UpdatedAt = now

	// Check for conflicts
	if conflicts := c.FindConflicts(event); len(conflicts) > 0 {
		return fmt.Errorf("event conflicts with existing events: %v", conflicts)
	}

	c.events[event.ID] = event
	return nil
}

// UpdateEvent updates an existing event
func (c *Calendar) UpdateEvent(event *Event) error {
	if event == nil {
		return fmt.Errorf("event cannot be nil")
	}

	if event.ID == "" {
		return fmt.Errorf("event ID is required for update")
	}

	// Check if event exists
	if _, exists := c.events[event.ID]; !exists {
		return fmt.Errorf("event with ID %s not found", event.ID)
	}

	// Validate event
	if err := c.validateEvent(event); err != nil {
		return fmt.Errorf("invalid event: %w", err)
	}

	// Update timestamp
	event.UpdatedAt = time.Now()

	c.events[event.ID] = event
	return nil
}

// RemoveEvent removes an event from the calendar
func (c *Calendar) RemoveEvent(eventID string) error {
	if eventID == "" {
		return fmt.Errorf("event ID is required")
	}

	if _, exists := c.events[eventID]; !exists {
		return fmt.Errorf("event with ID %s not found", eventID)
	}

	delete(c.events, eventID)
	return nil
}

// GetEvent retrieves an event by ID
func (c *Calendar) GetEvent(eventID string) (*Event, error) {
	if eventID == "" {
		return nil, fmt.Errorf("event ID is required")
	}

	event, exists := c.events[eventID]
	if !exists {
		return nil, fmt.Errorf("event with ID %s not found", eventID)
	}

	return event, nil
}

// ListEvents returns all events, optionally filtered by date range
func (c *Calendar) ListEvents(startDate, endDate *time.Time) []*Event {
	var events []*Event

	for _, event := range c.events {
		// Apply date filter if provided
		if startDate != nil && event.EndTime.Before(*startDate) {
			continue
		}
		if endDate != nil && event.StartTime.After(*endDate) {
			continue
		}

		events = append(events, event)
	}

	// Sort events by start time
	sort.Slice(events, func(i, j int) bool {
		return events[i].StartTime.Before(events[j].StartTime)
	})

	return events
}

// ListEventsForDay returns all events for a specific day
func (c *Calendar) ListEventsForDay(date time.Time) []*Event {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	return c.ListEvents(&startOfDay, &endOfDay)
}

// ListEventsForWeek returns all events for a specific week
func (c *Calendar) ListEventsForWeek(date time.Time) []*Event {
	// Find the start of the week (Monday)
	weekday := int(date.Weekday())
	if weekday == 0 { // Sunday
		weekday = 7
	}
	startOfWeek := date.AddDate(0, 0, -(weekday-1))
	startOfWeek = time.Date(startOfWeek.Year(), startOfWeek.Month(), startOfWeek.Day(), 0, 0, 0, 0, date.Location())
	endOfWeek := startOfWeek.Add(7 * 24 * time.Hour)

	return c.ListEvents(&startOfWeek, &endOfWeek)
}

// ListEventsForMonth returns all events for a specific month
func (c *Calendar) ListEventsForMonth(year int, month time.Month) []*Event {
	startOfMonth := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
	endOfMonth := startOfMonth.AddDate(0, 1, 0)

	return c.ListEvents(&startOfMonth, &endOfMonth)
}

// FindConflicts finds events that conflict with the given event
func (c *Calendar) FindConflicts(event *Event) []*Event {
	var conflicts []*Event

	for _, existingEvent := range c.events {
		// Skip the same event (for updates)
		if existingEvent.ID == event.ID {
			continue
		}

		// Check for time overlap
		if c.eventsOverlap(event, existingEvent) {
			conflicts = append(conflicts, existingEvent)
		}
	}

	return conflicts
}

// SearchEvents searches for events by title, description, or location
func (c *Calendar) SearchEvents(query string) []*Event {
	var results []*Event
	query = strings.ToLower(query)

	for _, event := range c.events {
		if strings.Contains(strings.ToLower(event.Title), query) ||
			strings.Contains(strings.ToLower(event.Description), query) ||
			strings.Contains(strings.ToLower(event.Location), query) {
			results = append(results, event)
		}
	}

	// Sort results by start time
	sort.Slice(results, func(i, j int) bool {
		return results[i].StartTime.Before(results[j].StartTime)
	})

	return results
}

// GetEventsByTag returns events that have the specified tag
func (c *Calendar) GetEventsByTag(tag string) []*Event {
	var events []*Event

	for _, event := range c.events {
		for _, eventTag := range event.Tags {
			if eventTag == tag {
				events = append(events, event)
				break
			}
		}
	}

	// Sort events by start time
	sort.Slice(events, func(i, j int) bool {
		return events[i].StartTime.Before(events[j].StartTime)
	})

	return events
}

// GetAllEvents returns all events in the calendar
func (c *Calendar) GetAllEvents() []*Event {
	return c.ListEvents(nil, nil)
}

// GetEventCount returns the total number of events
func (c *Calendar) GetEventCount() int {
	return len(c.events)
}

// validateEvent validates an event's data
func (c *Calendar) validateEvent(event *Event) error {
	if event.Title == "" {
		return fmt.Errorf("event title is required")
	}

	if event.StartTime.IsZero() {
		return fmt.Errorf("event start time is required")
	}

	if event.EndTime.IsZero() {
		return fmt.Errorf("event end time is required")
	}

	if event.EndTime.Before(event.StartTime) {
		return fmt.Errorf("event end time must be after start time")
	}

	// Validate recurrence rule
	if event.Recurrence.Type != RecurrenceNone {
		if err := c.validateRecurrenceRule(&event.Recurrence); err != nil {
			return fmt.Errorf("invalid recurrence rule: %w", err)
		}
	}

	return nil
}

// validateRecurrenceRule validates a recurrence rule
func (c *Calendar) validateRecurrenceRule(rule *RecurrenceRule) error {
	if rule.Interval <= 0 {
		return fmt.Errorf("recurrence interval must be positive")
	}

	switch rule.Type {
	case RecurrenceDaily, RecurrenceWeekly, RecurrenceMonthly, RecurrenceYearly:
		// Valid types
	case RecurrenceCustom:
		// Custom recurrence requires additional validation
		if len(rule.WeekDays) == 0 && rule.MonthDay == 0 {
			return fmt.Errorf("custom recurrence requires weekdays or month day specification")
		}
	default:
		return fmt.Errorf("invalid recurrence type: %s", rule.Type)
	}

	if rule.EndDate != nil && rule.Count > 0 {
		return fmt.Errorf("recurrence cannot have both end date and count")
	}

	if rule.MonthDay < 0 || rule.MonthDay > 31 {
		return fmt.Errorf("invalid month day: %d", rule.MonthDay)
	}

	return nil
}

// eventsOverlap checks if two events overlap in time
func (c *Calendar) eventsOverlap(event1, event2 *Event) bool {
	// Events overlap if one starts before the other ends
	return event1.StartTime.Before(event2.EndTime) && event2.StartTime.Before(event1.EndTime)
}

// GenerateRecurringEvents generates recurring event instances for a given time range
func (c *Calendar) GenerateRecurringEvents(event *Event, startDate, endDate time.Time) ([]*Event, error) {
	if event.Recurrence.Type == RecurrenceNone {
		return []*Event{event}, nil
	}

	var instances []*Event
	current := event.StartTime
	count := 0

	for current.Before(endDate) || current.Equal(endDate) {
		// Check if we've reached the end conditions
		if event.Recurrence.EndDate != nil && current.After(*event.Recurrence.EndDate) {
			break
		}
		if event.Recurrence.Count > 0 && count >= event.Recurrence.Count {
			break
		}

		// Create instance if it's within the requested range
		if (current.After(startDate) || current.Equal(startDate)) && current.Before(endDate) {
			instance := *event // Copy the event
			instance.ID = fmt.Sprintf("%s-%d", event.ID, count)
			duration := event.EndTime.Sub(event.StartTime)
			instance.StartTime = current
			instance.EndTime = current.Add(duration)
			instances = append(instances, &instance)
		}

		// Calculate next occurrence
		switch event.Recurrence.Type {
		case RecurrenceDaily:
			current = current.AddDate(0, 0, event.Recurrence.Interval)
		case RecurrenceWeekly:
			current = current.AddDate(0, 0, 7*event.Recurrence.Interval)
		case RecurrenceMonthly:
			current = current.AddDate(0, event.Recurrence.Interval, 0)
		case RecurrenceYearly:
			current = current.AddDate(event.Recurrence.Interval, 0, 0)
		default:
			return nil, fmt.Errorf("unsupported recurrence type: %s", event.Recurrence.Type)
		}

		count++

		// Safety check to prevent infinite loops
		if count > 1000 {
			return nil, fmt.Errorf("too many recurring instances (limit: 1000)")
		}
	}

	return instances, nil
}
