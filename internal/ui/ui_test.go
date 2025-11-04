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

	tm.Send(tea.KeyMsg{Type: tea.KeyDown})
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return len(bts) > 0
	})

	tm.Send(tea.KeyMsg{Type: tea.KeyUp})
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return len(bts) > 0
	})

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
}

func TestDiffModel_QuitKeys(t *testing.T) {
	edits := []diff.Edit{
		{Kind: diff.Equal, AIndex: 0, BIndex: 0, Content: "test"},
	}

	quitKeys := []tea.KeyType{
		tea.KeyRunes,
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

func TestMultiFileDiffModel_Init(t *testing.T) {
	files := []FileDiff{
		{
			Edits:   []diff.Edit{{Kind: diff.Equal, AIndex: 0, BIndex: 0, Content: "test"}},
			OldPath: "old/file1.go",
			NewPath: "new/file1.go",
		},
	}

	model := NewMultiFileDiffModel(files, false)

	cmd := model.Init()
	if cmd != nil {
		t.Errorf("Init() should return nil, got %v", cmd)
	}
}

func TestMultiFileDiffModel_View(t *testing.T) {
	files := []FileDiff{
		{
			Edits:   []diff.Edit{{Kind: diff.Equal, AIndex: 0, BIndex: 0, Content: "line 1"}},
			OldPath: "old/file1.go",
			NewPath: "new/file1.go",
		},
		{
			Edits:   []diff.Edit{{Kind: diff.Insert, AIndex: -1, BIndex: 0, Content: "line 2"}},
			OldPath: "old/file2.go",
			NewPath: "new/file2.go",
		},
	}

	model := NewMultiFileDiffModel(files, false)

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	model = updated.(MultiFileDiffModel)

	view := model.View()

	if !strings.Contains(view, "file1.go") {
		t.Error("View should contain first file path")
	}
	if !strings.Contains(view, "[1/2]") {
		t.Error("View should contain file indicator [1/2]")
	}
}

func TestMultiFileDiffModel_Pagination(t *testing.T) {
	files := []FileDiff{
		{
			Edits:   []diff.Edit{{Kind: diff.Equal, AIndex: 0, BIndex: 0, Content: "file 1"}},
			OldPath: "old/file1.go",
			NewPath: "new/file1.go",
		},
		{
			Edits:   []diff.Edit{{Kind: diff.Equal, AIndex: 0, BIndex: 0, Content: "file 2"}},
			OldPath: "old/file2.go",
			NewPath: "new/file2.go",
		},
		{
			Edits:   []diff.Edit{{Kind: diff.Equal, AIndex: 0, BIndex: 0, Content: "file 3"}},
			OldPath: "old/file3.go",
			NewPath: "new/file3.go",
		},
	}

	model := NewMultiFileDiffModel(files, false)

	tm := teatest.NewTestModel(t, model, teatest.WithInitialTermSize(80, 24))

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return len(bts) > 0
	})

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return len(bts) > 0
	})

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
}

func TestMultiFileDiffModel_EmptyFiles(t *testing.T) {
	files := []FileDiff{}

	model := NewMultiFileDiffModel(files, false)

	view := model.View()

	if !strings.Contains(view, "No files") {
		t.Error("View should indicate no files to display")
	}
}

func TestMultiFileDiffModel_SingleFile(t *testing.T) {
	files := []FileDiff{
		{
			Edits:   []diff.Edit{{Kind: diff.Equal, AIndex: 0, BIndex: 0, Content: "test"}},
			OldPath: "old/single.go",
			NewPath: "new/single.go",
		},
	}

	model := NewMultiFileDiffModel(files, false)

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	model = updated.(MultiFileDiffModel)

	view := model.View()

	if !strings.Contains(view, "single.go") {
		t.Error("View should contain file path")
	}
	if !strings.Contains(view, "[1/1]") {
		t.Error("View should show [1/1] for single file")
	}
}

func TestMultiFileDiffModel_UpdateViewport(t *testing.T) {
	files := []FileDiff{
		{
			Edits:   []diff.Edit{{Kind: diff.Equal, AIndex: 0, BIndex: 0, Content: "content 1"}},
			OldPath: "file1.go",
			NewPath: "file1.go",
		},
		{
			Edits:   []diff.Edit{{Kind: diff.Insert, AIndex: -1, BIndex: 0, Content: "content 2"}},
			OldPath: "file2.go",
			NewPath: "file2.go",
		},
	}

	model := NewMultiFileDiffModel(files, false)

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	model = updated.(MultiFileDiffModel)

	initialView := model.View()
	if !strings.Contains(initialView, "content 1") {
		t.Error("Initial view should contain content from first file")
	}

	model.paginator.NextPage()
	model.updateViewport()

	updatedView := model.View()
	if !strings.Contains(updatedView, "file2.go") {
		t.Error("Updated view should show second file path")
	}
}

func TestMultiFileDiffModel_RenderHeader(t *testing.T) {
	files := []FileDiff{
		{
			Edits:   []diff.Edit{{Kind: diff.Equal, AIndex: 0, BIndex: 0, Content: "test"}},
			OldPath: "old/test.go",
			NewPath: "new/test.go",
		},
	}

	model := NewMultiFileDiffModel(files, false)
	header := model.renderMultiFileHeader()

	if !strings.Contains(header, "old/test.go") {
		t.Error("Header should contain old file path")
	}
	if !strings.Contains(header, "new/test.go") {
		t.Error("Header should contain new file path")
	}
	if !strings.Contains(header, "[1/1]") {
		t.Error("Header should contain file indicator")
	}
}

func TestMultiFileDiffModel_RenderFooter(t *testing.T) {
	files := []FileDiff{
		{
			Edits:   []diff.Edit{{Kind: diff.Equal, AIndex: 0, BIndex: 0, Content: "test"}},
			OldPath: "test.go",
			NewPath: "test.go",
		},
	}

	model := NewMultiFileDiffModel(files, false)
	footer := model.renderMultiFileFooter()

	if !strings.Contains(footer, "h/l") {
		t.Error("Footer should contain navigation help for h/l keys")
	}
	if !strings.Contains(footer, "scroll") {
		t.Error("Footer should contain scroll help")
	}
	if !strings.Contains(footer, "quit") {
		t.Error("Footer should contain quit help")
	}
}
