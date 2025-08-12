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

// TaskResponse represents a task response with additional data
type TaskResponse struct {
	Task       *models.Task
	ParentTask *models.Task
	Subtasks   []models.Task
}

// Handlers contains the application's HTTP handlers and their dependencies
type Handlers struct {
	db        *database.DB
	templates *template.Template
	router    *mux.Router
}

// Router returns the configured router
func (h *Handlers) Router() *mux.Router {
	return h.router
}

// New creates a new handlers instance and sets up routes
func New(db *database.DB) *Handlers {
	r := mux.NewRouter()
	h := &Handlers{
		db:     db,
		router: r,
	}

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
		"formatDateTime": func(t *time.Time) string {
			if t == nil {
				return ""
			}
			return t.Format("Mon Jan 2 15:04")
		},
		"iterate": func(count int) []int {
			var i []int
			for j := 0; j < count; j++ {
				i = append(i, j)
			}
			return i
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
		"contains": func(slice interface{}, item interface{}) bool {
			switch s := slice.(type) {
			case []string:
				for _, v := range s {
					if v == item {
						return true
					}
				}
			case []int:
				target, ok := item.(int)
				if !ok {
					return false
				}
				for _, v := range s {
					if v == target {
						return true
					}
				}
			}
			return false
		},
		"formatCurrency": func(amount int) string {
			return fmt.Sprintf("$%d", amount)
		},
		"truncate": func(s string, l int) string {
			if len(s) <= l {
				return s
			}
			return s[:l] + "..."
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
	}

	h.templates = template.Must(template.New("").Funcs(funcMap).ParseGlob("templates/*.html"))

	// Register routes
	r.HandleFunc("/", h.handleIndex).Methods("GET")
	r.HandleFunc("/tasks/new", h.handleNewTask).Methods("GET")
	r.HandleFunc("/tasks", h.handleCreateTask).Methods("POST")
	r.HandleFunc("/tasks/{id}/toggle-status", h.handleToggleStatus).Methods("PUT")
	r.HandleFunc("/tasks/quick-add", h.handleQuickAdd).Methods("GET", "POST")

	// Serve static files
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	r.PathPrefix("/uploads/").Handler(http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))))

	return h
}

// handleIndex renders the main dashboard
func (h *Handlers) handleIndex(w http.ResponseWriter, r *http.Request) {
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

// handleNewTask returns the create task form
func (h *Handlers) handleNewTask(w http.ResponseWriter, r *http.Request) {
	if err := h.templates.ExecuteTemplate(w, "create_task_form.html", nil); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Failed to render form", http.StatusInternalServerError)
	}
}

// handleCreateTask handles task creation from a form submission
func (h *Handlers) handleCreateTask(w http.ResponseWriter, r *http.Request) {
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

// handleToggleStatus handles task status updates via HTMX
func (h *Handlers) handleToggleStatus(w http.ResponseWriter, r *http.Request) {
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

// handleQuickAdd handles quick task creation
func (h *Handlers) handleQuickAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		if err := h.templates.ExecuteTemplate(w, "quick_add.html", nil); err != nil {
			log.Printf("Error executing template: %v", err)
			http.Error(w, "Failed to render form", http.StatusInternalServerError)
		}
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	if title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	req := &models.CreateTaskRequest{
		Title:                 title,
		EstimatedDurationMins: 30, // default
		Priority:              2,  // default medium
		EnergyLevel:           2,  // default medium
		Difficulty:            2,  // default medium
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