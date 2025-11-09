---
title: Introduction
outline: deep
---

# Introduction

Storm is a CLI that keeps changelog entries close to your code. It grew from a
few principles:

1. **Plain text first.** Every unreleased change is a Markdown file that lives
   in your repo, making reviews and rebases simple.
2. **Deterministic releases.** Given the same `.changes` directory, Storm will
   always write the same `CHANGELOG.md` section.
3. **Interactive when helpful, scriptable everywhere.** TUIs exist for reviews
   and diffs, but every action prints machine-readable summaries for CI.

## Why Storm?

Storm sits between `git log` and `CHANGELOG.md`. It understands conventional
commits, keeps notes in version control, and prefers deterministic text files
over generated blobs. The CLI is written in Go, so it ships as a single binary
that runs anywhere your repository does.

- **Local-first workflow:** no external services or databases.
- **Deterministic releases:** `storm release` is idempotent and can run in CI.
- **Composable commands:** each subcommand prints useful summaries for scripts.

## Quick Preview

```sh
# Extract commits into .changes entries
storm generate --since v1.2.0 --interactive

# Review pending notes
storm unreleased review

# Cut a new release and update package.json
storm release --bump minor --toolchain package.json --tag
```

Need the details? Head to the [Quickstart](/quickstart) for a guided flow or
read the [manual](/manual) for every flag and exit code.

## Architecture Overview

```sh
.git/
.changes/
CHANGELOG.md
```

- `storm generate` and `storm unreleased add` populate `.changes/`.
- `storm check` and your CI ensure nothing merges without an entry.
- `storm release` merges the queue into `CHANGELOG.md`, optionally creates a
tag, and can update external manifests.

## Toolchain-aware versioning

The bump and release commands understand common ecosystem manifests:

| Manifest | Alias | Notes |
| -------- | ----- | ----- |
| `Cargo.toml` | `cargo`, `rust` | Updates `[package]` version. |
| `pyproject.toml` | `pyproject`, `python`, `poetry` | Supports `[project]` and `[tool.poetry]`. |
| `package.json` | `npm`, `node`, `package` | Edits the top-level `version` field. |
| `deno.json` | `deno` | Updates the root `version`. |

Pass specific paths or the literal `interactive` to launch the toolchain picker
TUI.

## TUIs everywhere

Storm shares a consistent palette (`internal/style`) across Bubble Tea
experiences:

- **Commit selector** for `storm generate --interactive`.
- **Unreleased review** for curating `.changes` entries.
- **Diff viewer** for `storm diff` with split/unified modes.
- **Toolchain picker** accessible via `--toolchain interactive`.

Each interface supports familiar Vim-style navigation (↑/↓, g/G, space to
select, `q` to quit) and degrades gracefully when no TTY is available.

## Suggested Workflow

1. Developers add `.changes` entries alongside feature branches.
2. Pull requests run `storm check --since <last-release>`.
3. Release engineers run `storm generate` (if needed) then `storm release`.
4. CI tags the release and publishes artifacts.

Need concrete steps? See the [Quickstart](/quickstart) or jump to the
[manual](/manual).
