/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// cliCmd represents the cli command
var cliCmd = &cobra.Command{
	Use:   "cli",
	Short: "Interact with the Draftsmith API using the CLI",
	Long: `A CLI for interacting with the Draftsmith API. Useful for scripting and automation.

This Can be used to create, read, update, and delete drafts and notes, as well as to manage tags and categories.

Use this as a reference for interacting with the API programmatically.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("cli called")
	},
}

func init() {
	rootCmd.AddCommand(cliCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// cliCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// cliCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
