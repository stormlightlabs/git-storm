package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/paginator"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/stormlightlabs/git-storm/internal/diff"
	"github.com/stormlightlabs/git-storm/internal/style"
)

// FileDiff represents a diff for a single file.
type FileDiff struct {
	Edits   []diff.Edit
	OldPath string
	NewPath string
}

// DiffModel holds the state for the side-by-side diff viewer.
type DiffModel struct {
	viewport viewport.Model
	content  string
	ready    bool
	oldPath  string
	newPath  string
}

// keyMap defines keyboard shortcuts for the diff viewer.
type keyMap struct {
	Up       key.Binding
	Down     key.Binding
	PageUp   key.Binding
	PageDown key.Binding
	HalfUp   key.Binding
	HalfDown key.Binding
	Top      key.Binding
	Bottom   key.Binding
	Quit     key.Binding
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	PageUp: key.NewBinding(
		key.WithKeys("pgup", "b"),
		key.WithHelp("pgup/b", "page up"),
	),
	PageDown: key.NewBinding(
		key.WithKeys("pgdown", "f", " "),
		key.WithHelp("pgdn/f/space", "page down"),
	),
	HalfUp: key.NewBinding(
		key.WithKeys("u", "ctrl+u"),
		key.WithHelp("u", "half page up"),
	),
	HalfDown: key.NewBinding(
		key.WithKeys("d", "ctrl+d"),
		key.WithHelp("d", "half page down"),
	),
	Top: key.NewBinding(
		key.WithKeys("home", "g"),
		key.WithHelp("g/home", "top"),
	),
	Bottom: key.NewBinding(
		key.WithKeys("end", "G"),
		key.WithHelp("G/end", "bottom"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "esc", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}

// NewDiffModel creates a new diff viewer model with the given edits.
func NewDiffModel(edits []diff.Edit, oldPath, newPath string, terminalWidth, terminalHeight int) DiffModel {
	formatter := &diff.SideBySideFormatter{
		TerminalWidth:   terminalWidth,
		ShowLineNumbers: true,
	}

	content := formatter.Format(edits)

	vp := viewport.New(terminalWidth, terminalHeight-2)
	vp.SetContent(content)

	return DiffModel{
		viewport: vp,
		content:  content,
		ready:    true,
		oldPath:  oldPath,
		newPath:  newPath,
	}
}

// Init initializes the model (required by Bubble Tea).
func (m DiffModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model state.
func (m DiffModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, keys.Up):
			m.viewport.ScrollUp(1)

		case key.Matches(msg, keys.Down):
			m.viewport.ScrollDown(1)

		case key.Matches(msg, keys.PageUp):
			m.viewport.PageUp()

		case key.Matches(msg, keys.PageDown):
			m.viewport.PageDown()

		case key.Matches(msg, keys.HalfUp):
			m.viewport.HalfPageUp()

		case key.Matches(msg, keys.HalfDown):
			m.viewport.HalfPageDown()

		case key.Matches(msg, keys.Top):
			m.viewport.GotoTop()

		case key.Matches(msg, keys.Bottom):
			m.viewport.GotoBottom()
		}

	case tea.WindowSizeMsg:
		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-2)
			m.viewport.SetContent(m.content)
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - 2
		}
	}

	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// View renders the current view of the diff viewer.
func (m DiffModel) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}

	header := m.renderHeader()
	footer := m.renderFooter()

	return fmt.Sprintf("%s\n%s\n%s", header, m.viewport.View(), footer)
}

// renderHeader creates the header bar showing file paths.
func (m DiffModel) renderHeader() string {
	headerStyle := lipgloss.NewStyle().Foreground(style.AccentBlue).Bold(true).Padding(0, 1)

	oldLabel := lipgloss.NewStyle().Foreground(style.RemovedColor).Render("−")
	newLabel := lipgloss.NewStyle().Foreground(style.AddedColor).Render("+")

	return headerStyle.Render(
		fmt.Sprintf("%s %s  %s %s", oldLabel, m.oldPath, newLabel, m.newPath),
	)
}

// renderFooter creates the footer bar with help text and scroll position.
func (m DiffModel) renderFooter() string {
	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6C7A89")).
		Faint(true).
		Padding(0, 1)

	helpText := "↑/↓: scroll • space/b: page • g/G: top/bottom • q: quit"

	scrollPercent := m.viewport.ScrollPercent()
	scrollInfo := fmt.Sprintf("%.0f%%", scrollPercent*100)

	totalWidth := m.viewport.Width
	helpWidth := lipgloss.Width(helpText)
	scrollWidth := lipgloss.Width(scrollInfo)
	padding := max(totalWidth-helpWidth-scrollWidth-2, 0)

	return footerStyle.Render(
		helpText + strings.Repeat(" ", padding) + scrollInfo,
	)
}

// MultiFileDiffModel holds the state for viewing diffs across multiple files with pagination.
type MultiFileDiffModel struct {
	files     []FileDiff
	paginator paginator.Model
	viewport  viewport.Model
	ready     bool
	width     int
	height    int
	expanded  bool // Controls whether unchanged blocks are compressed
	view      diff.DiffViewKind
}

// NewMultiFileDiffModel creates a new multi-file diff viewer with pagination.
func NewMultiFileDiffModel(files []FileDiff, expanded bool, view diff.DiffViewKind) MultiFileDiffModel {
	p := paginator.New()
	p.Type = paginator.Dots
	p.PerPage = 1
	p.SetTotalPages(len(files))
	p.ActiveDot = lipgloss.NewStyle().Foreground(style.AccentBlue).Render("•")
	p.InactiveDot = lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7A89")).Render("•")

	model := MultiFileDiffModel{
		files:     files,
		paginator: p,
		ready:     false,
		expanded:  expanded,
		view:      view,
	}

	return model
}

// Init initializes the multi-file diff model.
func (m MultiFileDiffModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the multi-file diff model state.
func (m MultiFileDiffModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, key.NewBinding(key.WithKeys("e"))):
			m.expanded = !m.expanded
			m.updateViewport()

		case key.Matches(msg, key.NewBinding(key.WithKeys("left", "h"))):
			m.paginator.PrevPage()
			m.updateViewport()
			m.viewport.GotoTop()

		case key.Matches(msg, key.NewBinding(key.WithKeys("right", "l"))):
			m.paginator.NextPage()
			m.updateViewport()
			m.viewport.GotoTop()

		case key.Matches(msg, keys.Up):
			m.viewport.ScrollUp(1)

		case key.Matches(msg, keys.Down):
			m.viewport.ScrollDown(1)

		case key.Matches(msg, keys.PageUp):
			m.viewport.PageUp()

		case key.Matches(msg, keys.PageDown):
			m.viewport.PageDown()

		case key.Matches(msg, keys.HalfUp):
			m.viewport.HalfPageUp()

		case key.Matches(msg, keys.HalfDown):
			m.viewport.HalfPageDown()

		case key.Matches(msg, keys.Top):
			m.viewport.GotoTop()

		case key.Matches(msg, keys.Bottom):
			m.viewport.GotoBottom()
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-4)
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - 4
		}

		m.updateViewport()
	}

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// View renders the current view of the multi-file diff viewer.
func (m MultiFileDiffModel) View() string {
	if !m.ready || len(m.files) == 0 {
		return "\n  No files to display"
	}

	header := m.renderMultiFileHeader()
	footer := m.renderMultiFileFooter()
	paginatorView := m.renderPaginator()

	return fmt.Sprintf("%s\n%s\n%s\n%s", header, m.viewport.View(), paginatorView, footer)
}

// updateViewport updates the viewport content to show the current file.
func (m *MultiFileDiffModel) updateViewport() {
	if len(m.files) == 0 {
		return
	}

	width := m.width
	if width <= 0 {
		width = 80
	}

	currentFile := m.files[m.paginator.Page]

	var content string

	switch m.view {
	case diff.ViewUnified:
		formatter := &diff.UnifiedFormatter{
			TerminalWidth:   width,
			ShowLineNumbers: true,
			Expanded:        m.expanded,
			EnableWordWrap:  false,
		}
		content = formatter.Format(currentFile.Edits)
	default:
		formatter := &diff.SideBySideFormatter{
			TerminalWidth:   width,
			ShowLineNumbers: true,
			Expanded:        m.expanded,
			EnableWordWrap:  false,
		}
		content = formatter.Format(currentFile.Edits)
	}
	m.viewport.SetContent(content)
}

// renderMultiFileHeader creates the header showing current file paths.
func (m MultiFileDiffModel) renderMultiFileHeader() string {
	if len(m.files) == 0 {
		return ""
	}

	currentFile := m.files[m.paginator.Page]

	headerStyle := lipgloss.NewStyle().
		Foreground(style.AccentBlue).
		Bold(true).
		Padding(0, 1)

	oldLabel := lipgloss.NewStyle().Foreground(style.RemovedColor).Render("−")
	newLabel := lipgloss.NewStyle().Foreground(style.AddedColor).Render("+")

	fileIndicator := fmt.Sprintf("[%d/%d]", m.paginator.Page+1, len(m.files))

	return headerStyle.Render(
		fmt.Sprintf("%s %s %s  %s %s", fileIndicator, oldLabel, currentFile.OldPath, newLabel, currentFile.NewPath),
	)
}

// renderPaginator renders the pagination dots.
func (m MultiFileDiffModel) renderPaginator() string {
	if len(m.files) <= 1 {
		return ""
	}

	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6C7A89")).
		Padding(0, 1).
		Render(m.paginator.View())
}

// renderMultiFileFooter creates the footer with help text and scroll position.
func (m MultiFileDiffModel) renderMultiFileFooter() string {
	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6C7A89")).
		Faint(true).
		Padding(0, 1)

	expandedIndicator := "compressed"
	if m.expanded {
		expandedIndicator = "expanded"
	}

	helpText := fmt.Sprintf("↑/↓: scroll • h/l: files • e: %s • q: quit", expandedIndicator)

	scrollPercent := m.viewport.ScrollPercent()
	scrollInfo := fmt.Sprintf("%.0f%%", scrollPercent*100)

	totalWidth := m.width
	helpWidth := lipgloss.Width(helpText)
	scrollWidth := lipgloss.Width(scrollInfo)
	padding := max(totalWidth-helpWidth-scrollWidth-2, 0)

	return footerStyle.Render(
		helpText + strings.Repeat(" ", padding) + scrollInfo,
	)
}
