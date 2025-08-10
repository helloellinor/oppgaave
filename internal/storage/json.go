package storage

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"

	"oppgaave/internal/calendar"
)

// JSONStorage provides JSON-based persistence for calendar data
type JSONStorage struct {
	dataDir    string
	calendarFile string
	mutex      sync.RWMutex
}

// CalendarData represents the structure of calendar data in JSON
type CalendarData struct {
	Events    map[string]*calendar.Event `json:"events"`
	Version   string                     `json:"version"`
	CreatedAt time.Time                  `json:"created_at"`
	UpdatedAt time.Time                  `json:"updated_at"`
}

// NewJSONStorage creates a new JSON storage instance
func NewJSONStorage(dataDir string) (*JSONStorage, error) {
	if dataDir == "" {
		return nil, fmt.Errorf("data directory cannot be empty")
	}

	// Create data directory if it doesn't exist
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	storage := &JSONStorage{
		dataDir:      dataDir,
		calendarFile: filepath.Join(dataDir, "calendar.json"),
	}

	return storage, nil
}

// SaveCalendar saves calendar data to JSON file
func (s *JSONStorage) SaveCalendar(cal *calendar.Calendar) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Get all events from calendar
	events := cal.GetAllEvents()
	eventMap := make(map[string]*calendar.Event)
	for _, event := range events {
		eventMap[event.ID] = event
	}

	// Create calendar data structure
	data := CalendarData{
		Events:    eventMap,
		Version:   "1.0",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Check if file exists to preserve creation time
	if existingData, err := s.loadCalendarData(); err == nil {
		data.CreatedAt = existingData.CreatedAt
	}

	// Create backup before saving
	if err := s.createBackup(); err != nil {
		// Log warning but don't fail the save operation
		fmt.Printf("Warning: failed to create backup: %v\n", err)
	}

	// Marshal to JSON with indentation for readability
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal calendar data: %w", err)
	}

	// Write to temporary file first, then rename (atomic operation)
	tempFile := s.calendarFile + ".tmp"
	if err := os.WriteFile(tempFile, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write calendar data: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tempFile, s.calendarFile); err != nil {
		os.Remove(tempFile) // Clean up temp file
		return fmt.Errorf("failed to save calendar data: %w", err)
	}

	return nil
}

// LoadCalendar loads calendar data from JSON file
func (s *JSONStorage) LoadCalendar() (*calendar.Calendar, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	cal := calendar.NewCalendar()

	// Check if file exists
	if _, err := os.Stat(s.calendarFile); os.IsNotExist(err) {
		// Return empty calendar if file doesn't exist
		return cal, nil
	}

	data, err := s.loadCalendarData()
	if err != nil {
		return nil, err
	}

	// Load events into calendar
	for _, event := range data.Events {
		if err := cal.AddEvent(event); err != nil {
			// Log warning but continue loading other events
			fmt.Printf("Warning: failed to load event %s: %v\n", event.ID, err)
		}
	}

	return cal, nil
}

// loadCalendarData loads raw calendar data from JSON file
func (s *JSONStorage) loadCalendarData() (*CalendarData, error) {
	jsonData, err := os.ReadFile(s.calendarFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read calendar file: %w", err)
	}

	var data CalendarData
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal calendar data: %w", err)
	}

	// Validate data version
	if data.Version == "" {
		data.Version = "1.0" // Default version for legacy data
	}

	return &data, nil
}

// createBackup creates a backup of the current calendar file
func (s *JSONStorage) createBackup() error {
	// Check if original file exists
	if _, err := os.Stat(s.calendarFile); os.IsNotExist(err) {
		return nil // No backup needed if original doesn't exist
	}

	// Create backup filename with timestamp
	timestamp := time.Now().Format("20060102_150405")
	backupFile := filepath.Join(s.dataDir, fmt.Sprintf("calendar_backup_%s.json", timestamp))

	// Copy original file to backup
	originalData, err := os.ReadFile(s.calendarFile)
	if err != nil {
		return fmt.Errorf("failed to read original file for backup: %w", err)
	}

	if err := os.WriteFile(backupFile, originalData, 0644); err != nil {
		return fmt.Errorf("failed to create backup file: %w", err)
	}

	// Clean up old backups (keep only last 10)
	if err := s.cleanupOldBackups(); err != nil {
		// Log warning but don't fail the backup operation
		fmt.Printf("Warning: failed to cleanup old backups: %v\n", err)
	}

	return nil
}

// cleanupOldBackups removes old backup files, keeping only the most recent ones
func (s *JSONStorage) cleanupOldBackups() error {
	pattern := filepath.Join(s.dataDir, "calendar_backup_*.json")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("failed to find backup files: %w", err)
	}

	// Keep only the 10 most recent backups
	if len(matches) <= 10 {
		return nil
	}

	// Sort by modification time (oldest first)
	type fileInfo struct {
		path    string
		modTime time.Time
	}

	var files []fileInfo
	for _, match := range matches {
		info, err := os.Stat(match)
		if err != nil {
			continue
		}
		files = append(files, fileInfo{path: match, modTime: info.ModTime()})
	}

	// Sort by modification time
	for i := 0; i < len(files)-1; i++ {
		for j := i + 1; j < len(files); j++ {
			if files[i].modTime.After(files[j].modTime) {
				files[i], files[j] = files[j], files[i]
			}
		}
	}

	// Remove oldest files
	filesToRemove := len(files) - 10
	for i := 0; i < filesToRemove; i++ {
		if err := os.Remove(files[i].path); err != nil {
			fmt.Printf("Warning: failed to remove old backup %s: %v\n", files[i].path, err)
		}
	}

	return nil
}

// ExportCalendar exports calendar data to a specified file
func (s *JSONStorage) ExportCalendar(cal *calendar.Calendar, exportPath string) error {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Get all events from calendar
	events := cal.GetAllEvents()
	eventMap := make(map[string]*calendar.Event)
	for _, event := range events {
		eventMap[event.ID] = event
	}

	// Create export data structure
	data := CalendarData{
		Events:    eventMap,
		Version:   "1.0",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Marshal to JSON with indentation
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal export data: %w", err)
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(exportPath), 0755); err != nil {
		return fmt.Errorf("failed to create export directory: %w", err)
	}

	// Write export file
	if err := os.WriteFile(exportPath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write export file: %w", err)
	}

	return nil
}

// ImportCalendar imports calendar data from a specified file
func (s *JSONStorage) ImportCalendar(importPath string) (*calendar.Calendar, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Check if import file exists
	if _, err := os.Stat(importPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("import file does not exist: %s", importPath)
	}

	// Read import file
	jsonData, err := os.ReadFile(importPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read import file: %w", err)
	}

	// Unmarshal data
	var data CalendarData
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal import data: %w", err)
	}

	// Create new calendar and load events
	cal := calendar.NewCalendar()
	for _, event := range data.Events {
		if err := cal.AddEvent(event); err != nil {
			fmt.Printf("Warning: failed to import event %s: %v\n", event.ID, err)
		}
	}

	return cal, nil
}

// GetStorageInfo returns information about the storage
func (s *JSONStorage) GetStorageInfo() (map[string]interface{}, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	info := make(map[string]interface{})
	info["data_directory"] = s.dataDir
	info["calendar_file"] = s.calendarFile

	// File information
	if stat, err := os.Stat(s.calendarFile); err == nil {
		info["file_size"] = stat.Size()
		info["last_modified"] = stat.ModTime()
	} else {
		info["file_exists"] = false
	}

	// Backup information
	pattern := filepath.Join(s.dataDir, "calendar_backup_*.json")
	if matches, err := filepath.Glob(pattern); err == nil {
		info["backup_count"] = len(matches)
	}

	return info, nil
}

// ValidateStorage validates the integrity of stored data
func (s *JSONStorage) ValidateStorage() error {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Check if file exists
	if _, err := os.Stat(s.calendarFile); os.IsNotExist(err) {
		return nil // No file to validate
	}

	// Try to load and validate data
	data, err := s.loadCalendarData()
	if err != nil {
		return fmt.Errorf("storage validation failed: %w", err)
	}

	// Validate each event
	for eventID, event := range data.Events {
		if event.ID != eventID {
			return fmt.Errorf("event ID mismatch: map key %s != event ID %s", eventID, event.ID)
		}

		if event.Title == "" {
			return fmt.Errorf("event %s has empty title", eventID)
		}

		if event.StartTime.IsZero() {
			return fmt.Errorf("event %s has zero start time", eventID)
		}

		if event.EndTime.IsZero() {
			return fmt.Errorf("event %s has zero end time", eventID)
		}

		if event.EndTime.Before(event.StartTime) {
			return fmt.Errorf("event %s has end time before start time", eventID)
		}
	}

	return nil
}

// RepairStorage attempts to repair corrupted storage data
func (s *JSONStorage) RepairStorage() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// First, try to validate current storage
	if err := s.ValidateStorage(); err == nil {
		return nil // No repair needed
	}

	// Look for the most recent backup
	pattern := filepath.Join(s.dataDir, "calendar_backup_*.json")
	matches, err := filepath.Glob(pattern)
	if err != nil || len(matches) == 0 {
		return fmt.Errorf("no backups available for repair")
	}

	// Find the most recent backup
	var mostRecent string
	var mostRecentTime time.Time
	for _, match := range matches {
		if info, err := os.Stat(match); err == nil {
			if mostRecent == "" || info.ModTime().After(mostRecentTime) {
				mostRecent = match
				mostRecentTime = info.ModTime()
			}
		}
	}

	if mostRecent == "" {
		return fmt.Errorf("no valid backup found for repair")
	}

	// Copy backup to main file
	backupData, err := os.ReadFile(mostRecent)
	if err != nil {
		return fmt.Errorf("failed to read backup for repair: %w", err)
	}

	if err := os.WriteFile(s.calendarFile, backupData, 0644); err != nil {
		return fmt.Errorf("failed to restore from backup: %w", err)
	}

	// Validate the restored data
	if err := s.ValidateStorage(); err != nil {
		return fmt.Errorf("restored data is still invalid: %w", err)
	}

	return nil
}
