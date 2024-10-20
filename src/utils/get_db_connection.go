package cmd

import (
	"database/sql"
	_ "embed"
	"fmt"
	"log"

	_ "github.com/lib/pq"
	"github.com/spf13/viper"
)

// Opens a new connection to the specified PostgreSQL database.
// using the connection details from the parent command
func Get_db(dbName string) *sql.DB {
	dbHost := viper.GetString("db_host")
	dbPort := viper.GetInt("db_port")
	dbUser := viper.GetString("db_user")
	dbPass := viper.GetString("db_pass")
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",

		dbHost, dbPort, dbUser, dbPass, dbName)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Error opening database connection: %v", err)
	}
	return db
}

// Drop the specified database
// This assumes there are no active connections to the database
// and that the database exists
// Also assumes there is a postgres database to connect to
func Drop_db() {
	fmt.Println("Dropping database...")

	dbName := "draftsmith"

	// Connect to the default database
	db := Get_db("postgres")
	defer db.Close()

	// Drop the specified database
	stmt := fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbName)
	_, err := db.Exec(stmt)
	if err != nil {
		log.Fatalf("Error dropping database: %v", err)
	}

	fmt.Println("Database dropped successfully")
}

// Create the new database (called "draftsmith")
// This assumes the database does not already exist
// Also assumes there is a postgres database to connect to
func Create_db() {
	dbName := "draftsmith"

	// Connect to the default database
	db := Get_db("postgres")
	defer db.Close()

	// Create the new database
	stmt := fmt.Sprintf("CREATE DATABASE %s", dbName)
	_, err := db.Exec(stmt)
	if err != nil {
		log.Fatalf("Error creating database: %v", err)
	}
}
