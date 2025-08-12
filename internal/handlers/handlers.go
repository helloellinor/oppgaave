package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"oppgaave/internal/database"
	"oppgaave/internal/models"

	"github.com/gorilla/mux"
)

type Handlers struct {
	db        *database.DB
	templates *template.Template
}

// New creates a new handlers instance
func New(db *database.DB) *Handlers {
	// Load templates with custom functions
	funcMap := template.FuncMap{
		"formatDuration": func(minutes int) string {
			if minutes < 60 {
				return fmt.Sprintf("%dm", minutes)
			}
			hours := minutes / 60
			mins := minutes % 60
			if mins == 0 {
				return fmt.Sprintf("%dh", hours)
			}
			return fmt.Sprintf("%dh%dm", hours, mins)
		},
		"formatTime": func(t *time.Time) string {
			if t == nil {
				return ""
			}
			return t.Format("15:04")
		},
		"formatDate": func(t *time.Time) string {
			if t == nil {
				return ""
			}
			return t.Format("Jan 2")
		},
		"formatCurrency": func(coins int) string {
			return fmt.Sprintf("$%d", coins)
		},
		"statusIcon": func(status models.TaskStatus) string {
			switch status {
			case models.StatusDone:
				return "✓"
			case models.StatusInProgress:
				return "⏳"
			case models.StatusBlocked:
				return "🚫"
			default:
				return "○"
			}
		},
		"priorityText": func(priority int) string {
			switch priority {
			case 3:
				return "High"
			case 2:
				return "Medium"
			default:
				return "Low"
			}
		},
		"energyText": func(energy int) string {
			switch energy {
			case 3:
				return "High Energy"
			case 2:
				return "Medium Energy"
			default:
				return "Low Energy"
			}
		},
		"mul": func(a, b int) int {
			return a * b
		},
		"div": func(a, b int) float64 {
			if b == 0 {
				return 0
			}
			return float64(a) / float64(b)
		},
		"percentage": func(spent, total int) float64 {
			if total == 0 {
				return 0
			}
			return (float64(spent) / float64(total)) * 100
		},
	}

	templates := template.Must(template.New("").Funcs(funcMap).ParseGlob("templates/*.html"))

	return &Handlers{
		db:        db,
		templates: templates,
	}
}

// Dashboard renders the main dashboard
func (h *Handlers) Dashboard(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.db.GetAllTasks()
	if err != nil {
		log.Printf("Error getting tasks: %v", err)
		http.Error(w, "Failed to load tasks", http.StatusInternalServerError)
		return
	}

	today := time.Now()
	budget, err := h.db.GetDailyBudget(today)
	if err != nil {
		log.Printf("Error getting daily budget: %v", err)
		http.Error(w, "Failed to load budget", http.StatusInternalServerError)
		return
	}

	// Calculate spent budget from pending/in-progress tasks
	spentCoins := 0
	var todayTasks []models.Task
	for _, task := range tasks {
		if task.Status == models.StatusPending || task.Status == models.StatusInProgress {
			spentCoins += task.MoneyCost
			todayTasks = append(todayTasks, task)
		}
	}

	budget.SpentCoins = spentCoins

	data := struct {
		Tasks       []models.Task
		TodayTasks  []models.Task
		Budget      *models.DailyBudget
		CurrentTime string
	}{
		Tasks:       tasks,
		TodayTasks:  todayTasks,
		Budget:      budget,
		CurrentTime: today.Format("15:04"),
	}

	if err := h.templates.ExecuteTemplate(w, "dashboard.html", data); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Failed to render dashboard", http.StatusInternalServerError)
	}
}

// GetTaskList returns the task list as HTML fragment for HTMX
func (h *Handlers) GetTaskList(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.db.GetAllTasks()
	if err != nil {
		log.Printf("Error getting tasks: %v", err)
		http.Error(w, "Failed to load tasks", http.StatusInternalServerError)
		return
	}

	if err := h.templates.ExecuteTemplate(w, "task_list.html", tasks); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Failed to render task list", http.StatusInternalServerError)
	}
}

// CreateTask handles task creation
func (h *Handlers) CreateTask(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		// Return the create task form
		if err := h.templates.ExecuteTemplate(w, "create_task_form.html", nil); err != nil {
			log.Printf("Error executing template: %v", err)
			http.Error(w, "Failed to render form", http.StatusInternalServerError)
		}
		return
	}

	if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		duration, _ := strconv.Atoi(r.FormValue("duration"))
		if duration == 0 {
			duration = 30 // default
		}

		priority, _ := strconv.Atoi(r.FormValue("priority"))
		if priority == 0 {
			priority = 2 // default medium
		}

		energy, _ := strconv.Atoi(r.FormValue("energy"))
		if energy == 0 {
			energy = 2 // default medium
		}

		difficulty, _ := strconv.Atoi(r.FormValue("difficulty"))
		if difficulty == 0 {
			difficulty = 2 // default medium
		}

		req := &models.CreateTaskRequest{
			Title:                 r.FormValue("title"),
			Description:           r.FormValue("description"),
			EstimatedDurationMins: duration,
			Priority:              priority,
			EnergyLevel:           energy,
			Difficulty:            difficulty,
		}

		task, err := h.db.CreateTask(req)
		if err != nil {
			log.Printf("Error creating task: %v", err)
			http.Error(w, "Failed to create task", http.StatusInternalServerError)
			return
		}

		// Return the new task as HTML fragment
		if err := h.templates.ExecuteTemplate(w, "task_item.html", task); err != nil {
			log.Printf("Error executing template: %v", err)
			http.Error(w, "Failed to render task", http.StatusInternalServerError)
		}
	}
}

// UpdateTaskStatus handles task status updates via HTMX
func (h *Handlers) UpdateTaskStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	status := r.FormValue("status")
	if status == "" {
		http.Error(w, "Status is required", http.StatusBadRequest)
		return
	}

	taskStatus := models.TaskStatus(status)
	if err := h.db.UpdateTaskStatus(taskID, taskStatus); err != nil {
		log.Printf("Error updating task status: %v", err)
		http.Error(w, "Failed to update task", http.StatusInternalServerError)
		return
	}

	// Get updated task and return HTML fragment
	task, err := h.db.GetTask(taskID)
	if err != nil {
		log.Printf("Error getting updated task: %v", err)
		http.Error(w, "Failed to get task", http.StatusInternalServerError)
		return
	}

	if err := h.templates.ExecuteTemplate(w, "task_item.html", task); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Failed to render task", http.StatusInternalServerError)
	}
}

// GetBudgetWidget returns the budget widget as HTML fragment
func (h *Handlers) GetBudgetWidget(w http.ResponseWriter, r *http.Request) {
	today := time.Now()
	budget, err := h.db.GetDailyBudget(today)
	if err != nil {
		log.Printf("Error getting daily budget: %v", err)
		http.Error(w, "Failed to load budget", http.StatusInternalServerError)
		return
	}

	// Calculate current spent amount from pending/in-progress tasks
	tasks, err := h.db.GetAllTasks()
	if err != nil {
		log.Printf("Error getting tasks: %v", err)
		http.Error(w, "Failed to load tasks", http.StatusInternalServerError)
		return
	}

	spentCoins := 0
	for _, task := range tasks {
		if task.Status == models.StatusPending || task.Status == models.StatusInProgress {
			spentCoins += task.MoneyCost
		}
	}
	budget.SpentCoins = spentCoins

	if err := h.templates.ExecuteTemplate(w, "budget_widget.html", budget); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Failed to render budget widget", http.StatusInternalServerError)
	}
}

// API endpoints for JSON responses

// GetTasksAPI returns tasks as JSON
func (h *Handlers) GetTasksAPI(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.db.GetAllTasks()
	if err != nil {
		log.Printf("Error getting tasks: %v", err)
		http.Error(w, "Failed to load tasks", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

// CreateTaskAPI creates a task via JSON API
func (h *Handlers) CreateTaskAPI(w http.ResponseWriter, r *http.Request) {
	var req models.CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	task, err := h.db.CreateTask(&req)
	if err != nil {
		log.Printf("Error creating task: %v", err)
		http.Error(w, "Failed to create task", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}