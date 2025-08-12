package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
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
		"taskTypeText": func(taskType models.TaskType) string {
			switch taskType {
			case models.TypeAppointment:
				return "Appointment"
			case models.TypeEvent:
				return "Event"
			case models.TypeConcert:
				return "Concert"
			case models.TypeMeeting:
				return "Meeting"
			default:
				return "Task"
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
	for i, task := range tasks {
		// Calculate radar position for each task
		tasks[i].CalculateRadarPosition()
		
		if task.Status == models.StatusPending || task.Status == models.StatusInProgress {
			spentCoins += task.MoneyCost
			todayTasks = append(todayTasks, tasks[i])
		}
	}

	budget.SpentCoins = spentCoins

	// Get contacts for the dashboard
	contacts, err := h.db.GetAllContacts()
	if err != nil {
		log.Printf("Error getting contacts: %v", err)
		// Don't fail the whole dashboard, just use empty contacts
		contacts = []models.Contact{}
	}

	data := struct {
		Tasks       []models.Task
		TodayTasks  []models.Task
		Budget      *models.DailyBudget
		CurrentTime string
		Contacts    []models.Contact
	}{
		Tasks:       tasks,
		TodayTasks:  todayTasks,
		Budget:      budget,
		CurrentTime: today.Format("15:04"),
		Contacts:    contacts,
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

// GetTaskRadar returns the radar visualization for tasks
func (h *Handlers) GetTaskRadar(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.db.GetAllTasks()
	if err != nil {
		log.Printf("Error getting tasks for radar: %v", err)
		http.Error(w, "Failed to load tasks", http.StatusInternalServerError)
		return
	}

	// Calculate radar positions for all tasks
	for i := range tasks {
		tasks[i].CalculateRadarPosition()
	}

	if err := h.templates.ExecuteTemplate(w, "task_radar.html", tasks); err != nil {
		log.Printf("Error executing radar template: %v", err)
		http.Error(w, "Failed to render radar", http.StatusInternalServerError)
	}
}

// GetTaskDetails returns detailed task information
func (h *Handlers) GetTaskDetails(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	task, err := h.db.GetTask(taskID)
	if err != nil {
		log.Printf("Error getting task details: %v", err)
		http.Error(w, "Failed to get task", http.StatusInternalServerError)
		return
	}

	if err := h.templates.ExecuteTemplate(w, "task_details.html", task); err != nil {
		log.Printf("Error executing task details template: %v", err)
		http.Error(w, "Failed to render task details", http.StatusInternalServerError)
	}
}

// GetContacts returns all contacts
func (h *Handlers) GetContacts(w http.ResponseWriter, r *http.Request) {
	contacts, err := h.db.GetAllContacts()
	if err != nil {
		log.Printf("Error getting contacts: %v", err)
		http.Error(w, "Failed to load contacts", http.StatusInternalServerError)
		return
	}

	if err := h.templates.ExecuteTemplate(w, "contact_threads.html", contacts); err != nil {
		log.Printf("Error executing contacts template: %v", err)
		http.Error(w, "Failed to render contacts", http.StatusInternalServerError)
	}
}

// CreateContact handles contact creation
func (h *Handlers) CreateContact(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		// Return contact creation form
		w.Write([]byte(`<div class="modal-content">
			<div class="modal-header">
				<h2>Add Contact</h2>
				<button class="modal-close" onclick="document.getElementById('contact-modal').innerHTML = ''">×</button>
			</div>
			<form hx-post="/contacts/create" hx-target="#contact-modal" hx-swap="innerHTML">
				<div class="form-group">
					<label>Name:</label>
					<input type="text" name="name" required>
				</div>
				<div class="form-group">
					<label>Email:</label>
					<input type="email" name="email">
				</div>
				<div class="form-group">
					<label>Phone:</label>
					<input type="tel" name="phone">
				</div>
				<div class="form-group">
					<label>Type:</label>
					<select name="type">
						<option value="person">Person</option>
						<option value="organization">Organization</option>
						<option value="venue">Venue</option>
					</select>
				</div>
				<div class="form-group">
					<label>Notes:</label>
					<textarea name="notes"></textarea>
				</div>
				<div class="form-actions">
					<button type="submit" class="btn btn-primary">Create Contact</button>
					<button type="button" class="btn btn-secondary" onclick="document.getElementById('contact-modal').innerHTML = ''">Cancel</button>
				</div>
			</form>
		</div>`))
		return
	}

	// Handle POST - create the contact
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	email := r.FormValue("email")
	phone := r.FormValue("phone")
	contactType := r.FormValue("type")
	notes := r.FormValue("notes")

	if name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	contact, err := h.db.CreateContact(name, email, phone, contactType, notes)
	if err != nil {
		log.Printf("Error creating contact: %v", err)
		http.Error(w, "Failed to create contact", http.StatusInternalServerError)
		return
	}

	// Return success message and refresh the contact list
	w.Write([]byte(fmt.Sprintf(`<div class="success-message">
		<p>✅ Contact "%s" created successfully!</p>
		<button type="button" class="btn btn-secondary" 
		        onclick="document.getElementById('contact-modal').innerHTML = ''; window.location.reload();">
			Close
		</button>
	</div>`, contact.Name)))
}

// GetContactThreads returns communication threads for a contact
func (h *Handlers) GetContactThreads(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	contactID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid contact ID", http.StatusBadRequest)
		return
	}

	// Get contact details
	contact, err := h.db.GetContact(contactID)
	if err != nil {
		log.Printf("Error getting contact: %v", err)
		http.Error(w, "Contact not found", http.StatusNotFound)
		return
	}

	threads, err := h.db.GetContactThreads(contactID)
	if err != nil {
		log.Printf("Error getting contact threads: %v", err)
		http.Error(w, "Failed to load threads", http.StatusInternalServerError)
		return
	}

	// Build HTML for thread display
	html := fmt.Sprintf(`<div class="thread-section">
		<div class="thread-header">
			<h3>💬 Communication with %s</h3>
			<button class="btn btn-primary" 
			        hx-get="/contacts/%d/message" 
			        hx-target="#message-modal"
			        hx-trigger="click">
				➕ Add Message
			</button>
		</div>
		<div class="thread-list">`, contact.Name, contactID)

	if len(threads) == 0 {
		html += `<div class="empty-state">
			<p>No communication threads yet. Start a conversation!</p>
		</div>`
	} else {
		for _, thread := range threads {
			directionIcon := "➡️"
			if thread.Direction == "inbound" {
				directionIcon = "⬅️"
			}
			
			typeIcon := "💬"
			switch thread.ThreadType {
			case "email":
				typeIcon = "📧"
			case "call":
				typeIcon = "📞"
			case "meeting":
				typeIcon = "💼"
			}

			html += fmt.Sprintf(`<div class="thread-item">
				<div class="thread-meta">
					<span class="thread-type">%s %s</span>
					<span class="thread-direction">%s</span>
					<span class="thread-date">%s</span>
				</div>
				<div class="thread-subject">%s</div>
				<div class="thread-message">%s</div>
			</div>`, typeIcon, thread.ThreadType, directionIcon, 
				thread.CreatedAt.Format("Jan 2, 15:04"), thread.Subject, thread.Message)
		}
	}

	html += `</div></div>`
	w.Write([]byte(html))
}

// CreateMessage handles creating new messages
func (h *Handlers) CreateMessage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	contactIDStr := vars["id"]
	contactID, err := strconv.Atoi(contactIDStr)
	if err != nil {
		http.Error(w, "Invalid contact ID", http.StatusBadRequest)
		return
	}

	if r.Method == "GET" {
		w.Write([]byte(fmt.Sprintf(`<div class="modal-content">
			<div class="modal-header">
				<h2>Send Message</h2>
				<button class="modal-close" onclick="document.getElementById('message-modal').innerHTML = ''">×</button>
			</div>
			<form hx-post="/contacts/%s/message" hx-target="#message-modal" hx-swap="innerHTML">
				<div class="form-group">
					<label>Subject:</label>
					<input type="text" name="subject">
				</div>
				<div class="form-group">
					<label>Message:</label>
					<textarea name="message" rows="4" required></textarea>
				</div>
				<div class="form-group">
					<label>Type:</label>
					<select name="type">
						<option value="message">Message</option>
						<option value="email">Email</option>
						<option value="call">Call Log</option>
						<option value="meeting">Meeting Notes</option>
					</select>
				</div>
				<div class="form-group">
					<label>Direction:</label>
					<select name="direction">
						<option value="outbound">Outbound</option>
						<option value="inbound">Inbound</option>
					</select>
				</div>
				<div class="form-actions">
					<button type="submit" class="btn btn-primary">Save</button>
					<button type="button" class="btn btn-secondary" onclick="document.getElementById('message-modal').innerHTML = ''">Cancel</button>
				</div>
			</form>
		</div>`, contactIDStr)))
		return
	}

	// Handle POST - create the message
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	subject := r.FormValue("subject")
	message := r.FormValue("message")
	threadType := r.FormValue("type")
	direction := r.FormValue("direction")

	if message == "" {
		http.Error(w, "Message is required", http.StatusBadRequest)
		return
	}

	// Create the thread entry
	_, err = h.db.CreateContactThread(contactID, nil, subject, message, threadType, direction)
	if err != nil {
		log.Printf("Error creating contact thread: %v", err)
		http.Error(w, "Failed to save message", http.StatusInternalServerError)
		return
	}

	// Return success message and close modal
	w.Write([]byte(`<div class="success-message">
		<p>✅ Message saved successfully!</p>
		<button type="button" class="btn btn-secondary" 
		        onclick="document.getElementById('message-modal').innerHTML = ''; document.getElementById('thread-viewer').innerHTML = '';">
			Close
		</button>
	</div>`))
}

// ForwardEmail handles email forwarding and parsing
func (h *Handlers) ForwardEmail(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	emailFrom := r.FormValue("from")
	_ = r.FormValue("to") // Currently unused but available for future features
	subject := r.FormValue("subject")
	body := r.FormValue("body")

	if emailFrom == "" || body == "" {
		http.Error(w, "From address and body are required", http.StatusBadRequest)
		return
	}

	// Try to find existing contact by email
	contact, err := h.db.GetContactByEmail(emailFrom)
	if err != nil {
		// Contact doesn't exist, create a new one
		// Extract name from email (before @)
		name := emailFrom
		if atIndex := strings.Index(emailFrom, "@"); atIndex > 0 {
			name = emailFrom[:atIndex]
		}
		
		contact, err = h.db.CreateContact(name, emailFrom, "", "person", "Created from forwarded email")
		if err != nil {
			log.Printf("Error creating contact from email: %v", err)
			http.Error(w, "Failed to create contact", http.StatusInternalServerError)
			return
		}
	}

	// Create thread entry for the forwarded email
	_, err = h.db.CreateContactThread(contact.ID, nil, subject, body, "email", "inbound")
	if err != nil {
		log.Printf("Error creating thread for forwarded email: %v", err)
		http.Error(w, "Failed to save email thread", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(fmt.Sprintf(`{"success": true, "message": "Email forwarded and saved", "contact_id": %d}`, contact.ID)))
}

// ParseEmailForm provides a form for manual email entry/forwarding
func (h *Handlers) ParseEmailForm(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		w.Write([]byte(`<!DOCTYPE html>
<html>
<head>
    <title>Forward Email to Contact System</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 800px; margin: 0 auto; padding: 20px; }
        .form-group { margin-bottom: 15px; }
        label { display: block; margin-bottom: 5px; font-weight: bold; }
        input, textarea { width: 100%; padding: 8px; border: 1px solid #ddd; border-radius: 4px; }
        textarea { height: 200px; }
        button { background: #007bff; color: white; padding: 10px 20px; border: none; border-radius: 4px; cursor: pointer; }
        .success { color: green; padding: 10px; background: #f0f8f0; border-radius: 4px; margin: 10px 0; }
        .error { color: red; padding: 10px; background: #f8f0f0; border-radius: 4px; margin: 10px 0; }
    </style>
</head>
<body>
    <h1>📧 Forward Email to Contact System</h1>
    <p>Use this form to forward emails and automatically create contacts and communication threads.</p>
    
    <form method="POST" action="/contacts/email/parse">
        <div class="form-group">
            <label for="from">From Email Address:</label>
            <input type="email" name="from" id="from" required 
                   placeholder="sender@example.com" />
        </div>
        
        <div class="form-group">
            <label for="to">To Email Address (optional):</label>
            <input type="email" name="to" id="to" 
                   placeholder="your@email.com" />
        </div>
        
        <div class="form-group">
            <label for="subject">Subject:</label>
            <input type="text" name="subject" id="subject" 
                   placeholder="Email subject line" />
        </div>
        
        <div class="form-group">
            <label for="body">Email Body:</label>
            <textarea name="body" id="body" required 
                      placeholder="Paste the email content here..."></textarea>
        </div>
        
        <button type="submit">📥 Forward Email</button>
        <a href="/" style="margin-left: 10px;">← Back to Dashboard</a>
    </form>
    
    <div style="margin-top: 30px; padding: 15px; background: #f8f9fa; border-radius: 4px;">
        <h3>💡 How it works:</h3>
        <ul>
            <li><strong>Auto-Contact Creation:</strong> If the sender email doesn't exist, a new contact will be created automatically</li>
            <li><strong>Thread Logging:</strong> The email will be saved as a communication thread entry</li>
            <li><strong>Email Association:</strong> The email address will be linked to the contact for future reference</li>
            <li><strong>AI-Ready:</strong> All data is structured for future AI processing and analysis</li>
        </ul>
    </div>
</body>
</html>`))
		return
	}

	// Handle POST - same as ForwardEmail but with HTML response
	if err := r.ParseForm(); err != nil {
		w.Write([]byte(`<div class="error">Failed to parse form data</div>`))
		return
	}

	emailFrom := r.FormValue("from")
	_ = r.FormValue("to") // Currently unused but available for future features
	subject := r.FormValue("subject")
	body := r.FormValue("body")

	if emailFrom == "" || body == "" {
		w.Write([]byte(`<div class="error">From address and body are required</div>`))
		return
	}

	// Try to find existing contact by email
	contact, err := h.db.GetContactByEmail(emailFrom)
	if err != nil {
		// Contact doesn't exist, create a new one
		// Extract name from email (before @)
		name := emailFrom
		if atIndex := strings.Index(emailFrom, "@"); atIndex > 0 {
			name = emailFrom[:atIndex]
		}
		
		contact, err = h.db.CreateContact(name, emailFrom, "", "person", "Created from forwarded email")
		if err != nil {
			log.Printf("Error creating contact from email: %v", err)
			w.Write([]byte(`<div class="error">Failed to create contact</div>`))
			return
		}
	}

	// Create thread entry for the forwarded email
	_, err = h.db.CreateContactThread(contact.ID, nil, subject, body, "email", "inbound")
	if err != nil {
		log.Printf("Error creating thread for forwarded email: %v", err)
		w.Write([]byte(`<div class="error">Failed to save email thread</div>`))
		return
	}

	w.Write([]byte(fmt.Sprintf(`<div class="success">
		✅ Email successfully forwarded and saved!<br>
		📞 Contact: %s (%s)<br>
		📧 Subject: %s<br>
		<a href="/">← Back to Dashboard</a> | 
		<a href="/contacts/email/parse">Forward Another Email</a>
	</div>`, contact.Name, contact.Email, subject)))
}