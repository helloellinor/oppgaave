-- ADHD Task Management System Database Schema

-- Drop existing tables if they exist
DROP TABLE IF EXISTS attachments;
DROP TABLE IF EXISTS tasks;

CREATE TABLE tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    type VARCHAR(50) NOT NULL DEFAULT 'other',
    parent_id INTEGER REFERENCES tasks(id),
    estimated_duration_minutes INTEGER DEFAULT 0,
    start_time DATETIME,
    deadline DATETIME,
    priority INTEGER DEFAULT 0,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    tags TEXT DEFAULT '[]',
    energy_level INTEGER DEFAULT 5,
    difficulty INTEGER DEFAULT 5,
    money_cost INTEGER DEFAULT 0,
    location TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    completed_at DATETIME
);

-- Create indexes for better performance
CREATE INDEX idx_tasks_start_time ON tasks(start_time);
CREATE INDEX idx_tasks_deadline ON tasks(deadline);
CREATE INDEX idx_tasks_type ON tasks(type);
CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_parent_id ON tasks(parent_id);
CREATE INDEX idx_tasks_priority ON tasks(priority);
CREATE INDEX idx_tasks_created_at ON tasks(created_at);

-- Search indexes
CREATE INDEX idx_tasks_title ON tasks(title);
CREATE INDEX idx_tasks_description ON tasks(description);

-- Attachments table
CREATE TABLE attachments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id INTEGER REFERENCES tasks(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL,
    path TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_attachments_task_id ON attachments(task_id);
CREATE INDEX idx_attachments_type ON attachments(type);

-- Task prerequisites (DAG structure)
CREATE TABLE IF NOT EXISTS task_prerequisites (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id INTEGER NOT NULL,
    prerequisite_task_id INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (task_id) REFERENCES tasks(id),
    FOREIGN KEY (prerequisite_task_id) REFERENCES tasks(id),
    UNIQUE(task_id, prerequisite_task_id)
);

-- Daily budgets for time management
CREATE TABLE IF NOT EXISTS daily_budgets (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date DATE NOT NULL UNIQUE,
    total_budget_coins INTEGER DEFAULT 500, -- Daily budget in "coins"
    spent_coins INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Task assignments to days
CREATE TABLE IF NOT EXISTS task_schedule (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id INTEGER NOT NULL,
    scheduled_date DATE NOT NULL,
    start_time TIME,
    estimated_end_time TIME,
    actual_start_time DATETIME,
    actual_end_time DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (task_id) REFERENCES tasks(id)
);

-- User settings and preferences
CREATE TABLE IF NOT EXISTS settings (
    key TEXT PRIMARY KEY,
    value TEXT,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Initial settings
INSERT OR REPLACE INTO settings (key, value) VALUES 
    ('daily_budget_coins', '500'),
    ('coin_per_minute', '10'),
    ('energy_multiplier', '1.5'),
    ('difficulty_multiplier', '1.3');

-- Sample tasks for demonstration
INSERT OR REPLACE INTO tasks (id, title, description, estimated_duration_minutes, priority, status, energy_level, difficulty, money_cost) VALUES
    (1, 'Morning Coffee & Journal', 'Start the day with coffee and journaling to set intentions', 15, 1, 'pending', 1, 1, 15),
    (2, 'Write Project Draft', 'Complete first draft of the project proposal', 120, 3, 'pending', 3, 3, 180),
    (3, 'Yoga Class', 'Attend morning yoga session for physical and mental wellness', 45, 2, 'pending', 2, 1, 45),
    (4, 'Call Landlord', 'Important call about lease renewal - deadline approaching', 30, 3, 'pending', 2, 2, 60),
    (5, 'Grocery Shopping', 'Buy ingredients for meal prep', 60, 2, 'pending', 2, 2, 60),
    (6, 'Meal Prep', 'Prepare meals for the week', 90, 2, 'pending', 3, 2, 90);

-- Add some prerequisites
INSERT OR REPLACE INTO task_prerequisites (task_id, prerequisite_task_id) VALUES
    (6, 5), -- Meal prep requires grocery shopping first
    (2, 1); -- Writing requires coffee/journal first for focus