package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stormlightlabs/git-storm/internal/changeset"
	"github.com/stormlightlabs/git-storm/internal/testutils"
)

func TestUnreleasedReviewWorkflow_Delete(t *testing.T) {
	tmpDir := t.TempDir()
	changesDir := filepath.Join(tmpDir, ".changes")

	entry1 := changeset.Entry{
		Type:    "added",
		Scope:   "test",
		Summary: "Entry to keep",
	}
	entry2 := changeset.Entry{
		Type:    "fixed",
		Scope:   "test",
		Summary: "Entry to delete",
	}

	filePath1, err := changeset.Write(changesDir, entry1)
	if err != nil {
		t.Fatalf("Failed to create entry1: %v", err)
	}
	filePath2, err := changeset.Write(changesDir, entry2)
	if err != nil {
		t.Fatalf("Failed to create entry2: %v", err)
	}

	filename2 := filepath.Base(filePath2)

	err = changeset.Delete(changesDir, filename2)
	if err != nil {
		t.Fatalf("Delete action failed: %v", err)
	}

	if _, err := os.Stat(filePath1); os.IsNotExist(err) {
		t.Error("Entry1 should still exist")
	}

	if _, err := os.Stat(filePath2); !os.IsNotExist(err) {
		t.Error("Entry2 should have been deleted")
	}

	entries, err := changeset.List(changesDir)
	if err != nil {
		t.Fatalf("Failed to list entries: %v", err)
	}

	testutils.Expect.Equal(t, len(entries), 1, "Should have 1 entry remaining")
	testutils.Expect.Equal(t, entries[0].Entry.Summary, "Entry to keep")
}

func TestUnreleasedReviewWorkflow_Edit(t *testing.T) {
	tmpDir := t.TempDir()
	changesDir := filepath.Join(tmpDir, ".changes")

	originalEntry := changeset.Entry{
		Type:       "added",
		Scope:      "cli",
		Summary:    "Original summary",
		Breaking:   false,
		CommitHash: "abc123",
	}

	filePath, err := changeset.Write(changesDir, originalEntry)
	if err != nil {
		t.Fatalf("Failed to create entry: %v", err)
	}

	filename := filepath.Base(filePath)

	editedEntry := changeset.Entry{
		Type:       "changed",
		Scope:      "api",
		Summary:    "Updated summary",
		Breaking:   true,
		CommitHash: "abc123",
	}

	err = changeset.Update(changesDir, filename, editedEntry)
	if err != nil {
		t.Fatalf("Update action failed: %v", err)
	}

	entries, err := changeset.List(changesDir)
	if err != nil {
		t.Fatalf("Failed to list entries: %v", err)
	}

	testutils.Expect.Equal(t, len(entries), 1, "Should still have 1 entry")
	testutils.Expect.Equal(t, entries[0].Entry.Type, "changed", "Type should be updated")
	testutils.Expect.Equal(t, entries[0].Entry.Scope, "api", "Scope should be updated")
	testutils.Expect.Equal(t, entries[0].Entry.Summary, "Updated summary", "Summary should be updated")
	testutils.Expect.Equal(t, entries[0].Entry.Breaking, true, "Breaking should be updated")
	testutils.Expect.Equal(t, entries[0].Entry.CommitHash, "abc123", "CommitHash should be preserved")
}

func TestUnreleasedReviewWorkflow_DeleteAndEdit(t *testing.T) {
	tmpDir := t.TempDir()
	changesDir := filepath.Join(tmpDir, ".changes")

	entry1 := changeset.Entry{
		Type:    "added",
		Summary: "Entry to delete",
	}
	entry2 := changeset.Entry{
		Type:    "fixed",
		Summary: "Entry to edit",
	}
	entry3 := changeset.Entry{
		Type:    "changed",
		Summary: "Entry to keep",
	}

	filePath1, err := changeset.Write(changesDir, entry1)
	if err != nil {
		t.Fatalf("Failed to create entry1: %v", err)
	}
	filePath2, err := changeset.Write(changesDir, entry2)
	if err != nil {
		t.Fatalf("Failed to create entry2: %v", err)
	}
	_, err = changeset.Write(changesDir, entry3)
	if err != nil {
		t.Fatalf("Failed to create entry3: %v", err)
	}

	filename1 := filepath.Base(filePath1)
	filename2 := filepath.Base(filePath2)

	err = changeset.Delete(changesDir, filename1)
	if err != nil {
		t.Fatalf("Delete action failed: %v", err)
	}

	editedEntry := changeset.Entry{
		Type:    "security",
		Scope:   "auth",
		Summary: "Updated security fix",
	}
	err = changeset.Update(changesDir, filename2, editedEntry)
	if err != nil {
		t.Fatalf("Update action failed: %v", err)
	}

	entries, err := changeset.List(changesDir)
	if err != nil {
		t.Fatalf("Failed to list entries: %v", err)
	}

	testutils.Expect.Equal(t, len(entries), 2, "Should have 2 entries remaining")

	var found bool
	for _, e := range entries {
		if e.Entry.Type == "security" {
			testutils.Expect.Equal(t, e.Entry.Scope, "auth")
			testutils.Expect.Equal(t, e.Entry.Summary, "Updated security fix")
			found = true
		}
	}

	if !found {
		t.Error("Edited entry not found in results")
	}

	if _, err := os.Stat(filePath1); !os.IsNotExist(err) {
		t.Error("Deleted entry should not exist")
	}
}

func TestUnreleasedReviewWorkflow_EmptyChanges(t *testing.T) {
	tmpDir := t.TempDir()
	changesDir := filepath.Join(tmpDir, ".changes")

	if err := os.MkdirAll(changesDir, 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	entries, err := changeset.List(changesDir)
	if err != nil {
		t.Fatalf("List should not error on empty directory: %v", err)
	}

	testutils.Expect.Equal(t, len(entries), 0, "Should have no entries")
}
