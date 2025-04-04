package main

import (
	"Roshamble/internal/handlers"
	"Roshamble/internal/routes"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

var db *sql.DB

func main() {
	initDB()

	r := gin.Default()

	// Load templates
	r.LoadHTMLGlob("templates/**/*")

	// Serve static files
	r.Static("/assets", "./assets")

	handler := &handlers.Handler{DB: db}

	// Add public routes
	routes.PublicRoutes(r, handler)

	// Add Protected routes
	routes.ProtectedRoutes(r, handler)

	port := ":4000"
	fmt.Printf("Server running at http://localhost%s\n", port)
	r.Run(port)
}

func initDB() {
	var err error
	connStr := "postgres://postgres:password@localhost:5432/postgres?sslmode=disable"
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Cron jobs to debug tournament behavior
	go func() {
		ticker := time.NewTicker(time.Second * 20)

		for range ticker.C {
			_, err := db.Exec("INSERT INTO tournaments (name, prize, prize_url, description) VALUES ($1, $2, $3, $4)", "Test", "A Rivian RT-1", "https://example.com", "A test tournament")
			if err != nil {
				slog.Error("Error creating test tournaments ", "error ", err.Error())
			}
			slog.Info("Created tournament")
		}
	}()

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(0)

	if err := db.Ping(); err != nil {
		log.Fatalf("Database ping failed: %v", err)
	}

	applyMigrations()

}

func executeDeleteFunction(db *sql.DB, functionName string) {
	_, err := db.Exec("SELECT " + functionName + "()")
	if err != nil {
		log.Printf("Error executing %s: %v", functionName, err)
	} else {
		log.Printf("%s executed successfully", functionName)
	}
}
func runMatchmaking(db *sql.DB) {
	_, err := db.Exec("SELECT find_and_match_players();")
	if err != nil {
		log.Printf("Error running matchmaking: %v\n", err)
	} else {
		fmt.Println("Matchmaking executed successfully")
	}
}

func applyMigrations() {
	migrationsDir := "migrations"
	if err := goose.Up(db, migrationsDir); err != nil {
		log.Fatalf("Failed to apply migrations: %v", err)
	}
	log.Println("Database migrations applied successfully.")
}
