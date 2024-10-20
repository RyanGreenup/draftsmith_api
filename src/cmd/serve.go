package cmd

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

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
	r.HandleFunc("/tags/hierarchy/{childId}", deleteTagHierarchyEntry).Methods("DELETE")
	r.HandleFunc("/notes/hierarchy/{childId}", deleteNoteHierarchyEntry).Methods("DELETE")
	r.HandleFunc("/notes/hierarchy/{childId}", updateNoteHierarchyEntry).Methods("PUT")
	r.HandleFunc("/tags/hierarchy/{childId}", updateTagHierarchyEntry).Methods("PUT")

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
    // TODO is this necessary if the DB will handle it?
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
func getNoteTree(w http.ResponseWriter, r *http.Request) {
	// First, get all notes
	rows, err := db.Query("SELECT id, title FROM notes")
	if err != nil {
		log.Printf("Error querying notes: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	noteMap := make(map[int]*NoteTree)
	for rows.Next() {
		var id int
		var title string
		if err := rows.Scan(&id, &title); err != nil {
			log.Printf("Error scanning note row: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		noteMap[id] = &NoteTree{ID: id, Title: title}
	}

	// Get the note hierarchy
	rows, err = db.Query("SELECT parent_note_id, child_note_id, hierarchy_type FROM note_hierarchy")
	if err != nil {
		log.Printf("Error querying note hierarchy: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var parentID, childID int
		var hierarchyType string
		if err := rows.Scan(&parentID, &childID, &hierarchyType); err != nil {
			log.Printf("Error scanning note hierarchy row: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		parent := noteMap[parentID]
		child := noteMap[childID]
		child.Type = hierarchyType
		parent.Children = append(parent.Children, child)
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
