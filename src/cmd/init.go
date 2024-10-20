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

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the database",
	Long: `This will initialize the database and create the necessary tables.

This requires the database to be dropped first, use the drop command to do this.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Initializing database...")

		// Get database connection details from viper
		dbHost := viper.GetString("db_host")
		dbPort := viper.GetInt("db_port")
		dbUser := viper.GetString("db_user")
		dbPass := viper.GetString("db_pass")
		dbName := viper.GetString("db_name")

		// Create connection string
		connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s sslmode=disable",
			dbHost, dbPort, dbUser, dbPass, dbName)

		// Open database connection
		db, err := sql.Open("postgres", connStr)
		if err != nil {
			log.Fatalf("Error opening database connection: %v", err)
		}
		defer db.Close()

		// Execute SQL commands
		_, err = db.Exec(sql_commands)
		if err != nil {
			log.Fatalf("Error executing SQL commands: %v", err)
		}

		fmt.Println("Database initialized successfully.")
	},
}

func init() {
	cliCmd.AddCommand(initCmd)

	// Add flags for database connection details
	initCmd.Flags().String("db_host", "localhost", "Database host")
	initCmd.Flags().Int("db_port", 5432, "Database port")
	initCmd.Flags().String("db_user", "postgres", "Database user")
	initCmd.Flags().String("db_pass", "postgres", "Database password")
	initCmd.Flags().String("db_name", "draftsmith", "Database name")

	// Bind flags to viper
	viper.BindPFlag("db_host", initCmd.Flags().Lookup("db_host"))
	viper.BindPFlag("db_port", initCmd.Flags().Lookup("db_port"))
	viper.BindPFlag("db_user", initCmd.Flags().Lookup("db_user"))
	viper.BindPFlag("db_pass", initCmd.Flags().Lookup("db_pass"))
	viper.BindPFlag("db_name", initCmd.Flags().Lookup("db_name"))
}
