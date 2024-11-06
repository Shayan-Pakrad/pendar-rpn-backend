package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

var DB *sql.DB

// Initialize database connection
func InitDB() {

	// use the code below to run project localy
	// connStr := "user=postgres password=admin dbname=rpn_webservice_database sslmode=disable"

	connStr := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable host=%s port=%d",
		"postgres",
		"admin",
		"rpn_webservice_database",
		"db",
		5432,
	)

	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	if err := DB.Ping(); err != nil {
		log.Fatalf("Database connection error: %v", err)
	}

	log.Println("Connected to the database successfully!")
}
