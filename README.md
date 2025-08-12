# 🧠 ADHD Task Management System

An ADHD-friendly task management system built with Go and HTMX, featuring time budgeting with a money allegory, recursive task structures, and consequence/reward awareness.

## Features

- **Time as Money**: Daily budget visualization using a wallet metaphor
- **Recursive Tasks**: Support for subtasks and prerequisites (DAG structure)
- **HTMX Interface**: Dynamic updates without page refreshes
- **ADHD-Friendly Design**: Calming colors, clear priorities, and gentle nudging
- **Consequence Awareness**: Visual warnings and rewards based on task completion

## Quick Start

1. **Run the application:**
   ```bash
   go run main.go
   ```

2. **Open your browser:**
   ```
   http://localhost:8080
   ```

3. **Start managing tasks!**
   - Click "➕ Add Task" to create new tasks
   - Click the status buttons (○, ⏳, ✓) to update task progress
   - Watch your daily budget update in real-time

## Architecture

- **Backend**: Go HTTP server with SQLite database
- **Frontend**: HTMX + CSS (minimal JavaScript)
- **Database**: SQLite with support for:
  - Tasks with recursive structure
  - Prerequisites (DAG relationships)
  - Daily budget tracking
  - Task scheduling

## Development Roadmap

### Phase 1 - MVP ✅
- [x] Basic task CRUD operations
- [x] Time budgeting with money allegory
- [x] HTMX-powered interface
- [x] Task prerequisites and subtasks
- [x] Daily budget visualization

### Phase 2 - Enhanced Features (Future)
- [ ] AI-powered task splitting and estimation
- [ ] Smart energy/difficulty matching
- [ ] Contact management
- [ ] Event integration
- [ ] Meal planning module

### Phase 3 - Advanced Features (Future)
- [ ] Modular addon system
- [ ] External integrations (calendar, weather, etc.)
- [ ] Advanced visualizations
- [ ] Gamification elements

### Phase 4 - Desktop Experience (Future)
- [ ] Raylib integration for rich visualizations
- [ ] Offline desktop app
- [ ] Advanced timeline interactions

## Technology Stack

- **Go**: HTTP server, business logic
- **SQLite**: Data persistence
- **HTMX**: Dynamic frontend interactions
- **CSS**: ADHD-friendly styling
- **Gorilla Mux**: HTTP routing

## Database Schema

The system uses SQLite with tables for:
- `tasks` - Main task data with recursive relationships
- `task_prerequisites` - DAG structure for dependencies
- `daily_budgets` - Time budget tracking
- `task_schedule` - Task scheduling and timing
- `settings` - User preferences

## Contributing

This is a prototype system designed to demonstrate ADHD-friendly task management concepts. Feel free to extend and improve!
