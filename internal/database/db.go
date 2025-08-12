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
		Status:                models.StatusPending,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}
	
	// Calculate money cost
	task.MoneyCost = task.CalculateMoneyCost()

	query := `
		INSERT INTO tasks (title, description, parent_id, estimated_duration_minutes, 
			deadline, priority, status, tags, energy_level, difficulty, money_cost, 
			created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := db.conn.Exec(query, task.Title, task.Description, task.ParentID,
		task.EstimatedDurationMins, task.Deadline, task.Priority, task.Status,
		task.Tags, task.EnergyLevel, task.Difficulty, task.MoneyCost,
		task.CreatedAt, task.UpdatedAt)
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
	query := `
		SELECT id, title, description, parent_id, estimated_duration_minutes,
			deadline, priority, status, tags, energy_level, difficulty, money_cost,
			created_at, updated_at, completed_at
		FROM tasks WHERE id = ?`

	err := db.conn.QueryRow(query, id).Scan(
		&task.ID, &task.Title, &task.Description, &task.ParentID,
		&task.EstimatedDurationMins, &task.Deadline, &task.Priority,
		&task.Status, &task.Tags, &task.EnergyLevel, &task.Difficulty,
		&task.MoneyCost, &task.CreatedAt, &task.UpdatedAt, &task.CompletedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	// Load prerequisites
	if err := db.loadTaskPrerequisites(task); err != nil {
		return nil, fmt.Errorf("failed to load prerequisites: %w", err)
	}

	// Load subtasks
	if err := db.loadTaskSubtasks(task); err != nil {
		return nil, fmt.Errorf("failed to load subtasks: %w", err)
	}

	return task, nil
}

// GetAllTasks retrieves all tasks
func (db *DB) GetAllTasks() ([]models.Task, error) {
	query := `
		SELECT id, title, description, parent_id, estimated_duration_minutes,
			deadline, priority, status, tags, energy_level, difficulty, money_cost,
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
		err := rows.Scan(
			&task.ID, &task.Title, &task.Description, &task.ParentID,
			&task.EstimatedDurationMins, &task.Deadline, &task.Priority,
			&task.Status, &task.Tags, &task.EnergyLevel, &task.Difficulty,
			&task.MoneyCost, &task.CreatedAt, &task.UpdatedAt, &task.CompletedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}

		// Load prerequisites for each task
		if err := db.loadTaskPrerequisites(&task); err != nil {
			return nil, fmt.Errorf("failed to load prerequisites: %w", err)
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
			t.money_cost, t.created_at, t.updated_at, t.completed_at
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
		err := rows.Scan(
			&prereq.ID, &prereq.Title, &prereq.Description, &prereq.ParentID,
			&prereq.EstimatedDurationMins, &prereq.Deadline, &prereq.Priority,
			&prereq.Status, &prereq.Tags, &prereq.EnergyLevel, &prereq.Difficulty,
			&prereq.MoneyCost, &prereq.CreatedAt, &prereq.UpdatedAt, &prereq.CompletedAt)
		if err != nil {
			return fmt.Errorf("failed to scan prerequisite: %w", err)
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
		err := rows.Scan(
			&subtask.ID, &subtask.Title, &subtask.Description, &subtask.ParentID,
			&subtask.EstimatedDurationMins, &subtask.Deadline, &subtask.Priority,
			&subtask.Status, &subtask.Tags, &subtask.EnergyLevel, &subtask.Difficulty,
			&subtask.MoneyCost, &subtask.CreatedAt, &subtask.UpdatedAt, &subtask.CompletedAt)
		if err != nil {
			return fmt.Errorf("failed to scan subtask: %w", err)
		}
		subtasks = append(subtasks, subtask)
	}

	task.Subtasks = subtasks
	return nil
}