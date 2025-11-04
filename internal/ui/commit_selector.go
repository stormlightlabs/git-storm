package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-git/go-git/v6/plumbing/object"
	"github.com/stormlightlabs/git-storm/internal/gitlog"
	"github.com/stormlightlabs/git-storm/internal/style"
)

// CommitItem wraps a commit with its selection state and parsed metadata.
type CommitItem struct {
	Commit   *object.Commit
	Meta     gitlog.CommitMeta
	Category string
	Selected bool
}

// CommitSelectorModel holds the state for the interactive commit selector TUI.
type CommitSelectorModel struct {
	viewport  viewport.Model
	items     []CommitItem
	cursor    int
	ready     bool
	fromRef   string
	toRef     string
	width     int
	height    int
	confirmed bool
	cancelled bool
}

// commitSelectorKeyMap defines keyboard shortcuts for the commit selector.
type commitSelectorKeyMap struct {
	Up          key.Binding
	Down        key.Binding
	PageUp      key.Binding
	PageDown    key.Binding
	Top         key.Binding
	Bottom      key.Binding
	Toggle      key.Binding
	SelectAll   key.Binding
	DeselectAll key.Binding
	Confirm     key.Binding
	Quit        key.Binding
}

var commitKeys = commitSelectorKeyMap{
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
	Toggle: key.NewBinding(
		key.WithKeys(" "),
		key.WithHelp("space", "toggle"),
	),
	SelectAll: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "select all"),
	),
	DeselectAll: key.NewBinding(
		key.WithKeys("A"),
		key.WithHelp("A", "deselect all"),
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

// NewCommitSelectorModel creates a new commit selector model.
func NewCommitSelectorModel(commits []*object.Commit, fromRef, toRef string, parser gitlog.CommitParser) CommitSelectorModel {
	items := make([]CommitItem, 0, len(commits))

	for _, commit := range commits {
		subject := commit.Message
		body := ""
		lines := strings.Split(commit.Message, "\n")
		if len(lines) > 0 {
			subject = lines[0]
			if len(lines) > 1 {
				body = strings.Join(lines[1:], "\n")
			}
		}

		meta, err := parser.Parse(commit.Hash.String(), subject, body, commit.Author.When)
		if err != nil {
			meta = gitlog.CommitMeta{
				Type:        "unknown",
				Description: subject,
				Body:        body,
			}
		}

		category := parser.Categorize(meta)

		items = append(items, CommitItem{
			Commit:   commit,
			Meta:     meta,
			Category: category,
			Selected: category != "",
		})
	}

	return CommitSelectorModel{
		items:   items,
		cursor:  0,
		fromRef: fromRef,
		toRef:   toRef,
		ready:   false,
	}
}

// Init initializes the model (required by Bubble Tea).
func (m CommitSelectorModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model state.
func (m CommitSelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, commitKeys.Quit):
			m.cancelled = true
			return m, tea.Quit

		case key.Matches(msg, commitKeys.Confirm):
			m.confirmed = true
			return m, tea.Quit

		case key.Matches(msg, commitKeys.Up):
			if m.cursor > 0 {
				m.cursor--
				m.ensureVisible()
			}

		case key.Matches(msg, commitKeys.Down):
			if m.cursor < len(m.items)-1 {
				m.cursor++
				m.ensureVisible()
			}

		case key.Matches(msg, commitKeys.PageUp):
			m.cursor -= m.viewport.Height
			if m.cursor < 0 {
				m.cursor = 0
			}
			m.ensureVisible()

		case key.Matches(msg, commitKeys.PageDown):
			m.cursor += m.viewport.Height
			if m.cursor >= len(m.items) {
				m.cursor = len(m.items) - 1
			}
			m.ensureVisible()

		case key.Matches(msg, commitKeys.Top):
			m.cursor = 0
			m.ensureVisible()

		case key.Matches(msg, commitKeys.Bottom):
			m.cursor = len(m.items) - 1
			m.ensureVisible()

		case key.Matches(msg, commitKeys.Toggle):
			if m.cursor >= 0 && m.cursor < len(m.items) {
				m.items[m.cursor].Selected = !m.items[m.cursor].Selected
				m.updateContent()
			}

		case key.Matches(msg, commitKeys.SelectAll):
			for i := range m.items {
				m.items[i].Selected = true
			}
			m.updateContent()

		case key.Matches(msg, commitKeys.DeselectAll):
			for i := range m.items {
				m.items[i].Selected = false
			}
			m.updateContent()
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

// View renders the current view of the commit selector.
func (m CommitSelectorModel) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}

	header := m.renderCommitHeader()
	footer := m.renderCommitFooter()

	return fmt.Sprintf("%s\n%s\n%s", header, m.viewport.View(), footer)
}

// GetSelectedCommits returns the list of selected commits.
func (m CommitSelectorModel) GetSelectedCommits() []*object.Commit {
	selected := make([]*object.Commit, 0)
	for _, item := range m.items {
		if item.Selected {
			selected = append(selected, item.Commit)
		}
	}
	return selected
}

// GetSelectedItems returns the list of selected commit items with metadata.
func (m CommitSelectorModel) GetSelectedItems() []CommitItem {
	selected := make([]CommitItem, 0)
	for _, item := range m.items {
		if item.Selected {
			selected = append(selected, item)
		}
	}
	return selected
}

// IsCancelled returns true if the user quit without confirming.
func (m CommitSelectorModel) IsCancelled() bool {
	return m.cancelled
}

// IsConfirmed returns true if the user confirmed their selection.
func (m CommitSelectorModel) IsConfirmed() bool {
	return m.confirmed
}

// ensureVisible scrolls the viewport to keep the cursor visible.
func (m *CommitSelectorModel) ensureVisible() {
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
func (m *CommitSelectorModel) updateContent() {
	if !m.ready {
		return
	}

	var content strings.Builder

	for i, item := range m.items {
		content.WriteString(m.renderCommitLine(i, item))
		content.WriteString("\n")
	}

	m.viewport.SetContent(content.String())
}

// renderCommitLine renders a single commit line with selection state.
func (m CommitSelectorModel) renderCommitLine(index int, item CommitItem) string {
	checkbox := "[ ]"
	if item.Selected {
		checkbox = "[✓]"
	}

	shortHash := item.Commit.Hash.String()[:7]
	subject := item.Meta.Description
	if subject == "" {
		subject = strings.Split(item.Commit.Message, "\n")[0]
	}

	maxSubjectLen := max(m.width-60, 20)
	if len(subject) > maxSubjectLen {
		subject = subject[:maxSubjectLen-3] + "..."
	}

	author := item.Commit.Author.Name
	if len(author) > 15 {
		author = author[:12] + "..."
	}

	timeAgo := fmtTimeAgo(item.Commit.Author.When)

	category := item.Category
	if category == "" {
		category = "skip"
	}

	categoryStyle := getCategoryStyle(category)
	lineStyle := lipgloss.NewStyle()
	checkboxStyle := lipgloss.NewStyle().Foreground(style.AccentBlue)
	hashStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7A89"))

	if index == m.cursor {
		lineStyle = lineStyle.Background(lipgloss.Color("#1f2428"))
		checkboxStyle = checkboxStyle.Bold(true)
	}

	line := fmt.Sprintf("%s %s %s %s %s %s",
		checkboxStyle.Render(checkbox),
		hashStyle.Render(shortHash),
		categoryStyle.Render(fmt.Sprintf("%-8s", category)),
		subject,
		lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7A89")).Render(author),
		lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7A89")).Faint(true).Render(timeAgo),
	)

	return lineStyle.Render(line)
}

// renderCommitHeader creates the header showing the range.
func (m CommitSelectorModel) renderCommitHeader() string {
	headerStyle := lipgloss.NewStyle().
		Foreground(style.AccentBlue).
		Bold(true).
		Padding(0, 1)

	return headerStyle.Render(
		fmt.Sprintf("Select commits to include (%s..%s)", m.fromRef, m.toRef),
	)
}

// renderCommitFooter creates the footer with help text and selection count.
func (m CommitSelectorModel) renderCommitFooter() string {
	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6C7A89")).
		Faint(true).
		Padding(0, 1)

	selectedCount := 0
	for _, item := range m.items {
		if item.Selected {
			selectedCount++
		}
	}

	helpText := "↑/↓: navigate • space: toggle • a/A: select/deselect all • enter: confirm • q: quit"
	selectionInfo := fmt.Sprintf("%d/%d selected", selectedCount, len(m.items))

	totalWidth := m.width
	helpWidth := lipgloss.Width(helpText)
	selWidth := lipgloss.Width(selectionInfo)
	padding := max(totalWidth-helpWidth-selWidth-2, 0)

	return footerStyle.Render(
		helpText + strings.Repeat(" ", padding) + selectionInfo,
	)
}

// fmtTimeAgo returns a human-readable relative time string.
func fmtTimeAgo(t time.Time) string {
	duration := time.Since(t)

	if duration < time.Minute {
		return "just now"
	} else if duration < time.Hour {
		minutes := int(duration.Minutes())
		return fmt.Sprintf("%dm ago", minutes)
	} else if duration < 24*time.Hour {
		hours := int(duration.Hours())
		return fmt.Sprintf("%dh ago", hours)
	} else if duration < 30*24*time.Hour {
		days := int(duration.Hours() / 24)
		return fmt.Sprintf("%dd ago", days)
	} else if duration < 365*24*time.Hour {
		months := int(duration.Hours() / 24 / 30)
		return fmt.Sprintf("%dmo ago", months)
	} else {
		years := int(duration.Hours() / 24 / 365)
		return fmt.Sprintf("%dy ago", years)
	}
}

func getCategoryStyle(c string) lipgloss.Style {
	s := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7A89"))
	switch c {
	case "added":
		s = lipgloss.NewStyle().Foreground(style.AddedColor)
	case "changed":
		s = lipgloss.NewStyle().Foreground(style.ChangedColor)
	case "fixed":
		s = lipgloss.NewStyle().Foreground(style.AccentBlue)
	case "removed":
		s = lipgloss.NewStyle().Foreground(style.RemovedColor)
	case "security":
		s = lipgloss.NewStyle().Foreground(lipgloss.Color("#BF616A"))
	}
	return s
}
