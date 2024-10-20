/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// dropCmd represents the drop command
var dropCmd = &cobra.Command{
	Use:   "drop",
	Short: "Drop the database",
	Long: `Drop the database and all of its contents. Use with caution.
This is primarily for development purposes and should not be used in production.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Dropping database...")

		DB_NAME := "draftsmith"

		// Get database connection details from viper
		dbHost := viper.GetString("db_host")
		dbPort := viper.GetInt("db_port")
		dbUser := viper.GetString("db_user")
		dbPass := viper.GetString("db_pass")

		// Connect to the default postgres database
		connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			dbHost, dbPort, dbUser, dbPass, "postgres")
		db, err := sql.Open("postgres", connStr)
		if err != nil {
			log.Fatalf("Error opening database connection: %v", err)
		}
		defer db.Close()

		// Drop the database
		stmt := fmt.Sprintf("DROP DATABASE IF EXISTS %s", DB_NAME)
		_, err = db.Exec(stmt)
		if err != nil {
			log.Fatalf("Error dropping database: %v", err)
		}

		fmt.Println("Database dropped successfully")
	},
}

func init() {
	cliCmd.AddCommand(dropCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// dropCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// dropCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
