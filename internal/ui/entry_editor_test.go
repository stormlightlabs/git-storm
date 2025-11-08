package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stormlightlabs/git-storm/internal/changeset"
)

func TestEntryEditorModel_Init(t *testing.T) {
	entry := changeset.EntryWithFile{
		Entry: changeset.Entry{
			Type:    "added",
			Scope:   "cli",
			Summary: "Test entry",
		},
		Filename: "test.md",
	}

	model := NewEntryEditorModel(entry)

	cmd := model.Init()
	if cmd == nil {
		t.Error("Init() should return textinput.Blink command")
	}
}

func TestEntryEditorModel_DefaultState(t *testing.T) {
	entry := changeset.EntryWithFile{
		Entry: changeset.Entry{
			Type:    "added",
			Scope:   "cli",
			Summary: "Test entry",
		},
		Filename: "test.md",
	}

	model := NewEntryEditorModel(entry)

	if model.confirmed {
		t.Error("Model should not be confirmed initially")
	}

	if model.cancelled {
		t.Error("Model should not be cancelled initially")
	}

	if model.focusIdx != 0 {
		t.Errorf("Focus should be on first input, got %d", model.focusIdx)
	}

	if model.typeIdx != 0 {
		t.Errorf("Type index should be 0 for 'added', got %d", model.typeIdx)
	}
}

func TestEntryEditorModel_TypeCycling(t *testing.T) {
	entry := changeset.EntryWithFile{
		Entry: changeset.Entry{
			Type:    "added",
			Scope:   "cli",
			Summary: "Test entry",
		},
		Filename: "test.md",
	}

	model := NewEntryEditorModel(entry)

	initialType := model.typeIdx

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyCtrlT})
	model = updated.(EntryEditorModel)

	if model.typeIdx == initialType {
		t.Error("Type should have cycled to next value")
	}

	expectedNext := (initialType + 1) % len(validTypes)
	if model.typeIdx != expectedNext {
		t.Errorf("Type index should be %d, got %d", expectedNext, model.typeIdx)
	}
}

func TestEntryEditorModel_Confirm(t *testing.T) {
	entry := changeset.EntryWithFile{
		Entry: changeset.Entry{
			Type:    "added",
			Scope:   "cli",
			Summary: "Test entry",
		},
		Filename: "test.md",
	}

	model := NewEntryEditorModel(entry)

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(EntryEditorModel)

	if !model.confirmed {
		t.Error("Model should be confirmed after pressing Enter")
	}

	if cmd == nil {
		t.Error("Confirm should return tea.Quit command")
	}
}

func TestEntryEditorModel_ConfirmWithCtrlS(t *testing.T) {
	entry := changeset.EntryWithFile{
		Entry: changeset.Entry{
			Type:    "added",
			Scope:   "cli",
			Summary: "Test entry",
		},
		Filename: "test.md",
	}

	model := NewEntryEditorModel(entry)

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	model = updated.(EntryEditorModel)

	if !model.confirmed {
		t.Error("Model should be confirmed after pressing Ctrl+S")
	}

	if cmd == nil {
		t.Error("Confirm should return tea.Quit command")
	}
}

func TestEntryEditorModel_Cancel(t *testing.T) {
	entry := changeset.EntryWithFile{
		Entry: changeset.Entry{
			Type:    "added",
			Scope:   "cli",
			Summary: "Test entry",
		},
		Filename: "test.md",
	}

	model := NewEntryEditorModel(entry)

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model = updated.(EntryEditorModel)

	if !model.cancelled {
		t.Error("Model should be cancelled after pressing Esc")
	}

	if cmd == nil {
		t.Error("Cancel should return tea.Quit command")
	}
}

func TestEntryEditorModel_FieldNavigation(t *testing.T) {
	entry := changeset.EntryWithFile{
		Entry: changeset.Entry{
			Type:    "added",
			Scope:   "cli",
			Summary: "Test entry",
		},
		Filename: "test.md",
	}

	model := NewEntryEditorModel(entry)

	if model.focusIdx != 0 {
		t.Fatalf("Initial focus should be on field 0, got %d", model.focusIdx)
	}

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(EntryEditorModel)

	if model.focusIdx != 1 {
		t.Errorf("Focus should move to field 1 after Tab, got %d", model.focusIdx)
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	model = updated.(EntryEditorModel)

	if model.focusIdx != 0 {
		t.Errorf("Focus should move back to field 0 after Shift+Tab, got %d", model.focusIdx)
	}
}

func TestEntryEditorModel_GetEditedEntry(t *testing.T) {
	entry := changeset.EntryWithFile{
		Entry: changeset.Entry{
			Type:       "added",
			Scope:      "cli",
			Summary:    "Test entry",
			Breaking:   false,
			CommitHash: "abc123",
			DiffHash:   "def456",
		},
		Filename: "test.md",
	}

	model := NewEntryEditorModel(entry)

	model.typeIdx = 1

	editedEntry := model.GetEditedEntry()

	if editedEntry.Type != validTypes[1] {
		t.Errorf("Expected type %s, got %s", validTypes[1], editedEntry.Type)
	}

	if editedEntry.CommitHash != entry.Entry.CommitHash {
		t.Error("CommitHash should be preserved")
	}

	if editedEntry.DiffHash != entry.Entry.DiffHash {
		t.Error("DiffHash should be preserved")
	}
}

func TestEntryEditorModel_IsConfirmed(t *testing.T) {
	entry := changeset.EntryWithFile{
		Entry:    changeset.Entry{Type: "added", Summary: "Test"},
		Filename: "test.md",
	}

	model := NewEntryEditorModel(entry)

	if model.IsConfirmed() {
		t.Error("Model should not be confirmed initially")
	}

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(EntryEditorModel)

	if !model.IsConfirmed() {
		t.Error("Model should be confirmed after Enter key")
	}
}

func TestEntryEditorModel_IsCancelled(t *testing.T) {
	entry := changeset.EntryWithFile{
		Entry:    changeset.Entry{Type: "added", Summary: "Test"},
		Filename: "test.md",
	}

	model := NewEntryEditorModel(entry)

	if model.IsCancelled() {
		t.Error("Model should not be cancelled initially")
	}

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model = updated.(EntryEditorModel)

	if !model.IsCancelled() {
		t.Error("Model should be cancelled after Esc key")
	}
}

func TestEntryEditorModel_WindowSize(t *testing.T) {
	entry := changeset.EntryWithFile{
		Entry:    changeset.Entry{Type: "added", Summary: "Test"},
		Filename: "test.md",
	}

	model := NewEntryEditorModel(entry)

	if model.width != 0 || model.height != 0 {
		t.Error("Initial window size should be 0")
	}

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	model = updated.(EntryEditorModel)

	if model.width != 100 {
		t.Errorf("Width should be 100, got %d", model.width)
	}

	if model.height != 30 {
		t.Errorf("Height should be 30, got %d", model.height)
	}
}

func TestEntryEditorModel_TypeIndexForDifferentTypes(t *testing.T) {
	tests := []struct {
		entryType     string
		expectedIndex int
	}{
		{"added", 0},
		{"changed", 1},
		{"fixed", 2},
		{"removed", 3},
		{"security", 4},
	}

	for _, tt := range tests {
		t.Run(tt.entryType, func(t *testing.T) {
			entry := changeset.EntryWithFile{
				Entry: changeset.Entry{
					Type:    tt.entryType,
					Summary: "Test entry",
				},
				Filename: "test.md",
			}

			model := NewEntryEditorModel(entry)

			if model.typeIdx != tt.expectedIndex {
				t.Errorf("Type index for %s should be %d, got %d", tt.entryType, tt.expectedIndex, model.typeIdx)
			}
		})
	}
}
