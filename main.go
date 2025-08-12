package main

import (
	"log"
	"net/http"
	"os"

	"oppgaave/internal/database"
	"oppgaave/internal/handlers"
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

	port := getEnv("PORT", "8080")
	log.Printf("🚀 ADHD Task Manager starting on port %s", port)
	log.Printf("📊 Dashboard: http://localhost:%s", port)

	if err := http.ListenAndServe(":"+port, h.Router()); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

// getEnv gets an environment variable with a fallback default
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}