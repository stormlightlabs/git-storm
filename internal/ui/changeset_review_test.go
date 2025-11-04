package ui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/stormlightlabs/git-storm/internal/changeset"
)

func createMockEntry(filename, entryType, scope, summary string) changeset.EntryWithFile {
	return changeset.EntryWithFile{
		Entry: changeset.Entry{
			Type:     entryType,
			Scope:    scope,
			Summary:  summary,
			Breaking: false,
		},
		Filename: filename,
	}
}

func TestChangesetReviewModel_Init(t *testing.T) {
	entries := []changeset.EntryWithFile{
		createMockEntry("test.md", "added", "cli", "Test entry"),
	}

	model := NewChangesetReviewModel(entries)

	cmd := model.Init()
	if cmd != nil {
		t.Errorf("Init() should return nil, got %v", cmd)
	}
}

func TestChangesetReviewModel_DefaultActions(t *testing.T) {
	entries := []changeset.EntryWithFile{
		createMockEntry("test1.md", "added", "cli", "Test entry 1"),
		createMockEntry("test2.md", "fixed", "", "Test entry 2"),
	}

	model := NewChangesetReviewModel(entries)

	for i, item := range model.items {
		if item.Action != ActionKeep {
			t.Errorf("Item %d should default to ActionKeep, got %v", i, item.Action)
		}
	}
}

func TestChangesetReviewModel_MarkDelete(t *testing.T) {
	entries := []changeset.EntryWithFile{
		createMockEntry("test.md", "added", "cli", "Test entry"),
	}

	model := NewChangesetReviewModel(entries)

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	model = updated.(ChangesetReviewModel)

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	model = updated.(ChangesetReviewModel)

	if model.items[0].Action != ActionDelete {
		t.Errorf("Item should be marked for deletion, got action %v", model.items[0].Action)
	}
}

func TestChangesetReviewModel_MarkEdit(t *testing.T) {
	entries := []changeset.EntryWithFile{
		createMockEntry("test.md", "added", "cli", "Test entry"),
	}

	model := NewChangesetReviewModel(entries)

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	model = updated.(ChangesetReviewModel)

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	model = updated.(ChangesetReviewModel)

	if model.items[0].Action != ActionEdit {
		t.Errorf("Item should be marked for editing, got action %v", model.items[0].Action)
	}
}

func TestChangesetReviewModel_KeepAction(t *testing.T) {
	entries := []changeset.EntryWithFile{
		createMockEntry("test.md", "added", "cli", "Test entry"),
	}

	model := NewChangesetReviewModel(entries)

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	model = updated.(ChangesetReviewModel)

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	model = updated.(ChangesetReviewModel)

	if model.items[0].Action != ActionDelete {
		t.Fatal("Setup failed: item should be marked for deletion")
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeySpace})
	model = updated.(ChangesetReviewModel)

	if model.items[0].Action != ActionKeep {
		t.Errorf("Item should be marked for keeping, got action %v", model.items[0].Action)
	}
}

func TestChangesetReviewModel_Navigation(t *testing.T) {
	entries := []changeset.EntryWithFile{
		createMockEntry("test1.md", "added", "cli", "Test 1"),
		createMockEntry("test2.md", "fixed", "", "Test 2"),
		createMockEntry("test3.md", "changed", "api", "Test 3"),
	}

	model := NewChangesetReviewModel(entries)

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	model = updated.(ChangesetReviewModel)

	if model.cursor != 0 {
		t.Error("Cursor should start at 0")
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = updated.(ChangesetReviewModel)

	if model.cursor != 1 {
		t.Errorf("Cursor should be at 1 after down, got %d", model.cursor)
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyUp})
	model = updated.(ChangesetReviewModel)

	if model.cursor != 0 {
		t.Errorf("Cursor should be at 0 after up, got %d", model.cursor)
	}
}

func TestChangesetReviewModel_TopBottom(t *testing.T) {
	entries := make([]changeset.EntryWithFile, 10)
	for i := range entries {
		entries[i] = createMockEntry("test.md", "added", "", "Test")
	}

	model := NewChangesetReviewModel(entries)

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	model = updated.(ChangesetReviewModel)

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}})
	model = updated.(ChangesetReviewModel)

	if model.cursor != len(entries)-1 {
		t.Errorf("Cursor should be at bottom (index %d), got %d", len(entries)-1, model.cursor)
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
	model = updated.(ChangesetReviewModel)

	if model.cursor != 0 {
		t.Errorf("Cursor should be at top (index 0), got %d", model.cursor)
	}
}

func TestChangesetReviewModel_Confirm(t *testing.T) {
	entries := []changeset.EntryWithFile{
		createMockEntry("test.md", "added", "cli", "Test entry"),
	}

	model := NewChangesetReviewModel(entries)
	tm := teatest.NewTestModel(t, model)

	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))

	finalModel := tm.FinalModel(t, teatest.WithFinalTimeout(time.Second))
	reviewModel, ok := finalModel.(ChangesetReviewModel)
	if !ok {
		t.Fatal("Expected ChangesetReviewModel")
	}

	if !reviewModel.IsConfirmed() {
		t.Error("Model should be confirmed after pressing enter")
	}
	if reviewModel.IsCancelled() {
		t.Error("Model should not be cancelled")
	}
}

func TestChangesetReviewModel_QuitKeys(t *testing.T) {
	entries := []changeset.EntryWithFile{
		createMockEntry("test.md", "added", "cli", "Test entry"),
	}

	quitKeys := []struct {
		name    string
		keyType tea.KeyType
		runes   []rune
	}{
		{"q", tea.KeyRunes, []rune{'q'}},
		{"esc", tea.KeyEsc, nil},
		{"ctrl+c", tea.KeyCtrlC, nil},
	}

	for _, tc := range quitKeys {
		t.Run(tc.name, func(t *testing.T) {
			model := NewChangesetReviewModel(entries)
			tm := teatest.NewTestModel(t, model)

			var msg tea.Msg
			if tc.runes != nil {
				msg = tea.KeyMsg{Type: tc.keyType, Runes: tc.runes}
			} else {
				msg = tea.KeyMsg{Type: tc.keyType}
			}

			tm.Send(msg)
			tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))

			finalModel := tm.FinalModel(t, teatest.WithFinalTimeout(time.Second))
			reviewModel, ok := finalModel.(ChangesetReviewModel)
			if !ok {
				t.Fatal("Expected ChangesetReviewModel")
			}

			if !reviewModel.IsCancelled() {
				t.Error("Model should be cancelled after quit key")
			}
			if reviewModel.IsConfirmed() {
				t.Error("Model should not be confirmed")
			}
		})
	}
}

func TestChangesetReviewModel_GetReviewedItems(t *testing.T) {
	entries := []changeset.EntryWithFile{
		createMockEntry("test1.md", "added", "cli", "Test 1"),
		createMockEntry("test2.md", "fixed", "", "Test 2"),
		createMockEntry("test3.md", "changed", "api", "Test 3"),
	}

	model := NewChangesetReviewModel(entries)

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	model = updated.(ChangesetReviewModel)

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	model = updated.(ChangesetReviewModel)

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = updated.(ChangesetReviewModel)

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	model = updated.(ChangesetReviewModel)

	items := model.GetReviewedItems()

	if len(items) != 3 {
		t.Errorf("Expected 3 items, got %d", len(items))
	}

	if items[0].Action != ActionDelete {
		t.Errorf("First item should be ActionDelete, got %v", items[0].Action)
	}

	if items[1].Action != ActionEdit {
		t.Errorf("Second item should be ActionEdit, got %v", items[1].Action)
	}

	if items[2].Action != ActionKeep {
		t.Errorf("Third item should be ActionKeep, got %v", items[2].Action)
	}
}

func TestChangesetReviewModel_View(t *testing.T) {
	entries := []changeset.EntryWithFile{
		createMockEntry("test.md", "added", "cli", "Test entry"),
	}

	model := NewChangesetReviewModel(entries)

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	model = updated.(ChangesetReviewModel)

	view := model.View()

	if !strings.Contains(view, "Review unreleased changes") {
		t.Error("View should contain header text")
	}
	if !strings.Contains(view, "navigate") {
		t.Error("View should contain navigation help")
	}
	if !strings.Contains(view, "keep") {
		t.Error("View should contain action counts")
	}
}

func TestChangesetReviewModel_RenderHeader(t *testing.T) {
	entries := []changeset.EntryWithFile{
		createMockEntry("test1.md", "added", "cli", "Test 1"),
		createMockEntry("test2.md", "fixed", "", "Test 2"),
	}

	model := NewChangesetReviewModel(entries)
	model.width = 100

	header := model.renderReviewHeader()

	if !strings.Contains(header, "Review unreleased changes") {
		t.Error("Header should contain title")
	}
	if !strings.Contains(header, "2") {
		t.Error("Header should contain entry count")
	}
}

func TestChangesetReviewModel_RenderFooter(t *testing.T) {
	entries := []changeset.EntryWithFile{
		createMockEntry("test1.md", "added", "cli", "Test 1"),
		createMockEntry("test2.md", "fixed", "", "Test 2"),
		createMockEntry("test3.md", "changed", "api", "Test 3"),
	}

	model := NewChangesetReviewModel(entries)
	model.width = 100

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	model = updated.(ChangesetReviewModel)

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	model = updated.(ChangesetReviewModel)

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = updated.(ChangesetReviewModel)

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	model = updated.(ChangesetReviewModel)

	footer := model.renderReviewFooter()

	if !strings.Contains(footer, "keep: 1") {
		t.Error("Footer should show 1 keep action")
	}
	if !strings.Contains(footer, "delete: 1") {
		t.Error("Footer should show 1 delete action")
	}
	if !strings.Contains(footer, "edit: 1") {
		t.Error("Footer should show 1 edit action")
	}
}

func TestChangesetReviewModel_RenderReviewLine(t *testing.T) {
	entries := []changeset.EntryWithFile{
		createMockEntry("test.md", "added", "cli", "Test entry"),
	}

	model := NewChangesetReviewModel(entries)
	model.width = 100

	line := model.renderReviewLine(0, model.items[0])

	if !strings.Contains(line, "[") {
		t.Error("Line should contain action icon")
	}
	if !strings.Contains(line, "added") {
		t.Error("Line should contain type")
	}
	if !strings.Contains(line, "cli") {
		t.Error("Line should contain scope")
	}
	if !strings.Contains(line, "Test entry") {
		t.Error("Line should contain summary")
	}
}

func TestChangesetReviewModel_EmptyEntries(t *testing.T) {
	entries := []changeset.EntryWithFile{}

	model := NewChangesetReviewModel(entries)

	if len(model.items) != 0 {
		t.Errorf("Expected 0 items, got %d", len(model.items))
	}

	items := model.GetReviewedItems()
	if len(items) != 0 {
		t.Errorf("Expected 0 reviewed items, got %d", len(items))
	}
}

func TestChangesetReviewModel_WindowResize(t *testing.T) {
	entries := []changeset.EntryWithFile{
		createMockEntry("test.md", "added", "cli", "Test entry"),
	}

	model := NewChangesetReviewModel(entries)

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	model = updated.(ChangesetReviewModel)

	if model.width != 80 || model.height != 24 {
		t.Errorf("Expected dimensions 80x24, got %dx%d", model.width, model.height)
	}
	if !model.ready {
		t.Error("Model should be ready after window size message")
	}

	updated, _ = model.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	model = updated.(ChangesetReviewModel)

	if model.width != 120 || model.height != 40 {
		t.Errorf("Expected dimensions 120x40, got %dx%d", model.width, model.height)
	}
}

func TestChangesetReviewModel_ActionCounts(t *testing.T) {
	entries := []changeset.EntryWithFile{
		createMockEntry("test1.md", "added", "cli", "Test 1"),
		createMockEntry("test2.md", "fixed", "", "Test 2"),
		createMockEntry("test3.md", "changed", "api", "Test 3"),
		createMockEntry("test4.md", "removed", "", "Test 4"),
	}

	model := NewChangesetReviewModel(entries)

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	model = updated.(ChangesetReviewModel)

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	model = updated.(ChangesetReviewModel)

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = updated.(ChangesetReviewModel)

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	model = updated.(ChangesetReviewModel)

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = updated.(ChangesetReviewModel)

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	model = updated.(ChangesetReviewModel)

	items := model.GetReviewedItems()

	deleteCount := 0
	editCount := 0
	keepCount := 0

	for _, item := range items {
		switch item.Action {
		case ActionDelete:
			deleteCount++
		case ActionEdit:
			editCount++
		case ActionKeep:
			keepCount++
		}
	}

	if deleteCount != 2 {
		t.Errorf("Expected 2 delete actions, got %d", deleteCount)
	}
	if editCount != 1 {
		t.Errorf("Expected 1 edit action, got %d", editCount)
	}
	if keepCount != 1 {
		t.Errorf("Expected 1 keep action, got %d", keepCount)
	}
}
