package notes

import (
	"time"
)

// Note represents an individual note with its content and metadata
type Note struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"` // Markdown content
	Images    []Image   `json:"images,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Tags      []string  `json:"tags,omitempty"`
}

// Image represents an image embedded in a note
type Image struct {
	ID       string `json:"id"`
	Path     string `json:"path"`     // Path to the image on disk
	Caption  string `json:"caption"`  // Optional caption
	AltText  string `json:"alt_text"` // Alternative text for accessibility
	Position int    `json:"position"` // Position in the note
}

// NewNote creates a new note with default values
func NewNote(title string) *Note {
	now := time.Now()
	return &Note{
		ID:        generateID(),
		Title:     title,
		CreatedAt: now,
		UpdatedAt: now,
		Images:    []Image{},
		Tags:      []string{},
	}
}

// AddImage adds an image to a note
func (n *Note) AddImage(path string, caption string, altText string) {
	// Ensure the Images slice is initialized
	if n.Images == nil {
		n.Images = []Image{}
	}

	// Add the image to the note
	n.Images = append(n.Images, Image{
		Path:    path,
		Caption: caption,
		AltText: altText,
	})

	// Update the modification timestamp
	n.UpdatedAt = time.Now()
}

// AddTag adds a new tag to the note
func (n *Note) AddTag(tag string) {
	for _, t := range n.Tags {
		if t == tag {
			return // Tag already exists
		}
	}
	n.Tags = append(n.Tags, tag)
	n.UpdatedAt = time.Now()
}

// RemoveTag removes a tag from the note
func (n *Note) RemoveTag(tag string) {
	for i, t := range n.Tags {
		if t == tag {
			n.Tags = append(n.Tags[:i], n.Tags[i+1:]...)
			n.UpdatedAt = time.Now()
			return
		}
	}
}

// Utility function to generate a unique ID
func generateID() string {
	return time.Now().Format("20060102150405") + randomString(6)
}

// Generates a random string of length n
func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}
