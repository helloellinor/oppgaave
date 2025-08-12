package database

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
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

// initSchema reads and executes the schema.sql file
func (db *DB) initSchema() error {
	schemaPath := filepath.Join(".", "schema.sql")
	schema, err := ioutil.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("failed to read schema file: %w", err)
	}

	if _, err := db.conn.Exec(string(schema)); err != nil {
		return fmt.Errorf("failed to execute schema: %w", err)
	}

	log.Println("Database schema initialized successfully")
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