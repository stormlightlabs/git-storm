package toolchain

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stormlightlabs/git-storm/internal/testutils"
)

func TestSelectorModel_DefaultSelection(t *testing.T) {
	manifests := []Manifest{{RelPath: "Cargo.toml"}, {RelPath: "package.json"}}
	model := newSelectorModel(manifests)
	final := testutils.RunModelWithInteraction(t, model, []tea.Msg{tea.KeyMsg{Type: tea.KeyEnter}})
	selector := final.(selectorModel)
	selected := selector.selectedManifests()
	if len(selected) != 1 || selected[0].RelPath != "Cargo.toml" {
		t.Fatalf("expected first manifest to be selected, got %#v", selected)
	}
}

func TestSelectorModel_ToggleSelection(t *testing.T) {
	manifests := []Manifest{{RelPath: "Cargo.toml"}, {RelPath: "package.json"}}
	model := newSelectorModel(manifests)
	msgs := []tea.Msg{
		tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}},
		tea.KeyMsg{Type: tea.KeySpace},
		tea.KeyMsg{Type: tea.KeyEnter},
	}
	final := testutils.RunModelWithInteraction(t, model, msgs)
	selector := final.(selectorModel)
	selected := selector.selectedManifests()
	if len(selected) != 1 || selected[0].RelPath != "package.json" {
		t.Fatalf("expected second manifest to be selected, got %#v", selected)
	}
}
