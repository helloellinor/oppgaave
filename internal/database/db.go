package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"oppgaave/internal/models"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	conn *sql.DB
}

// New creates a new database connection and initializes schema
func New(dbPath string) (*DB, error) {
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db := &DB{conn: conn}
	
	if err := db.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return db, nil
}

// initSchema initializes database schema and runs migrations
func (db *DB) initSchema() error {
	// First, create core tables
	if err := db.createCoreTables(); err != nil {
		return fmt.Errorf("failed to create core tables: %w", err)
	}

	// Run migrations to add new columns
	if err := db.runMigrations(); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Insert sample data
	if err := db.insertSampleData(); err != nil {
		return fmt.Errorf("failed to insert sample data: %w", err)
	}

	log.Println("Database schema initialized successfully")
	return nil
}

// createCoreTables creates the core database tables
func (db *DB) createCoreTables() error {
	coreSchema := `
-- ADHD Task Management System Database Schema

-- Tasks table with recursive structure
CREATE TABLE IF NOT EXISTS tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    description TEXT,
    parent_id INTEGER, -- For subtasks
    estimated_duration_minutes INTEGER DEFAULT 30,
    deadline DATETIME,
    priority INTEGER DEFAULT 1, -- 1=low, 2=medium, 3=high
    status TEXT DEFAULT 'pending', -- pending, in_progress, done, blocked
    tags TEXT, -- JSON array of tags
    energy_level INTEGER DEFAULT 2, -- 1=low, 2=medium, 3=high energy needed
    difficulty INTEGER DEFAULT 2, -- 1=easy, 2=medium, 3=hard
    money_cost INTEGER DEFAULT 0, -- Time budget cost in "coins"
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    completed_at DATETIME,
    FOREIGN KEY (parent_id) REFERENCES tasks(id)
);

-- Task prerequisites (DAG structure)
CREATE TABLE IF NOT EXISTS task_prerequisites (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id INTEGER NOT NULL,
    prerequisite_task_id INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (task_id) REFERENCES tasks(id),
    FOREIGN KEY (prerequisite_task_id) REFERENCES tasks(id),
    UNIQUE(task_id, prerequisite_task_id)
);

-- Daily budgets for time management
CREATE TABLE IF NOT EXISTS daily_budgets (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date DATE NOT NULL UNIQUE,
    total_budget_coins INTEGER DEFAULT 500, -- Daily budget in "coins"
    spent_coins INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Task assignments to days
CREATE TABLE IF NOT EXISTS task_schedule (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id INTEGER NOT NULL,
    scheduled_date DATE NOT NULL,
    start_time TIME,
    estimated_end_time TIME,
    actual_start_time DATETIME,
    actual_end_time DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (task_id) REFERENCES tasks(id)
);

-- User settings and preferences
CREATE TABLE IF NOT EXISTS settings (
    key TEXT PRIMARY KEY,
    value TEXT,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Contacts for communication and task management
CREATE TABLE IF NOT EXISTS contacts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    email TEXT,
    phone TEXT,
    type TEXT DEFAULT 'person', -- person, organization, venue
    notes TEXT,
    avatar_url TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Contact threads for communication history
CREATE TABLE IF NOT EXISTS contact_threads (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    contact_id INTEGER NOT NULL,
    task_id INTEGER, -- Optional link to task
    subject TEXT,
    message TEXT NOT NULL,
    thread_type TEXT DEFAULT 'message', -- message, email, call, meeting
    direction TEXT DEFAULT 'outbound', -- inbound, outbound
    status TEXT DEFAULT 'sent', -- sent, received, pending, failed
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (contact_id) REFERENCES contacts(id),
    FOREIGN KEY (task_id) REFERENCES tasks(id)
);

-- Attachments for tasks and events
CREATE TABLE IF NOT EXISTS attachments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id INTEGER,
    contact_id INTEGER,
    filename TEXT NOT NULL,
    original_filename TEXT NOT NULL,
    file_path TEXT NOT NULL,
    file_size INTEGER,
    mime_type TEXT,
    description TEXT,
    attachment_type TEXT DEFAULT 'document', -- document, image, audio, video, link
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (task_id) REFERENCES tasks(id),
    FOREIGN KEY (contact_id) REFERENCES contacts(id)
);

-- Task contacts relationship (many-to-many)
CREATE TABLE IF NOT EXISTS task_contacts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id INTEGER NOT NULL,
    contact_id INTEGER NOT NULL,
    role TEXT DEFAULT 'participant', -- organizer, participant, venue, vendor
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (task_id) REFERENCES tasks(id),
    FOREIGN KEY (contact_id) REFERENCES contacts(id),
    UNIQUE(task_id, contact_id)
);
`

	if _, err := db.conn.Exec(coreSchema); err != nil {
		return fmt.Errorf("failed to execute core schema: %w", err)
	}

	return nil
}

// runMigrations adds new columns for existing databases
func (db *DB) runMigrations() error {
	migrations := []string{
		// Add new task columns for radar visualization and events
		`ALTER TABLE tasks ADD COLUMN task_type TEXT DEFAULT 'task'`,
		`ALTER TABLE tasks ADD COLUMN event_location TEXT`,
		`ALTER TABLE tasks ADD COLUMN event_start DATETIME`,
		`ALTER TABLE tasks ADD COLUMN event_end DATETIME`,
		`ALTER TABLE tasks ADD COLUMN radar_position_x REAL DEFAULT 0`,
		`ALTER TABLE tasks ADD COLUMN radar_position_y REAL DEFAULT 0`,
	}

	for _, migration := range migrations {
		if _, err := db.conn.Exec(migration); err != nil {
			// Ignore "duplicate column name" errors - column already exists
			if !isColumnExistsError(err) {
				return fmt.Errorf("failed to run migration '%s': %w", migration, err)
			}
		}
	}

	return nil
}

// isColumnExistsError checks if the error is due to column already existing
func isColumnExistsError(err error) bool {
	return err != nil && (
		err.Error() == "duplicate column name: task_type" ||
		err.Error() == "duplicate column name: event_location" ||
		err.Error() == "duplicate column name: event_start" ||
		err.Error() == "duplicate column name: event_end" ||
		err.Error() == "duplicate column name: radar_position_x" ||
		err.Error() == "duplicate column name: radar_position_y")
}

// insertSampleData inserts initial settings and sample data
func (db *DB) insertSampleData() error {
	sampleData := `
-- Initial settings
INSERT OR REPLACE INTO settings (key, value) VALUES 
    ('daily_budget_coins', '500'),
    ('coin_per_minute', '10'),
    ('energy_multiplier', '1.5'),
    ('difficulty_multiplier', '1.3');

-- Sample tasks for demonstration
INSERT OR REPLACE INTO tasks (id, title, description, estimated_duration_minutes, priority, status, energy_level, difficulty, money_cost, task_type, event_start, event_end) VALUES
    (1, 'Morning Coffee & Journal', 'Start the day with coffee and journaling to set intentions', 15, 1, 'pending', 1, 1, 15, 'task', NULL, NULL),
    (2, 'Write Project Draft', 'Complete first draft of the project proposal', 120, 3, 'pending', 3, 3, 180, 'task', NULL, NULL),
    (3, 'Yoga Class', 'Attend morning yoga session for physical and mental wellness', 45, 2, 'pending', 2, 1, 45, 'appointment', '2025-08-12 09:00:00', '2025-08-12 09:45:00'),
    (4, 'Call Landlord', 'Important call about lease renewal - deadline approaching', 30, 3, 'pending', 2, 2, 60, 'task', NULL, NULL),
    (5, 'Grocery Shopping', 'Buy ingredients for meal prep', 60, 2, 'pending', 2, 2, 60, 'task', NULL, NULL),
    (6, 'Meal Prep', 'Prepare meals for the week', 90, 2, 'pending', 3, 2, 90, 'task', NULL, NULL),
    (7, 'Team Meeting', 'Weekly standup with development team', 60, 2, 'pending', 2, 1, 60, 'meeting', '2025-08-12 14:00:00', '2025-08-12 15:00:00'),
    (8, 'Concert Planning', 'Plan upcoming jazz concert attendance', 30, 1, 'pending', 1, 1, 30, 'event', '2025-08-15 19:00:00', '2025-08-15 22:00:00');

-- Sample contacts
INSERT OR REPLACE INTO contacts (id, name, email, phone, type, notes) VALUES
    (1, 'Dr. Sarah Johnson', 'sarah.johnson@yogastudio.com', '+1-555-0123', 'person', 'Yoga instructor'),
    (2, 'Development Team', 'team@company.com', NULL, 'organization', 'Work team for standups'),
    (3, 'Jazz Venue', 'info@jazzclub.com', '+1-555-0456', 'venue', 'Downtown jazz club'),
    (4, 'Property Manager', 'landlord@property.com', '+1-555-0789', 'person', 'Lease renewal contact');

-- Link contacts to tasks
INSERT OR REPLACE INTO task_contacts (task_id, contact_id, role) VALUES
    (3, 1, 'organizer'), -- Yoga class with instructor
    (7, 2, 'participant'), -- Team meeting
    (8, 3, 'venue'); -- Concert at jazz venue

-- Add some prerequisites
INSERT OR REPLACE INTO task_prerequisites (task_id, prerequisite_task_id) VALUES
    (6, 5), -- Meal prep requires grocery shopping first
    (2, 1); -- Writing requires coffee/journal first for focus
`

	if _, err := db.conn.Exec(sampleData); err != nil {
		return fmt.Errorf("failed to insert sample data: %w", err)
	}

	return nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.conn.Close()
}

// CreateTask creates a new task
func (db *DB) CreateTask(req *models.CreateTaskRequest) (*models.Task, error) {
	task := &models.Task{
		Title:                 req.Title,
		Description:           req.Description,
		ParentID:              req.ParentID,
		EstimatedDurationMins: req.EstimatedDurationMins,
		Deadline:              req.Deadline,
		Priority:              req.Priority,
		Tags:                  models.Tags(req.Tags),
		EnergyLevel:           req.EnergyLevel,
		Difficulty:            req.Difficulty,
		TaskType:              req.TaskType,
		EventLocation:         req.EventLocation,
		EventStart:            req.EventStart,
		EventEnd:              req.EventEnd,
		Status:                models.StatusPending,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}
	
	// Calculate money cost
	task.MoneyCost = task.CalculateMoneyCost()
	
	// Calculate radar position
	task.CalculateRadarPosition()

	query := `
		INSERT INTO tasks (title, description, parent_id, estimated_duration_minutes, 
			deadline, priority, status, tags, energy_level, difficulty, money_cost,
			task_type, event_location, event_start, event_end, radar_position_x, radar_position_y,
			created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := db.conn.Exec(query, task.Title, task.Description, task.ParentID,
		task.EstimatedDurationMins, task.Deadline, task.Priority, task.Status,
		task.Tags, task.EnergyLevel, task.Difficulty, task.MoneyCost,
		task.TaskType, task.EventLocation, task.EventStart, task.EventEnd,
		task.RadarPositionX, task.RadarPositionY, task.CreatedAt, task.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get task ID: %w", err)
	}

	task.ID = int(id)
	return task, nil
}

// GetTask retrieves a task by ID with its prerequisites and subtasks
func (db *DB) GetTask(id int) (*models.Task, error) {
	task := &models.Task{}
	var (
		parentID sql.NullInt64
		deadline, eventStart, eventEnd, completedAt sql.NullTime
		description, eventLocation sql.NullString
	)
	
	query := `
		SELECT id, title, description, parent_id, estimated_duration_minutes,
			deadline, priority, status, tags, energy_level, difficulty, money_cost,
			task_type, event_location, event_start, event_end, radar_position_x, radar_position_y,
			created_at, updated_at, completed_at
		FROM tasks WHERE id = ?`

	err := db.conn.QueryRow(query, id).Scan(
		&task.ID, &task.Title, &description, &parentID,
		&task.EstimatedDurationMins, &deadline, &task.Priority,
		&task.Status, &task.Tags, &task.EnergyLevel, &task.Difficulty,
		&task.MoneyCost, &task.TaskType, &eventLocation, &eventStart,
		&eventEnd, &task.RadarPositionX, &task.RadarPositionY,
		&task.CreatedAt, &task.UpdatedAt, &completedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	// Handle nullable fields
	if parentID.Valid {
		task.ParentID = &[]int{int(parentID.Int64)}[0]
	}
	if deadline.Valid {
		task.Deadline = &deadline.Time
	}
	if description.Valid {
		task.Description = description.String
	}
	if eventLocation.Valid {
		task.EventLocation = eventLocation.String
	}
	if eventStart.Valid {
		task.EventStart = &eventStart.Time
	}
	if eventEnd.Valid {
		task.EventEnd = &eventEnd.Time
	}
	if completedAt.Valid {
		task.CompletedAt = &completedAt.Time
	}

	// Load prerequisites
	if err := db.loadTaskPrerequisites(task); err != nil {
		return nil, fmt.Errorf("failed to load prerequisites: %w", err)
	}

	// Load subtasks
	if err := db.loadTaskSubtasks(task); err != nil {
		return nil, fmt.Errorf("failed to load subtasks: %w", err)
	}

	// Load contacts
	if err := db.loadTaskContacts(task); err != nil {
		return nil, fmt.Errorf("failed to load contacts: %w", err)
	}

	// Load attachments
	if err := db.loadTaskAttachments(task); err != nil {
		return nil, fmt.Errorf("failed to load attachments: %w", err)
	}

	return task, nil
}

// GetAllTasks retrieves all tasks
func (db *DB) GetAllTasks() ([]models.Task, error) {
	query := `
		SELECT id, title, description, parent_id, estimated_duration_minutes,
			deadline, priority, status, tags, energy_level, difficulty, money_cost,
			task_type, event_location, event_start, event_end, radar_position_x, radar_position_y,
			created_at, updated_at, completed_at
		FROM tasks ORDER BY priority DESC, deadline ASC`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks: %w", err)
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var task models.Task
		var (
			parentID sql.NullInt64
			deadline, eventStart, eventEnd, completedAt sql.NullTime
			description, eventLocation sql.NullString
		)
		
		err := rows.Scan(
			&task.ID, &task.Title, &description, &parentID,
			&task.EstimatedDurationMins, &deadline, &task.Priority,
			&task.Status, &task.Tags, &task.EnergyLevel, &task.Difficulty,
			&task.MoneyCost, &task.TaskType, &eventLocation, &eventStart,
			&eventEnd, &task.RadarPositionX, &task.RadarPositionY,
			&task.CreatedAt, &task.UpdatedAt, &completedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}

		// Handle nullable fields
		if parentID.Valid {
			task.ParentID = &[]int{int(parentID.Int64)}[0]
		}
		if deadline.Valid {
			task.Deadline = &deadline.Time
		}
		if description.Valid {
			task.Description = description.String
		}
		if eventLocation.Valid {
			task.EventLocation = eventLocation.String
		}
		if eventStart.Valid {
			task.EventStart = &eventStart.Time
		}
		if eventEnd.Valid {
			task.EventEnd = &eventEnd.Time
		}
		if completedAt.Valid {
			task.CompletedAt = &completedAt.Time
		}

		// Load prerequisites for each task
		if err := db.loadTaskPrerequisites(&task); err != nil {
			return nil, fmt.Errorf("failed to load prerequisites: %w", err)
		}

		// Load contacts for each task
		if err := db.loadTaskContacts(&task); err != nil {
			return nil, fmt.Errorf("failed to load contacts: %w", err)
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}

// UpdateTaskStatus updates a task's status
func (db *DB) UpdateTaskStatus(id int, status models.TaskStatus) error {
	var completedAt *time.Time
	if status == models.StatusDone {
		now := time.Now()
		completedAt = &now
	}

	query := `UPDATE tasks SET status = ?, completed_at = ?, updated_at = ? WHERE id = ?`
	_, err := db.conn.Exec(query, status, completedAt, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}

	return nil
}

// GetDailyBudget gets or creates a daily budget for the given date
func (db *DB) GetDailyBudget(date time.Time) (*models.DailyBudget, error) {
	dateStr := date.Format("2006-01-02")
	
	budget := &models.DailyBudget{}
	query := `SELECT id, date, total_budget_coins, spent_coins, created_at, updated_at 
		FROM daily_budgets WHERE date = ?`

	err := db.conn.QueryRow(query, dateStr).Scan(
		&budget.ID, &budget.Date, &budget.TotalBudgetCoins,
		&budget.SpentCoins, &budget.CreatedAt, &budget.UpdatedAt)
	
	if err == sql.ErrNoRows {
		// Create new budget for the day
		return db.CreateDailyBudget(date)
	} else if err != nil {
		return nil, fmt.Errorf("failed to get daily budget: %w", err)
	}

	return budget, nil
}

// CreateDailyBudget creates a new daily budget
func (db *DB) CreateDailyBudget(date time.Time) (*models.DailyBudget, error) {
	dateStr := date.Format("2006-01-02")
	now := time.Now()
	
	query := `INSERT INTO daily_budgets (date, total_budget_coins, spent_coins, created_at, updated_at)
		VALUES (?, 500, 0, ?, ?)`
	
	result, err := db.conn.Exec(query, dateStr, now, now)
	if err != nil {
		return nil, fmt.Errorf("failed to create daily budget: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get budget ID: %w", err)
	}

	return &models.DailyBudget{
		ID:               int(id),
		Date:            date,
		TotalBudgetCoins: 500,
		SpentCoins:      0,
		CreatedAt:       now,
		UpdatedAt:       now,
	}, nil
}

// loadTaskPrerequisites loads prerequisites for a task
func (db *DB) loadTaskPrerequisites(task *models.Task) error {
	query := `
		SELECT t.id, t.title, t.description, t.parent_id, t.estimated_duration_minutes,
			t.deadline, t.priority, t.status, t.tags, t.energy_level, t.difficulty, 
			t.money_cost, t.task_type, t.event_location, t.event_start, t.event_end,
			t.radar_position_x, t.radar_position_y, t.created_at, t.updated_at, t.completed_at
		FROM tasks t
		JOIN task_prerequisites tp ON t.id = tp.prerequisite_task_id
		WHERE tp.task_id = ?`

	rows, err := db.conn.Query(query, task.ID)
	if err != nil {
		return fmt.Errorf("failed to query prerequisites: %w", err)
	}
	defer rows.Close()

	var prerequisites []models.Task
	for rows.Next() {
		var prereq models.Task
		var (
			parentID sql.NullInt64
			deadline, eventStart, eventEnd, completedAt sql.NullTime
			description, eventLocation sql.NullString
		)
		
		err := rows.Scan(
			&prereq.ID, &prereq.Title, &description, &parentID,
			&prereq.EstimatedDurationMins, &deadline, &prereq.Priority,
			&prereq.Status, &prereq.Tags, &prereq.EnergyLevel, &prereq.Difficulty,
			&prereq.MoneyCost, &prereq.TaskType, &eventLocation, &eventStart,
			&eventEnd, &prereq.RadarPositionX, &prereq.RadarPositionY,
			&prereq.CreatedAt, &prereq.UpdatedAt, &completedAt)
		if err != nil {
			return fmt.Errorf("failed to scan prerequisite: %w", err)
		}

		// Handle nullable fields
		if parentID.Valid {
			prereq.ParentID = &[]int{int(parentID.Int64)}[0]
		}
		if deadline.Valid {
			prereq.Deadline = &deadline.Time
		}
		if description.Valid {
			prereq.Description = description.String
		}
		if eventLocation.Valid {
			prereq.EventLocation = eventLocation.String
		}
		if eventStart.Valid {
			prereq.EventStart = &eventStart.Time
		}
		if eventEnd.Valid {
			prereq.EventEnd = &eventEnd.Time
		}
		if completedAt.Valid {
			prereq.CompletedAt = &completedAt.Time
		}
		
		prerequisites = append(prerequisites, prereq)
	}

	task.Prerequisites = prerequisites
	return nil
}

// loadTaskSubtasks loads subtasks for a task
func (db *DB) loadTaskSubtasks(task *models.Task) error {
	query := `
		SELECT id, title, description, parent_id, estimated_duration_minutes,
			deadline, priority, status, tags, energy_level, difficulty, money_cost,
			task_type, event_location, event_start, event_end, radar_position_x, radar_position_y,
			created_at, updated_at, completed_at
		FROM tasks WHERE parent_id = ?`

	rows, err := db.conn.Query(query, task.ID)
	if err != nil {
		return fmt.Errorf("failed to query subtasks: %w", err)
	}
	defer rows.Close()

	var subtasks []models.Task
	for rows.Next() {
		var subtask models.Task
		var (
			parentID sql.NullInt64
			deadline, eventStart, eventEnd, completedAt sql.NullTime
			description, eventLocation sql.NullString
		)
		
		err := rows.Scan(
			&subtask.ID, &subtask.Title, &description, &parentID,
			&subtask.EstimatedDurationMins, &deadline, &subtask.Priority,
			&subtask.Status, &subtask.Tags, &subtask.EnergyLevel, &subtask.Difficulty,
			&subtask.MoneyCost, &subtask.TaskType, &eventLocation, &eventStart,
			&eventEnd, &subtask.RadarPositionX, &subtask.RadarPositionY,
			&subtask.CreatedAt, &subtask.UpdatedAt, &completedAt)
		if err != nil {
			return fmt.Errorf("failed to scan subtask: %w", err)
		}

		// Handle nullable fields
		if parentID.Valid {
			subtask.ParentID = &[]int{int(parentID.Int64)}[0]
		}
		if deadline.Valid {
			subtask.Deadline = &deadline.Time
		}
		if description.Valid {
			subtask.Description = description.String
		}
		if eventLocation.Valid {
			subtask.EventLocation = eventLocation.String
		}
		if eventStart.Valid {
			subtask.EventStart = &eventStart.Time
		}
		if eventEnd.Valid {
			subtask.EventEnd = &eventEnd.Time
		}
		if completedAt.Valid {
			subtask.CompletedAt = &completedAt.Time
		}
		
		subtasks = append(subtasks, subtask)
	}

	task.Subtasks = subtasks
	return nil
}

// loadTaskContacts loads contacts associated with a task
func (db *DB) loadTaskContacts(task *models.Task) error {
	query := `
		SELECT c.id, c.name, c.email, c.phone, c.type, c.notes, c.avatar_url, c.created_at, c.updated_at
		FROM contacts c
		JOIN task_contacts tc ON c.id = tc.contact_id
		WHERE tc.task_id = ?`

	rows, err := db.conn.Query(query, task.ID)
	if err != nil {
		return fmt.Errorf("failed to query task contacts: %w", err)
	}
	defer rows.Close()

	var contacts []models.Contact
	for rows.Next() {
		var contact models.Contact
		var (
			email, phone, notes, avatarURL sql.NullString
		)
		
		err := rows.Scan(
			&contact.ID, &contact.Name, &email, &phone,
			&contact.Type, &notes, &avatarURL,
			&contact.CreatedAt, &contact.UpdatedAt)
		if err != nil {
			return fmt.Errorf("failed to scan contact: %w", err)
		}

		// Handle nullable fields
		if email.Valid {
			contact.Email = email.String
		}
		if phone.Valid {
			contact.Phone = phone.String
		}
		if notes.Valid {
			contact.Notes = notes.String
		}
		if avatarURL.Valid {
			contact.AvatarURL = avatarURL.String
		}
		
		contacts = append(contacts, contact)
	}

	task.Contacts = contacts
	return nil
}

// loadTaskAttachments loads attachments for a task
func (db *DB) loadTaskAttachments(task *models.Task) error {
	query := `
		SELECT id, task_id, contact_id, filename, original_filename, file_path,
			file_size, mime_type, description, attachment_type, created_at
		FROM attachments WHERE task_id = ?`

	rows, err := db.conn.Query(query, task.ID)
	if err != nil {
		return fmt.Errorf("failed to query attachments: %w", err)
	}
	defer rows.Close()

	var attachments []models.Attachment
	for rows.Next() {
		var attachment models.Attachment
		err := rows.Scan(
			&attachment.ID, &attachment.TaskID, &attachment.ContactID,
			&attachment.Filename, &attachment.OriginalFilename, &attachment.FilePath,
			&attachment.FileSize, &attachment.MimeType, &attachment.Description,
			&attachment.AttachmentType, &attachment.CreatedAt)
		if err != nil {
			return fmt.Errorf("failed to scan attachment: %w", err)
		}
		attachments = append(attachments, attachment)
	}

	task.Attachments = attachments
	return nil
}

// GetAllContacts retrieves all contacts
func (db *DB) GetAllContacts() ([]models.Contact, error) {
	query := `SELECT id, name, email, phone, type, notes, avatar_url, created_at, updated_at FROM contacts ORDER BY name`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get contacts: %w", err)
	}
	defer rows.Close()

	var contacts []models.Contact
	for rows.Next() {
		var contact models.Contact
		var (
			email, phone, notes, avatarURL sql.NullString
		)
		
		err := rows.Scan(
			&contact.ID, &contact.Name, &email, &phone,
			&contact.Type, &notes, &avatarURL,
			&contact.CreatedAt, &contact.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan contact: %w", err)
		}

		// Handle nullable fields
		if email.Valid {
			contact.Email = email.String
		}
		if phone.Valid {
			contact.Phone = phone.String
		}
		if notes.Valid {
			contact.Notes = notes.String
		}
		if avatarURL.Valid {
			contact.AvatarURL = avatarURL.String
		}
		
		contacts = append(contacts, contact)
	}

	return contacts, nil
}

// GetContactThreads retrieves communication threads for a contact
func (db *DB) GetContactThreads(contactID int) ([]models.ContactThread, error) {
	query := `
		SELECT id, contact_id, task_id, subject, message, thread_type, direction, status, created_at
		FROM contact_threads WHERE contact_id = ? ORDER BY created_at DESC`

	rows, err := db.conn.Query(query, contactID)
	if err != nil {
		return nil, fmt.Errorf("failed to get contact threads: %w", err)
	}
	defer rows.Close()

	var threads []models.ContactThread
	for rows.Next() {
		var thread models.ContactThread
		err := rows.Scan(
			&thread.ID, &thread.ContactID, &thread.TaskID, &thread.Subject,
			&thread.Message, &thread.ThreadType, &thread.Direction,
			&thread.Status, &thread.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan thread: %w", err)
		}
		threads = append(threads, thread)
	}

	return threads, nil
}

// CreateContact creates a new contact
func (db *DB) CreateContact(name, email, phone, contactType, notes string) (*models.Contact, error) {
	query := `
		INSERT INTO contacts (name, email, phone, type, notes, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`

	result, err := db.conn.Exec(query, name, email, phone, contactType, notes)
	if err != nil {
		return nil, fmt.Errorf("failed to create contact: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get contact ID: %w", err)
	}

	// Get the created contact
	return db.GetContact(int(id))
}

// CreateContactThread creates a new communication thread entry
func (db *DB) CreateContactThread(contactID int, taskID *int, subject, message, threadType, direction string) (*models.ContactThread, error) {
	query := `
		INSERT INTO contact_threads (contact_id, task_id, subject, message, thread_type, direction, status, created_at)
		VALUES (?, ?, ?, ?, ?, ?, 'sent', CURRENT_TIMESTAMP)`

	result, err := db.conn.Exec(query, contactID, taskID, subject, message, threadType, direction)
	if err != nil {
		return nil, fmt.Errorf("failed to create contact thread: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get thread ID: %w", err)
	}

	// Get the created thread
	thread := &models.ContactThread{
		ID:         int(id),
		ContactID:  contactID,
		TaskID:     taskID,
		Subject:    subject,
		Message:    message,
		ThreadType: threadType,
		Direction:  direction,
		Status:     "sent",
		CreatedAt:  time.Now(),
	}

	return thread, nil
}

// GetContact retrieves a specific contact by ID
func (db *DB) GetContact(id int) (*models.Contact, error) {
	query := `SELECT id, name, email, phone, type, notes, avatar_url, created_at, updated_at FROM contacts WHERE id = ?`

	var contact models.Contact
	var (
		email, phone, notes, avatarURL sql.NullString
	)
	
	err := db.conn.QueryRow(query, id).Scan(
		&contact.ID, &contact.Name, &email, &phone,
		&contact.Type, &notes, &avatarURL,
		&contact.CreatedAt, &contact.UpdatedAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("contact not found")
		}
		return nil, fmt.Errorf("failed to get contact: %w", err)
	}

	// Handle nullable fields
	if email.Valid {
		contact.Email = email.String
	}
	if phone.Valid {
		contact.Phone = phone.String
	}
	if notes.Valid {
		contact.Notes = notes.String
	}
	if avatarURL.Valid {
		contact.AvatarURL = avatarURL.String
	}

	return &contact, nil
}

// GetContactByEmail finds a contact by email address
func (db *DB) GetContactByEmail(email string) (*models.Contact, error) {
	query := `SELECT id, name, email, phone, type, notes, avatar_url, created_at, updated_at FROM contacts WHERE email = ?`

	var contact models.Contact
	var (
		emailValue, phone, notes, avatarURL sql.NullString
	)
	
	err := db.conn.QueryRow(query, email).Scan(
		&contact.ID, &contact.Name, &emailValue, &phone,
		&contact.Type, &notes, &avatarURL,
		&contact.CreatedAt, &contact.UpdatedAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("contact not found")
		}
		return nil, fmt.Errorf("failed to get contact: %w", err)
	}

	// Handle nullable fields
	if emailValue.Valid {
		contact.Email = emailValue.String
	}
	if phone.Valid {
		contact.Phone = phone.String
	}
	if notes.Valid {
		contact.Notes = notes.String
	}
	if avatarURL.Valid {
		contact.AvatarURL = avatarURL.String
	}

	return &contact, nil
}

// CreateAttachment creates a new attachment record
func (db *DB) CreateAttachment(taskID, contactID *int, filename, originalFilename, filePath, mimeType, description, attachmentType string, fileSize int64) (*models.Attachment, error) {
	query := `
		INSERT INTO attachments (task_id, contact_id, filename, original_filename, file_path, file_size, mime_type, description, attachment_type, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`

	result, err := db.conn.Exec(query, taskID, contactID, filename, originalFilename, filePath, fileSize, mimeType, description, attachmentType)
	if err != nil {
		return nil, fmt.Errorf("failed to create attachment: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get attachment ID: %w", err)
	}

	// Return the created attachment
	attachment := &models.Attachment{
		ID:               int(id),
		TaskID:           taskID,
		ContactID:        contactID,
		Filename:         filename,
		OriginalFilename: originalFilename,
		FilePath:         filePath,
		FileSize:         fileSize,
		MimeType:         mimeType,
		Description:      description,
		AttachmentType:   attachmentType,
		CreatedAt:        time.Now(),
	}

	return attachment, nil
}