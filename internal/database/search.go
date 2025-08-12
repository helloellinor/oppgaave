package database

import (
	"oppgaave/internal/models"
)

// SearchTasks performs a full-text search on tasks
func (db *DB) SearchTasks(query string) ([]models.Task, error) {
	searchQuery := `
		WITH RECURSIVE search_results AS (
			SELECT t.*, rank
			FROM tasks t
			JOIN (
				SELECT rowid, rank
				FROM tasks_fts
				WHERE tasks_fts MATCH ?
				ORDER BY rank
			) fts ON t.id = fts.rowid

			UNION ALL

			SELECT t.*, sr.rank
			FROM tasks t
			JOIN search_results sr ON t.parent_id = sr.id
		)
		SELECT DISTINCT id, title, description, type, parent_id,
		       estimated_duration_minutes, start_time, deadline,
		       priority, status, tags, energy_level, difficulty,
		       money_cost, location, created_at, updated_at, completed_at
		FROM search_results
		ORDER BY rank, start_time ASC NULLS LAST, deadline ASC NULLS LAST`

	rows, err := db.conn.Query(searchQuery, query+"*")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var task models.Task
		err := rows.Scan(
			&task.ID, &task.Title, &task.Description, &task.Type, &task.ParentID,
			&task.EstimatedDurationMins, &task.StartTime, &task.Deadline,
			&task.Priority, &task.Status, &task.Tags, &task.EnergyLevel,
			&task.Difficulty, &task.MoneyCost, &task.Location,
			&task.CreatedAt, &task.UpdatedAt, &task.CompletedAt,
		)
		if err != nil {
			return nil, err
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
