package ui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/object"
	"github.com/stormlightlabs/git-storm/internal/gitlog"
)

type mockParser struct{}

func (p *mockParser) Parse(hash, subject, body string, date time.Time) (gitlog.CommitMeta, error) {
	meta := gitlog.CommitMeta{
		Type:        "feat",
		Scope:       "",
		Description: subject,
		Body:        body,
		Breaking:    false,
		Footers:     make(map[string]string),
	}
	return meta, nil
}

func (p *mockParser) IsValidType(kind gitlog.CommitKind) bool {
	return kind != gitlog.CommitTypeUnknown
}

func (p *mockParser) Categorize(meta gitlog.CommitMeta) string {
	switch meta.Type {
	case "feat":
		return "added"
	case "fix":
		return "fixed"
	default:
		return "changed"
	}
}

func createMockCommit(hash, message string, when time.Time) *object.Commit {
	return &object.Commit{
		Hash:    plumbing.NewHash(hash),
		Message: message,
		Author: object.Signature{
			Name:  "Test Author",
			Email: "test@example.com",
			When:  when,
		},
	}
}

func TestCommitSelectorModel_Init(t *testing.T) {
	commits := []*object.Commit{
		createMockCommit("a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2", "feat: add feature", time.Now()),
	}

	parser := &mockParser{}
	model := NewCommitSelectorModel(commits, "v1.0.0", "HEAD", parser)

	cmd := model.Init()
	if cmd != nil {
		t.Errorf("Init() should return nil, got %v", cmd)
	}
}

func TestCommitSelectorModel_AutoSelect(t *testing.T) {
	now := time.Now()
	commits := []*object.Commit{
		createMockCommit("a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2", "feat: add feature", now),
		createMockCommit("b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3", "fix: bug fix", now),
	}

	parser := &gitlog.ConventionalParser{}
	model := NewCommitSelectorModel(commits, "v1.0.0", "HEAD", parser)

	selectedItems := model.GetSelectedItems()
	if len(selectedItems) != 2 {
		t.Errorf("Expected 2 auto-selected items, got %d", len(selectedItems))
	}
}

func TestCommitSelectorModel_GetSelectedCommits(t *testing.T) {
	now := time.Now()
	commits := []*object.Commit{
		createMockCommit("a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2", "feat: feature 1", now),
		createMockCommit("b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3", "feat: feature 2", now),
		createMockCommit("c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4", "feat: feature 3", now),
	}

	parser := &mockParser{}
	model := NewCommitSelectorModel(commits, "v1.0.0", "HEAD", parser)

	model.items[1].Selected = false

	selected := model.GetSelectedCommits()
	if len(selected) != 2 {
		t.Errorf("Expected 2 selected commits, got %d", len(selected))
	}

	if selected[0].Hash != commits[0].Hash {
		t.Error("First selected commit should match first commit")
	}
	if selected[1].Hash != commits[2].Hash {
		t.Error("Second selected commit should match third commit")
	}
}

func TestCommitSelectorModel_ToggleSelection(t *testing.T) {
	now := time.Now()
	commits := []*object.Commit{
		createMockCommit("a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2", "feat: test", now),
	}

	parser := &mockParser{}
	model := NewCommitSelectorModel(commits, "v1.0.0", "HEAD", parser)

	initialSelected := model.items[0].Selected

	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	model = updatedModel.(CommitSelectorModel)

	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeySpace})
	model = updatedModel.(CommitSelectorModel)

	if model.items[0].Selected == initialSelected {
		t.Error("Selection should have been toggled")
	}
}

func TestCommitSelectorModel_SelectAll(t *testing.T) {
	now := time.Now()
	commits := []*object.Commit{
		createMockCommit("a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2", "feat: test 1", now),
		createMockCommit("b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3", "feat: test 2", now),
		createMockCommit("c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4", "feat: test 3", now),
	}

	parser := &mockParser{}
	model := NewCommitSelectorModel(commits, "v1.0.0", "HEAD", parser)

	for i := range model.items {
		model.items[i].Selected = false
	}

	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	model = updatedModel.(CommitSelectorModel)

	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	model = updatedModel.(CommitSelectorModel)

	for i, item := range model.items {
		if !item.Selected {
			t.Errorf("Item %d should be selected", i)
		}
	}
}

func TestCommitSelectorModel_DeselectAll(t *testing.T) {
	now := time.Now()
	commits := []*object.Commit{
		createMockCommit("a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2", "feat: test 1", now),
		createMockCommit("b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3", "feat: test 2", now),
	}

	parser := &mockParser{}
	model := NewCommitSelectorModel(commits, "v1.0.0", "HEAD", parser)

	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	model = updatedModel.(CommitSelectorModel)

	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'A'}})
	model = updatedModel.(CommitSelectorModel)

	for i, item := range model.items {
		if item.Selected {
			t.Errorf("Item %d should be deselected", i)
		}
	}
}

func TestCommitSelectorModel_Navigation(t *testing.T) {
	now := time.Now()
	commits := make([]*object.Commit, 50)
	for i := range commits {
		hash := strings.Repeat("a", 40)
		commits[i] = createMockCommit(hash, "feat: test", now)
	}

	parser := &mockParser{}
	model := NewCommitSelectorModel(commits, "v1.0.0", "HEAD", parser)

	tm := teatest.NewTestModel(t, model, teatest.WithInitialTermSize(100, 20))

	tm.Send(tea.KeyMsg{Type: tea.KeyDown})
	tm.Send(tea.KeyMsg{Type: tea.KeyUp})
	tm.Send(tea.KeyMsg{Type: tea.KeyPgDown})
	tm.Send(tea.KeyMsg{Type: tea.KeyPgUp})
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
}

func TestCommitSelectorModel_TopBottom(t *testing.T) {
	now := time.Now()
	commits := make([]*object.Commit, 20)
	for i := range commits {
		hash := strings.Repeat("a", 40)
		commits[i] = createMockCommit(hash, "feat: test", now)
	}

	parser := &mockParser{}
	model := NewCommitSelectorModel(commits, "v1.0.0", "HEAD", parser)

	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 20})
	model = updatedModel.(CommitSelectorModel)

	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}})
	model = updatedModel.(CommitSelectorModel)

	if model.cursor != len(commits)-1 {
		t.Errorf("Cursor should be at bottom (index %d), got %d", len(commits)-1, model.cursor)
	}

	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
	model = updatedModel.(CommitSelectorModel)

	if model.cursor != 0 {
		t.Errorf("Cursor should be at top (index 0), got %d", model.cursor)
	}
}

func TestCommitSelectorModel_Confirm(t *testing.T) {
	now := time.Now()
	commits := []*object.Commit{
		createMockCommit("a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2", "feat: test", now),
	}

	parser := &mockParser{}
	model := NewCommitSelectorModel(commits, "v1.0.0", "HEAD", parser)

	tm := teatest.NewTestModel(t, model)

	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))

	finalModel := tm.FinalModel(t, teatest.WithFinalTimeout(time.Second))
	selectorModel, ok := finalModel.(CommitSelectorModel)
	if !ok {
		t.Fatal("Expected CommitSelectorModel")
	}

	if !selectorModel.IsConfirmed() {
		t.Error("Model should be confirmed after pressing enter")
	}
	if selectorModel.IsCancelled() {
		t.Error("Model should not be cancelled")
	}
}

func TestCommitSelectorModel_QuitKeys(t *testing.T) {
	now := time.Now()
	commits := []*object.Commit{
		createMockCommit("a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2", "feat: test", now),
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
			parser := &mockParser{}
			model := NewCommitSelectorModel(commits, "v1.0.0", "HEAD", parser)
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
			selectorModel, ok := finalModel.(CommitSelectorModel)
			if !ok {
				t.Fatal("Expected CommitSelectorModel")
			}

			if !selectorModel.IsCancelled() {
				t.Error("Model should be cancelled after quit key")
			}
			if selectorModel.IsConfirmed() {
				t.Error("Model should not be confirmed")
			}
		})
	}
}

func TestCommitSelectorModel_View(t *testing.T) {
	now := time.Now()
	commits := []*object.Commit{
		createMockCommit("a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2", "feat: add feature", now),
	}

	parser := &mockParser{}
	model := NewCommitSelectorModel(commits, "v1.0.0", "HEAD", parser)

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	model = updated.(CommitSelectorModel)

	view := model.View()

	if !strings.Contains(view, "v1.0.0") {
		t.Error("View should contain fromRef")
	}
	if !strings.Contains(view, "HEAD") {
		t.Error("View should contain toRef")
	}
	if !strings.Contains(view, "a1b2c3d") {
		t.Error("View should contain commit hash")
	}
}

func TestCommitSelectorModel_RenderHeader(t *testing.T) {
	now := time.Now()
	commits := []*object.Commit{
		createMockCommit("a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2", "feat: test", now),
	}

	parser := &mockParser{}
	model := NewCommitSelectorModel(commits, "v1.0.0", "HEAD", parser)
	model.width = 100

	header := model.renderCommitHeader()

	if !strings.Contains(header, "v1.0.0") {
		t.Error("Header should contain fromRef")
	}
	if !strings.Contains(header, "HEAD") {
		t.Error("Header should contain toRef")
	}
	if !strings.Contains(header, "Select commits") {
		t.Error("Header should contain instruction text")
	}
}

func TestCommitSelectorModel_RenderFooter(t *testing.T) {
	now := time.Now()
	commits := []*object.Commit{
		createMockCommit("a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2", "feat: test 1", now),
		createMockCommit("b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3", "feat: test 2", now),
	}

	parser := &mockParser{}
	model := NewCommitSelectorModel(commits, "v1.0.0", "HEAD", parser)
	model.width = 100

	footer := model.renderCommitFooter()

	if !strings.Contains(footer, "navigate") {
		t.Error("Footer should contain navigation help")
	}
	if !strings.Contains(footer, "toggle") {
		t.Error("Footer should contain toggle help")
	}
	if !strings.Contains(footer, "confirm") {
		t.Error("Footer should contain confirm help")
	}
	if !strings.Contains(footer, "selected") {
		t.Error("Footer should contain selection count")
	}
}

func TestCommitSelectorModel_RenderCommitLine(t *testing.T) {
	now := time.Now()
	commits := []*object.Commit{
		createMockCommit("a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2", "feat: add feature", now),
	}

	parser := &mockParser{}
	model := NewCommitSelectorModel(commits, "v1.0.0", "HEAD", parser)
	model.width = 100

	line := model.renderCommitLine(0, model.items[0])

	if !strings.Contains(line, "[") && !strings.Contains(line, "]") {
		t.Error("Line should contain checkbox")
	}
	if !strings.Contains(line, "a1b2c3d") {
		t.Error("Line should contain short commit hash")
	}
	if !strings.Contains(line, "added") {
		t.Error("Line should contain category")
	}
}

func TestFormatTimeAgo(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		time     time.Time
		expected string
	}{
		{"just now", now, "just now"},
		{"minutes ago", now.Add(-5 * time.Minute), "5m ago"},
		{"hours ago", now.Add(-2 * time.Hour), "2h ago"},
		{"days ago", now.Add(-3 * 24 * time.Hour), "3d ago"},
		{"months ago", now.Add(-45 * 24 * time.Hour), "1mo ago"},
		{"years ago", now.Add(-400 * 24 * time.Hour), "1y ago"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := fmtTimeAgo(tc.time)
			if result != tc.expected {
				t.Errorf("Expected %q, got %q", tc.expected, result)
			}
		})
	}
}

func TestCommitSelectorModel_EmptyCommits(t *testing.T) {
	commits := []*object.Commit{}

	parser := &mockParser{}
	model := NewCommitSelectorModel(commits, "v1.0.0", "HEAD", parser)

	if len(model.items) != 0 {
		t.Errorf("Expected 0 items, got %d", len(model.items))
	}

	selected := model.GetSelectedCommits()
	if len(selected) != 0 {
		t.Errorf("Expected 0 selected commits, got %d", len(selected))
	}
}

func TestCommitSelectorModel_WindowResize(t *testing.T) {
	now := time.Now()
	commits := []*object.Commit{
		createMockCommit("a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2", "feat: test", now),
	}

	parser := &mockParser{}
	model := NewCommitSelectorModel(commits, "v1.0.0", "HEAD", parser)

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	model = updated.(CommitSelectorModel)

	if model.width != 80 || model.height != 24 {
		t.Errorf("Expected dimensions 80x24, got %dx%d", model.width, model.height)
	}
	if !model.ready {
		t.Error("Model should be ready after window size message")
	}

	updated, _ = model.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	model = updated.(CommitSelectorModel)

	if model.width != 120 || model.height != 40 {
		t.Errorf("Expected dimensions 120x40, got %dx%d", model.width, model.height)
	}
}
