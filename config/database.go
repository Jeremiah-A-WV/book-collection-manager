package config

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql" // MySQL driver
	"github.com/joho/godotenv"
	"log"
	"os"
)

// DB is the global database connection
var DB *sql.DB

// LoadDBConfig loads the database configuration and initializes the connection
func LoadDBConfig() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found — continuing with system env vars")
	}

	// Use environment variables
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	// Data source
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", dbUser, dbPassword, dbHost, dbPort, dbName)
	log.Printf("Connecting to DB as %s@%s:%s/%s", dbUser, dbHost, dbPort, dbName)

	// Connect to the database
	var err error
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	// Verify the connection
	if err := DB.Ping(); err != nil {
		log.Fatalf("Failed to ping the database: %v", err)
	}

	log.Println("Database connection established successfully")
}
