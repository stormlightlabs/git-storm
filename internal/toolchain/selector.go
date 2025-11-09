package toolchain

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/stormlightlabs/git-storm/internal/style"
	"golang.org/x/term"
)

var (
	cursorStyle   = lipgloss.NewStyle().Foreground(style.AccentBlue).Bold(true)
	selectedStyle = lipgloss.NewStyle().Foreground(style.AddedColor)
	mutedStyle    = lipgloss.NewStyle().Foreground(style.AccentSteel)
)

// SelectManifests launches an interactive selector so users can pick which manifests to update.
func SelectManifests(manifests []Manifest) ([]Manifest, error) {
	if len(manifests) == 0 {
		return nil, fmt.Errorf("no toolchain manifests detected")
	}
	if !isTTY() {
		return nil, fmt.Errorf("interactive selection requires a TTY; pass specific --toolchain paths instead")
	}

	program := tea.NewProgram(newSelectorModel(manifests))
	finalModel, err := program.Run()
	if err != nil {
		return nil, err
	}

	model, ok := finalModel.(selectorModel)
	if !ok {
		return nil, fmt.Errorf("unexpected selector model type %T", finalModel)
	}
	if model.cancelled {
		return nil, fmt.Errorf("selection cancelled")
	}
	return model.selectedManifests(), nil
}

func isTTY() bool {
	return term.IsTerminal(int(os.Stdout.Fd())) && term.IsTerminal(int(os.Stdin.Fd()))
}

type selectorModel struct {
	manifests []Manifest
	cursor    int
	selected  map[int]struct{}
	done      bool
	cancelled bool
}

func newSelectorModel(manifests []Manifest) selectorModel {
	return selectorModel{
		manifests: manifests,
		selected:  make(map[int]struct{}),
	}
}

func (m selectorModel) Init() tea.Cmd { return nil }

func (m selectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			m.cancelled = true
			return m, tea.Quit
		case "enter":
			if len(m.selected) == 0 && len(m.manifests) > 0 {
				m.selected[m.cursor] = struct{}{}
			}
			m.done = true
			return m, tea.Quit
		case " ":
			if _, ok := m.selected[m.cursor]; ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}
		case "j", "down":
			if m.cursor < len(m.manifests)-1 {
				m.cursor++
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "home", "g":
			m.cursor = 0
		case "end", "G":
			if len(m.manifests) > 0 {
				m.cursor = len(m.manifests) - 1
			}
		}
	}

	return m, nil
}

func (m selectorModel) View() string {
	if len(m.manifests) == 0 {
		return "No manifests available"
	}

	var view strings.Builder
	view.WriteString(style.StyleHeadline.Render("Select toolchain manifests to bump"))
	view.WriteString("\n\n")

	for i, manifest := range m.manifests {
		cursor := " "
		if i == m.cursor {
			cursor = cursorStyle.Render("›")
		}
		checkbox := "[ ]"
		if _, ok := m.selected[i]; ok {
			checkbox = selectedStyle.Render("[x]")
		}
		line := fmt.Sprintf("%s %s %s", cursor, checkbox, manifest.DisplayLabel())
		view.WriteString(line)
		view.WriteString("\n")
	}

	view.WriteString("\n")
	view.WriteString(mutedStyle.Render("space: toggle • enter: confirm • q: cancel"))
	return view.String()
}

func (m selectorModel) selectedManifests() []Manifest {
	if len(m.manifests) == 0 {
		return nil
	}
	var chosen []Manifest
	for idx, manifest := range m.manifests {
		if _, ok := m.selected[idx]; ok {
			chosen = append(chosen, manifest)
		}
	}
	if len(chosen) == 0 && !m.cancelled && len(m.manifests) > 0 {
		return []Manifest{m.manifests[m.cursor]}
	}
	return chosen
}
