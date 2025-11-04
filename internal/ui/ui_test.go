package ui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/stormlightlabs/git-storm/internal/diff"
)

func TestDiffModel_Init(t *testing.T) {
	edits := []diff.Edit{
		{Kind: diff.Equal, AIndex: 0, BIndex: 0, Content: "line 1"},
	}

	model := NewDiffModel(edits, "old.txt", "new.txt", 80, 24)

	cmd := model.Init()
	if cmd != nil {
		t.Errorf("Init() should return nil, got %v", cmd)
	}
}

func TestDiffModel_View(t *testing.T) {
	edits := []diff.Edit{
		{Kind: diff.Equal, AIndex: 0, BIndex: 0, Content: "unchanged line"},
		{Kind: diff.Delete, AIndex: 1, BIndex: -1, Content: "removed line"},
		{Kind: diff.Insert, AIndex: -1, BIndex: 1, Content: "added line"},
	}

	model := NewDiffModel(edits, "old.txt", "new.txt", 100, 30)

	view := model.View()

	if !strings.Contains(view, "old.txt") {
		t.Error("View should contain old file path")
	}
	if !strings.Contains(view, "new.txt") {
		t.Error("View should contain new file path")
	}
	if !strings.Contains(view, "unchanged line") {
		t.Error("View should contain unchanged line")
	}
}

func TestDiffModel_KeyboardNavigation(t *testing.T) {
	edits := make([]diff.Edit, 100)
	for i := range edits {
		edits[i] = diff.Edit{
			Kind:    diff.Equal,
			AIndex:  i,
			BIndex:  i,
			Content: "line content",
		}
	}

	model := NewDiffModel(edits, "old.txt", "new.txt", 80, 20)

	tm := teatest.NewTestModel(t, model, teatest.WithInitialTermSize(80, 20))

	// Test down movement
	tm.Send(tea.KeyMsg{Type: tea.KeyDown})
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return len(bts) > 0
	})

	// Test up movement
	tm.Send(tea.KeyMsg{Type: tea.KeyUp})
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return len(bts) > 0
	})

	// Test quit
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
}

func TestDiffModel_QuitKeys(t *testing.T) {
	edits := []diff.Edit{
		{Kind: diff.Equal, AIndex: 0, BIndex: 0, Content: "test"},
	}

	quitKeys := []tea.KeyType{
		tea.KeyRunes, // 'q'
		tea.KeyEsc,
		tea.KeyCtrlC,
	}

	for _, keyType := range quitKeys {
		t.Run(keyType.String(), func(t *testing.T) {
			model := NewDiffModel(edits, "old.txt", "new.txt", 80, 20)
			tm := teatest.NewTestModel(t, model)

			var msg tea.Msg
			if keyType == tea.KeyRunes {
				msg = tea.KeyMsg{Type: keyType, Runes: []rune{'q'}}
			} else {
				msg = tea.KeyMsg{Type: keyType}
			}

			tm.Send(msg)
			tm.WaitFinished(t, teatest.WithFinalTimeout(3*time.Second))
		})
	}
}

func TestDiffModel_RenderHeader(t *testing.T) {
	edits := []diff.Edit{}
	model := NewDiffModel(edits, "src/old.go", "src/new.go", 80, 20)

	header := model.renderHeader()

	if !strings.Contains(header, "old.go") {
		t.Error("Header should contain old file path")
	}
	if !strings.Contains(header, "new.go") {
		t.Error("Header should contain new file path")
	}
}

func TestDiffModel_RenderFooter(t *testing.T) {
	edits := []diff.Edit{
		{Kind: diff.Equal, AIndex: 0, BIndex: 0, Content: "test"},
	}
	model := NewDiffModel(edits, "old.txt", "new.txt", 80, 20)

	footer := model.renderFooter()

	if !strings.Contains(footer, "scroll") {
		t.Error("Footer should contain help text about scrolling")
	}
	if !strings.Contains(footer, "quit") {
		t.Error("Footer should contain help text about quitting")
	}
}
