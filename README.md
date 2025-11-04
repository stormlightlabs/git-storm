# `storm`

> A Go-based changelog manager built for clarity, speed, and interaction.

## Goals

- Use Git as a data source, not a dependency.
- Store unreleased notes locally (`.changes/*.md`) in a simple, editable format.
- Provide a terminal UI for reviewing commits and changes interactively.
- Generate Markdown in strict [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) format.

### Architecture

- **Git integration:** Uses `go-git` for commit history and tag resolution — no shell calls.
- **Diffing:** Custom lightweight diff engine for readable line-by-line output.
- **Unreleased storage:** Simple Markdown files with YAML frontmatter (no external formats).
- **Interactive mode:** Bubble Tea model for categorizing and confirming changes.
- **Output:** Always produces Keep a Changelog–compliant Markdown.

## Core Packages

```sh
.
├── cmd
├── internal
│   ├── gitlog       # Parse and structure commit history via `go-git`
│   ├── diff         # Minimal line diff for display and review
│   ├── changeset    # Manage `.changes/*.md` files
│   ├── changelog    # Build and update `CHANGELOG.md` sections
│   ├── tui          # Bubble Tea–based interactive interface
│   └── style        # Centralized Lip Gloss palette and formatting
├── PROJECT.md
└── README.md
```

## Command Model

### Unreleased Changes

```sh
storm unreleased add --type added --scope cli --summary "Add changelog command"
storm unreleased list
storm unreleased review
```

Adds or reviews pending `.changes/*.md` entries.

### Generate From Git

```sh
storm generate <from> <to> [--interactive]
```

Pulls commits between refs, categorizes them by prefix, and optionally opens an interactive review.

### Release

```sh
storm release --version 1.3.0 [--tag]
```

Merges `.changes/*.md` into the changelog, writes a new section, and optionally tags the repository.

## Development Guidance

1. Composable
   Each subsystem (`diff`, `gitlog`, `tui`, etc.) should work standalone and be callable from tests or other Go programs.
2. Frontmatter

    ```yaml
    type: added
    scope: cli
    summary: Add changelog command
    ```

3. Consistent Palette

    | Type     | Color     |
    | -------- | --------- |
    | Added    | `#10b981` |
    | Changed  | `#0ea5e9` |
    | Fixed    | `#f43f5e` |
    | Removed  | `#f59e0b` |
    | Security | `#9333ea` |

4. Commands should chain naturally and script cleanly:

    ```sh
    storm unreleased list --json
    storm generate --since v1.2.0 --interactive
    storm release --version 1.3.0
    ```

5. Tests
    - Research testing bubbletea programs
    - Use golden files for diff/changelog output.
    - Use in-memory `go-git` repos in unit tests.

## Roadmap

| Phase | Deliverable                                    |
| ----- | ---------------------------------------------- |
| 1     | Core CLI (`generate`, `unreleased`, `release`) |
| 2     | Git integration and commit parsing             |
| 3     | Diff engine and styling                        |
| 4     | `.changes` storage and parsing                 |
| 5     | Interactive TUI                                |
| 6     | Keep a Changelog writer                        |
| 7     | Git tagging and CI integration                 |

## Notes

- No external dependencies beyond `cobra`, `go-git`, `bubbletea`, `lipgloss`, and `yaml.v3`.
- Keep the workflow simple and reproducible so changelogs can be deterministically derived from local data.
- Make sure interactive sessions degrade gracefully in non-TTY environments.

## Conventional Commits

### Structure

| Element                   | Format                                   | Description                              |
| ------------------------- | ---------------------------------------- | ---------------------------------------- |
| Header                    | `<type>(<scope>): <description>`         | The main commit message line.            |
| Scope                     | Optional, e.g. `api`, `cli`, `deps`      | Indicates part of the codebase affected. |
| Breaking Change Indicator | `!` after type/scope, e.g. `feat(api)!:` | Marks a breaking API change.             |
| Body                      | (Optional) one blank line then body text | Explanation of what & why.               |
| Footer                    | (Optional) one blank line then meta info | Issue refs, `BREAKING CHANGE: …`, etc    |

### Types

| Type       | Description                                                             |
| ---------- | ----------------------------------------------------------------------- |
| `feat`     | A new feature.                                                          |
| `fix`      | A bug fix.                                                              |
| `docs`     | Documentation only changes.                                             |
| `style`    | Code style changes (formatting, whitespace) that don’t affect behavior. |
| `refactor` | Code changes that neither fix a bug nor add a feature.                  |
| `perf`     | Performance improvements.                                               |
| `test`     | Adding or updating tests.                                               |
| `build`    | Changes that affect the build system or dependencies.                   |
| `ci`       | Changes to CI configuration and scripts.                                |
| `chore`    | Other changes that don’t touch src/test (e.g., tooling, config).        |
| `revert`   | Reverts a previous commit.                                              |

### Examples

```text
feat(api): add pagination endpoint

fix(ui): correct button alignment issue

docs: update README installation instructions

perf(core): optimize user query performance

refactor: restructure payment module for clarity

style: apply consistent formatting

test(auth): add integration tests for OAuth flow

build(deps): bump dependencies to latest versions

ci: add GitHub Actions workflow for CI

chore: update .gitignore and clean up obsolete files

feat(api)!: remove support for legacy endpoints

BREAKING CHANGE: API no longer accepts XML-formatted requests.
```

### Reference

<https://www.conventionalcommits.org/en/v1.0.0/> "Conventional Commits"
