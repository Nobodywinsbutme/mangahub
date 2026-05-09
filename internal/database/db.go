package database

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite" // blank import: registers the SQLite driver
)

// DB is the global database connection pool.
// Using a package-level variable lets all handlers share one connection.
var DB *sql.DB

func Init(dbPath string) error {
	// Ensure the directory exists (e.g., ./data/)
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return err
	}

	var err error
	DB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}

	// Verify the connection is actually alive
	if err = DB.Ping(); err != nil {
		return err
	}

	log.Println("✓ Database connection established")
	return runMigrations()
}

func runMigrations() error {
	schema := `
    CREATE TABLE IF NOT EXISTS users (
        id            TEXT PRIMARY KEY,
        username      TEXT UNIQUE NOT NULL,
        email         TEXT UNIQUE NOT NULL,
        password_hash TEXT NOT NULL,
        created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );
    CREATE TABLE IF NOT EXISTS manga (
        id             TEXT PRIMARY KEY,
        title          TEXT NOT NULL,
        author         TEXT NOT NULL,
        genres         TEXT DEFAULT '[]',
        status         TEXT NOT NULL,
        total_chapters INTEGER DEFAULT 0,
        description    TEXT DEFAULT ''
    );
    CREATE TABLE IF NOT EXISTS user_progress (
        user_id         TEXT NOT NULL,
        manga_id        TEXT NOT NULL,
        current_chapter INTEGER DEFAULT 0,
        status          TEXT NOT NULL,
        updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        PRIMARY KEY (user_id, manga_id),
        FOREIGN KEY (user_id) REFERENCES users(id),
        FOREIGN KEY (manga_id) REFERENCES manga(id)
    );`

	_, err := DB.Exec(schema)
	if err != nil {
		return err
	}

	log.Println("✓ Database schema ready")
	return nil
}
