---
title: Quickstart
outline: deep
---

# Quickstart

This walkthrough gets you from zero to a published changelog entry in a few
minutes. It mirrors the default workflow baked into the CLI.

## 1. Install the CLI

```sh
go install github.com/stormlightlabs/git-storm/cmd/storm@latest
```

Verify the binary is available:

```sh
storm version
```

## 2. Capture unreleased changes

Create a `.changes` entry manually or generate them from commits.

### Option A — Manual entry

```sh
storm unreleased add \
  --type added \
  --scope cli \
  --summary "Add bump command"
```

### Option B — From git history

```sh
storm generate --since v1.2.0 --interactive
```

Use the commit selector TUI to pick which commits become entries. Storm writes
Markdown files such as `.changes/2025-03-01-add-bump-command.md`.

## 3. Review pending entries

```sh
storm unreleased review
```

The Bubble Tea UI lets you edit summaries, delete noise, or mark entries as
ready. In non-interactive environments, fall back to
`storm unreleased list --json`.

## 4. Dry-run a release

```sh
storm release --bump patch --dry-run
```

This prints the new `CHANGELOG` section without modifying files. When the
output looks right, re-run without `--dry-run`.

## 5. Publish and tag

```sh
storm release --bump patch --toolchain package.json --tag
```

- `--bump patch` derives the next version from the previous release.
- `--toolchain package.json` keeps your npm manifest in sync.
- `--tag` creates an annotated git tag containing the release notes.

Follow up with standard git commands:

```sh
git add CHANGELOG.md package.json .changes
git commit -m "Release v$(storm bump --bump patch)"
git push origin main --tags
```

## 6. Enforce entries in CI (optional)

```sh
storm check --since v1.2.0
```

The command exits non-zero when commits are missing `.changes` files, making it
ideal for pre-merge checks.

## Next steps

- Skim the [Introduction](/introduction) to understand the design.
- Explore every flag in the [manual](/manual).
