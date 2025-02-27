-- DROP DATABASE IF EXISTS draftsmith;
-- CREATE DATABASE draftsmith;
-- \c draftsmith

-- Enable full-text search
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE EXTENSION IF NOT EXISTS unaccent;

-- Table to store notes with a full-text search
CREATE TABLE notes (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    modified_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    fts tsvector GENERATED ALWAYS AS (to_tsvector('english', coalesce(title,'') || ' ' || coalesce(content,''))) STORED
);

CREATE INDEX notes_fts_idx ON notes USING gin(fts);

-- Table to store modified dates
CREATE TABLE note_modifications (
    note_id INT REFERENCES notes(id),
    modified_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Table for categories
CREATE TABLE categories (
    id SERIAL PRIMARY KEY,
    name TEXT UNIQUE NOT NULL
);

-- Table for tags
CREATE TABLE tags (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL
);

-- Many-to-many relationship tables for categories and tags
CREATE TABLE note_categories (
    note_id INT REFERENCES notes(id),
    category_id INT REFERENCES categories(id),
    PRIMARY KEY (note_id, category_id)
);

CREATE TABLE note_tags (
    note_id INT REFERENCES notes(id),
    tag_id INT REFERENCES tags(id),
    PRIMARY KEY (note_id, tag_id)
);

-- Table for tag hierarchy
CREATE TABLE tag_hierarchy (
    id SERIAL PRIMARY KEY,
    parent_tag_id INT REFERENCES tags(id),
    child_tag_id INT REFERENCES tags(id),
    UNIQUE(parent_tag_id, child_tag_id)
);

-- Table for assets
CREATE TABLE assets (
    id SERIAL PRIMARY KEY,
    note_id INT REFERENCES notes(id),
    asset_type TEXT NOT NULL,
    location TEXT NOT NULL,
    description TEXT
);

-- Table for misc attributes
CREATE TABLE attributes (
    id SERIAL PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    description TEXT
);

CREATE TABLE note_attributes (
    id SERIAL PRIMARY KEY,
    note_id INT REFERENCES notes(id),
    attribute_id INT REFERENCES attributes(id),
    VALUE TEXT NOT NULL
);

-- Table for note types
CREATE TABLE note_types (
    id SERIAL PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    description TEXT
);

CREATE TABLE note_type_mappings (
    note_id INT REFERENCES notes(id),
    type_id INT REFERENCES note_types(id),
    PRIMARY KEY (note_id, type_id)
);

-- Table for handling note hierarchy
CREATE TABLE note_hierarchy (
    id SERIAL PRIMARY KEY,
    parent_note_id INT REFERENCES notes(id),
    child_note_id INT REFERENCES notes(id),
    hierarchy_type TEXT CHECK (hierarchy_type IN ('page', 'block', 'subpage'))
);

-- Table for journal/calendar view (optional)
CREATE TABLE journal_entries (
    id SERIAL PRIMARY KEY,
    note_id INT REFERENCES notes(id),
    entry_date DATE NOT NULL
);

-- Populate initial data for note types
INSERT INTO note_types (name, description) VALUES
    ('asset', 'Asset related notes'),
    ('bookmark', 'Bookmark related notes'),
    ('contact', 'Contact information'),
    ('page', 'A standalone page'),
    ('block', 'A block of information within a page'),
    ('subpage', 'A subpage within a note');

-- Populate initial data for categories
INSERT INTO categories (name) VALUES
    ('Personal'),
    ('Work'),
    ('Ideas'),
    ('Journal');

-- Populate initial data for tags
INSERT INTO tags (name) VALUES
    ('important'),
    ('urgent'),
    ('todo'),
    ('done');

-- Populate initial data for attributes
INSERT INTO attributes (name, description) VALUES
    ('location', 'Location of the note'),
    ('author', 'Author of the note'),
    ('source', 'Source of the note');

-- Populate initial data for notes
INSERT INTO notes (title, content) VALUES
    ('First note', 'This is the first note in the system.'),
    ('Second note', 'This is the second note in the system.');
