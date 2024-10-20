/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"database/sql"
	_ "embed"
	"fmt"
	"log"

	_ "github.com/lib/pq"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

//go:embed draftsmith.sql
var sql_commands string

func getArg(name string, cmd cobra.Command) string {
	arg, err := cmd.Flags().GetString("db_host")
	if err != nil {
		log.Fatalf("Error getting db_host: %v", err)
	}
    return arg

}

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the database",
	Long: `This will initialize the database and create the necessary tables.

This requires the database to be dropped first, use the drop command to do this.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Initializing database...")

		DB_NAME := "draftsmith"

        // TODO get the parent flags

		// Connect to the default postgres database
		connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			dbHost, dbPort, dbUser, dbPass, "postgres")
		print(connStr)


		// Open database connection
		db, err := sql.Open("postgres", connStr)
		if err != nil {
			log.Fatalf("Error opening database connection: %v", err)
		}
		defer db.Close()

		// Create the database
		stmt := fmt.Sprintf("DROP DATABASE IF EXISTS %s", DB_NAME) // TODO move to drop command
		_, err = db.Exec(stmt)                                     // TODO parameterize
		if err != nil {
			log.Fatalf("Error dropping database: %v", err)

		}

		stmt = fmt.Sprintf("CREATE DATABASE %s", DB_NAME)
		_, err = db.Exec(stmt) // TODO parameterize
		if err != nil {
			log.Fatalf("Error creating database: %v", err)

		}

		db.Close()
		// Connect to the new database
		connStr = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			dbHost, dbPort, dbUser, dbPass, DB_NAME)
		db, err = sql.Open("postgres", connStr)
		if err != nil {
			log.Fatalf("Error opening database connection: %v", err)
		}
		defer db.Close()

		// Execute SQL commands
		print(sql_commands)
		_, err = db.Exec(sql_commands)
		if err != nil {
			log.Fatalf("Error executing SQL commands: %v", err)
		}

		fmt.Println("Database initialized successfully.")
	},
}

func init() {
	cliCmd.AddCommand(initCmd)

	// // Add flags for database connection details
	// initCmd.Flags().String("db_host", "localhost", "Database host")
	// initCmd.Flags().Int("db_port", 5432, "Database port")
	// initCmd.Flags().String("db_user", "postgres", "Database user")
	// initCmd.Flags().String("db_pass", "postgres", "Database password")
	// // initCmd.Flags().String("db_name", "draftsmith", "Database name") // TODO
	//
	// // Bind flags to viper
	// viper.BindPFlag("db_host", initCmd.Flags().Lookup("db_host"))
	// viper.BindPFlag("db_port", initCmd.Flags().Lookup("db_port"))
	// viper.BindPFlag("db_user", initCmd.Flags().Lookup("db_user"))
	// viper.BindPFlag("db_pass", initCmd.Flags().Lookup("db_pass"))
	// viper.BindPFlag("db_name", initCmd.Flags().Lookup("db_name"))
}
