package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"oppgaave/internal/models"
	"github.com/gorilla/mux"
)

// handleCurrentTime returns the current time for the time display
func (h *Handlers) handleCurrentTime(w http.ResponseWriter, r *http.Request) {
	currentTime := time.Now().Format("15:04")
	w.Write([]byte(currentTime))
}

// handleTimeline renders the timeline view with tasks
func (h *Handlers) handleTimeline(w http.ResponseWriter, r *http.Request) {
	// Get tasks for today and upcoming
	tasks, err := h.db.GetTasksByTimeRange(time.Now(), time.Now().AddDate(0, 0, 7))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Tasks []models.Task
	}{
		Tasks: tasks,
	}

	h.templates.ExecuteTemplate(w, "timeline_view.html", data)
}

// handleTaskDetail renders the task detail view
func (h *Handlers) handleTaskDetail(w http.ResponseWriter, r *http.Request) {
	taskID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	task, err := h.db.GetTaskByID(taskID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get parent task if exists
	var parentTask *models.Task
	if task.ParentID != nil {
		parent, err := h.db.GetTaskByID(*task.ParentID)
		if err == nil {
			parentTask = &parent
		}
	}

	// Get subtasks
	subtasks, err := h.db.GetSubtasks(taskID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := TaskResponse{
		Task:       &task,
		ParentTask: parentTask,
		Subtasks:   subtasks,
	}

	h.templates.ExecuteTemplate(w, "task_detail.html", data)
}

// handleUploadAttachment handles file uploads for tasks
func (h *Handlers) handleUploadAttachment(w http.ResponseWriter, r *http.Request) {
	taskID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	// Parse multipart form
	err = r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("attachment")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Create uploads directory if it doesn't exist
	uploadDir := "./uploads"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create a unique filename
	filename := filepath.Join(uploadDir, time.Now().Format("20060102150405")+"-"+handler.Filename)
	
	// Create the file
	dst, err := os.Create(filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// Copy the uploaded file
	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create attachment record
	attachment := models.Attachment{
		TaskID:    taskID,
		Name:      handler.Filename,
		Type:      handler.Header.Get("Content-Type"),
		Path:      filename,
		CreatedAt: time.Now(),
	}

	err = h.db.CreateAttachment(&attachment)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the updated attachments list
	task, err := h.db.GetTaskByID(taskID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.templates.ExecuteTemplate(w, "attachments", struct{ Task *models.Task }{Task: &task})
}

// handleUpdateTaskField updates a single field of a task
func (h *Handlers) handleUpdateTaskField(w http.ResponseWriter, r *http.Request) {
	taskID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	field := mux.Vars(r)["field"]
	if field == "" {
		http.Error(w, "Field name required", http.StatusBadRequest)
		return
	}

	var update struct {
		Value interface{} `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	task, err := h.db.GetTaskByID(taskID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Update the specified field
	switch field {
	case "energy_level":
		if val, ok := update.Value.(float64); ok {
			task.EnergyLevel = int(val)
		}
	case "difficulty":
		if val, ok := update.Value.(float64); ok {
			task.Difficulty = int(val)
		}
	case "description":
		if val, ok := update.Value.(string); ok {
			task.Description = val
		}
	case "money_cost":
		if val, ok := update.Value.(float64); ok {
			task.MoneyCost = int(val)
		}
	default:
		http.Error(w, "Invalid field", http.StatusBadRequest)
		return
	}

	task.UpdatedAt = time.Now()
	if err := h.db.UpdateTask(&task); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the updated field HTML
	h.templates.ExecuteTemplate(w, "task_field_"+field, task)
}
