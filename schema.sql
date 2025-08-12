-- ADHD Task Management System Database Schema

-- Tasks table with recursive structure
CREATE TABLE IF NOT EXISTS tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    description TEXT,
    parent_id INTEGER, -- For subtasks
    estimated_duration_minutes INTEGER DEFAULT 30,
    deadline DATETIME,
    priority INTEGER DEFAULT 1, -- 1=low, 2=medium, 3=high
    status TEXT DEFAULT 'pending', -- pending, in_progress, done, blocked
    tags TEXT, -- JSON array of tags
    energy_level INTEGER DEFAULT 2, -- 1=low, 2=medium, 3=high energy needed
    difficulty INTEGER DEFAULT 2, -- 1=easy, 2=medium, 3=hard
    money_cost INTEGER DEFAULT 0, -- Time budget cost in "coins"
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    completed_at DATETIME,
    FOREIGN KEY (parent_id) REFERENCES tasks(id)
);

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