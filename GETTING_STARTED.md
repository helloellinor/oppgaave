# Getting Started with ADHD Task Manager

## Quick Start

1. **Start the application:**
   ```bash
   go run main.go
   ```

2. **Open your browser and go to:**
   ```
   http://localhost:8080
   ```

## What You'll See

### 💰 Daily Budget Section
- **Wallet Metaphor**: Your day starts with $500 in "coins"
- **Visual Budget Bar**: Shows how much you've "spent" on tasks
- **Color-coded warnings**: 
  - Green: Budget looks good
  - Yellow: Running low
  - Red: Over budget

### 📅 Today's Focus
- Tasks planned for today
- Blocked tasks are clearly marked (🚫)
- Prerequisites shown for dependent tasks

### 📋 All Tasks Section
- Complete list of all your tasks
- **Status buttons**: Click ○ → ⏳ → ✓ to update progress
- **Add Task**: Click "➕ Add Task" to create new tasks

## Understanding the Money System

Each task has a "cost" calculated from:
- **Base time**: Duration in minutes
- **Energy multiplier**: Low energy tasks cost less
- **Difficulty multiplier**: Harder tasks cost more  
- **Priority multiplier**: Higher priority = higher cost (urgency)

## Task Features

### Prerequisites
- Tasks can depend on other tasks
- Blocked tasks show 🚫 until prerequisites are done
- Example: "Meal Prep" requires "Grocery Shopping" first

### Visual Priorities
- **High Priority**: Red border, high cost
- **Medium Priority**: Orange/yellow indicators
- **Low Priority**: Gray/muted colors

### Energy Levels
- **High Energy**: Tasks that require focus/motivation
- **Medium Energy**: Standard tasks
- **Low Energy**: Easy, low-effort tasks

## ADHD-Friendly Features

### Visual Design
- **Calming colors** to reduce overwhelm
- **Clear visual hierarchy** with priorities
- **Gentle animations** for status changes
- **Generous spacing** to prevent crowding

### Cognitive Support
- **Money metaphor** makes time budgets tangible
- **Consequence awareness** without guilt
- **Progress visualization** for motivation
- **Task blocking** prevents overwhelm

### Interaction Design
- **One-click status updates**
- **Modal forms** keep context
- **Real-time updates** via HTMX
- **No page refreshes** to maintain flow

## API Usage

The system also provides a JSON API:

### Get all tasks:
```bash
curl http://localhost:8080/api/tasks
```

### Create a new task:
```bash
curl -X POST -H "Content-Type: application/json" \
  -d '{"title":"My Task","description":"Task details","estimated_duration_minutes":30,"priority":2}' \
  http://localhost:8080/api/tasks
```

## Database

Tasks are stored in SQLite (`tasks.db`) with:
- Task hierarchy (parent/child relationships)
- Prerequisites (DAG structure)
- Daily budgets
- Task scheduling information

## Customization

Edit these files to customize:
- `static/style.css` - Visual styling
- `schema.sql` - Database structure and sample data
- `templates/` - HTML templates
- `internal/models/task.go` - Cost calculation logic

## Next Steps

This MVP demonstrates the core ADHD-friendly concepts. Future enhancements could include:
- AI-powered task estimation
- Calendar integration
- Habit tracking
- Gamification elements
- Mobile app
- Raylib desktop visualizations