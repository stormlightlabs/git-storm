package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/stormlightlabs/git-storm/internal/changeset"
	"github.com/stormlightlabs/git-storm/internal/style"
)

// EntryEditorModel holds the state for the inline entry editor TUI.
type EntryEditorModel struct {
	entry     changeset.Entry
	filename  string
	inputs    []textinput.Model
	focusIdx  int
	typeIdx   int // index in validTypes array
	confirmed bool
	cancelled bool
	width     int
	height    int
}

// validTypes defines the allowed changelog entry types.
var validTypes = []string{"added", "changed", "fixed", "removed", "security"}

// editorKeyMap defines keyboard shortcuts for the entry editor.
type editorKeyMap struct {
	Next      key.Binding
	Prev      key.Binding
	Confirm   key.Binding
	Quit      key.Binding
	CycleType key.Binding
}

var editorKeys = editorKeyMap{
	Next: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next field"),
	),
	Prev: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "prev field"),
	),
	Confirm: key.NewBinding(
		key.WithKeys("ctrl+s"),
		key.WithHelp("ctrl+s", "save"),
	),
	Quit: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "cancel"),
	),
	CycleType: key.NewBinding(
		key.WithKeys("ctrl+t"),
		key.WithHelp("ctrl+t", "cycle type"),
	),
}

// NewEntryEditorModel creates a new editor initialized with the given entry.
func NewEntryEditorModel(entry changeset.EntryWithFile) EntryEditorModel {
	m := EntryEditorModel{
		entry:    entry.Entry,
		filename: entry.Filename,
		inputs:   make([]textinput.Model, 2),
	}

	for i, t := range validTypes {
		if t == entry.Entry.Type {
			m.typeIdx = i
			break
		}
	}

	m.inputs[0] = textinput.New()
	m.inputs[0].Placeholder = "optional scope (e.g., cli, api)"
	m.inputs[0].SetValue(entry.Entry.Scope)
	m.inputs[0].CharLimit = 50
	m.inputs[0].Width = 50

	m.inputs[1] = textinput.New()
	m.inputs[1].Placeholder = "brief description of the change"
	m.inputs[1].SetValue(entry.Entry.Summary)
	m.inputs[1].CharLimit = 200
	m.inputs[1].Width = 80

	m.inputs[0].Focus()
	return m
}

// Init implements tea.Model.
func (m EntryEditorModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update implements tea.Model.
func (m EntryEditorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, editorKeys.Quit):
			m.cancelled = true
			return m, tea.Quit
		case key.Matches(msg, editorKeys.Confirm):
			m.confirmed = true
			return m, tea.Quit
		case key.Matches(msg, editorKeys.CycleType):
			m.typeIdx = (m.typeIdx + 1) % len(validTypes)
			return m, nil
		case key.Matches(msg, editorKeys.Next):
			m.nextField()
			return m, nil
		case key.Matches(msg, editorKeys.Prev):
			m.prevField()
			return m, nil
		case msg.String() == "enter":
			m.confirmed = true
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	cmd := m.updateInputs(msg)
	return m, cmd
}

// View implements tea.Model.
func (m EntryEditorModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	var b strings.Builder

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(style.AccentBlue).
		Render(fmt.Sprintf("Editing: %s", m.filename))
	b.WriteString(title)
	b.WriteString("\n\n")

	typeLabel := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7A89")).Render("Type:")
	typeValue := getCategoryStyle(validTypes[m.typeIdx]).Render(validTypes[m.typeIdx])
	b.WriteString(fmt.Sprintf("%s %s (ctrl+t to cycle)\n", typeLabel, typeValue))

	scopeLabel := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7A89")).Render("Scope:")
	b.WriteString(fmt.Sprintf("\n%s\n%s\n", scopeLabel, m.inputs[0].View()))

	summaryLabel := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7A89")).Render("Summary:")
	b.WriteString(fmt.Sprintf("\n%s\n%s\n", summaryLabel, m.inputs[1].View()))

	breakingLabel := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7A89")).Render("Breaking:")
	breakingValue := "no"
	if m.entry.Breaking {
		breakingValue = style.StyleRemoved.Render("yes")
	}
	b.WriteString(fmt.Sprintf("\n%s %s\n", breakingLabel, breakingValue))

	b.WriteString("\n")
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7A89"))
	b.WriteString(helpStyle.Render("tab: next • shift+tab: prev • ctrl+t: cycle type • enter/ctrl+s: save • esc: cancel"))
	return b.String()
}

// GetEditedEntry returns the entry with updated values.
func (m EntryEditorModel) GetEditedEntry() changeset.Entry {
	return changeset.Entry{
		Type:       validTypes[m.typeIdx],
		Scope:      strings.TrimSpace(m.inputs[0].Value()),
		Summary:    strings.TrimSpace(m.inputs[1].Value()),
		Breaking:   m.entry.Breaking,
		CommitHash: m.entry.CommitHash,
		DiffHash:   m.entry.DiffHash,
	}
}

// IsConfirmed returns true if the user confirmed the edit.
func (m EntryEditorModel) IsConfirmed() bool {
	return m.confirmed
}

// IsCancelled returns true if the user cancelled the edit.
func (m EntryEditorModel) IsCancelled() bool {
	return m.cancelled
}

// nextField moves focus to the next input field.
func (m *EntryEditorModel) nextField() {
	m.inputs[m.focusIdx].Blur()
	m.focusIdx = (m.focusIdx + 1) % len(m.inputs)
	m.inputs[m.focusIdx].Focus()
}

// prevField moves focus to the previous input field.
func (m *EntryEditorModel) prevField() {
	m.inputs[m.focusIdx].Blur()
	m.focusIdx--
	if m.focusIdx < 0 {
		m.focusIdx = len(m.inputs) - 1
	}
	m.inputs[m.focusIdx].Focus()
}

// updateInputs handles updates for text input fields.
func (m *EntryEditorModel) updateInputs(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.inputs[m.focusIdx], cmd = m.inputs[m.focusIdx].Update(msg)
	return cmd
}
