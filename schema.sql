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
    task_type TEXT DEFAULT 'task', -- task, appointment, event, concert, meeting
    event_location TEXT, -- For events and appointments
    event_start DATETIME, -- For scheduled events
    event_end DATETIME, -- For scheduled events
    radar_position_x REAL DEFAULT 0, -- X position on radar (time axis)
    radar_position_y REAL DEFAULT 0, -- Y position on radar (priority/energy axis)
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

-- Contacts for communication and task management
CREATE TABLE IF NOT EXISTS contacts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    email TEXT,
    phone TEXT,
    type TEXT DEFAULT 'person', -- person, organization, venue
    notes TEXT,
    avatar_url TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Contact threads for communication history
CREATE TABLE IF NOT EXISTS contact_threads (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    contact_id INTEGER NOT NULL,
    task_id INTEGER, -- Optional link to task
    subject TEXT,
    message TEXT NOT NULL,
    thread_type TEXT DEFAULT 'message', -- message, email, call, meeting
    direction TEXT DEFAULT 'outbound', -- inbound, outbound
    status TEXT DEFAULT 'sent', -- sent, received, pending, failed
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (contact_id) REFERENCES contacts(id),
    FOREIGN KEY (task_id) REFERENCES tasks(id)
);

-- Attachments for tasks and events
CREATE TABLE IF NOT EXISTS attachments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id INTEGER,
    contact_id INTEGER,
    filename TEXT NOT NULL,
    original_filename TEXT NOT NULL,
    file_path TEXT NOT NULL,
    file_size INTEGER,
    mime_type TEXT,
    description TEXT,
    attachment_type TEXT DEFAULT 'document', -- document, image, audio, video, link
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (task_id) REFERENCES tasks(id),
    FOREIGN KEY (contact_id) REFERENCES contacts(id)
);

-- Task contacts relationship (many-to-many)
CREATE TABLE IF NOT EXISTS task_contacts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id INTEGER NOT NULL,
    contact_id INTEGER NOT NULL,
    role TEXT DEFAULT 'participant', -- organizer, participant, venue, vendor
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (task_id) REFERENCES tasks(id),
    FOREIGN KEY (contact_id) REFERENCES contacts(id),
    UNIQUE(task_id, contact_id)
);

-- Initial settings
INSERT OR REPLACE INTO settings (key, value) VALUES 
    ('daily_budget_coins', '500'),
    ('coin_per_minute', '10'),
    ('energy_multiplier', '1.5'),
    ('difficulty_multiplier', '1.3');

-- Sample tasks for demonstration
INSERT OR REPLACE INTO tasks (id, title, description, estimated_duration_minutes, priority, status, energy_level, difficulty, money_cost, task_type, event_start, event_end) VALUES
    (1, 'Morning Coffee & Journal', 'Start the day with coffee and journaling to set intentions', 15, 1, 'pending', 1, 1, 15, 'task', NULL, NULL),
    (2, 'Write Project Draft', 'Complete first draft of the project proposal', 120, 3, 'pending', 3, 3, 180, 'task', NULL, NULL),
    (3, 'Yoga Class', 'Attend morning yoga session for physical and mental wellness', 45, 2, 'pending', 2, 1, 45, 'appointment', '2025-08-12 09:00:00', '2025-08-12 09:45:00'),
    (4, 'Call Landlord', 'Important call about lease renewal - deadline approaching', 30, 3, 'pending', 2, 2, 60, 'task', NULL, NULL),
    (5, 'Grocery Shopping', 'Buy ingredients for meal prep', 60, 2, 'pending', 2, 2, 60, 'task', NULL, NULL),
    (6, 'Meal Prep', 'Prepare meals for the week', 90, 2, 'pending', 3, 2, 90, 'task', NULL, NULL),
    (7, 'Team Meeting', 'Weekly standup with development team', 60, 2, 'pending', 2, 1, 60, 'meeting', '2025-08-12 14:00:00', '2025-08-12 15:00:00'),
    (8, 'Concert Planning', 'Plan upcoming jazz concert attendance', 30, 1, 'pending', 1, 1, 30, 'event', '2025-08-15 19:00:00', '2025-08-15 22:00:00');

-- Sample contacts
INSERT OR REPLACE INTO contacts (id, name, email, phone, type, notes) VALUES
    (1, 'Dr. Sarah Johnson', 'sarah.johnson@yogastudio.com', '+1-555-0123', 'person', 'Yoga instructor'),
    (2, 'Development Team', 'team@company.com', NULL, 'organization', 'Work team for standups'),
    (3, 'Jazz Venue', 'info@jazzclub.com', '+1-555-0456', 'venue', 'Downtown jazz club'),
    (4, 'Property Manager', 'landlord@property.com', '+1-555-0789', 'person', 'Lease renewal contact');

-- Link contacts to tasks
INSERT OR REPLACE INTO task_contacts (task_id, contact_id, role) VALUES
    (3, 1, 'organizer'), -- Yoga class with instructor
    (7, 2, 'participant'), -- Team meeting
    (8, 3, 'venue'); -- Concert at jazz venue

-- Add some prerequisites
INSERT OR REPLACE INTO task_prerequisites (task_id, prerequisite_task_id) VALUES
    (6, 5), -- Meal prep requires grocery shopping first
    (2, 1); -- Writing requires coffee/journal first for focus