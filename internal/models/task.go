package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"math"
	"time"
)

// TaskStatus represents the current state of a task
type TaskStatus string

const (
	StatusPending    TaskStatus = "pending"
	StatusInProgress TaskStatus = "in_progress"
	StatusDone       TaskStatus = "done"
	StatusBlocked    TaskStatus = "blocked"
)

// TaskType represents the type of task/event
type TaskType string

const (
	TypeTask        TaskType = "task"
	TypeAppointment TaskType = "appointment"
	TypeEvent       TaskType = "event"
	TypeConcert     TaskType = "concert"
	TypeMeeting     TaskType = "meeting"
)

// Task represents a task in our ADHD-friendly system
type Task struct {
	ID                     int       `json:"id" db:"id"`
	Title                  string    `json:"title" db:"title"`
	Description            string    `json:"description" db:"description"`
	ParentID               *int      `json:"parent_id" db:"parent_id"`
	EstimatedDurationMins  int       `json:"estimated_duration_minutes" db:"estimated_duration_minutes"`
	Deadline               *time.Time `json:"deadline" db:"deadline"`
	Priority               int       `json:"priority" db:"priority"`
	Status                 TaskStatus `json:"status" db:"status"`
	Tags                   Tags      `json:"tags" db:"tags"`
	EnergyLevel            int       `json:"energy_level" db:"energy_level"`
	Difficulty             int       `json:"difficulty" db:"difficulty"`
	MoneyCost              int       `json:"money_cost" db:"money_cost"`
	TaskType               TaskType  `json:"task_type" db:"task_type"`
	EventLocation          string    `json:"event_location" db:"event_location"`
	EventStart             *time.Time `json:"event_start" db:"event_start"`
	EventEnd               *time.Time `json:"event_end" db:"event_end"`
	RadarPositionX         float64   `json:"radar_position_x" db:"radar_position_x"`
	RadarPositionY         float64   `json:"radar_position_y" db:"radar_position_y"`
	CreatedAt              time.Time `json:"created_at" db:"created_at"`
	UpdatedAt              time.Time `json:"updated_at" db:"updated_at"`
	CompletedAt            *time.Time `json:"completed_at" db:"completed_at"`
	
	// Computed fields
	Subtasks      []Task      `json:"subtasks,omitempty"`
	Prerequisites []Task      `json:"prerequisites,omitempty"`
	Contacts      []Contact   `json:"contacts,omitempty"`
	Attachments   []Attachment `json:"attachments,omitempty"`
}

// Tags represents a list of task tags
type Tags []string

// Value implements the driver.Valuer interface for database storage
func (t Tags) Value() (driver.Value, error) {
	if len(t) == 0 {
		return "[]", nil
	}
	return json.Marshal(t)
}

// Scan implements the sql.Scanner interface for database retrieval
func (t *Tags) Scan(value interface{}) error {
	if value == nil {
		*t = Tags{}
		return nil
	}
	
	switch v := value.(type) {
	case string:
		return json.Unmarshal([]byte(v), t)
	case []byte:
		return json.Unmarshal(v, t)
	default:
		return fmt.Errorf("cannot scan %T into Tags", value)
	}
}

// DailyBudget represents the time/money budget for a day
type DailyBudget struct {
	ID              int       `json:"id" db:"id"`
	Date            time.Time `json:"date" db:"date"`
	TotalBudgetCoins int      `json:"total_budget_coins" db:"total_budget_coins"`
	SpentCoins      int       `json:"spent_coins" db:"spent_coins"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// RemainingCoins calculates remaining budget
func (db *DailyBudget) RemainingCoins() int {
	return db.TotalBudgetCoins - db.SpentCoins
}

// TaskSchedule represents when a task is scheduled
type TaskSchedule struct {
	ID               int        `json:"id" db:"id"`
	TaskID           int        `json:"task_id" db:"task_id"`
	ScheduledDate    time.Time  `json:"scheduled_date" db:"scheduled_date"`
	StartTime        *time.Time `json:"start_time" db:"start_time"`
	EstimatedEndTime *time.Time `json:"estimated_end_time" db:"estimated_end_time"`
	ActualStartTime  *time.Time `json:"actual_start_time" db:"actual_start_time"`
	ActualEndTime    *time.Time `json:"actual_end_time" db:"actual_end_time"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	
	// Associated task
	Task *Task `json:"task,omitempty"`
}

// TaskPrerequisite represents a prerequisite relationship
type TaskPrerequisite struct {
	ID                 int       `json:"id" db:"id"`
	TaskID             int       `json:"task_id" db:"task_id"`
	PrerequisiteTaskID int       `json:"prerequisite_task_id" db:"prerequisite_task_id"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
}

// CreateTaskRequest represents the request to create a new task
type CreateTaskRequest struct {
	Title                 string     `json:"title"`
	Description           string     `json:"description"`
	ParentID              *int       `json:"parent_id"`
	EstimatedDurationMins int        `json:"estimated_duration_minutes"`
	Deadline              *time.Time `json:"deadline"`
	Priority              int        `json:"priority"`
	Tags                  []string   `json:"tags"`
	EnergyLevel           int        `json:"energy_level"`
	Difficulty            int        `json:"difficulty"`
	TaskType              TaskType   `json:"task_type"`
	EventLocation         string     `json:"event_location"`
	EventStart            *time.Time `json:"event_start"`
	EventEnd              *time.Time `json:"event_end"`
}

// Contact represents a person or organization
type Contact struct {
	ID        int       `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Email     string    `json:"email" db:"email"`
	Phone     string    `json:"phone" db:"phone"`
	Type      string    `json:"type" db:"type"` // person, organization, venue
	Notes     string    `json:"notes" db:"notes"`
	AvatarURL string    `json:"avatar_url" db:"avatar_url"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// ContactThread represents a communication thread with a contact
type ContactThread struct {
	ID         int       `json:"id" db:"id"`
	ContactID  int       `json:"contact_id" db:"contact_id"`
	TaskID     *int      `json:"task_id" db:"task_id"`
	Subject    string    `json:"subject" db:"subject"`
	Message    string    `json:"message" db:"message"`
	ThreadType string    `json:"thread_type" db:"thread_type"` // message, email, call, meeting
	Direction  string    `json:"direction" db:"direction"`     // inbound, outbound
	Status     string    `json:"status" db:"status"`           // sent, received, pending, failed
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	
	// Associated contact
	Contact *Contact `json:"contact,omitempty"`
}

// Attachment represents a file attachment
type Attachment struct {
	ID               int       `json:"id" db:"id"`
	TaskID           *int      `json:"task_id" db:"task_id"`
	ContactID        *int      `json:"contact_id" db:"contact_id"`
	Filename         string    `json:"filename" db:"filename"`
	OriginalFilename string    `json:"original_filename" db:"original_filename"`
	FilePath         string    `json:"file_path" db:"file_path"`
	FileSize         int64     `json:"file_size" db:"file_size"`
	MimeType         string    `json:"mime_type" db:"mime_type"`
	Description      string    `json:"description" db:"description"`
	AttachmentType   string    `json:"attachment_type" db:"attachment_type"` // document, image, audio, video, link
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
}

// TaskContact represents the relationship between a task and contact
type TaskContact struct {
	ID        int       `json:"id" db:"id"`
	TaskID    int       `json:"task_id" db:"task_id"`
	ContactID int       `json:"contact_id" db:"contact_id"`
	Role      string    `json:"role" db:"role"` // organizer, participant, venue, vendor
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// CalculateMoneyCost calculates the "cost" of a task in our money allegory
func (t *Task) CalculateMoneyCost() int {
	baseCost := t.EstimatedDurationMins
	
	// Apply energy multiplier
	energyMultiplier := 1.0
	switch t.EnergyLevel {
	case 1: // Low energy
		energyMultiplier = 0.8
	case 2: // Medium energy
		energyMultiplier = 1.0
	case 3: // High energy
		energyMultiplier = 1.5
	}
	
	// Apply difficulty multiplier
	difficultyMultiplier := 1.0
	switch t.Difficulty {
	case 1: // Easy
		difficultyMultiplier = 0.9
	case 2: // Medium
		difficultyMultiplier = 1.0
	case 3: // Hard
		difficultyMultiplier = 1.3
	}
	
	// Apply priority multiplier (higher priority costs more to reflect urgency)
	priorityMultiplier := 1.0
	switch t.Priority {
	case 1: // Low priority
		priorityMultiplier = 0.8
	case 2: // Medium priority
		priorityMultiplier = 1.0
	case 3: // High priority
		priorityMultiplier = 1.2
	}
	
	cost := float64(baseCost) * energyMultiplier * difficultyMultiplier * priorityMultiplier
	return int(cost)
}

// GetUrgencyColor returns a CSS class based on deadline proximity and priority
func (t *Task) GetUrgencyColor() string {
	if t.Deadline == nil {
		switch t.Priority {
		case 3:
			return "high-priority"
		case 2:
			return "medium-priority"
		default:
			return "low-priority"
		}
	}
	
	timeUntilDeadline := time.Until(*t.Deadline)
	
	if timeUntilDeadline < 0 {
		return "overdue"
	} else if timeUntilDeadline < 24*time.Hour {
		return "urgent"
	} else if timeUntilDeadline < 3*24*time.Hour {
		return "soon"
	}
	
	return "normal"
}

// IsBlocked checks if a task is blocked by incomplete prerequisites
func (t *Task) IsBlocked() bool {
	for _, prereq := range t.Prerequisites {
		if prereq.Status != StatusDone {
			return true
		}
	}
	return false
}

// IsEvent returns true if this is an event-type task
func (t *Task) IsEvent() bool {
	return t.TaskType == TypeAppointment || t.TaskType == TypeEvent || 
		   t.TaskType == TypeConcert || t.TaskType == TypeMeeting
}

// GetEventDuration returns the duration of an event
func (t *Task) GetEventDuration() time.Duration {
	if t.EventStart != nil && t.EventEnd != nil {
		return t.EventEnd.Sub(*t.EventStart)
	}
	return time.Duration(t.EstimatedDurationMins) * time.Minute
}

// GetTaskTypeIcon returns an icon for the task type
func (t *Task) GetTaskTypeIcon() string {
	switch t.TaskType {
	case TypeAppointment:
		return "📅"
	case TypeEvent:
		return "🎉"
	case TypeConcert:
		return "🎵"
	case TypeMeeting:
		return "💼"
	default:
		return "📋"
	}
}

// CalculateRadarPosition calculates the radar position based on time and priority
func (t *Task) CalculateRadarPosition() {
	// X-axis: time-based (distance from now)
	now := time.Now()
	var timeDistance float64
	
	if t.EventStart != nil {
		timeDistance = t.EventStart.Sub(now).Hours()
	} else if t.Deadline != nil {
		timeDistance = t.Deadline.Sub(now).Hours()
	} else {
		timeDistance = 24 // Default to 1 day out
	}
	
	// Normalize to 0-100 range (0-168 hours = 1 week)
	t.RadarPositionX = math.Max(0, math.Min(100, (timeDistance/168)*100))
	
	// Y-axis: priority and energy level combination
	priorityWeight := float64(t.Priority) * 20    // 0-60
	energyWeight := float64(t.EnergyLevel) * 15   // 0-45
	t.RadarPositionY = math.Min(100, priorityWeight + energyWeight)
}