/*
Copyright © 2024 Ryan Greenup

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "src",
	Short: "Draftsmith API and CLI",
	Long: `Draftsmith is a tool for managing your drafts and notes.

This tool serves the API and provides a simple CLI for interacting with it.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.src.yaml)")
	rootCmd.PersistentFlags().String("token", "secret", "The token to use for authentication")
	rootCmd.PersistentFlags().Int("port", 37238, "The port to run the server on")
	rootCmd.PersistentFlags().Int("db_port", 5432, "The Database Port")
	rootCmd.PersistentFlags().String("db_host", "localhost", "The Database Host")
	rootCmd.PersistentFlags().String("db_user", "postgres", "The Database User")
	rootCmd.PersistentFlags().String("db_pass", "postgres", "The Database Password")
	rootCmd.PersistentFlags().String("db_name", "draftsmith", "The Database Name")

	// Register with viper
	viper.BindPFlag("token", rootCmd.PersistentFlags().Lookup("token"))
	viper.BindPFlag("port", rootCmd.PersistentFlags().Lookup("port"))
	viper.BindPFlag("db_port", rootCmd.PersistentFlags().Lookup("db_port"))
	viper.BindPFlag("db_host", rootCmd.PersistentFlags().Lookup("db_host"))
	viper.BindPFlag("db_user", rootCmd.PersistentFlags().Lookup("db_user"))
	viper.BindPFlag("db_pass", rootCmd.PersistentFlags().Lookup("db_pass"))
	viper.BindPFlag("db_name", rootCmd.PersistentFlags().Lookup("db_name"))

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".src" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".src")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
