package database

import (
	"database/sql"
	"encoding/json"
	"time"

	"oppgaave/internal/models"
)

// GetTasksByTimeRange returns tasks within the specified time range
func (db *DB) GetTasksByTimeRange(start, end time.Time) ([]models.Task, error) {
	query := `
		SELECT id, title, description, type, parent_id, estimated_duration_minutes,
		       start_time, deadline, priority, status, tags, energy_level,
		       difficulty, money_cost, location, created_at, updated_at, completed_at
		FROM tasks
		WHERE (start_time BETWEEN ? AND ?) OR (deadline BETWEEN ? AND ?)
		ORDER BY COALESCE(start_time, deadline) ASC`

	rows, err := db.conn.Query(query, start, end, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var task models.Task
		var tagsJSON sql.NullString
		err := rows.Scan(
			&task.ID, &task.Title, &task.Description, &task.Type, &task.ParentID,
			&task.EstimatedDurationMins, &task.StartTime, &task.Deadline,
			&task.Priority, &task.Status, &tagsJSON, &task.EnergyLevel,
			&task.Difficulty, &task.MoneyCost, &task.Location,
			&task.CreatedAt, &task.UpdatedAt, &task.CompletedAt,
		)
		if err != nil {
			return nil, err
		}

		if tagsJSON.Valid {
			if err := json.Unmarshal([]byte(tagsJSON.String), &task.Tags); err != nil {
				return nil, err
			}
		}

		// Load attachments
		attachments, err := db.GetAttachments(task.ID)
		if err != nil {
			return nil, err
		}
		task.Attachments = attachments

		tasks = append(tasks, task)
	}

	return tasks, nil
}

// GetTaskByID returns a single task by ID
func (db *DB) GetTaskByID(id int) (models.Task, error) {
	query := `
		SELECT id, title, description, type, parent_id, estimated_duration_minutes,
		       start_time, deadline, priority, status, tags, energy_level,
		       difficulty, money_cost, location, created_at, updated_at, completed_at
		FROM tasks
		WHERE id = ?`

	var task models.Task
	var tagsJSON sql.NullString
	err := db.conn.QueryRow(query, id).Scan(
		&task.ID, &task.Title, &task.Description, &task.Type, &task.ParentID,
		&task.EstimatedDurationMins, &task.StartTime, &task.Deadline,
		&task.Priority, &task.Status, &tagsJSON, &task.EnergyLevel,
		&task.Difficulty, &task.MoneyCost, &task.Location,
		&task.CreatedAt, &task.UpdatedAt, &task.CompletedAt,
	)
	if err != nil {
		return task, err
	}

	if tagsJSON.Valid {
		if err := json.Unmarshal([]byte(tagsJSON.String), &task.Tags); err != nil {
			return task, err
		}
	}

	// Load attachments
	attachments, err := db.GetAttachments(task.ID)
	if err != nil {
		return task, err
	}
	task.Attachments = attachments

	return task, nil
}

// GetSubtasks returns all subtasks for a given task
func (db *DB) GetSubtasks(parentID int) ([]models.Task, error) {
	query := `
		SELECT id, title, description, type, parent_id, estimated_duration_minutes,
		       start_time, deadline, priority, status, tags, energy_level,
		       difficulty, money_cost, location, created_at, updated_at, completed_at
		FROM tasks
		WHERE parent_id = ?
		ORDER BY created_at ASC`

	rows, err := db.conn.Query(query, parentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var task models.Task
		var tagsJSON sql.NullString
		err := rows.Scan(
			&task.ID, &task.Title, &task.Description, &task.Type, &task.ParentID,
			&task.EstimatedDurationMins, &task.StartTime, &task.Deadline,
			&task.Priority, &task.Status, &tagsJSON, &task.EnergyLevel,
			&task.Difficulty, &task.MoneyCost, &task.Location,
			&task.CreatedAt, &task.UpdatedAt, &task.CompletedAt,
		)
		if err != nil {
			return nil, err
		}

		if tagsJSON.Valid {
			if err := json.Unmarshal([]byte(tagsJSON.String), &task.Tags); err != nil {
				return nil, err
			}
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}

// UpdateTask updates an existing task
func (db *DB) UpdateTask(task *models.Task) error {
	tagsJSON, err := json.Marshal(task.Tags)
	if err != nil {
		return err
	}

	query := `
		UPDATE tasks
		SET title = ?, description = ?, type = ?, parent_id = ?,
		    estimated_duration_minutes = ?, start_time = ?, deadline = ?,
		    priority = ?, status = ?, tags = ?, energy_level = ?,
		    difficulty = ?, money_cost = ?, location = ?,
		    updated_at = ?, completed_at = ?
		WHERE id = ?`

	_, err = db.conn.Exec(query,
		task.Title, task.Description, task.Type, task.ParentID,
		task.EstimatedDurationMins, task.StartTime, task.Deadline,
		task.Priority, task.Status, tagsJSON, task.EnergyLevel,
		task.Difficulty, task.MoneyCost, task.Location,
		task.UpdatedAt, task.CompletedAt, task.ID,
	)

	return err
}

// InsertTask inserts a task directly into the database
func (db *DB) InsertTask(task *models.Task) error {
	tagsJSON, err := json.Marshal(task.Tags)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO tasks (
			title, description, type, parent_id, estimated_duration_minutes,
			start_time, deadline, priority, status, tags, energy_level,
			difficulty, money_cost, location, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := db.conn.Exec(query,
		task.Title, task.Description, task.Type, task.ParentID,
		task.EstimatedDurationMins, task.StartTime, task.Deadline,
		task.Priority, task.Status, tagsJSON, task.EnergyLevel,
		task.Difficulty, task.MoneyCost, task.Location,
		task.CreatedAt, task.UpdatedAt,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	task.ID = int(id)
	return nil
}

// GetAttachments returns all attachments for a task
func (db *DB) GetAttachments(taskID int) ([]models.Attachment, error) {
	query := `
		SELECT id, task_id, name, type, path, created_at
		FROM attachments
		WHERE task_id = ?
		ORDER BY created_at ASC`

	rows, err := db.conn.Query(query, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var attachments []models.Attachment
	for rows.Next() {
		var attachment models.Attachment
		err := rows.Scan(
			&attachment.ID, &attachment.TaskID, &attachment.Name,
			&attachment.Type, &attachment.Path, &attachment.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		attachments = append(attachments, attachment)
	}

	return attachments, nil
}

// CreateAttachment creates a new attachment
func (db *DB) CreateAttachment(attachment *models.Attachment) error {
	query := `
		INSERT INTO attachments (task_id, name, type, path, created_at)
		VALUES (?, ?, ?, ?, ?)`

	result, err := db.conn.Exec(query,
		attachment.TaskID, attachment.Name, attachment.Type,
		attachment.Path, attachment.CreatedAt,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	attachment.ID = int(id)
	return nil
}
