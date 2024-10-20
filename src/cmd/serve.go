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
    Content string `json:"content"`
    Created_at string `json:"created_at"`
    Modified_at string `json:"modified_at"`
}

// NoteUpdate represents the structure for updating a note
type NoteUpdate struct {
    Title   string `json:"title"`
    Content string `json:"content"`
}

// NewNote represents the structure for creating a new note
type NewNote struct {
    Title   string `json:"title"`
    Content string `json:"content"`
}

// NewTag represents the structure for creating a new tag
type NewTag struct {
    Name string `json:"name"`
}

// Tag represents a tag structure
type Tag struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}

// AddTagToNote represents the structure for adding a tag to a note
type AddTagToNote struct {
    TagID int `json:"tag_id"`
}

// TagHierarchyEntry represents the structure for adding a tag hierarchy entry
type TagHierarchyEntry struct {
    ParentTagID int `json:"parent_tag_id"`
    ChildTagID  int `json:"child_tag_id"`
}

// Category represents a category structure
type Category struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}

// NewCategory represents the structure for creating a new category
type NewCategory struct {
    Name string `json:"name"`
}

// NoteHierarchyEntry represents the structure for adding a hierarchy entry
type NoteHierarchyEntry struct {
    ParentNoteID  int    `json:"parent_note_id"`
    ChildNoteID   int    `json:"child_note_id"`
    HierarchyType string `json:"hierarchy_type"`
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
	r.HandleFunc("/notes", getNoteTitles).Methods("GET")
	r.HandleFunc("/notes/{id}", updateNote).Methods("PUT")
	r.HandleFunc("/notes", createNote).Methods("POST")
	r.HandleFunc("/tags", createTag).Methods("POST")
	r.HandleFunc("/tags", listTags).Methods("GET")
	r.HandleFunc("/notes/{id}/tags", addTagToNote).Methods("POST")
	r.HandleFunc("/categories", listCategories).Methods("GET")
	r.HandleFunc("/categories", createCategory).Methods("POST")
	r.HandleFunc("/notes/hierarchy", addNoteHierarchyEntry).Methods("POST")
	r.HandleFunc("/tags/hierarchy", addTagHierarchyEntry).Methods("POST")

	portStr := fmt.Sprintf(":%d", port)
	fmt.Printf("Server is running on http://localhost%s\n", portStr)
	log.Fatal(http.ListenAndServe(portStr, r))
}

func getNoteTitles(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, title, content, created_at, modified_at FROM notes")
	if err != nil {
		log.Printf("Error querying database: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var notes []Note
	for rows.Next() {
		var n Note
		if err := rows.Scan(&n.ID, &n.Title, &n.Content, &n.Created_at, &n.Modified_at); err != nil {
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

func updateNote(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id := vars["id"]

    var update NoteUpdate
    err := json.NewDecoder(r.Body).Decode(&update)
    if err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    _, err = db.Exec("UPDATE notes SET title = $1, content = $2, modified_at = CURRENT_TIMESTAMP WHERE id = $3",
        update.Title, update.Content, id)
    if err != nil {
        log.Printf("Error updating note: %v", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"message": "Note updated successfully"})
}

func createNote(w http.ResponseWriter, r *http.Request) {
    var newNote NewNote
    err := json.NewDecoder(r.Body).Decode(&newNote)
    if err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    var noteID int
    err = db.QueryRow("INSERT INTO notes (title, content) VALUES ($1, $2) RETURNING id",
        newNote.Title, newNote.Content).Scan(&noteID)
    if err != nil {
        log.Printf("Error creating note: %v", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "message": "Note created successfully",
        "id":      noteID,
    })
}

func addTagHierarchyEntry(w http.ResponseWriter, r *http.Request) {
    var entry TagHierarchyEntry
    err := json.NewDecoder(r.Body).Decode(&entry)
    if err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    // Insert the new tag hierarchy entry
    var entryID int
    err = db.QueryRow(`
        INSERT INTO tag_hierarchy (parent_tag_id, child_tag_id)
        VALUES ($1, $2)
        RETURNING id
    `, entry.ParentTagID, entry.ChildTagID).Scan(&entryID)

    if err != nil {
        log.Printf("Error adding tag hierarchy entry: %v", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "message": "Tag hierarchy entry added successfully",
        "id":      entryID,
    })
}

func addNoteHierarchyEntry(w http.ResponseWriter, r *http.Request) {
    var entry NoteHierarchyEntry
    err := json.NewDecoder(r.Body).Decode(&entry)
    if err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    // Validate hierarchy_type
    if entry.HierarchyType != "page" && entry.HierarchyType != "block" && entry.HierarchyType != "subpage" {
        http.Error(w, "Invalid hierarchy_type. Must be 'page', 'block', or 'subpage'", http.StatusBadRequest)
        return
    }

    // Insert the new hierarchy entry
    var entryID int
    err = db.QueryRow(`
        INSERT INTO note_hierarchy (parent_note_id, child_note_id, hierarchy_type)
        VALUES ($1, $2, $3)
        RETURNING id
    `, entry.ParentNoteID, entry.ChildNoteID, entry.HierarchyType).Scan(&entryID)

    if err != nil {
        log.Printf("Error adding note hierarchy entry: %v", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "message": "Note hierarchy entry added successfully",
        "id":      entryID,
    })
}

func listTags(w http.ResponseWriter, r *http.Request) {
    rows, err := db.Query("SELECT id, name FROM tags ORDER BY name")
    if err != nil {
        log.Printf("Error querying tags: %v", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var tags []Tag
    for rows.Next() {
        var t Tag
        if err := rows.Scan(&t.ID, &t.Name); err != nil {
            log.Printf("Error scanning tag row: %v", err)
            http.Error(w, "Internal server error", http.StatusInternalServerError)
            return
        }
        tags = append(tags, t)
    }

    if err := rows.Err(); err != nil {
        log.Printf("Error after scanning tag rows: %v", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(tags)
}

func createTag(w http.ResponseWriter, r *http.Request) {
    var newTag NewTag
    err := json.NewDecoder(r.Body).Decode(&newTag)
    if err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    var tagID int
    err = db.QueryRow("INSERT INTO tags (name) VALUES ($1) RETURNING id", newTag.Name).Scan(&tagID)
    if err != nil {
        log.Printf("Error creating tag: %v", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "message": "Tag created successfully",
        "id":      tagID,
    })
}

func addTagToNote(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    noteID := vars["id"]

    var addTag AddTagToNote
    err := json.NewDecoder(r.Body).Decode(&addTag)
    if err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    _, err = db.Exec("INSERT INTO note_tags (note_id, tag_id) VALUES ($1, $2)", noteID, addTag.TagID)
    if err != nil {
        log.Printf("Error adding tag to note: %v", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"message": "Tag added to note successfully"})
}

func listCategories(w http.ResponseWriter, r *http.Request) {
    rows, err := db.Query("SELECT id, name FROM categories ORDER BY name")
    if err != nil {
        log.Printf("Error querying categories: %v", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var categories []Category
    for rows.Next() {
        var c Category
        if err := rows.Scan(&c.ID, &c.Name); err != nil {
            log.Printf("Error scanning category row: %v", err)
            http.Error(w, "Internal server error", http.StatusInternalServerError)
            return
        }
        categories = append(categories, c)
    }

    if err := rows.Err(); err != nil {
        log.Printf("Error after scanning category rows: %v", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(categories)
}

func createCategory(w http.ResponseWriter, r *http.Request) {
    var newCategory NewCategory
    err := json.NewDecoder(r.Body).Decode(&newCategory)
    if err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    var categoryID int
    err = db.QueryRow("INSERT INTO categories (name) VALUES ($1) RETURNING id", newCategory.Name).Scan(&categoryID)
    if err != nil {
        log.Printf("Error creating category: %v", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "message": "Category created successfully",
        "id":      categoryID,
    })
}
