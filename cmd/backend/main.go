package main

import (
	"database/sql"
	"fmt"
	"mangahub/internal/auth"
	"mangahub/internal/handler"
	"mangahub/internal/repo"
	"mangahub/internal/tcp"
	"os"

	"github.com/gin-gonic/gin"
	_ "modernc.org/sqlite"
)

func main() {
	jwtService, err := auth.NewJWTService()
	if err != nil {
		fmt.Println("JWT configuration error:", err)
		return
	}

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

	tcpServer := tcp.NewServer(":9090", jwtService)

	//allow tcp server and gin to working asynchronously
	go func() {
		if err := tcpServer.Start(); err != nil {
			fmt.Println("TCP server error:", err)
		}
	}()

	userRepo := repo.NewUserRepo(db)
	libRepo := repo.NewLibraryRepo(db)

	authHandler := handler.NewAuthHandler(userRepo, jwtService)
	libHandler := handler.NewLibraryHandler(libRepo, tcpServer)

	var version string
	err = db.QueryRow("SELECT sqlite_version()").Scan(&version)
	if err != nil {
		fmt.Println("Database error:", err)
		return
	}

	fmt.Println("Connected to backend.db! SQLite version:", version)

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

	r := gin.Default()

	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	auth := r.Group("/api/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
	}

	api := r.Group("/api")
	api.Use(handler.AuthMiddleware(jwtService))
	{
		api.GET("/library", libHandler.GetLibrary)
		api.POST("/library", libHandler.AddToLibrary)
		api.PUT("/library/:mangaId/progress", libHandler.UpdateProgress)
		api.DELETE("/library/:mangaId", libHandler.DeleteManga)
	}

	fmt.Println("Starting Gin on port 8080")
	if err := r.Run(":8080"); err != nil {
		fmt.Println("Failed to start server:", err)
	}
}
