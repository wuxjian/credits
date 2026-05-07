package database

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func Init(dbPath string) {
	var err error
	DB, err = sql.Open("sqlite", dbPath+"?_journal_mode=WAL&_foreign_keys=on")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	createTables()
	log.Println("Database initialized successfully")
}

func createTables() {
	schema := `
	CREATE TABLE IF NOT EXISTS tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		points INTEGER NOT NULL DEFAULT 0,
		task_type TEXT NOT NULL CHECK(task_type IN ('once','daily')),
		deleted INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS child_progress (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		task_id INTEGER NOT NULL,
		completed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (task_id) REFERENCES tasks(id)
	);

	CREATE TABLE IF NOT EXISTS daily_task_status (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		task_id INTEGER NOT NULL,
		date TEXT NOT NULL,
		completed INTEGER NOT NULL DEFAULT 1,
		UNIQUE(task_id, date),
		FOREIGN KEY (task_id) REFERENCES tasks(id)
	);

	CREATE TABLE IF NOT EXISTS point_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		points_change INTEGER NOT NULL,
		reason TEXT NOT NULL DEFAULT '',
		type TEXT NOT NULL CHECK(type IN ('task','manual','redeem')),
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS redeem_items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		points_required INTEGER NOT NULL DEFAULT 0,
		active INTEGER NOT NULL DEFAULT 1
	);

	CREATE TABLE IF NOT EXISTS current_points (
		id INTEGER PRIMARY KEY CHECK(id = 1),
		total_points INTEGER NOT NULL DEFAULT 0
	);

	INSERT OR IGNORE INTO current_points (id, total_points) VALUES (1, 0);
	`

	_, err := DB.Exec(schema)
	if err != nil {
		log.Fatalf("Failed to create tables: %v", err)
	}
}
