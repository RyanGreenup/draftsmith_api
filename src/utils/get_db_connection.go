package cmd

import (
	"database/sql"
	_ "embed"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func get_db(dbHost string, dbPort int, dbUser string, dbPass string, dbName string) *sql.DB {
    connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
        dbHost, dbPort, dbUser, dbPass, dbName)
    db, err := sql.Open("postgres", connStr)
    if err != nil {
        log.Fatalf("Error opening database connection: %v", err)
    }
    return db
}
