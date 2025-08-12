#!/bin/bash

# Check required commands
if ! command -v sqlite3 > /dev/null; then
    echo "❌ sqlite3 is required but not installed."
    echo "Please install it with: brew install sqlite3"
    exit 1
fi

if ! command -v go > /dev/null; then
    echo "❌ Go is required but not installed."
    echo "Please install it with: brew install go"
    exit 1
fi

# Create uploads directory if it doesn't exist
if [ ! -d uploads ]; then
    mkdir -p uploads
fi

# Create database and start server
echo "🗄️ Setting up database..."
rm -f tasks.db

# Check if schema.sql exists
if [ ! -f schema.sql ]; then
    echo "❌ schema.sql not found!"
    exit 1
fi

# Initialize database
if ! sqlite3 tasks.db < schema.sql; then
    echo "❌ Failed to initialize database!"
    exit 1
fi

# Add sample data
sqlite3 tasks.db << 'ENDSQL'
-- Doctor's appointment
INSERT INTO tasks (title, description, type, estimated_duration_minutes, start_time, location, energy_level, difficulty, money_cost, status)
VALUES ('Annual Checkup',
        'Regular health checkup with Dr. Smith

Remember to bring:
- Insurance card
- List of medications
- Recent test results',
        'appointment', 60, datetime('now', '+3 days', '+10 hours'),
        'City Medical Center, Room 305', 7, 3, 50, 'pending');

-- Concert event
INSERT INTO tasks (title, description, type, estimated_duration_minutes, start_time, location, energy_level, difficulty, money_cost, status)
VALUES ('Taylor Swift Concert',
        'The Eras Tour!

Doors open at 6:30 PM
Show starts at 8:00 PM

Meet friends at entrance',
        'event', 240, datetime('now', '+7 days', '+19 hours'),
        'Stadium Arena, Gate B', 9, 2, 200, 'pending');

-- Workout routine
INSERT INTO tasks (title, description, type, estimated_duration_minutes, start_time, location, energy_level, difficulty, status, tags)
VALUES ('Morning Yoga',
        'Gentle morning flow

- Sun salutations
- Standing poses
- Cool down stretches',
        'workout', 45, datetime('now', '+1 days', '+7 hours'),
        'Home Yoga Studio', 6, 4, 'pending', '["exercise","morning-routine"]');

-- Work project
INSERT INTO tasks (title, description, type, estimated_duration_minutes, deadline, location, energy_level, difficulty, status)
VALUES ('Quarterly Report',
        'Prepare Q3 financial analysis

Sections:
- Revenue breakdown
- Expense analysis
- Growth projections',
        'work', 180, datetime('now', '+5 days', '+17 hours'),
        'Home Office', 8, 7, 'pending');

-- Social event
INSERT INTO tasks (title, description, type, estimated_duration_minutes, start_time, location, energy_level, difficulty, money_cost, status)
VALUES ('Dinner with Friends',
        'Catching up with college buddies

Bring wine!',
        'social', 120, datetime('now', '+2 days', '+18 hours'),
        'Italian Restaurant Downtown', 5, 2, 40, 'pending');

-- Daily chore with subtasks
INSERT INTO tasks (title, description, type, estimated_duration_minutes, start_time, location, energy_level, difficulty, status)
VALUES ('House Cleaning',
        'Weekly deep clean',
        'chore', 120, datetime('now', '+1 days', '+14 hours'),
        'Home', 7, 5, 'pending');

-- Store the parent ID for subtasks
CREATE TEMPORARY TABLE vars(parent_id INTEGER);
INSERT INTO vars(parent_id) VALUES(last_insert_rowid());

-- Add subtasks
INSERT INTO tasks (title, type, parent_id, estimated_duration_minutes, energy_level, difficulty, status)
SELECT 'Vacuum all rooms', 'chore', parent_id, 30, 6, 4, 'pending' FROM vars
UNION ALL
SELECT 'Clean bathrooms', 'chore', parent_id, 45, 8, 6, 'pending' FROM vars
UNION ALL
SELECT 'Dust furniture', 'chore', parent_id, 20, 4, 3, 'pending' FROM vars
UNION ALL
SELECT 'Mop floors', 'chore', parent_id, 25, 7, 5, 'pending' FROM vars;

DROP TABLE vars;
ENDSQL

# Build and run the server
echo "🔨 Building server..."
if ! go build -o oppgaave main.go; then
    echo "❌ Failed to build server!"
    exit 1
fi

echo "🚀 Starting server..."
echo "Open http://localhost:8080 in your browser"
./oppgaave
