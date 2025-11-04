package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/stormlightlabs/git-storm/internal/changeset"
	"github.com/stormlightlabs/git-storm/internal/style"
)

// ReviewAction represents an action to perform on a changeset entry.
type ReviewAction int

const (
	ActionKeep ReviewAction = iota
	ActionDelete
	ActionEdit
)

// ReviewItem wraps a changeset entry with its review state.
type ReviewItem struct {
	Entry  changeset.EntryWithFile
	Action ReviewAction
}

// ChangesetReviewModel holds the state for the interactive changeset review TUI.
type ChangesetReviewModel struct {
	viewport  viewport.Model
	items     []ReviewItem
	cursor    int
	ready     bool
	width     int
	height    int
	confirmed bool
	cancelled bool
}

// changesetReviewKeyMap defines keyboard shortcuts for the changeset reviewer.
type changesetReviewKeyMap struct {
	Up       key.Binding
	Down     key.Binding
	PageUp   key.Binding
	PageDown key.Binding
	Top      key.Binding
	Bottom   key.Binding
	Delete   key.Binding
	Edit     key.Binding
	Keep     key.Binding
	Confirm  key.Binding
	Quit     key.Binding
}

var reviewKeys = changesetReviewKeyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	PageUp: key.NewBinding(
		key.WithKeys("pgup", "u"),
		key.WithHelp("pgup/u", "page up"),
	),
	PageDown: key.NewBinding(
		key.WithKeys("pgdown", "d"),
		key.WithHelp("pgdn/d", "page down"),
	),
	Top: key.NewBinding(
		key.WithKeys("g", "home"),
		key.WithHelp("g/home", "top"),
	),
	Bottom: key.NewBinding(
		key.WithKeys("G", "end"),
		key.WithHelp("G/end", "bottom"),
	),
	Delete: key.NewBinding(
		key.WithKeys("x"),
		key.WithHelp("x", "mark delete"),
	),
	Edit: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "mark edit"),
	),
	Keep: key.NewBinding(
		key.WithKeys(" "),
		key.WithHelp("space", "keep"),
	),
	Confirm: key.NewBinding(
		key.WithKeys("enter", "c"),
		key.WithHelp("enter/c", "confirm"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "esc", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}

// NewChangesetReviewModel creates a new changeset review model.
func NewChangesetReviewModel(entries []changeset.EntryWithFile) ChangesetReviewModel {
	items := make([]ReviewItem, 0, len(entries))

	for _, entry := range entries {
		items = append(items, ReviewItem{
			Entry:  entry,
			Action: ActionKeep,
		})
	}

	return ChangesetReviewModel{
		items:  items,
		cursor: 0,
		ready:  false,
	}
}

// Init initializes the model (required by Bubble Tea).
func (m ChangesetReviewModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model state.
func (m ChangesetReviewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, reviewKeys.Quit):
			m.cancelled = true
			return m, tea.Quit

		case key.Matches(msg, reviewKeys.Confirm):
			m.confirmed = true
			return m, tea.Quit

		case key.Matches(msg, reviewKeys.Up):
			if m.cursor > 0 {
				m.cursor--
				m.ensureVisible()
			}

		case key.Matches(msg, reviewKeys.Down):
			if m.cursor < len(m.items)-1 {
				m.cursor++
				m.ensureVisible()
			}

		case key.Matches(msg, reviewKeys.PageUp):
			m.cursor -= m.viewport.Height
			if m.cursor < 0 {
				m.cursor = 0
			}
			m.ensureVisible()

		case key.Matches(msg, reviewKeys.PageDown):
			m.cursor += m.viewport.Height
			if m.cursor >= len(m.items) {
				m.cursor = len(m.items) - 1
			}
			m.ensureVisible()

		case key.Matches(msg, reviewKeys.Top):
			m.cursor = 0
			m.ensureVisible()

		case key.Matches(msg, reviewKeys.Bottom):
			m.cursor = len(m.items) - 1
			m.ensureVisible()

		case key.Matches(msg, reviewKeys.Delete):
			if m.cursor >= 0 && m.cursor < len(m.items) {
				m.items[m.cursor].Action = ActionDelete
				m.updateContent()
			}

		case key.Matches(msg, reviewKeys.Edit):
			if m.cursor >= 0 && m.cursor < len(m.items) {
				m.items[m.cursor].Action = ActionEdit
				m.updateContent()
			}

		case key.Matches(msg, reviewKeys.Keep):
			if m.cursor >= 0 && m.cursor < len(m.items) {
				m.items[m.cursor].Action = ActionKeep
				m.updateContent()
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-4)
			m.ready = true
			m.updateContent()
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - 4
			m.updateContent()
		}
	}

	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// View renders the current view of the changeset reviewer.
func (m ChangesetReviewModel) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}

	header := m.renderReviewHeader()
	footer := m.renderReviewFooter()

	return fmt.Sprintf("%s\n%s\n%s", header, m.viewport.View(), footer)
}

// GetReviewedItems returns all items with their review actions.
func (m ChangesetReviewModel) GetReviewedItems() []ReviewItem {
	return m.items
}

// IsCancelled returns true if the user quit without confirming.
func (m ChangesetReviewModel) IsCancelled() bool {
	return m.cancelled
}

// IsConfirmed returns true if the user confirmed their review.
func (m ChangesetReviewModel) IsConfirmed() bool {
	return m.confirmed
}

// ensureVisible scrolls the viewport to keep the cursor visible.
func (m *ChangesetReviewModel) ensureVisible() {
	lineHeight := 1
	cursorY := m.cursor * lineHeight

	if cursorY < m.viewport.YOffset {
		m.viewport.YOffset = cursorY
	} else if cursorY >= m.viewport.YOffset+m.viewport.Height {
		m.viewport.YOffset = cursorY - m.viewport.Height + 1
	}

	m.updateContent()
}

// updateContent regenerates the viewport content.
func (m *ChangesetReviewModel) updateContent() {
	if !m.ready {
		return
	}

	var content strings.Builder

	for i, item := range m.items {
		content.WriteString(m.renderReviewLine(i, item))
		content.WriteString("\n")
	}

	m.viewport.SetContent(content.String())
}

// renderReviewLine renders a single changeset entry line with action state.
func (m ChangesetReviewModel) renderReviewLine(index int, item ReviewItem) string {
	var actionIcon string
	var actionStyle lipgloss.Style

	switch item.Action {
	case ActionKeep:
		actionIcon = "[✓]"
		actionStyle = lipgloss.NewStyle().Foreground(style.AddedColor)
	case ActionDelete:
		actionIcon = "[✗]"
		actionStyle = lipgloss.NewStyle().Foreground(style.RemovedColor)
	case ActionEdit:
		actionIcon = "[✎]"
		actionStyle = lipgloss.NewStyle().Foreground(style.SecurityColor)
	}

	categoryStyle := getCategoryStyle(item.Entry.Entry.Type)
	lineStyle := lipgloss.NewStyle()

	if index == m.cursor {
		lineStyle = lineStyle.Background(lipgloss.Color("#1f2428"))
		actionStyle = actionStyle.Bold(true)
	}

	typeLabel := fmt.Sprintf("%-8s", item.Entry.Entry.Type)
	scopePart := ""
	if item.Entry.Entry.Scope != "" {
		scopePart = fmt.Sprintf("(%s) ", item.Entry.Entry.Scope)
	}

	maxSummaryLen := max(m.width-40, 20)
	summary := item.Entry.Entry.Summary
	if len(summary) > maxSummaryLen {
		summary = summary[:maxSummaryLen-3] + "..."
	}

	line := fmt.Sprintf("%s %s %s%s",
		actionStyle.Render(actionIcon),
		categoryStyle.Render(typeLabel),
		scopePart,
		summary,
	)

	return lineStyle.Render(line)
}

// renderReviewHeader creates the header showing entry count.
func (m ChangesetReviewModel) renderReviewHeader() string {
	headerStyle := lipgloss.NewStyle().
		Foreground(style.AccentBlue).
		Bold(true).
		Padding(0, 1)

	return headerStyle.Render(
		fmt.Sprintf("Review unreleased changes (%d entries)", len(m.items)),
	)
}

// renderReviewFooter creates the footer with help text and action summary.
func (m ChangesetReviewModel) renderReviewFooter() string {
	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6C7A89")).
		Faint(true).
		Padding(0, 1)

	keepCount := 0
	deleteCount := 0
	editCount := 0

	for _, item := range m.items {
		switch item.Action {
		case ActionKeep:
			keepCount++
		case ActionDelete:
			deleteCount++
		case ActionEdit:
			editCount++
		}
	}

	helpText := "↑/↓: navigate • space: keep • x: delete • e: edit • enter: confirm • q: quit"
	actionInfo := fmt.Sprintf("keep: %d | delete: %d | edit: %d", keepCount, deleteCount, editCount)

	totalWidth := m.width
	helpWidth := lipgloss.Width(helpText)
	actionWidth := lipgloss.Width(actionInfo)
	padding := max(totalWidth-helpWidth-actionWidth-2, 0)

	return footerStyle.Render(
		helpText + strings.Repeat(" ", padding) + actionInfo,
	)
}
