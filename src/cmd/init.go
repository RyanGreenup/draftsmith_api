/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	_ "embed"
	"fmt"
	"log"

	utils "draftsmith/src/utils"
	_ "github.com/lib/pq"
	"github.com/spf13/cobra"
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

        // Create the default Database
        utils.Create_db()

        // Connect to the new database
        db = utils.Get_db(DB_NAME)
        defer db.Close()

		// Execute SQL commands
		fmt.Println(sql_commands)
        _, err := db.Exec(sql_commands)
		if err != nil {
			log.Fatalf("Error executing SQL commands: %v", err)
		}

		fmt.Println("Database initialized successfully.")
	},
}

func init() {
	cliCmd.AddCommand(initCmd)
}
