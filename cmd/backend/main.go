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
}