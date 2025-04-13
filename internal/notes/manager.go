package notes

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// NotesManager manages the collection of notes and their saving/loading
type NotesManager struct {
	Notes       []*Note
	StoragePath string
	ImageDir    string
}

// NewNotesManager creates a new notes manager
func NewNotesManager(storagePath string) (*NotesManager, error) {
	// Create storage directory if it doesn't exist
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		return nil, fmt.Errorf("unable to create storage directory: %w", err)
	}

	imageDir := filepath.Join(storagePath, "images")
	if err := os.MkdirAll(imageDir, 0755); err != nil {
		return nil, fmt.Errorf("unable to create images directory: %w", err)
	}

	manager := &NotesManager{
		Notes:       []*Note{},
		StoragePath: storagePath,
		ImageDir:    imageDir,
	}

	// Load existing notes
	err := manager.LoadNotes()
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("error loading notes: %w", err)
	}

	return manager, nil
}

// CreateNote creates a new note and adds it to the manager
func (m *NotesManager) CreateNote(title string) *Note {
	note := NewNote(title)
	m.Notes = append(m.Notes, note)
	return note
}

// GetNoteByID retrieves a note by its ID
func (m *NotesManager) GetNoteByID(id string) (*Note, error) {
	for _, note := range m.Notes {
		if note.ID == id {
			return note, nil
		}
	}
	return nil, errors.New("note not found")
}

// UpdateNote updates an existing note
func (m *NotesManager) UpdateNote(note *Note) {
	note.UpdatedAt = time.Now()
	m.SaveNotes() // Automatic save after update
}

// DeleteNote deletes a note by its ID
func (m *NotesManager) DeleteNote(id string) error {
	for i, note := range m.Notes {
		if note.ID == id {
			// Remove note from the list
			m.Notes = append(m.Notes[:i], m.Notes[i+1:]...)
			return m.SaveNotes()
		}
	}
	return errors.New("note not found")
}

// SearchNotes searches for notes by title or content
func (m *NotesManager) SearchNotes(query string) []*Note {
	if query == "" {
		return m.Notes
	}

	query = strings.ToLower(query)
	results := []*Note{}

	for _, note := range m.Notes {
		if strings.Contains(strings.ToLower(note.Title), query) ||
			strings.Contains(strings.ToLower(note.Content), query) {
			results = append(results, note)
		}
	}

	return results
}

// FilterByTags filters notes by tags
func (m *NotesManager) FilterByTags(tags []string) []*Note {
	if len(tags) == 0 {
		return m.Notes
	}

	results := []*Note{}

	for _, note := range m.Notes {
		match := false
		for _, noteTag := range note.Tags {
			for _, filterTag := range tags {
				if noteTag == filterTag {
					match = true
					break
				}
			}
			if match {
				break
			}
		}
		if match {
			results = append(results, note)
		}
	}

	return results
}

// ImportImage imports an image into the images directory and adds it to a note
func (m *NotesManager) ImportImage(noteID string, sourcePath, caption, altText string) error {
	note, err := m.GetNoteByID(noteID)
	if err != nil {
		return err
	}

	// Generate a unique name for the image
	ext := filepath.Ext(sourcePath)
	newFilename := fmt.Sprintf("%s%s", generateID(), ext)
	destPath := filepath.Join(m.ImageDir, newFilename)

	// Copy the image file
	source, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("unable to open source image: %w", err)
	}
	defer source.Close()

	destination, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("unable to create destination file: %w", err)
	}
	defer destination.Close()

	data, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("unable to read source image: %w", err)
	}

	if _, err := destination.Write(data); err != nil {
		return fmt.Errorf("unable to write image: %w", err)
	}

	// Add image to the note
	note.AddImage(newFilename, caption, altText)
	m.UpdateNote(note)

	return nil
}

// SaveNotes saves all notes to a JSON file
func (m *NotesManager) SaveNotes() error {
	// Sort notes by update date (most recent first)
	sort.Slice(m.Notes, func(i, j int) bool {
		return m.Notes[i].UpdatedAt.After(m.Notes[j].UpdatedAt)
	})

	data, err := json.MarshalIndent(m.Notes, "", "  ")
	if err != nil {
		return fmt.Errorf("error serializing notes: %w", err)
	}

	notesFile := filepath.Join(m.StoragePath, "notes.json")
	if err := os.WriteFile(notesFile, data, 0644); err != nil {
		return fmt.Errorf("error writing notes file: %w", err)
	}

	return nil
}

// LoadNotes loads all notes from a JSON file
func (m *NotesManager) LoadNotes() error {
	notesFile := filepath.Join(m.StoragePath, "notes.json")

	data, err := os.ReadFile(notesFile)
	if err != nil {
		return err
	}

	var notes []*Note
	if err := json.Unmarshal(data, &notes); err != nil {
		return fmt.Errorf("error deserializing notes: %w", err)
	}

	m.Notes = notes
	return nil
}

// GetAllTags retrieves all unique tags used in notes
func (m *NotesManager) GetAllTags() []string {
	tagsMap := make(map[string]bool)

	for _, note := range m.Notes {
		for _, tag := range note.Tags {
			tagsMap[tag] = true
		}
	}

	tags := make([]string, 0, len(tagsMap))
	for tag := range tagsMap {
		tags = append(tags, tag)
	}

	sort.Strings(tags)
	return tags
}
