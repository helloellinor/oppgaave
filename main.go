package main

import (
	"log"
	"net/http"
	"os"

	"oppgaave/internal/database"
	"oppgaave/internal/handlers"

	"github.com/gorilla/mux"
)

func main() {
	// Initialize database
	dbPath := getEnv("DATABASE_PATH", "./tasks.db")
	db, err := database.New(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize handlers
	h := handlers.New(db)

	// Setup routes
	r := mux.NewRouter()

	// Static files
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	// Main dashboard
	r.HandleFunc("/", h.Dashboard).Methods("GET")

	// HTMX endpoints for dynamic content
	r.HandleFunc("/tasks", h.GetTaskList).Methods("GET")
	r.HandleFunc("/tasks/create", h.CreateTask).Methods("GET", "POST")
	r.HandleFunc("/tasks/{id}/status", h.UpdateTaskStatus).Methods("POST")
	r.HandleFunc("/budget-widget", h.GetBudgetWidget).Methods("GET")

	// JSON API endpoints
	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/tasks", h.GetTasksAPI).Methods("GET")
	api.HandleFunc("/tasks", h.CreateTaskAPI).Methods("POST")

	port := getEnv("PORT", "8080")
	log.Printf("🚀 ADHD Task Manager starting on port %s", port)
	log.Printf("📊 Dashboard: http://localhost:%s", port)
	log.Printf("🔧 API: http://localhost:%s/api/tasks", port)

	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// getEnv gets an environment variable with a fallback default
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}