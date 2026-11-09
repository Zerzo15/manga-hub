package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "modernc.org/sqlite"
)

func main() {
	if err := os.MkdirAll("./internal/database", 0755); err != nil {
		fmt.Println("Error creating database: ", err)
		return
	}

	db, err := sql.Open("sqlite", "./internal/database/backend.db")
	if err != nil {
		fmt.Println("Error open database: ", err)
		return
	}

	defer db.Close()

	var version string
	err = db.QueryRow("SELECT sqlite_version()").Scan(&version)
	if err != nil {
		fmt.Println("Database error:", err)
		return
	}

	fmt.Println("Connected! SQLite version:", version)

	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS user_library (
		user_id INTEGER NOT NULL,
		manga_id INTEGER NOT NULL,
		status TEXT NOT NULL,
		current_chap INTEGER DEFAULT 0,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		version INTEGER DEFAULT 1,
		PRIMARY KEY (user_id, manga_id),
		FOREIGN KEY (user_id) REFERENCES users(id)
	);
	`

	if _, err := db.Exec(schema); err != nil {
		fmt.Println("Error using schema: ", err)
		return
	}
	

	fmt.Println("Created table!")
}