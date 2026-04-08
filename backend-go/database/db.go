package database

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func InitDB() {
	var err error
	workDir, _ := os.Getwd()
	dbPath := filepath.Join(workDir, "data.db") + "?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)"

	DB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	// Increasing max open connections to allow parallel LogScan calls.
	// With WAL mode, SQLite can handle multiple concurrent readers/writers better.
	DB.SetMaxOpenConns(25)

	createTables := `
	CREATE TABLE IF NOT EXISTS domains (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		hostname TEXT UNIQUE NOT NULL,
		ssl_expiry DATETIME,
		domain_expiry DATETIME,
		last_scan DATETIME,
		status TEXT DEFAULT 'pending',
		nameservers TEXT,
		security_rating TEXT,
		status_availability TEXT,
		last_whois_raw TEXT
	);

	CREATE INDEX IF NOT EXISTS idx_ssl_expiry ON domains (ssl_expiry);
	CREATE INDEX IF NOT EXISTS idx_domain_expiry ON domains (domain_expiry);
	CREATE UNIQUE INDEX IF NOT EXISTS idx_hostname_unique ON domains (hostname);

	CREATE TABLE IF NOT EXISTS access (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		provider TEXT NOT NULL,
		config TEXT NOT NULL, -- JSON string
		is_encrypted INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_access_name ON access (name);

	CREATE TABLE IF NOT EXISTS certificates (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		domain TEXT NOT NULL,
		certificate TEXT NOT NULL,
		private_key TEXT NOT NULL,
		issuer TEXT,
		expiry DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_cert_domain ON certificates (domain);

	CREATE TABLE IF NOT EXISTS workflows (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		domain_id INTEGER,
		certificate_id INTEGER,
		access_id INTEGER, -- Primary deployment target
		type TEXT, -- 'acme', 'deploy', 'both'
		status TEXT DEFAULT 'idle',
		last_run DATETIME,
		config TEXT, -- JSON workflow steps
		FOREIGN KEY(domain_id) REFERENCES domains(id),
		FOREIGN KEY(certificate_id) REFERENCES certificates(id),
		FOREIGN KEY(access_id) REFERENCES access(id)
	);

	CREATE INDEX IF NOT EXISTS idx_workflow_domain ON workflows (domain_id);
	CREATE INDEX IF NOT EXISTS idx_workflow_type ON workflows (type);

	CREATE TABLE IF NOT EXISTS settings (
		key TEXT PRIMARY KEY,
		value TEXT
	);

	INSERT OR IGNORE INTO settings (key, value) VALUES ('webhook_url', '');
	INSERT OR IGNORE INTO settings (key, value) VALUES ('notification_threshold', '30');
	INSERT OR IGNORE INTO settings (key, value) VALUES ('telegram_token', '');
	INSERT OR IGNORE INTO settings (key, value) VALUES ('telegram_chat_id', '');
	INSERT OR IGNORE INTO settings (key, value) VALUES ('discord_webhook_url', '');
	INSERT OR IGNORE INTO settings (key, value) VALUES ('slack_webhook_url', '');
	INSERT OR IGNORE INTO settings (key, value) VALUES ('smtp_host', '');
	INSERT OR IGNORE INTO settings (key, value) VALUES ('smtp_port', '587');
	INSERT OR IGNORE INTO settings (key, value) VALUES ('smtp_user', '');
	INSERT OR IGNORE INTO settings (key, value) VALUES ('smtp_pass', '');
	INSERT OR IGNORE INTO settings (key, value) VALUES ('smtp_from', '');
	INSERT OR IGNORE INTO settings (key, value) VALUES ('smtp_to', '');
	INSERT OR IGNORE INTO settings (key, value) VALUES ('scan_interval', '24');

	CREATE TABLE IF NOT EXISTS scan_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		hostname TEXT,
		type TEXT,
		status TEXT,
		message TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_logs_hostname ON scan_logs (hostname);
	CREATE INDEX IF NOT EXISTS idx_logs_created ON scan_logs (created_at);
	`

	_, err = DB.Exec(createTables)
	if err != nil {
		log.Printf("[DB] Note: Table creation handled: %v", err)
	}

	// Safety: Add columns manually if they don't exist (Migrations)
	DB.Exec("ALTER TABLE access ADD COLUMN is_encrypted INTEGER DEFAULT 0")

	// DEDUPLICATION: Remove duplicates from domains table strictly
	_, err = DB.Exec("DELETE FROM domains WHERE id NOT IN (SELECT MIN(id) FROM domains GROUP BY LOWER(hostname))")
	if err == nil {
		log.Println("[DB] Deduplication complete. Only unique hostnames remain (Case Insensitive).")
	}

	log.Println("[DB] SQLite database initialized successfully.")
}

func LogScan(hostname, logType, status, message string) {
	DB.Exec("INSERT INTO scan_logs (hostname, type, status, message) VALUES (?, ?, ?, ?)", hostname, logType, status, message)
}
