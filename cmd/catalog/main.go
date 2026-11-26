package main

import (
	"database/sql"
	"fmt"
	"os"

	"mangahub/internal/catalog"

	_ "modernc.org/sqlite"
)

func main() {
	if err := os.MkdirAll("./internal/database", 0755); err != nil {
		fmt.Println("Error creating database directory:", err)
		return
	}

	db, err := sql.Open("sqlite", "./internal/database/catalog.db")
	if err != nil {
		fmt.Println("Error opening catalog.db:", err)
		return
	}
	defer db.Close()

	repo := catalog.NewCatalogRepo(db)

	if err := repo.InitSchema(); err != nil {
		fmt.Println("Error initializing schema:", err)
		return
	}

	if err := repo.SeedIfEmpty(); err != nil {
		fmt.Println("Error seeding catalog:", err)
		return
	}

	fmt.Println("Catalog DB ready!")
}
