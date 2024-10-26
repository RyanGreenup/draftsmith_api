package cmd

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var db *sql.DB

// Note represents a simplified note structure
type Note struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Content     string `json:"content"`
	Created_at  string `json:"created_at"`
	Modified_at string `json:"modified_at"`
}

// NoteUpdate represents the structure for updating a note
type NoteUpdate struct {
	Title   *string `json:"title,omitempty"`
	Content *string `json:"content,omitempty"`
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

// NoteInfo represents basic note information
type NoteInfo struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

// TagTree represents a tag and its children in a tree structure
type TagTree struct {
	ID       int        `json:"id"`
	Name     string     `json:"name"`
	Notes    []NoteInfo `json:"notes"`
	Children []*TagTree `json:"children,omitempty"`
}

type FileUpload struct {
    NoteID      int    `json:"note_id"`
    AssetType   string `json:"asset_type"`
    Description string `json:"description"`
}

// NoteTree represents a note and its children in a tree structure
type NoteTree struct {
	ID       int         `json:"id"`
	Title    string      `json:"title"`
	Type     string      `json:"type"`
	Children []*NoteTree `json:"children,omitempty"`
}

// TagWithNotes represents a tag and its associated notes
type TagWithNotes struct {
	ID    int        `json:"tag_id"`
	Name  string     `json:"tag_name"`
	Notes []NoteInfo `json:"notes"`
}

// UpdateTagRequest represents the request body for updating a tag
type UpdateTagRequest struct {
	Name string `json:"name"`
}

// NewTask represents the structure for creating a new task
type NewTask struct {
	NoteID           int     `json:"note_id"`
	Status           string  `json:"status"`
	EffortEstimate   float64 `json:"effort_estimate"`
	ActualEffort     float64 `json:"actual_effort"`
	Deadline         string  `json:"deadline"`
	Priority         int     `json:"priority"`
	AllDay           bool    `json:"all_day"`
	GoalRelationship int     `json:"goal_relationship"`
}

// UpdateTask represents the structure for updating an existing task
type UpdateTask struct {
	Status           *string  `json:"status,omitempty"`
	EffortEstimate   *float64 `json:"effort_estimate,omitempty"`
	ActualEffort     *float64 `json:"actual_effort,omitempty"`
	Deadline         *string  `json:"deadline,omitempty"`
	Priority         *int     `json:"priority,omitempty"`
	AllDay           *bool    `json:"all_day,omitempty"`
	GoalRelationship *int     `json:"goal_relationship,omitempty"`
}

type TaskWithDetails struct {
	ID               int            `json:"id"`
	NoteID           int            `json:"note_id"`
	Status           string         `json:"status"`
	EffortEstimate   float64        `json:"effort_estimate"`
	ActualEffort     float64        `json:"actual_effort"`
	Deadline         string         `json:"deadline"`
	Priority         int            `json:"priority"`
	AllDay           bool           `json:"all_day"`
	GoalRelationship int            `json:"goal_relationship"`
	CreatedAt        string         `json:"created_at"`
	ModifiedAt       string         `json:"modified_at"`
	Schedules        []TaskSchedule `json:"schedules"`
	Clocks           []TaskClock    `json:"clocks"`
}

type TaskSchedule struct {
	ID            int    `json:"id"`
	StartDatetime string `json:"start_datetime"`
	EndDatetime   string `json:"end_datetime"`
}

type TaskClock struct {
	ID       int    `json:"id"`
	ClockIn  string `json:"clock_in"`
	ClockOut string `json:"clock_out"`
}

type NewTaskSchedule struct {
	TaskID        int    `json:"task_id"`
	StartDatetime string `json:"start_datetime"`
	EndDatetime   string `json:"end_datetime"`
}

type NewTaskClock struct {
	TaskID   int    `json:"task_id"`
	ClockIn  string `json:"clock_in"`
	ClockOut string `json:"clock_out,omitempty"` // ClockOut can be optional
}

type UpdateTaskSchedule struct {
	StartDatetime string `json:"start_datetime,omitempty"`
	EndDatetime   string `json:"end_datetime,omitempty"`
}

type UpdateTaskClock struct {
	ClockIn  string `json:"clock_in,omitempty"`
	ClockOut string `json:"clock_out,omitempty"`
}

type TaskTreeNode struct {
	Task     TaskWithDetails `json:"task"`
	Children []*TaskTreeNode `json:"children,omitempty"`
}

func searchNotes(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Query parameter 'q' is required", http.StatusBadRequest)
		return
	}

	rows, err := db.Query(`
        SELECT id, title
        FROM notes
        WHERE to_tsvector('english', title || ' ' || content) @@ plainto_tsquery('english', $1)
        ORDER BY ts_rank(to_tsvector('english', title || ' ' || content), plainto_tsquery('english', $1)) DESC
    `, query)
	if err != nil {
		log.Printf("Error querying database: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var notes []struct {
		ID    int    `json:"id"`
		Title string `json:"title"`
	}

	for rows.Next() {
		var note struct {
			ID    int    `json:"id"`
			Title string `json:"title"`
		}
		if err := rows.Scan(&note.ID, &note.Title); err != nil {
			log.Printf("Error scanning row: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		notes = append(notes, note)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error after scanning rows: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notes)
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
	HierarchyType string `json:"hierarchy_type"`
	ChildNoteID   int    `json:"child_note_id"`
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
	r.HandleFunc("/notes/{id}", deleteNote).Methods("DELETE")
	r.HandleFunc("/tags", createTag).Methods("POST")
	r.HandleFunc("/tags", listTags).Methods("GET")
	r.HandleFunc("/notes/{id}/tags", addTagToNote).Methods("POST")
	r.HandleFunc("/categories", listCategories).Methods("GET")
	r.HandleFunc("/categories", createCategory).Methods("POST")
	r.HandleFunc("/notes/hierarchy", addNoteHierarchyEntry).Methods("POST")
	r.HandleFunc("/tags/hierarchy", addTagHierarchyEntry).Methods("POST")
	r.HandleFunc("/tags/tree", getTagTree).Methods("GET")
	r.HandleFunc("/notes/tree", getNoteTree).Methods("GET")
	r.HandleFunc("/tags/with-notes", getTagsWithNotes).Methods("GET")
	r.HandleFunc("/notes/no-content", getNoteTitlesAndIDs).Methods("GET")
	r.HandleFunc("/notes/search", searchNotes).Methods("GET")
	r.HandleFunc("/tags/{id}", updateTag).Methods("PUT")
	r.HandleFunc("/tags/{id}", deleteTag).Methods("DELETE")
	r.HandleFunc("/tags/hierarchy/{childId}", deleteTagHierarchyEntry).Methods("DELETE")
	r.HandleFunc("/notes/hierarchy/{childId}", deleteNoteHierarchyEntry).Methods("DELETE")
	r.HandleFunc("/notes/hierarchy/{childId}", updateNoteHierarchyEntry).Methods("PUT")
	r.HandleFunc("/tags/hierarchy/{childId}", updateTagHierarchyEntry).Methods("PUT")
	r.HandleFunc("/tasks", createTask).Methods("POST")
	r.HandleFunc("/tasks/{id}", updateTask).Methods("PUT")
	r.HandleFunc("/tasks/{id}", deleteTask).Methods("DELETE")
	r.HandleFunc("/tasks/details", getTasksWithDetails).Methods("GET")
	r.HandleFunc("/task_schedules", createTaskSchedule).Methods("POST")
	r.HandleFunc("/task_clocks", createTaskClock).Methods("POST")
	r.HandleFunc("/task_schedules/{id}", updateTaskSchedule).Methods("PUT")
	r.HandleFunc("/task_schedules/{id}", deleteTaskSchedule).Methods("DELETE")
	r.HandleFunc("/task_clocks/{id}", deleteTaskClock).Methods("DELETE")
	r.HandleFunc("/task_clocks/{id}", updateTaskClock).Methods("PUT")
	r.HandleFunc("/tasks/tree", getTasksWithDetailsAsTree).Methods("GET")
	r.HandleFunc("/upload", uploadFile).Methods("POST")
	r.HandleFunc("/assets/{id}", deleteFile).Methods("DELETE")

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

	// Start building the SQL query
	query := "UPDATE notes SET modified_at = CURRENT_TIMESTAMP"
	var args []interface{}
	var argIndex int = 1

	if update.Title != nil {
		query += fmt.Sprintf(", title = $%d", argIndex)
		args = append(args, *update.Title)
		argIndex++
	}

	if update.Content != nil {
		query += fmt.Sprintf(", content = $%d", argIndex)
		args = append(args, *update.Content)
		argIndex++
	}

	query += fmt.Sprintf(" WHERE id = $%d", argIndex)
	args = append(args, id)

	// Execute the query
	result, err := db.Exec(query, args...)
	if err != nil {
		log.Printf("Error updating note: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "Note not found", http.StatusNotFound)
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

func getTagTree(w http.ResponseWriter, r *http.Request) {
	// First, get all tags
	rows, err := db.Query("SELECT id, name FROM tags")
	if err != nil {
		log.Printf("Error querying tags: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	tagMap := make(map[int]*TagTree)
	for rows.Next() {
		var id int
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			log.Printf("Error scanning tag row: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		tagMap[id] = &TagTree{ID: id, Name: name}
	}

	// Get the tag hierarchy
	rows, err = db.Query("SELECT parent_tag_id, child_tag_id FROM tag_hierarchy")
	if err != nil {
		log.Printf("Error querying tag hierarchy: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var parentID, childID int
		if err := rows.Scan(&parentID, &childID); err != nil {
			log.Printf("Error scanning tag hierarchy row: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		parent := tagMap[parentID]
		child := tagMap[childID]
		parent.Children = append(parent.Children, child)
	}

	// Get notes for each tag
	rows, err = db.Query(`
        SELECT nt.tag_id, n.id, n.title
        FROM note_tags nt
        JOIN notes n ON nt.note_id = n.id
    `)
	if err != nil {
		log.Printf("Error querying notes for tags: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var tagID, noteID int
		var noteTitle string
		if err := rows.Scan(&tagID, &noteID, &noteTitle); err != nil {
			log.Printf("Error scanning note info row: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		if tag, ok := tagMap[tagID]; ok {
			tag.Notes = append(tag.Notes, NoteInfo{ID: noteID, Title: noteTitle})
		}
	}

	// Find root tags (tags without parents)
	var rootTags []*TagTree
	for _, tag := range tagMap {
		isChild := false
		for _, potentialParent := range tagMap {
			for _, child := range potentialParent.Children {
				if child.ID == tag.ID {
					isChild = true
					break
				}
			}
			if isChild {
				break
			}
		}
		if !isChild {
			rootTags = append(rootTags, tag)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rootTags)
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

	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Fetch all existing hierarchies
	rows, err := tx.Query("SELECT parent_note_id, child_note_id FROM note_hierarchy")
	if err != nil {
		log.Printf("Error fetching note hierarchies: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var parents, children []int
	for rows.Next() {
		var parent, child int
		if err := rows.Scan(&parent, &child); err != nil {
			log.Printf("Error scanning row: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		parents = append(parents, parent)
		children = append(children, child)
	}

	// Add the new relationship to check
	parents = append(parents, entry.ParentNoteID)
	children = append(children, entry.ChildNoteID)

	// Check for cycles
	if detectCycle(parents, children) {
		http.Error(w, "Operation would create a cycle in the hierarchy", http.StatusBadRequest)
		return
	}

	// Insert the new hierarchy entry
	var entryID int
	err = tx.QueryRow(`
        INSERT INTO note_hierarchy (parent_note_id, child_note_id, hierarchy_type)
        VALUES ($1, $2, $3)
        RETURNING id
    `, entry.ParentNoteID, entry.ChildNoteID, entry.HierarchyType).Scan(&entryID)

	if err != nil {
		log.Printf("Error adding note hierarchy entry: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
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
    noteID, err := strconv.Atoi(vars["id"])
    if err != nil {
        log.Printf("Invalid note ID: %v", err)
        http.Error(w, "Invalid note ID", http.StatusBadRequest)
        return
    }

    var addTag AddTagToNote
    err = json.NewDecoder(r.Body).Decode(&addTag)
    if err != nil {
        log.Printf("Invalid request body: %v", err)
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    log.Printf("Adding tag %d to note %d", addTag.TagID, noteID)

    // Check if the note exists
    var noteExists bool
    err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM notes WHERE id = $1)", noteID).Scan(&noteExists)
    if err != nil {
        log.Printf("Error checking note existence: %v", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }
    if !noteExists {
        log.Printf("Note %d not found", noteID)
        http.Error(w, "Note not found", http.StatusNotFound)
        return
    }

    // Check if the tag exists
    var tagExists bool
    err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM tags WHERE id = $1)", addTag.TagID).Scan(&tagExists)
    if err != nil {
        log.Printf("Error checking tag existence: %v", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }
    if !tagExists {
        log.Printf("Tag %d not found", addTag.TagID)
        http.Error(w, "Tag not found", http.StatusNotFound)
        return
    }

    // Insert the new relationship
    _, err = db.Exec("INSERT INTO note_tags (note_id, tag_id) VALUES ($1, $2)", noteID, addTag.TagID)
    if err != nil {
        log.Printf("Error adding tag to note: %v", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    log.Printf("Successfully added tag %d to note %d", addTag.TagID, noteID)
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
func getNoteTree(w http.ResponseWriter, r *http.Request) {
	rootNotes, err := buildNoteTree(db)
	if err != nil {
		log.Printf("Error building note tree: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rootNotes)
}

func getTagsWithNotes(w http.ResponseWriter, r *http.Request) {
	// Query to get tags and their associated notes
	rows, err := db.Query(`
        SELECT t.id, t.name, n.id, n.title
        FROM tags t
        LEFT JOIN note_tags nt ON t.id = nt.tag_id
        LEFT JOIN notes n ON nt.note_id = n.id
        ORDER BY t.name, n.title
    `)
	if err != nil {
		log.Printf("Error querying tags and notes: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	tagsMap := make(map[int]*TagWithNotes)
	for rows.Next() {
		var tagID int
		var tagName string
		var noteID sql.NullInt64
		var noteTitle sql.NullString

		if err := rows.Scan(&tagID, &tagName, &noteID, &noteTitle); err != nil {
			log.Printf("Error scanning row: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		tag, exists := tagsMap[tagID]
		if !exists {
			tag = &TagWithNotes{ID: tagID, Name: tagName}
			tagsMap[tagID] = tag
		}

		if noteID.Valid {
			tag.Notes = append(tag.Notes, NoteInfo{ID: int(noteID.Int64), Title: noteTitle.String})
		}
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error after scanning rows: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Convert map to slice
	var tagsWithNotes []TagWithNotes
	for _, tag := range tagsMap {
		tagsWithNotes = append(tagsWithNotes, *tag)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tagsWithNotes)
}

func getNoteTitlesAndIDs(w http.ResponseWriter, r *http.Request) {
	// I couldn't get arguments to work so I'm just going to hardcode the route.

	type Note struct {
		ID         int    `json:"id"`
		Title      string `json:"title"`
		CreatedAt  string `json:"created_at"`
		ModifiedAt string `json:"modified_at"`
	}

	rows, err := db.Query("SELECT id, title, created_at, modified_at FROM notes ORDER BY id")
	if err != nil {
		log.Printf("Error querying database: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var notes []Note

	for rows.Next() {
		var note Note
		if err := rows.Scan(&note.ID, &note.Title, &note.CreatedAt, &note.ModifiedAt); err != nil {
			log.Printf("Error scanning row: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		notes = append(notes, note)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Row iteration error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(notes); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func updateTag(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tagID := vars["id"]

	var req UpdateTagRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Tag name cannot be empty", http.StatusBadRequest)
		return
	}

	result, err := db.Exec("UPDATE tags SET name = $1 WHERE id = $2", req.Name, tagID)
	if err != nil {
		log.Printf("Error updating tag: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "Tag not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Tag updated successfully"})
}

func deleteTagHierarchyEntry(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	childTagID := vars["childId"]

	// Execute the delete query
	result, err := db.Exec("DELETE FROM tag_hierarchy WHERE child_tag_id = $1", childTagID)
	if err != nil {
		log.Printf("Error deleting tag hierarchy entry: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "Tag hierarchy entry not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Tag hierarchy entry deleted successfully"})
}

func deleteNoteHierarchyEntry(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	childNoteID := vars["childId"]

	// Execute the delete query
	result, err := db.Exec("DELETE FROM note_hierarchy WHERE child_note_id = $1", childNoteID)
	if err != nil {
		log.Printf("Error deleting note hierarchy entry: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "Note hierarchy entry not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Note hierarchy entry deleted successfully"})
}

func updateNoteHierarchyEntry(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	childNoteID := vars["childId"]

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

	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Fetch all existing hierarchies
	rows, err := tx.Query("SELECT parent_note_id, child_note_id FROM note_hierarchy")
	if err != nil {
		log.Printf("Error fetching note hierarchies: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var parents, children []int
	for rows.Next() {
		var parent, child int
		if err := rows.Scan(&parent, &child); err != nil {
			log.Printf("Error scanning row: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		parents = append(parents, parent)
		children = append(children, child)
	}

	// Add the new relationship to check
	childID, err := strconv.Atoi(childNoteID)
	if err != nil {
		http.Error(w, "Invalid child note ID", http.StatusBadRequest)
		return
	}
	parents = append(parents, entry.ParentNoteID)
	children = append(children, childID)

	// Check for cycles
	if detectCycle(parents, children) {
		http.Error(w, "Operation would create a cycle in the hierarchy", http.StatusBadRequest)
		return
	}

	// Update the existing entry
	result, err := tx.Exec(`
        UPDATE note_hierarchy
        SET parent_note_id = $1, hierarchy_type = $2
        WHERE child_note_id = $3
    `, entry.ParentNoteID, entry.HierarchyType, childNoteID)

	if err != nil {
		log.Printf("Error updating note hierarchy entry: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "Note hierarchy entry not found", http.StatusNotFound)
		return
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Note hierarchy entry updated successfully"})
}

func updateTagHierarchyEntry(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	childTagID := vars["childId"]

	var entry struct {
		ParentTagID int `json:"parent_tag_id"`
	}
	err := json.NewDecoder(r.Body).Decode(&entry)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Fetch all existing hierarchies
	rows, err := tx.Query("SELECT parent_tag_id, child_tag_id FROM tag_hierarchy")
	if err != nil {
		log.Printf("Error fetching tag hierarchies: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var parents, children []int
	for rows.Next() {
		var parent, child int
		if err := rows.Scan(&parent, &child); err != nil {
			log.Printf("Error scanning row: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		parents = append(parents, parent)
		children = append(children, child)
	}

	// Add the new relationship to check
	childID, err := strconv.Atoi(childTagID)
	if err != nil {
		http.Error(w, "Invalid child tag ID", http.StatusBadRequest)
		return
	}
	parents = append(parents, entry.ParentTagID)
	children = append(children, childID)

	// Check for cycles
	if detectCycle(parents, children) {
		http.Error(w, "Operation would create a cycle in the hierarchy", http.StatusBadRequest)
		return
	}

	// Update the existing entry
	result, err := tx.Exec(`
        UPDATE tag_hierarchy
        SET parent_tag_id = $1
        WHERE child_tag_id = $2
    `, entry.ParentTagID, childTagID)

	if err != nil {
		log.Printf("Error updating tag hierarchy entry: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "Tag hierarchy entry not found", http.StatusNotFound)
		return
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Tag hierarchy entry updated successfully"})
}

func createTask(w http.ResponseWriter, r *http.Request) {
	var newTask NewTask
	err := json.NewDecoder(r.Body).Decode(&newTask)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate the status
	validStatuses := []string{"todo", "done", "wait", "hold", "idea", "kill", "proj", "event"}
	if !contains(validStatuses, newTask.Status) {
		http.Error(w, "Invalid status", http.StatusBadRequest)
		return
	}

	// Validate the priority
	if newTask.Priority < 1 || newTask.Priority > 5 {
		http.Error(w, "Priority must be between 1 and 5", http.StatusBadRequest)
		return
	}

	// Validate the goal relationship
	if newTask.GoalRelationship < 1 || newTask.GoalRelationship > 5 {
		http.Error(w, "Goal relationship must be between 1 and 5", http.StatusBadRequest)
		return
	}

	var taskID int
	err = db.QueryRow(`
        INSERT INTO tasks (note_id, status, effort_estimate, actual_effort, deadline, priority, all_day, goal_relationship)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        RETURNING id
    `, newTask.NoteID, newTask.Status, newTask.EffortEstimate, newTask.ActualEffort, newTask.Deadline, newTask.Priority, newTask.AllDay, newTask.GoalRelationship).Scan(&taskID)

	if err != nil {
		log.Printf("Error creating task: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Task created successfully",
		"id":      taskID,
	})
}

func updateTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["id"]

	var update UpdateTask
	err := json.NewDecoder(r.Body).Decode(&update)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Start building the SQL query
	query := "UPDATE tasks SET modified_at = CURRENT_TIMESTAMP"
	var args []interface{}
	var argIndex int = 1

	// Validate and add fields to update
	if update.Status != nil {
		validStatuses := []string{"todo", "done", "wait", "hold", "idea", "kill", "proj", "event"}
		if !contains(validStatuses, *update.Status) {
			http.Error(w, "Invalid status", http.StatusBadRequest)
			return
		}
		query += fmt.Sprintf(", status = $%d", argIndex)
		args = append(args, *update.Status)
		argIndex++
	}

	if update.EffortEstimate != nil {
		query += fmt.Sprintf(", effort_estimate = $%d", argIndex)
		args = append(args, *update.EffortEstimate)
		argIndex++
	}

	if update.ActualEffort != nil {
		query += fmt.Sprintf(", actual_effort = $%d", argIndex)
		args = append(args, *update.ActualEffort)
		argIndex++
	}

	if update.Deadline != nil {
		query += fmt.Sprintf(", deadline = $%d", argIndex)
		args = append(args, *update.Deadline)
		argIndex++
	}

	if update.Priority != nil {
		if *update.Priority < 1 || *update.Priority > 5 {
			http.Error(w, "Priority must be between 1 and 5", http.StatusBadRequest)
			return
		}
		query += fmt.Sprintf(", priority = $%d", argIndex)
		args = append(args, *update.Priority)
		argIndex++
	}

	if update.AllDay != nil {
		query += fmt.Sprintf(", all_day = $%d", argIndex)
		args = append(args, *update.AllDay)
		argIndex++
	}

	if update.GoalRelationship != nil {
		if *update.GoalRelationship < 1 || *update.GoalRelationship > 5 {
			http.Error(w, "Goal relationship must be between 1 and 5", http.StatusBadRequest)
			return
		}
		query += fmt.Sprintf(", goal_relationship = $%d", argIndex)
		args = append(args, *update.GoalRelationship)
		argIndex++
	}

	query += fmt.Sprintf(" WHERE id = $%d", argIndex)
	args = append(args, taskID)

	// Execute the query
	result, err := db.Exec(query, args...)
	if err != nil {
		log.Printf("Error updating task: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Task updated successfully"})
}

func deleteTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["id"]

	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Delete the task
	result, err := tx.Exec("DELETE FROM tasks WHERE id = $1", taskID)
	if err != nil {
		log.Printf("Error deleting task: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Task deleted successfully"})
}

func getTasksWithDetails(w http.ResponseWriter, r *http.Request) {
    tasks, err := buildTasksWithDetails(db)

	if err != nil {
		log.Printf("Error building task list: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

func createTaskSchedule(w http.ResponseWriter, r *http.Request) {
	var newSchedule NewTaskSchedule
	err := json.NewDecoder(r.Body).Decode(&newSchedule)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate the input
	if newSchedule.TaskID == 0 {
		http.Error(w, "Task ID is required", http.StatusBadRequest)
		return
	}
	if newSchedule.StartDatetime == "" || newSchedule.EndDatetime == "" {
		http.Error(w, "Start and end datetimes are required", http.StatusBadRequest)
		return
	}

	// Insert the new schedule
	var scheduleID int
	err = db.QueryRow(`
        INSERT INTO task_schedules (task_id, start_datetime, end_datetime)
        VALUES ($1, $2, $3)
        RETURNING id
    `, newSchedule.TaskID, newSchedule.StartDatetime, newSchedule.EndDatetime).Scan(&scheduleID)

	if err != nil {
		log.Printf("Error creating task schedule: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Task schedule created successfully",
		"id":      scheduleID,
	})
}

func createTaskClock(w http.ResponseWriter, r *http.Request) {
	var newClock NewTaskClock
	var err error
	err = json.NewDecoder(r.Body).Decode(&newClock)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate the input
	if newClock.TaskID == 0 {
		http.Error(w, "Task ID is required", http.StatusBadRequest)
		return
	}
	if newClock.ClockIn == "" {
		http.Error(w, "Clock in time is required", http.StatusBadRequest)
		return
	}

	// Insert the new clock entry
	var clockID int
	if newClock.ClockOut == "" {
		err = db.QueryRow(`
            INSERT INTO task_clocks (task_id, clock_in)
            VALUES ($1, $2)
            RETURNING id
        `, newClock.TaskID, newClock.ClockIn).Scan(&clockID)
	} else {
		err = db.QueryRow(`
            INSERT INTO task_clocks (task_id, clock_in, clock_out)
            VALUES ($1, $2, $3)
            RETURNING id
        `, newClock.TaskID, newClock.ClockIn, newClock.ClockOut).Scan(&clockID)
	}

	if err != nil {
		log.Printf("Error creating task clock entry: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Task clock entry created successfully",
		"id":      clockID,
	})
}

func updateTaskSchedule(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	scheduleID := vars["id"]

	var update UpdateTaskSchedule
	err := json.NewDecoder(r.Body).Decode(&update)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Start building the SQL query
	query := "UPDATE task_schedules SET"
	var args []interface{}
	var setFields []string

	if update.StartDatetime != "" {
		args = append(args, update.StartDatetime)
		setFields = append(setFields, fmt.Sprintf("start_datetime = $%d", len(args)))
	}

	if update.EndDatetime != "" {
		args = append(args, update.EndDatetime)
		setFields = append(setFields, fmt.Sprintf("end_datetime = $%d", len(args)))
	}

	if len(setFields) == 0 {
		http.Error(w, "No fields to update", http.StatusBadRequest)
		return
	}

	query += " " + strings.Join(setFields, ", ")
	query += fmt.Sprintf(" WHERE id = $%d", len(args)+1)
	args = append(args, scheduleID)

	// Execute the query
	result, err := db.Exec(query, args...)
	if err != nil {
		log.Printf("Error updating task schedule: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "Task schedule not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Task schedule updated successfully"})
}

func deleteTaskSchedule(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	scheduleID := vars["id"]

	// Execute the delete query
	result, err := db.Exec("DELETE FROM task_schedules WHERE id = $1", scheduleID)
	if err != nil {
		log.Printf("Error deleting task schedule: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "Task schedule not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Task schedule deleted successfully"})
}

func deleteTaskClock(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clockID := vars["id"]

	// Execute the delete query
	result, err := db.Exec("DELETE FROM task_clocks WHERE id = $1", clockID)
	if err != nil {
		log.Printf("Error deleting task clock entry: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "Task clock entry not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Task clock entry deleted successfully"})
}

func updateTaskClock(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clockID := vars["id"]

	var update UpdateTaskClock
	err := json.NewDecoder(r.Body).Decode(&update)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Start building the SQL query
	query := "UPDATE task_clocks SET"
	var args []interface{}
	var setFields []string

	if update.ClockIn != "" {
		args = append(args, update.ClockIn)
		setFields = append(setFields, fmt.Sprintf("clock_in = $%d", len(args)))
	}

	if update.ClockOut != "" {
		args = append(args, update.ClockOut)
		setFields = append(setFields, fmt.Sprintf("clock_out = $%d", len(args)))
	}

	if len(setFields) == 0 {
		http.Error(w, "No fields to update", http.StatusBadRequest)
		return
	}

	query += " " + strings.Join(setFields, ", ")
	query += fmt.Sprintf(" WHERE id = $%d", len(args)+1)
	args = append(args, clockID)

	// Execute the query
	result, err := db.Exec(query, args...)
	if err != nil {
		log.Printf("Error updating task clock: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "Task clock entry not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Task clock entry updated successfully"})
}

func getTasksWithDetailsAsTree(w http.ResponseWriter, r *http.Request) {
	// Get the flat list of tasks
	tasks, err := buildTasksWithDetails(db)
	if err != nil {
		log.Printf("Error building task list: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Get the note tree
	noteHierarchy, err := buildNoteTree(db)
	if err != nil {
		log.Printf("Error building note tree: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Create a set of NoteIDs that are referenced in tasks
	taskNoteIDs := make(map[int]bool)
	for _, task := range tasks {
		taskNoteIDs[task.NoteID] = true
	}

	// Filter the note hierarchy
	filteredNotes := filterNotesForTasks(noteHierarchy, taskNoteIDs)

	// Respond with the filtered note tree
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(filteredNotes); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// filterNotesForTasks filters the note tree to include only notes with IDs present in the taskNoteIDs set
func filterNotesForTasks(notes []*NoteTree, taskNoteIDs map[int]bool) []*NoteTree {
	var filtered []*NoteTree
	for _, note := range notes {
		if taskNoteIDs[note.ID] {
			// Recursively filter children
			note.Children = filterNotesForTasks(note.Children, taskNoteIDs)
			filtered = append(filtered, note)
		} else {
			// Recursively check children and promote valid sub-trees
			children := filterNotesForTasks(note.Children, taskNoteIDs)
			// If children are retained, keep this node to maintain hierarchy
			if len(children) > 0 {
				note.Children = children
				filtered = append(filtered, note)
			}
		}
	}
	return filtered
}



func isDescendant(root *TaskTreeNode, node *TaskTreeNode) bool {
	if root == node {
		return true
	}
	for _, child := range root.Children {
		if isDescendant(child, node) {
			return true
		}
	}
	return false
}

// Helper function to check if a string is in a slice
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func deleteTag(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tagID := vars["id"]

	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Delete related entries in note_tags
	_, err = tx.Exec("DELETE FROM note_tags WHERE tag_id = $1", tagID)
	if err != nil {
		log.Printf("Error deleting from note_tags: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Delete related entries in tag_hierarchy
	_, err = tx.Exec("DELETE FROM tag_hierarchy WHERE parent_tag_id = $1 OR child_tag_id = $1", tagID)
	if err != nil {
		log.Printf("Error deleting from tag_hierarchy: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Delete the tag
	result, err := tx.Exec("DELETE FROM tags WHERE id = $1", tagID)
	if err != nil {
		log.Printf("Error deleting tag: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "Tag not found", http.StatusNotFound)
		return
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Tag deleted successfully"})
}

func deleteNote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	noteID := vars["id"]

	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Delete related entries in note_tags
	_, err = tx.Exec("DELETE FROM note_tags WHERE note_id = $1", noteID)
	if err != nil {
		log.Printf("Error deleting from note_tags: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Delete related entries in note_categories
	_, err = tx.Exec("DELETE FROM note_categories WHERE note_id = $1", noteID)
	if err != nil {
		log.Printf("Error deleting from note_categories: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Delete related entries in note_hierarchy
	_, err = tx.Exec("DELETE FROM note_hierarchy WHERE parent_note_id = $1 OR child_note_id = $1", noteID)
	if err != nil {
		log.Printf("Error deleting from note_hierarchy: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Delete the note
	result, err := tx.Exec("DELETE FROM notes WHERE id = $1", noteID)
	if err != nil {
		log.Printf("Error deleting note: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "Note not found", http.StatusNotFound)
		return
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Note deleted successfully"})
}


func buildTasksWithDetails(db *sql.DB) ([]*TaskWithDetails, error) {
	// Query to get tasks with their schedules and clocks
	query := `
        SELECT
            t.id, t.note_id, t.status, t.effort_estimate, t.actual_effort,
            t.deadline, t.priority, t.all_day, t.goal_relationship,
            t.created_at, t.modified_at,
            ts.id, ts.start_datetime, ts.end_datetime,
            tc.id, tc.clock_in, tc.clock_out
        FROM tasks t
        LEFT JOIN task_schedules ts ON t.id = ts.task_id
        LEFT JOIN task_clocks tc ON t.id = tc.task_id
        ORDER BY t.id, ts.id, tc.id
    `
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error querying tasks: %w", err)
	}
	defer rows.Close()

	tasksMap := make(map[int]*TaskWithDetails)

	for rows.Next() {
		var task TaskWithDetails
		var schedule TaskSchedule
		var clock TaskClock
		var scheduleID, clockID sql.NullInt64
		var startDatetime, endDatetime, clockIn, clockOut sql.NullString

		err := rows.Scan(
			&task.ID, &task.NoteID, &task.Status, &task.EffortEstimate, &task.ActualEffort,
			&task.Deadline, &task.Priority, &task.AllDay, &task.GoalRelationship,
			&task.CreatedAt, &task.ModifiedAt,
			&scheduleID, &startDatetime, &endDatetime,
			&clockID, &clockIn, &clockOut,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}

		if existingTask, ok := tasksMap[task.ID]; ok {
			task = *existingTask
		} else {
			tasksMap[task.ID] = &task
		}

		if scheduleID.Valid {
			schedule.ID = int(scheduleID.Int64)
			schedule.StartDatetime = startDatetime.String
			schedule.EndDatetime = endDatetime.String
			task.Schedules = append(task.Schedules, schedule)
		}

		if clockID.Valid {
			clock.ID = int(clockID.Int64)
			clock.ClockIn = clockIn.String
			clock.ClockOut = clockOut.String
			task.Clocks = append(task.Clocks, clock)
		}

		tasksMap[task.ID] = &task
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error after scanning rows: %w", err)
	}

	tasks := make([]*TaskWithDetails, 0, len(tasksMap))
	for _, task := range tasksMap {
		tasks = append(tasks, task)
	}

	return tasks, nil
}



func buildNoteTree(db *sql.DB) ([]*NoteTree, error) {

	// Query all notes
	rows, err := db.Query("SELECT id, title FROM notes")
	if err != nil {
		return nil, fmt.Errorf("error querying notes: %w", err)
	}
	defer rows.Close()

	noteMap := make(map[int]*NoteTree)
	for rows.Next() {
		var id int
		var title string
		if err := rows.Scan(&id, &title); err != nil {
			return nil, fmt.Errorf("error scanning note row: %w", err)
		}
		noteMap[id] = &NoteTree{ID: id, Title: title}
	}

	// Query the note hierarchy
	rows, err = db.Query("SELECT parent_note_id, child_note_id, hierarchy_type FROM note_hierarchy")
	if err != nil {
		return nil, fmt.Errorf("error querying note hierarchy: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var parentID, childID int
		var hierarchyType string
		if err := rows.Scan(&parentID, &childID, &hierarchyType); err != nil {
			return nil, fmt.Errorf("error scanning note hierarchy row: %w", err)
		}
		parent := noteMap[parentID]
		child := noteMap[childID]
		if child != nil {
			child.Type = hierarchyType
		}
		if parent != nil {
			parent.Children = append(parent.Children, child)
		}
	}

	// Find root notes (notes without parents)
	var rootNotes []*NoteTree
	for _, note := range noteMap {
		isChild := false
		for _, potentialParent := range noteMap {
			for _, child := range potentialParent.Children {
				if child.ID == note.ID {
					isChild = true
					break
				}
			}
			if isChild {
				break
			}
		}
		if !isChild {
			rootNotes = append(rootNotes, note)
		}
	}

	return rootNotes, nil

}

func detectCycle(parents, children []int) bool {
	graph := make(map[int][]int)
	for i := range parents {
		graph[parents[i]] = append(graph[parents[i]], children[i])
	}

	visited := make(map[int]bool)
	recStack := make(map[int]bool)

	var dfs func(node int) bool
	dfs = func(node int) bool {
		visited[node] = true
		recStack[node] = true

		for _, neighbor := range graph[node] {
			if !visited[neighbor] {
				if dfs(neighbor) {
					return true
				}
			} else if recStack[neighbor] {
				return true
			}
		}

		recStack[node] = false
		return false
	}

	for node := range graph {
		if !visited[node] {
			if dfs(node) {
				return true
			}
		}
	}

	return false
}

func addTagHierarchyEntry(w http.ResponseWriter, r *http.Request) {
	var entry struct {
		ParentTagID int `json:"parent_tag_id"`
		ChildTagID  int `json:"child_tag_id"`
	}
	err := json.NewDecoder(r.Body).Decode(&entry)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Fetch all existing hierarchies
	rows, err := tx.Query("SELECT parent_tag_id, child_tag_id FROM tag_hierarchy")
	if err != nil {
		log.Printf("Error fetching tag hierarchies: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var parents, children []int
	for rows.Next() {
		var parent, child int
		if err := rows.Scan(&parent, &child); err != nil {
			log.Printf("Error scanning row: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		parents = append(parents, parent)
		children = append(children, child)
	}

	// Add the new relationship to check
	parents = append(parents, entry.ParentTagID)
	children = append(children, entry.ChildTagID)

	// Check for cycles
	if detectCycle(parents, children) {
		http.Error(w, "Operation would create a cycle in the hierarchy", http.StatusBadRequest)
		return
	}

	// Insert the new entry
	_, err = tx.Exec(`
        INSERT INTO tag_hierarchy (parent_tag_id, child_tag_id)
        VALUES ($1, $2)
    `, entry.ParentTagID, entry.ChildTagID)

	if err != nil {
		log.Printf("Error inserting tag hierarchy entry: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Tag hierarchy entry added successfully"})
}

func uploadFile(w http.ResponseWriter, r *http.Request) {
    // Parse the multipart form
    err := r.ParseMultipartForm(10 << 29) // 5 GB max (Bitshifting 10*2**29)
    if err != nil {
        http.Error(w, "Unable to parse form", http.StatusBadRequest)
        return
    }

    // Get the file from the form
    file, header, err := r.FormFile("file")
    if err != nil {
        http.Error(w, "Error retrieving file", http.StatusBadRequest)
        return
    }
    defer file.Close()

    // Create the uploads directory if it doesn't exist
    uploadsDir := "uploads"
    if err := os.MkdirAll(uploadsDir, os.ModePerm); err != nil {
        http.Error(w, "Unable to create upload directory", http.StatusInternalServerError)
        return
    }

    // Generate a unique filename
    filename := header.Filename
    extension := filepath.Ext(filename)
    nameWithoutExt := filename[:len(filename)-len(extension)]
    counter := 1
    for {
        if _, err := os.Stat(filepath.Join(uploadsDir, filename)); os.IsNotExist(err) {
            break
        }
        filename = fmt.Sprintf("%s_%d%s", nameWithoutExt, counter, extension)
        counter++
    }

    // Create a new file in the uploads directory
    dst, err := os.Create(filepath.Join(uploadsDir, filename))
    if err != nil {
        http.Error(w, "Error creating destination file", http.StatusInternalServerError)
        return
    }
    defer dst.Close()

    // Copy the uploaded file to the destination file
    if _, err := io.Copy(dst, file); err != nil {
        http.Error(w, "Error copying file", http.StatusInternalServerError)
        return
    }

    // Get other form values
    assetType := r.FormValue("asset_type")
    description := r.FormValue("description")

    // Insert the file information into the database and get the generated ID
    var id int
    err = db.QueryRow(`
        INSERT INTO assets (asset_type, location, description)
        VALUES ($1, $2, $3)
        RETURNING id
    `, assetType, filepath.Join(uploadsDir, filename), description).Scan(&id)

    if err != nil {
        http.Error(w, "Error saving to database", http.StatusInternalServerError)
        return
    }

    // Respond with the created status and return the generated ID and filename
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "message": "File uploaded successfully",
        "filename": filename,
        "id": id,
    })
}


func deleteFile(w http.ResponseWriter, r *http.Request) {
    // Get the asset ID from the URL parameters
    vars := mux.Vars(r)
    assetID := vars["id"]

    // Start a transaction
    tx, err := db.Begin()
    if err != nil {
        log.Printf("Error starting transaction: %v", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }
    defer tx.Rollback()

    // Get the file location from the database
    var fileLocation string
    err = tx.QueryRow("SELECT location FROM assets WHERE id = $1", assetID).Scan(&fileLocation)
    if err != nil {
        if err == sql.ErrNoRows {
            http.Error(w, "Asset not found", http.StatusNotFound)
        } else {
            log.Printf("Error querying asset: %v", err)
            http.Error(w, "Internal server error", http.StatusInternalServerError)
        }
        return
    }

    // Delete the file from the filesystem
    err = os.Remove(fileLocation)
    if err != nil {
        log.Printf("Error deleting file: %v", err)
        http.Error(w, "Error deleting file", http.StatusInternalServerError)
        return
    }

    // Delete the asset record from the database
    _, err = tx.Exec("DELETE FROM assets WHERE id = $1", assetID)
    if err != nil {
        log.Printf("Error deleting asset record: %v", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    // Commit the transaction
    if err := tx.Commit(); err != nil {
        log.Printf("Error committing transaction: %v", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"message": "File deleted successfully"})
}
