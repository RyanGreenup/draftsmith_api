package cmd

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var db *sql.DB

// Note represents a simplified note structure
type Note struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serve the REST API for Draftsmith",
	Long:  `This starts a server that serves the REST API for Draftsmith.`,
	Run: func(cmd *cobra.Command, args []string) {
		serve()
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}

func serve() {
	// Get database connection details from viper
	dbHost := viper.GetString("db_host")
	dbPort := viper.GetInt("db_port")
	dbUser := viper.GetString("db_user")
	dbPass := viper.GetString("db_pass")
	dbName := viper.GetString("db_name")
	port := viper.GetInt("port")

	// Create the database connection string
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPass, dbName)

	// Open database connection
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Error opening database connection: %v", err)
	}
	defer db.Close()

	// Test the database connection
	err = db.Ping()
	if err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/titles", getNoteTitles).Methods("GET")

	portStr := fmt.Sprintf(":%d", port)
	fmt.Printf("Server is running on http://localhost%s\n", portStr)
	log.Fatal(http.ListenAndServe(portStr, r))
}

func getNoteTitles(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, title FROM notes")
	if err != nil {
		log.Printf("Error querying database: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var notes []Note
	for rows.Next() {
		var n Note
		if err := rows.Scan(&n.ID, &n.Title); err != nil {
			log.Printf("Error scanning row: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		notes = append(notes, n)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error after scanning rows: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notes)
}
