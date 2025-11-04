package diff

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/stormlightlabs/git-storm/internal/style"
)

const (
	// Layout constants for side-by-side view
	lineNumWidth = 4
	gutterWidth  = 3
	minPaneWidth = 40
)

// SideBySideFormatter renders diff edits in a split-pane layout with syntax highlighting.
type SideBySideFormatter struct {
	// TerminalWidth is the total available width for rendering
	TerminalWidth int
	// ShowLineNumbers controls whether line numbers are displayed
	ShowLineNumbers bool
}

// Format renders the edits as a styled side-by-side diff string.
//
// The left pane shows the old content (deletions and unchanged lines).
// The right pane shows the new content (insertions and unchanged lines).
// Line numbers and color coding help visualize the changes.
func (f *SideBySideFormatter) Format(edits []Edit) string {
	if len(edits) == 0 {
		return style.StyleText.Render("No changes")
	}

	paneWidth := f.calculatePaneWidth()

	var sb strings.Builder
	lineNumStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7A89")).Faint(true)

	for _, edit := range edits {
		left, right := f.renderEdit(edit, paneWidth)

		if f.ShowLineNumbers {
			leftNum := f.formatLineNum(edit.AIndex, lineNumStyle)
			rightNum := f.formatLineNum(edit.BIndex, lineNumStyle)

			sb.WriteString(leftNum)
			sb.WriteString(left)
			sb.WriteString(f.renderGutter(edit.Kind))
			sb.WriteString(rightNum)
			sb.WriteString(right)
		} else {
			sb.WriteString(left)
			sb.WriteString(f.renderGutter(edit.Kind))
			sb.WriteString(right)
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// calculatePaneWidth determines the width available for each content pane.
func (f *SideBySideFormatter) calculatePaneWidth() int {
	usedWidth := gutterWidth
	if f.ShowLineNumbers {
		usedWidth += 2 * lineNumWidth
	}

	availableWidth := f.TerminalWidth - usedWidth
	paneWidth := availableWidth / 2

	if paneWidth < minPaneWidth {
		paneWidth = minPaneWidth
	}

	return paneWidth
}

// renderEdit formats a single edit operation for both left and right panes.
func (f *SideBySideFormatter) renderEdit(edit Edit, paneWidth int) (left, right string) {
	content := f.truncateContent(edit.Content, paneWidth)

	switch edit.Kind {
	case Equal:
		// Show on both sides with neutral styling
		leftStyled := style.StyleText.Width(paneWidth).Render(content)
		rightStyled := style.StyleText.Width(paneWidth).Render(content)
		return leftStyled, rightStyled

	case Delete:
		// Show on left in red, empty right
		leftStyled := style.StyleRemoved.Width(paneWidth).Render(content)
		rightStyled := lipgloss.NewStyle().Width(paneWidth).Render("")
		return leftStyled, rightStyled

	case Insert:
		// Empty left, show on right in green
		leftStyled := lipgloss.NewStyle().Width(paneWidth).Render("")
		rightStyled := style.StyleAdded.Width(paneWidth).Render(content)
		return leftStyled, rightStyled

	default:
		// Fallback for unknown edit kinds
		return lipgloss.NewStyle().Width(paneWidth).Render(content),
			lipgloss.NewStyle().Width(paneWidth).Render(content)
	}
}

// renderGutter creates the visual separator between left and right panes.
func (f *SideBySideFormatter) renderGutter(kind EditKind) string {
	var symbol string
	var st lipgloss.Style

	switch kind {
	case Equal:
		symbol = " │ "
		st = lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7A89"))
	case Delete:
		symbol = " < "
		st = style.StyleRemoved
	case Insert:
		symbol = " > "
		st = style.StyleAdded
	default:
		symbol = " │ "
		st = lipgloss.NewStyle()
	}

	return st.Render(symbol)
}

// formatLineNum renders a line number with styling.
func (f *SideBySideFormatter) formatLineNum(index int, st lipgloss.Style) string {
	if index < 0 {
		return st.Width(lineNumWidth).Render("")
	}
	return st.Width(lineNumWidth).Render(fmt.Sprintf("%4d", index+1))
}

// truncateContent ensures content fits within the pane width.
func (f *SideBySideFormatter) truncateContent(content string, maxWidth int) string {
	// Remove trailing whitespace but preserve leading indentation
	content = strings.TrimRight(content, " \t\r\n")

	if len(content) <= maxWidth {
		return content
	}

	if maxWidth <= 3 {
		return content[:maxWidth]
	}

	return content[:maxWidth-3] + "..."
}
