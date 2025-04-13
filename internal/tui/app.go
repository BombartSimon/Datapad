package tui

import (
	"datapad/internal/notes"
	"fmt"
	"regexp"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yuin/goldmark"
)

// Mode represents the current state of the user interface
type Mode int

const (
	ModeList Mode = iota
	ModeView
	ModeEdit
	ModeNew
	ModeSearch
	ModeAddImage
	ModeHelp
	ModeAddTag
	ModeFilterByTag
)

// KeyMap defines the shortcut keys for the application
type KeyMap struct {
	Up            key.Binding
	Down          key.Binding
	Enter         key.Binding
	Back          key.Binding
	Quit          key.Binding
	New           key.Binding
	Edit          key.Binding
	Delete        key.Binding
	Save          key.Binding
	AddImage      key.Binding
	Search        key.Binding
	Help          key.Binding
	AddTag        key.Binding
	FilterByTag   key.Binding
	TogglePreview key.Binding
}

// DefaultKeyMap returns the default key mapping
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c", "q"),
			key.WithHelp("ctrl+c/q", "quit"),
		),
		New: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "new note"),
		),
		Edit: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "edit"),
		),
		Delete: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete"),
		),
		Save: key.NewBinding(
			key.WithKeys("ctrl+s"),
			key.WithHelp("ctrl+s", "save"),
		),
		AddImage: key.NewBinding(
			key.WithKeys("i"),
			key.WithHelp("i", "add image"),
		),
		Search: key.NewBinding(
			key.WithKeys("ctrl+f", "/"),
			key.WithHelp("ctrl+f", "search"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		AddTag: key.NewBinding(
			key.WithKeys("t"),
			key.WithHelp("t", "add tag"),
		),
		FilterByTag: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "filter by tag"),
		),
		TogglePreview: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "toggle preview"),
		),
	}
}

// Model contains the complete state of the application
type Model struct {
	notesManager  *notes.NotesManager
	mode          Mode
	noteList      list.Model
	textArea      textarea.Model
	titleInput    textinput.Model
	imagePath     textinput.Model
	imageCaption  textinput.Model
	searchInput   textinput.Model
	tagInput      textinput.Model
	selectedNote  *notes.Note
	keys          KeyMap
	help          help.Model
	showPreview   bool
	width, height int
	statusMsg     string
	markdown      goldmark.Markdown
}

// NewModel creates a new application model
func NewModel(notesManager *notes.NotesManager) Model {
	keys := DefaultKeyMap()
	helpModel := help.New()

	// Configure the notes list
	noteItems := []list.Item{}
	for _, note := range notesManager.Notes {
		noteItems = append(noteItems, NoteItem{note})
	}

	noteList := list.New(noteItems, list.NewDefaultDelegate(), 0, 0)
	noteList.Title = "Notes"
	noteList.SetShowHelp(false)

	// Configure the text editor
	ta := textarea.New()
	ta.Placeholder = "Write your note here..."
	ta.CharLimit = 0
	ta.SetWidth(80)
	ta.SetHeight(20)
	ta.ShowLineNumbers = true

	// Configure the title field
	ti := textinput.New()
	ti.Placeholder = "Note title"
	ti.CharLimit = 100
	ti.Width = 40

	// Configure image fields
	imagePath := textinput.New()
	imagePath.Placeholder = "Path to image"
	imagePath.CharLimit = 500
	imagePath.Width = 40

	imageCaption := textinput.New()
	imageCaption.Placeholder = "Image caption"
	imageCaption.CharLimit = 100
	imageCaption.Width = 40

	// Configure search field
	searchInput := textinput.New()
	searchInput.Placeholder = "Search..."
	searchInput.CharLimit = 100
	searchInput.Width = 40

	// Configure tag field
	tagInput := textinput.New()
	tagInput.Placeholder = "Tag name"
	tagInput.CharLimit = 50
	tagInput.Width = 30

	return Model{
		notesManager: notesManager,
		mode:         ModeList,
		noteList:     noteList,
		textArea:     ta,
		titleInput:   ti,
		imagePath:    imagePath,
		imageCaption: imageCaption,
		searchInput:  searchInput,
		tagInput:     tagInput,
		keys:         keys,
		help:         helpModel,
		showPreview:  false,
		markdown:     goldmark.New(),
	}
}

// NoteItem is a wrapper to adapt Note to the list.Item interface
type NoteItem struct {
	*notes.Note
}

// Title returns the title of a note for display in the list
func (n NoteItem) Title() string {
	return n.Note.Title
}

// Description returns a description of the note for display in the list
func (n NoteItem) Description() string {
	content := n.Note.Content
	if len(content) > 50 {
		content = content[:50] + "..."
	}
	tags := strings.Join(n.Note.Tags, ", ")
	if tags != "" {
		tags = "[" + tags + "]"
	}
	return fmt.Sprintf("%s %s", content, lipgloss.NewStyle().Foreground(lipgloss.Color("#5f5")).Render(tags))
}

// FilterValue returns the value to use for filtering notes
func (n NoteItem) FilterValue() string {
	return n.Note.Title + " " + n.Note.Content + " " + strings.Join(n.Note.Tags, " ")
}

// TagItem represents a tag in the tag list
type TagItem struct {
	Tag string
}

// Title returns the tag name for display in the list
func (t TagItem) Title() string {
	return t.Tag
}

// Description returns an empty description for tags
func (t TagItem) Description() string {
	return ""
}

// FilterValue returns the value to use for filtering tags
func (t TagItem) FilterValue() string {
	return t.Tag
}

// Init initializes the application model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update updates the application model based on received messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.noteList.SetWidth(msg.Width)
		m.noteList.SetHeight(msg.Height - 4) // Reserve space for status
		m.textArea.SetWidth(msg.Width)
		m.textArea.SetHeight(msg.Height - 6)
		return m, nil

	case tea.KeyMsg:
		// Handle global keys
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		}

		// Handle keys based on mode
		switch m.mode {
		case ModeAddTag:
			if key.Matches(msg, m.keys.Back) {
				m.mode = ModeView
				return m, nil
			} else if key.Matches(msg, m.keys.Enter) {
				// Add tag to the note
				if m.tagInput.Value() != "" {
					m.selectedNote.AddTag(m.tagInput.Value())
					m.notesManager.UpdateNote(m.selectedNote)
					m.statusMsg = "Tag added successfully"
					m.mode = ModeView
				}
				return m, nil
			}
			m.tagInput, cmd = m.tagInput.Update(msg)
			cmds = append(cmds, cmd)

		case ModeFilterByTag:
			if key.Matches(msg, m.keys.Back) {
				m.mode = ModeList
				return m, nil
			} else if key.Matches(msg, m.keys.Enter) {
				// Get all tags
				tags := m.notesManager.GetAllTags()

				// If no tags exist, return to the list
				if len(tags) == 0 {
					m.statusMsg = "No tags available"
					m.mode = ModeList
					return m, nil
				}

				// Select the tag at the current index
				index := m.noteList.Index()
				if index >= 0 && index < len(tags) {
					selectedTag := tags[index]

					// Filter notes by this tag
					filteredNotes := m.notesManager.FilterByTags([]string{selectedTag})

					// Update the list of notes
					items := []list.Item{}
					for _, n := range filteredNotes {
						items = append(items, NoteItem{n})
					}
					m.noteList.SetItems(items)

					m.statusMsg = fmt.Sprintf("Notes filtered by tag: %s", selectedTag)
					m.mode = ModeList
				}
				return m, nil
			}
			// Handle navigation in the tag list
			m.noteList, cmd = m.noteList.Update(msg)
			return m, cmd
		case ModeList:
			return m.updateListMode(msg)
		case ModeView:
			return m.updateViewMode(msg)
		case ModeEdit, ModeNew:
			if key.Matches(msg, m.keys.Save) {
				return m.saveNote()
			} else if key.Matches(msg, m.keys.Back) {
				if m.mode == ModeNew {
					m.mode = ModeList
				} else {
					m.mode = ModeView
				}
				return m, nil
			} else if key.Matches(msg, m.keys.TogglePreview) {
				m.showPreview = !m.showPreview
				return m, nil
			}

			if m.titleInput.Focused() {
				m.titleInput, cmd = m.titleInput.Update(msg)
				cmds = append(cmds, cmd)
			} else {
				m.textArea, cmd = m.textArea.Update(msg)
				cmds = append(cmds, cmd)
			}

			// Switch focus between title and content with tab
			if msg.String() == "tab" {
				if m.titleInput.Focused() {
					m.titleInput.Blur()
					m.textArea.Focus()
				} else {
					m.textArea.Blur()
					m.titleInput.Focus()
				}
			}

		case ModeSearch:
			if key.Matches(msg, m.keys.Back) {
				m.mode = ModeList
				return m, nil
			} else if key.Matches(msg, m.keys.Enter) {
				items := []list.Item{}
				notes := m.notesManager.SearchNotes(m.searchInput.Value())
				for _, note := range notes {
					items = append(items, NoteItem{note})
				}
				m.noteList.SetItems(items)
				m.mode = ModeList
				return m, nil
			}

			m.searchInput, cmd = m.searchInput.Update(msg)
			cmds = append(cmds, cmd)

		case ModeAddImage:
			if key.Matches(msg, m.keys.Back) {
				m.mode = ModeView
				return m, nil
			} else if key.Matches(msg, m.keys.Enter) {
				// Add the image to the note
				err := m.notesManager.ImportImage(
					m.selectedNote.ID,
					m.imagePath.Value(),
					m.imageCaption.Value(),
					"", // No alt text for now
				)

				if err != nil {
					m.statusMsg = fmt.Sprintf("Error: %s", err)
				} else {
					m.statusMsg = "Image added successfully"
					m.imagePath.Reset()
					m.imageCaption.Reset()
					m.mode = ModeView
				}
				return m, nil
			}

			if m.imagePath.Focused() {
				m.imagePath, cmd = m.imagePath.Update(msg)
				cmds = append(cmds, cmd)
			} else {
				m.imageCaption, cmd = m.imageCaption.Update(msg)
				cmds = append(cmds, cmd)
			}

			// Switch focus
			if msg.String() == "tab" {
				if m.imagePath.Focused() {
					m.imagePath.Blur()
					m.imageCaption.Focus()
				} else {
					m.imageCaption.Blur()
					m.imagePath.Focus()
				}
			}
		}

	}

	return m, tea.Batch(cmds...)
}

// updateListMode handles updates in list mode
func (m Model) updateListMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch {
	case key.Matches(msg, m.keys.New):
		m.mode = ModeNew
		m.titleInput.Reset()
		m.textArea.Reset()
		m.titleInput.Focus()
		return m, nil

	case key.Matches(msg, m.keys.Enter):
		if len(m.noteList.Items()) == 0 {
			return m, nil
		}
		item, ok := m.noteList.SelectedItem().(NoteItem)
		if ok {
			m.selectedNote = item.Note
			m.mode = ModeView
			return m, nil
		}

	case key.Matches(msg, m.keys.Search):
		m.mode = ModeSearch
		m.searchInput.Reset()
		m.searchInput.Focus()
		return m, nil

	case key.Matches(msg, m.keys.FilterByTag):
		// Get all tags
		tags := m.notesManager.GetAllTags()

		// If no tags exist, display a message
		if len(tags) == 0 {
			m.statusMsg = "No tags available"
			return m, nil
		}

		// Create TagItem elements for the list
		items := []list.Item{}
		for _, tag := range tags {
			items = append(items, TagItem{Tag: tag})
		}

		// Configure the list to display tags
		m.noteList.SetItems(items)
		m.mode = ModeFilterByTag
		m.statusMsg = "Select a tag"
		return m, nil
	}

	// Update the list model
	m.noteList, cmd = m.noteList.Update(msg)
	return m, cmd
}

// updateViewMode handles updates in view mode
func (m Model) updateViewMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Back):
		m.mode = ModeList
		return m, nil

	case key.Matches(msg, m.keys.Edit):
		m.mode = ModeEdit
		m.titleInput.SetValue(m.selectedNote.Title)
		m.textArea.SetValue(m.selectedNote.Content)
		m.titleInput.Focus()
		return m, nil

	case key.Matches(msg, m.keys.Delete):
		m.notesManager.DeleteNote(m.selectedNote.ID)

		// Update the list
		items := []list.Item{}
		for _, note := range m.notesManager.Notes {
			items = append(items, NoteItem{note})
		}
		m.noteList.SetItems(items)

		m.mode = ModeList
		m.statusMsg = "Note deleted"
		return m, nil

	case key.Matches(msg, m.keys.AddImage):
		m.mode = ModeAddImage
		m.imagePath.Reset()
		m.imageCaption.Reset()
		m.imagePath.Focus()
		return m, nil

	case key.Matches(msg, m.keys.AddTag):
		m.mode = ModeAddTag
		m.tagInput.Reset()
		m.tagInput.Focus()
		return m, nil
	}

	return m, nil
}

// saveNote saves the note being edited
func (m Model) saveNote() (tea.Model, tea.Cmd) {
	if m.mode == ModeNew {
		note := m.notesManager.CreateNote(m.titleInput.Value())
		note.Content = m.textArea.Value()
		m.notesManager.UpdateNote(note)
		m.selectedNote = note

		// Update the list
		items := []list.Item{}
		for _, n := range m.notesManager.Notes {
			items = append(items, NoteItem{n})
		}
		m.noteList.SetItems(items)

		m.mode = ModeView
		m.statusMsg = "Note created successfully"
	} else {
		// Edit mode
		m.selectedNote.Title = m.titleInput.Value()
		m.selectedNote.Content = m.textArea.Value()
		m.notesManager.UpdateNote(m.selectedNote)

		// Update the list
		items := []list.Item{}
		for _, n := range m.notesManager.Notes {
			items = append(items, NoteItem{n})
		}
		m.noteList.SetItems(items)

		m.mode = ModeView
		m.statusMsg = "Note updated successfully"
	}

	return m, nil
}

// View returns the user interface display
func (m Model) View() string {
	switch m.mode {
	case ModeList:
		return lipgloss.JoinVertical(
			lipgloss.Left,
			m.noteList.View(),
			m.statusBar(),
			m.helpView(),
		)

	case ModeView:
		return m.viewNote()

	case ModeEdit, ModeNew:
		return m.viewEditor()

	case ModeSearch:
		return lipgloss.JoinVertical(
			lipgloss.Left,
			"Search:",
			m.searchInput.View(),
			m.statusBar(),
			"Press Enter to search, Esc to cancel",
		)

	case ModeAddImage:
		return lipgloss.JoinVertical(
			lipgloss.Left,
			"Add an image:",
			"Image path:",
			m.imagePath.View(),
			"Caption (optional):",
			m.imageCaption.View(),
			m.statusBar(),
			"Press Enter to add, Esc to cancel",
			"",
		)

	case ModeAddTag:
		return lipgloss.JoinVertical(
			lipgloss.Left,
			"Add a tag:",
			m.tagInput.View(),
			m.statusBar(),
			"Press Enter to add, Esc to cancel",
		)

	case ModeFilterByTag:
		return lipgloss.JoinVertical(
			lipgloss.Left,
			"Filter by tag:",
			m.noteList.View(),
			m.statusBar(),
			"Press Enter to filter, Esc to cancel",
		)

	default:
		return "Unknown mode"
	}
}

// viewNote displays a note in view mode
func (m Model) viewNote() string {
	if m.selectedNote == nil {
		return "No note selected"
	}

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFA500")).
		MarginBottom(1)

	contentStyle := lipgloss.NewStyle().
		MarginTop(1).
		MarginBottom(1).
		Width(m.width)

	metadataStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		MarginTop(1)

	tagsStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#5f5"))

	title := titleStyle.Render(m.selectedNote.Title)
	content := contentStyle.Render(m.selectedNote.Content)
	created := metadataStyle.Render(fmt.Sprintf("Created on: %s", m.selectedNote.CreatedAt.Format("02/01/2006 15:04")))
	updated := metadataStyle.Render(fmt.Sprintf("Updated on: %s", m.selectedNote.UpdatedAt.Format("02/01/2006 15:04")))

	tags := ""
	if len(m.selectedNote.Tags) > 0 {
		tags = tagsStyle.Render("Tags: " + strings.Join(m.selectedNote.Tags, ", "))
	}

	imagesSection := ""
	if len(m.selectedNote.Images) > 0 {
		imagesSection = "Images:\n"
		for _, img := range m.selectedNote.Images {
			caption := img.Caption
			if caption == "" {
				caption = "(no caption)"
			}
			imagesSection += fmt.Sprintf("- %s: %s\n", img.Path, caption)
		}
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		content,
		imagesSection,
		tags,
		created,
		updated,
		"",
		m.statusBar(),
		m.helpView(),
	)
}

// viewEditor displays the note editor
func (m Model) viewEditor() string {
	modeText := "Editing"
	if m.mode == ModeNew {
		modeText = "New note"
	}

	// If preview is enabled, split the screen into two parts
	if m.showPreview {
		// Calculate widths for editor and preview
		editorWidth := m.width / 2
		previewWidth := m.width - editorWidth - 1

		// Create styles
		editorStyle := lipgloss.NewStyle().Width(editorWidth)
		previewStyle := lipgloss.NewStyle().
			Width(previewWidth).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#5f5")).
			Padding(0, 1)

		// Editor section
		editorSection := lipgloss.JoinVertical(
			lipgloss.Left,
			modeText,
			"Title:",
			m.titleInput.View(),
			"Content:",
			m.textArea.View(),
		)

		// Preview section
		previewContent := m.renderMarkdown(m.textArea.Value())
		previewTitle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFA500")).Render(m.titleInput.Value())

		previewSection := lipgloss.JoinVertical(
			lipgloss.Left,
			"Preview:",
			previewTitle,
			"",
			previewContent,
		)

		// Join the two sections horizontally
		content := lipgloss.JoinHorizontal(
			lipgloss.Top,
			editorStyle.Render(editorSection),
			"│",
			previewStyle.Render(previewSection),
		)

		return lipgloss.JoinVertical(
			lipgloss.Left,
			content,
			m.statusBar(),
			"Ctrl+S to save, Esc to cancel, P to toggle preview",
		)
	}

	// Normal display (without preview)
	return lipgloss.JoinVertical(
		lipgloss.Left,
		modeText+" (Tab to switch between title and content)",
		"Title:",
		m.titleInput.View(),
		"Content:",
		m.textArea.View(),
		m.statusBar(),
		"Ctrl+S to save, Esc to cancel, P for preview",
	)
}

// renderMarkdown renders Markdown content as formatted text
func (m Model) renderMarkdown(content string) string {
	if content == "" {
		return ""
	}

	// Convert markdown to HTML
	var htmlBuf strings.Builder
	if err := m.markdown.Convert([]byte(content), &htmlBuf); err != nil {
		return fmt.Sprintf("Error rendering Markdown: %s", err)
	}

	html := htmlBuf.String()

	// Apply styles for common HTML elements
	// Define styles
	h1Style := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF0000")).MarginBottom(1)
	h2Style := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF5500")).MarginBottom(1)
	h3Style := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFAA00"))
	boldStyle := lipgloss.NewStyle().Bold(true)
	italicStyle := lipgloss.NewStyle().Italic(true)
	codeStyle := lipgloss.NewStyle().Background(lipgloss.Color("#333")).Foreground(lipgloss.Color("#FFF"))

	// Replace HTML tags with formatted text
	// Headings
	h1Regex := regexp.MustCompile(`<h1[^>]*>(.*?)</h1>`)
	html = h1Regex.ReplaceAllStringFunc(html, func(match string) string {
		content := h1Regex.FindStringSubmatch(match)[1]
		return h1Style.Render(content)
	})

	h2Regex := regexp.MustCompile(`<h2[^>]*>(.*?)</h2>`)
	html = h2Regex.ReplaceAllStringFunc(html, func(match string) string {
		content := h2Regex.FindStringSubmatch(match)[1]
		return h2Style.Render(content)
	})

	h3Regex := regexp.MustCompile(`<h3[^>]*>(.*?)</h3>`)
	html = h3Regex.ReplaceAllStringFunc(html, func(match string) string {
		content := h3Regex.FindStringSubmatch(match)[1]
		return h3Style.Render(content)
	})

	// Bold
	boldRegex := regexp.MustCompile(`<(?:strong|b)[^>]*>(.*?)</(?:strong|b)>`)
	html = boldRegex.ReplaceAllStringFunc(html, func(match string) string {
		content := boldRegex.FindStringSubmatch(match)[1]
		return boldStyle.Render(content)
	})

	// Italic
	italicRegex := regexp.MustCompile(`<(?:em|i)[^>]*>(.*?)</(?:em|i)>`)
	html = italicRegex.ReplaceAllStringFunc(html, func(match string) string {
		content := italicRegex.FindStringSubmatch(match)[1]
		return italicStyle.Render(content)
	})

	// Code
	codeRegex := regexp.MustCompile(`<code[^>]*>(.*?)</code>`)
	html = codeRegex.ReplaceAllStringFunc(html, func(match string) string {
		content := codeRegex.FindStringSubmatch(match)[1]
		return codeStyle.Render(content)
	})

	// Lists
	html = strings.ReplaceAll(html, "<ul>", "")
	html = strings.ReplaceAll(html, "</ul>", "\n")
	html = strings.ReplaceAll(html, "<ol>", "")
	html = strings.ReplaceAll(html, "</ol>", "\n")

	// List items
	liRegex := regexp.MustCompile(`<li[^>]*>(.*?)</li>`)
	html = liRegex.ReplaceAllStringFunc(html, func(match string) string {
		content := liRegex.FindStringSubmatch(match)[1]
		return "• " + content + "\n"
	})

	// Paragraphs
	html = strings.ReplaceAll(html, "<p>", "")
	html = strings.ReplaceAll(html, "</p>", "\n\n")

	// Links
	linkRegex := regexp.MustCompile(`<a[^>]*href="([^"]*)"[^>]*>(.*?)</a>`)
	html = linkRegex.ReplaceAllStringFunc(html, func(match string) string {
		parts := linkRegex.FindStringSubmatch(match)
		url := parts[1]
		text := parts[2]
		return fmt.Sprintf("%s (%s)", text, url)
	})

	// Clean up excessive line breaks
	html = strings.ReplaceAll(html, "\n\n\n", "\n\n")

	// Remove remaining HTML tags
	cleanRegex := regexp.MustCompile("<[^>]*>")
	html = cleanRegex.ReplaceAllString(html, "")

	// Decode HTML entities
	html = strings.ReplaceAll(html, "&lt;", "<")
	html = strings.ReplaceAll(html, "&gt;", ">")
	html = strings.ReplaceAll(html, "&amp;", "&")
	html = strings.ReplaceAll(html, "&quot;", "\"")

	return html
}

// statusBar displays the status bar at the bottom of the screen
func (m Model) statusBar() string {
	status := m.statusMsg
	if status == "" {
		status = "Ready"
	}

	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#555555")).
		Padding(0, 1).
		Width(m.width).
		Render(status)
}

// helpView displays navigation help
func (m Model) helpView() string {
	switch m.mode {
	case ModeList:
		return m.help.ShortHelpView([]key.Binding{
			m.keys.Up,
			m.keys.Down,
			m.keys.Enter,
			m.keys.New,
			m.keys.Search,
			m.keys.Quit,
		})
	case ModeView:
		return m.help.ShortHelpView([]key.Binding{
			m.keys.Back,
			m.keys.Edit,
			m.keys.Delete,
			m.keys.AddImage,
			m.keys.AddTag,
			m.keys.Quit,
		})
	default:
		return ""
	}
}

// App launches the TUI application
func App(storagePath string) error {
	notesManager, err := notes.NewNotesManager(storagePath)
	if err != nil {
		return fmt.Errorf("error initializing notes manager: %w", err)
	}

	p := tea.NewProgram(NewModel(notesManager), tea.WithAltScreen())
	_, err = p.Run()
	return err
}
