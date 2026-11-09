package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "modernc.org/sqlite" //important import for sqlite
)

func main() {
	if err := os.MkdirAll("./internal/database", 0755); err != nil {
		fmt.Println("Error creating database:", err)
		return
	}

	db, err := sql.Open("sqlite", "./internal/database/catalog.db")
	if err != nil {
		fmt.Println("Error opening table:", err)
		return
	}
	defer db.Close()

	fmt.Println("Connected to catalog.db")

	schema := `
	CREATE TABLE IF NOT EXISTS manga (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		author TEXT NOT NULL, 
		genre TEXT NOT NULL, 
		status TEXT NOT NULL, 
		total_chap INTEGER NOT NULL DEFAULT 0,
		description TEXT 
	)
	`

	if _, err := db.Exec(schema); err != nil {
		fmt.Println("Error using schema:", err)
		return
	}

	fmt.Println("Created table!!")
}