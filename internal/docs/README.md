---
title: Testing Workflow
updated: 2025-01-08
version: 1
---

"Ride the lightning."

This document provides a comprehensive testing workflow for the `storm` changelog manager.
All tests should be run within this repository to validate functionality against real Git history.

## Setup

```bash
# Build the CLI
task build
```

## Core Workflow

### Manual Entry Creation (`unreleased add`)

Create entries manually without linking to commits.

#### Basic entry creation

```bash
storm unreleased add --type added --summary "Test manual entry"
```

**Expected:**

- Creates `.changes/<timestamp>-test-manual-entry.md`
- File contains YAML frontmatter with type and summary
- Styled success message displays created file path

#### Entry with scope

```bash
storm unreleased add --type fixed --scope api --summary "Fix authentication bug"
```

**Expected:**

- Includes `scope: api` in frontmatter
- Filename slugifies to `...-fix-authentication-bug.md`

#### Collision handling

```bash
# Run same command twice rapidly
storm unreleased add --type added --summary "Duplicate test"
storm unreleased add --type added --summary "Duplicate test"
```

**Expected:**

- Two different files created (second has `-1` suffix)
- Both files exist and are readable

**Edge Cases:**

- Invalid type (should error with helpful message)
- Missing required flags (should error)
- Very long summary (should truncate to 50 chars)
- Special characters in summary (should slugify correctly)
- Empty summary (should error)

### Commit-Linked Entry Creation (`unreleased partial`)

Create entries linked to specific commits with auto-detection.

#### Basic partial from commit

```bash
# Use a recent commit hash
storm unreleased partial HEAD
```

**Expected:**

- Auto-detects type from conventional commit format
- Creates `.changes/<sha7>.<type>.md`
- Includes `commit_hash` in frontmatter
- Shows styled success message

#### Override auto-detection

```bash
storm unreleased partial HEAD~1 --type fixed --summary "Custom summary"
```

**Expected:**

- Uses provided type instead of auto-detected
- Uses custom summary
- Preserves commit hash in frontmatter

#### Non-conventional commit

```bash
# Try a commit without conventional format
storm unreleased partial <old-commit>
```

**Expected:**

- Error message: "could not auto-detect change type"
- Suggests using `--type` flag

#### Duplicate prevention

```bash
storm unreleased partial HEAD
storm unreleased partial HEAD  # Run again
```

**Expected:**

- Second command fails with "file already exists" error

**Edge Cases:**

- Invalid commit ref (should error)
- Merge commit (should handle gracefully)
- Initial commit with no parent (should work)
- Commit with multi-line message (should parse correctly)
- Commit with breaking change marker (should set `breaking: true`)

### Listing Entries (`unreleased list`)

Display all unreleased changes.

#### Text output

```bash
storm unreleased list
```

**Expected:**

- Color-coded type labels ([added], [fixed], etc.)
- Shows scope if present
- Displays filename
- Shows breaking change indicator if applicable
- Empty state message if no entries

#### JSON output

```bash
storm unreleased list --json
```

**Expected:**

- Valid JSON array
- Each entry has type, scope, summary, filename
- Can be piped to `jq` for processing

**Edge Cases:**

- Empty `.changes/` directory
- Malformed YAML in entry file
- Mixed entry types (manual + partial)

### Generating Entries from Git History (`generate`)

Scan commit ranges and create changelog entries.

#### Range generation

```bash
# Generate from last 5 commits
storm generate HEAD~5 HEAD
```

**Expected:**

- Lists N commits found
- Creates entries for conventional commits
- Skips non-conventional commits
- Shows created count and skipped count
- Uses diff-based deduplication

#### Interactive selection

```bash
storm generate HEAD~10 HEAD --interactive
```

**Expected:**

- Launches TUI with commit list
- Shows parsed metadata (type, scope, summary)
- Allows selection/deselection
- Creates only selected entries
- Handles cancellation (Ctrl+C)

#### Since tag

```bash
storm generate --since v0.1.0
```

**Expected:**

- Generates entries from v0.1.0 to HEAD
- Auto-detects tag as starting point

#### Deduplication

```bash
storm generate HEAD~3 HEAD
storm generate HEAD~3 HEAD  # Run again
```

**Expected:**

- First run creates N entries
- Second run shows "Skipped N duplicates"
- No duplicate files created

#### Rebased commits

```bash
# Simulate rebase by checking metadata
storm generate <range-with-rebased-commits>
```

**Expected:**

- Detects same diff, different commit hash
- Updates metadata with new commit hash
- Shows "Updated N rebased commits"

**Edge Cases:**

- No commits in range (should show "No commits found")
- Range with only merge commits
- Range with revert commits (should skip)
- Commits with `[nochanges]` marker (should skip)
- Non-existent refs (should error)

### Reviewing Entries (`unreleased review`)

Interactive TUI for reviewing unreleased changes.

#### Basic review

```bash
storm unreleased review
```

**Expected:**

- Launches TUI with list of entries
- Shows entry details on selection
- Keyboard navigation works (j/k or arrows)
- Can mark for delete/edit (not yet implemented)
- Exit with q or ESC

**Edge Cases:**

- Empty changes directory (should show message, not crash)
- Corrupted entry file (should handle gracefully)
- Non-TTY environment (should detect and warn)

### CI Validation (`check`)

Validate that commits have changelog entries.

#### All commits documented

```bash
# After running generate for a range
storm check HEAD~5 HEAD
```

**Expected:**

- Shows "✓ All commits have changelog entries"
- Exit code 0

#### Missing entries

```bash
# Create new commits without entries
git commit --allow-empty -m "feat: undocumented feature"
storm check HEAD~1 HEAD
```

**Expected:**

- Shows "✗ N commits missing changelog entries"
- Lists missing commit SHAs and subjects
- Suggests commands to fix
- Exit code 1

#### Skip markers

```bash
git commit --allow-empty -m "chore: update deps [nochanges]"
storm check HEAD~1 HEAD
```

**Expected:**

- Skips commit with marker
- Shows "Skipped N commits with [nochanges] marker"
- Exit code 0

#### Since tag

```bash
storm check --since v0.1.0
```

**Expected:**

- Checks all commits since tag
- Reports missing entries

**Edge Cases:**

- Empty commit range (should succeed with 0 checks)
- Range with all skipped commits
- Invalid tag/ref (should error)

### Release Generation (`release`)

Promote unreleased changes to CHANGELOG.

#### Basic release

```bash
storm release --version 1.2.0
```

**Expected:**

- Creates/updates CHANGELOG.md
- Adds version header with date
- Groups entries by type (Added, Changed, Fixed, etc.)
- Maintains Keep a Changelog format
- Preserves existing changelog content

#### Dry run

```bash
storm release --version 1.2.0 --dry-run
```

**Expected:**

- Shows preview of changes
- No files modified
- Styled output shows what would be written

#### Clear changes

```bash
storm release --version 1.2.0 --clear-changes
```

**Expected:**

- Moves entries from `.changes/` to CHANGELOG
- Deletes `.changes/*.md` files after release
- Keeps `.changes/data/` metadata

#### Git tagging

```bash
storm release --version 1.2.0 --tag
```

**Expected:**

- Creates annotated Git tag `v1.2.0`
- Includes release notes in tag message
- Validates tag doesn't exist

**Edge Cases:**

- No unreleased entries (should warn)
- Existing version in CHANGELOG (should append)
- Malformed CHANGELOG.md (should handle)
- Tag already exists (should error)
- Custom date format with `--date`

### Diff Viewing (`diff`)

Display inline diffs between refs.

#### Basic diff

```bash
storm diff HEAD~1 HEAD
```

**Expected:**

- Shows unified diff with syntax highlighting
- Iceberg theme colors
- Context lines displayed
- File headers shown

#### File filtering

```bash
storm diff HEAD~1 HEAD -- "*.go"
```

**Expected:**

- Shows only Go file changes
- Respects glob patterns

**Edge Cases:**

- No changes between refs
- Binary files (should indicate)
- Large diffs (should handle gracefully)
