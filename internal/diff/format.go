package diff

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
	"github.com/stormlightlabs/git-storm/internal/style"
)

const (
	SymbolAdd          = "┃" // addition
	SymbolChange       = "▎" // modification/change
	SymbolDeleteLine   = "_" // line removed
	SymbolTopDelete    = "‾" // deletion at top (overline)
	SymbolChangeDelete = "~" // change + delete (hunk combined)
	SymbolUntracked    = "┆" // untracked lines/files

	AsciiSymbolAdd          = "|" // addition
	AsciiSymbolChange       = "|" // modification (same as add fallback)
	AsciiSymbolDeleteLine   = "-" // deletion line
	AsciiSymbolTopDelete    = "^" // “top delete” fallback
	AsciiSymbolChangeDelete = "~" // change+delete still ~
	AsciiSymbolUntracked    = ":" // untracked fallback

	lineNumWidth        = 4
	gutterWidth         = 3
	minPaneWidth        = 40
	contextLines        = 3  // Lines to show before/after changes
	minUnchangedToHide  = 10 // Minimum unchanged lines before hiding
	compressedIndicator = "⋮"
)

// SideBySideFormatter renders diff edits in a split-pane layout with syntax highlighting.
type SideBySideFormatter struct {
	// TerminalWidth is the total available width for rendering
	TerminalWidth int
	// ShowLineNumbers controls whether line numbers are displayed
	ShowLineNumbers bool
	// Expanded controls whether to show all unchanged lines or compress them
	Expanded bool
	// EnableWordWrap enables word wrapping for long lines
	EnableWordWrap bool
}

// Format renders the edits as a styled side-by-side diff string.
//
// The left pane shows the old content (deletions and unchanged lines).
// The right pane shows the new content (insertions and unchanged lines).
func (f *SideBySideFormatter) Format(edits []Edit) string {
	if len(edits) == 0 {
		return style.StyleText.Render("No changes")
	}

	processedEdits := MergeReplacements(edits)

	if !f.Expanded {
		processedEdits = f.compressUnchangedBlocks(processedEdits)
	}

	paneWidth := f.calculatePaneWidth()

	var sb strings.Builder
	lineNumStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7A89")).Faint(true)

	for _, edit := range processedEdits {
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
	if availableWidth < 0 {
		availableWidth = 0
	}

	paneWidth := availableWidth / 2

	if paneWidth < minPaneWidth {
		totalNeeded := usedWidth + (2 * minPaneWidth)
		if totalNeeded > f.TerminalWidth {
			return paneWidth
		}
		return minPaneWidth
	}

	return paneWidth
}

// renderEdit formats a single edit operation for both left and right panes.
func (f *SideBySideFormatter) renderEdit(edit Edit, paneWidth int) (left, right string) {
	content := detab(edit.Content, 8)
	content = f.truncateContent(content, paneWidth)

	if edit.AIndex == -2 && edit.BIndex == -2 {
		compressedStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6C7A89")).
			Faint(true).
			Italic(true)
		styled := f.padToWidth(compressedStyle.Render(content), paneWidth)
		return styled, styled
	}

	switch edit.Kind {
	case Equal:
		leftStyled := f.padToWidth(style.StyleText.Render(content), paneWidth)
		rightStyled := f.padToWidth(style.StyleText.Render(content), paneWidth)
		return leftStyled, rightStyled

	case Delete:
		leftStyled := f.padToWidth(style.StyleRemoved.Render(content), paneWidth)
		rightStyled := f.padToWidth("", paneWidth)
		return leftStyled, rightStyled

	case Insert:
		leftStyled := f.padToWidth("", paneWidth)
		rightStyled := f.padToWidth(style.StyleAdded.Render(content), paneWidth)
		return leftStyled, rightStyled

	case Replace:
		newContent := detab(edit.NewContent, 8)
		newContent = f.truncateContent(newContent, paneWidth)
		leftStyled := f.padToWidth(style.StyleRemoved.Render(content), paneWidth)
		rightStyled := f.padToWidth(style.StyleAdded.Render(newContent), paneWidth)
		return leftStyled, rightStyled

	default:
		return f.padToWidth(content, paneWidth),
			f.padToWidth(content, paneWidth)
	}
}

// padToWidth pads a string with spaces to reach the target width.
// If the string exceeds the target width, it truncates it.
func (f *SideBySideFormatter) padToWidth(s string, targetWidth int) string {
	currentWidth := lipgloss.Width(s)

	if currentWidth > targetWidth {
		return truncateToWidth(s, targetWidth)
	}

	if currentWidth == targetWidth {
		return s
	}

	padding := strings.Repeat(" ", targetWidth-currentWidth)
	return s + padding
}

// renderGutter creates the visual separator between left and right panes.
func (f *SideBySideFormatter) renderGutter(kind EditKind) string {
	var symbol string
	var st lipgloss.Style

	switch kind {
	case Equal:
		symbol = " " + SymbolUntracked + " "
		st = lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7A89"))
	case Delete:
		symbol = " " + SymbolDeleteLine + " "
		st = style.StyleRemoved
	case Insert:
		symbol = " " + SymbolAdd + " "
		st = style.StyleAdded
	case Replace:
		symbol = " " + SymbolChange + " "
		st = style.StyleChanged
	default:
		symbol = " " + SymbolUntracked + " "
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

// truncateContent ensures content fits within the pane width using proper display width.
func (f *SideBySideFormatter) truncateContent(content string, maxWidth int) string {
	content = strings.TrimRight(content, " \t\r\n")

	if f.EnableWordWrap {
		wrapped := wordwrap.String(content, maxWidth)
		lines := strings.Split(wrapped, "\n")
		if len(lines) > 0 {
			return lines[0]
		}
		return wrapped
	}

	displayWidth := lipgloss.Width(content)

	if displayWidth <= maxWidth {
		return content
	}

	if maxWidth <= 3 {
		return truncateToWidth(content, maxWidth)
	}

	targetWidth := maxWidth - 3
	truncated := truncateToWidth(content, targetWidth)
	return truncated + "..."
}

// truncateToWidth truncates a string to a specific display width.
func truncateToWidth(s string, width int) string {
	if width <= 0 {
		return ""
	}

	var result strings.Builder
	currentWidth := 0

	for _, r := range s {
		runeWidth := lipgloss.Width(string(r))

		if currentWidth+runeWidth > width {
			break
		}

		result.WriteRune(r)
		currentWidth += runeWidth
	}

	return result.String()
}

// compressUnchangedBlocks compresses large blocks of unchanged lines.
//
// It keeps contextLines before and after changes, and replaces large
// blocks of unchanged lines with a single compressed indicator.
func (f *SideBySideFormatter) compressUnchangedBlocks(edits []Edit) []Edit {
	if len(edits) == 0 {
		return edits
	}

	var result []Edit
	var unchangedRun []Edit

	for i, edit := range edits {
		if edit.Kind == Equal {
			unchangedRun = append(unchangedRun, edit)

			isLast := i == len(edits)-1
			nextIsChanged := !isLast && edits[i+1].Kind != Equal

			if isLast || nextIsChanged {
				if len(unchangedRun) >= minUnchangedToHide {
					for j := 0; j < contextLines && j < len(unchangedRun); j++ {
						result = append(result, unchangedRun[j])
					}

					hiddenCount := len(unchangedRun) - (2 * contextLines)
					if hiddenCount > 0 {
						result = append(result, Edit{
							Kind:    Equal,
							AIndex:  -2,
							BIndex:  -2,
							Content: fmt.Sprintf("%s %d unchanged lines", compressedIndicator, hiddenCount),
						})
					}

					start := max(len(unchangedRun)-contextLines, contextLines)
					for j := start; j < len(unchangedRun); j++ {
						result = append(result, unchangedRun[j])
					}
				} else {
					result = append(result, unchangedRun...)
				}
				unchangedRun = nil
			}
		} else {
			if len(unchangedRun) > 0 {
				if len(unchangedRun) >= minUnchangedToHide {
					for j := 0; j < contextLines && j < len(unchangedRun); j++ {
						result = append(result, unchangedRun[j])
					}

					hiddenCount := len(unchangedRun) - (2 * contextLines)
					if hiddenCount > 0 {
						result = append(result, Edit{
							Kind:    Equal,
							AIndex:  -2,
							BIndex:  -2,
							Content: fmt.Sprintf("%s %d unchanged lines", compressedIndicator, hiddenCount),
						})
					}

					start := max(len(unchangedRun)-contextLines, contextLines)
					for j := start; j < len(unchangedRun); j++ {
						result = append(result, unchangedRun[j])
					}
				} else {
					result = append(result, unchangedRun...)
				}
				unchangedRun = nil
			}

			result = append(result, edit)
		}
	}

	return result
}

// detab replaces tabs with spaces so alignment stays consistent across terminals.
func detab(s string, tabWidth int) string {
	if tabWidth <= 0 {
		tabWidth = 4
	}
	return strings.ReplaceAll(s, "\t", strings.Repeat(" ", tabWidth))
}

// UnifiedFormatter renders diff edits in a traditional unified diff layout.
type UnifiedFormatter struct {
	// TerminalWidth is the total available width for rendering
	TerminalWidth int
	// ShowLineNumbers controls whether line numbers are displayed
	ShowLineNumbers bool
	// Expanded controls whether to show all unchanged lines or compress them
	Expanded bool
	// EnableWordWrap enables word wrapping for long lines
	EnableWordWrap bool
}

// Format renders the edits as a styled unified diff string.
//
// The output shows deletions with "-" prefix, insertions with "+" prefix, and unchanged lines with " " prefix.
func (f *UnifiedFormatter) Format(edits []Edit) string {
	if len(edits) == 0 {
		return style.StyleText.Render("No changes")
	}

	processedEdits := MergeReplacements(edits)

	if !f.Expanded {
		processedEdits = f.compressUnchangedBlocks(processedEdits)
	}

	contentWidth := f.calculateContentWidth()

	var sb strings.Builder
	lineNumStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7A89")).Faint(true)

	for _, edit := range processedEdits {
		line := f.renderEdit(edit, contentWidth, lineNumStyle)
		sb.WriteString(line)
		sb.WriteString("\n")

		if edit.Kind == Replace {
			newLine := f.renderReplaceNew(edit, contentWidth, lineNumStyle)
			sb.WriteString(newLine)
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// calculateContentWidth determines the width available for content.
func (f *UnifiedFormatter) calculateContentWidth() int {
	usedWidth := 2
	if f.ShowLineNumbers {
		usedWidth += 2*lineNumWidth + 2
	}
	return max(f.TerminalWidth-usedWidth, minPaneWidth)
}

// renderEdit formats a single edit operation.
func (f *UnifiedFormatter) renderEdit(edit Edit, contentWidth int, lineNumStyle lipgloss.Style) string {
	var sb strings.Builder

	if edit.AIndex == -2 && edit.BIndex == -2 {
		compressedStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6C7A89")).
			Faint(true).
			Italic(true)
		if f.ShowLineNumbers {
			sb.WriteString(lineNumStyle.Width(lineNumWidth).Render(""))
			sb.WriteString(" ")
			sb.WriteString(lineNumStyle.Width(lineNumWidth).Render(""))
			sb.WriteString(" ")
		}
		sb.WriteString(compressedStyle.Render(edit.Content))
		return sb.String()
	}

	if f.ShowLineNumbers {
		oldNum := f.formatLineNum(edit.AIndex, lineNumStyle)
		newNum := f.formatLineNum(edit.BIndex, lineNumStyle)
		sb.WriteString(oldNum)
		sb.WriteString(" ")
		sb.WriteString(newNum)
		sb.WriteString(" ")
	}

	content := detab(edit.Content, 8)
	content = f.truncateContent(content, contentWidth)

	switch edit.Kind {
	case Equal:
		sb.WriteString(style.StyleText.Render(" " + content))
	case Delete:
		sb.WriteString(style.StyleRemoved.Render("-" + content))
	case Insert:
		sb.WriteString(style.StyleAdded.Render("+" + content))
	case Replace:
		sb.WriteString(style.StyleRemoved.Render("-" + content))
	default:
		sb.WriteString(" " + content)
	}

	return sb.String()
}

// renderReplaceNew renders the new content line for a Replace operation.
func (f *UnifiedFormatter) renderReplaceNew(edit Edit, contentWidth int, lineNumStyle lipgloss.Style) string {
	var sb strings.Builder

	if f.ShowLineNumbers {
		sb.WriteString(lineNumStyle.Width(lineNumWidth).Render(""))
		sb.WriteString(" ")
		sb.WriteString(f.formatLineNum(edit.BIndex, lineNumStyle))
		sb.WriteString(" ")
	}

	content := detab(edit.NewContent, 8)
	content = f.truncateContent(content, contentWidth)
	sb.WriteString(style.StyleAdded.Render("+" + content))

	return sb.String()
}

// formatLineNum renders a line number with styling.
func (f *UnifiedFormatter) formatLineNum(index int, st lipgloss.Style) string {
	if index < 0 {
		return st.Width(lineNumWidth).Render("")
	}
	return st.Width(lineNumWidth).Render(fmt.Sprintf("%4d", index+1))
}

// truncateContent ensures content fits within the available width.
func (f *UnifiedFormatter) truncateContent(content string, maxWidth int) string {
	content = strings.TrimRight(content, " \t\r\n")

	if f.EnableWordWrap {
		wrapped := wordwrap.String(content, maxWidth)
		lines := strings.Split(wrapped, "\n")
		if len(lines) > 0 {
			return lines[0]
		}
		return wrapped
	}

	displayWidth := lipgloss.Width(content)

	if displayWidth <= maxWidth {
		return content
	}

	if maxWidth <= 3 {
		return truncateToWidth(content, maxWidth)
	}

	return truncateToWidth(content, maxWidth-3) + "..."
}

// compressUnchangedBlocks compresses large blocks of unchanged lines.
func (f *UnifiedFormatter) compressUnchangedBlocks(edits []Edit) []Edit {
	if len(edits) == 0 {
		return edits
	}

	var result []Edit
	var unchangedRun []Edit

	for i, edit := range edits {
		if edit.Kind == Equal {
			unchangedRun = append(unchangedRun, edit)

			isLast := i == len(edits)-1
			nextIsChanged := !isLast && edits[i+1].Kind != Equal

			if isLast || nextIsChanged {
				if len(unchangedRun) >= minUnchangedToHide {
					for j := 0; j < contextLines && j < len(unchangedRun); j++ {
						result = append(result, unchangedRun[j])
					}

					hiddenCount := len(unchangedRun) - (2 * contextLines)
					if hiddenCount > 0 {
						result = append(result, Edit{
							Kind:    Equal,
							AIndex:  -2,
							BIndex:  -2,
							Content: fmt.Sprintf("%s %d unchanged lines", compressedIndicator, hiddenCount),
						})
					}

					start := max(len(unchangedRun)-contextLines, contextLines)
					for j := start; j < len(unchangedRun); j++ {
						result = append(result, unchangedRun[j])
					}
				} else {
					result = append(result, unchangedRun...)
				}
				unchangedRun = nil
			}
		} else {
			if len(unchangedRun) > 0 {
				if len(unchangedRun) >= minUnchangedToHide {
					for j := 0; j < contextLines && j < len(unchangedRun); j++ {
						result = append(result, unchangedRun[j])
					}

					hiddenCount := len(unchangedRun) - (2 * contextLines)
					if hiddenCount > 0 {
						result = append(result, Edit{
							Kind:    Equal,
							AIndex:  -2,
							BIndex:  -2,
							Content: fmt.Sprintf("%s %d unchanged lines", compressedIndicator, hiddenCount),
						})
					}

					start := max(len(unchangedRun)-contextLines, contextLines)
					for j := start; j < len(unchangedRun); j++ {
						result = append(result, unchangedRun[j])
					}
				} else {
					result = append(result, unchangedRun...)
				}
				unchangedRun = nil
			}

			result = append(result, edit)
		}
	}

	return result
}
