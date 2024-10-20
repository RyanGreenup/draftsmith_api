/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"database/sql"
	_ "embed"
	"fmt"
	"log"

    utils "draftsmith/src/utils"
	_ "github.com/lib/pq"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

//go:embed draftsmith.sql
var sql_commands string

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the database",
	Long: `This will initialize the database and create the necessary tables.

This requires the database to be dropped first, use the drop command to do this.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Initializing database...")

		DB_NAME := "draftsmith"

		// // Get database connection details from viper
		// dbHost := viper.GetString("db_host")
		// dbPort := viper.GetInt("db_port")
		// dbUser := viper.GetString("db_user")
		// dbPass := viper.GetString("db_pass")
		//
		// // Connect to the default postgres database
		// connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		// 	dbHost, dbPort, dbUser, dbPass, "postgres")
		//
		// // Open database connection
		// db, err := sql.Open("postgres", connStr)
		// if err != nil {
		// 	log.Fatalf("Error opening database connection: %v", err)
		// }
        utils.Get_db()
        db := utils.get_db()
		defer db.Close()

		// Create the database
        stmt := fmt.Sprintf("CREATE DATABASE %s", DB_NAME)
		_, err = db.Exec(stmt) // TODO parameterize
		if err != nil {
			log.Fatalf("Error creating database: %v", err)
		}

		// Connect to the new database
		connStr = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			dbHost, dbPort, dbUser, dbPass, DB_NAME)
		db, err = sql.Open("postgres", connStr)
		if err != nil {
			log.Fatalf("Error opening database connection: %v", err)
		}
		defer db.Close()

		// Execute SQL commands
		fmt.Println(sql_commands)
		_, err = db.Exec(sql_commands)
		if err != nil {
			log.Fatalf("Error executing SQL commands: %v", err)
		}

		fmt.Println("Database initialized successfully.")
	},
}

func init() {
	cliCmd.AddCommand(initCmd)
}
