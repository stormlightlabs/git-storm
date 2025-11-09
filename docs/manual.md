---
title: Storm CLI Manual
---

# NAME

**storm** is a git powered aware changelog manager for Go projects.

## SYNOPSIS

```text
storm [--repo <path>] [--output <file>] <command> [flags]
```

## DESCRIPTION

Storm keeps unreleased notes in `.changes/*.md`, promotes them into
`CHANGELOG.md`, and offers TUIs for reviewing diffs and entries. The binary
is composed of self-contained subcommands that chain well inside scripts
or CI jobs.

### GLOBAL FLAGS

| Flag                    | Description                                              |
| ----------------------- | -------------------------------------------------------- |
| `--repo <path>`         | Working tree to operate on (default: current directory). |
| `-o`, `--output <file>` | Target changelog (default: `CHANGELOG.md`).              |

### COMMANDS

#### `storm bump`

Calculate the next semantic version by inspecting `CHANGELOG.md`.

```text
storm bump --bump <major|minor|patch> [--toolchain value...]
```

##### Flags

| Flag                         | Description                                                                                                   |
| ---------------------------- | ------------------------------------------------------------------------------------------------------------- |
| `--bump <type>` _(required)_ | Which semver component to increment.                                                                          |
| `--toolchain <value>`        | Update language manifests (`Cargo.toml`, `pyproject.toml`, `package.json`, `deno.json`).                      |
|                              | Accepts explicit paths, type aliases like `cargo`/`npm`, or the literal `interactive` to launch a picker TUI. |

#### `storm release`

Promote `.changes/*.md` into the changelog and optionally tag the repo.

```text
storm release (--version X.Y.Z | --bump <type>) [flags]
```

##### Flags

| Flag                  | Description                                                                         |
| --------------------- | ----------------------------------------------------------------------------------- |
| `--version <X.Y.Z>`   | Explicit version for the new changelog entry.                                       |
| `--bump <type>`       | Derive the version from the previous release (mutually exclusive with `--version`). |
| `--date <YYYY-MM-DD>` | Override the release date (default: today).                                         |
| `--clear-changes`     | Remove `.changes/*.md` files after a successful release.                            |
| `--dry-run`           | Render a preview without touching any files.                                        |
| `--tag`               | Create an annotated git tag containing the release notes.                           |
| `--toolchain <value>` | Update manifest files just like in `storm bump`.                                    |
| `--output-json`       | Emit machine-readable JSON instead of styled text.                                  |

#### `storm generate`

Create `.changes/*.md` files from commit history, with optional TUI review.

```text
storm generate <from> <to>
storm generate --since <tag> [to]
```

##### Flags

| Flag                  | Description                                        |
| --------------------- | -------------------------------------------------- |
| `-i`, `--interactive` | Open a commit selector TUI for choosing entries.   |
| `--since <tag>`       | Shortcut for `<from>`; defaults `<to>` to `HEAD`.  |
| `--output-json`       | Emit machine-readable JSON instead of styled text. |

#### `storm diff`

Side-by-side or unified diff with TUI navigation.

```text
storm diff <from>..<to> [flags]
storm diff <from> <to> [flags]
```

| Flag                            | Description                                           |
| ------------------------------- | ----------------------------------------------------- |
| `-f`, `--file <path>`           | Restrict the diff to a single file.                   |
| `-e`, `--expanded`              | Show all unchanged lines instead of compressed hunks. |
| `-v`, `--view <split\|unified>` | Rendering style (default: split).                     |

#### `storm check`

Verify every commit in a range has a corresponding unreleased entry.

```text
storm check <from> <to>
storm check --since <tag> [to]
```

| Flag            | Description                                                |
| --------------- | ---------------------------------------------------------- |
| `--since <tag>` | Start range at the provided tag and default end to `HEAD`. |

Non-zero exit status indicates missing entries. Messages containing
`[nochanges]` or `[skip changelog]` are ignored.

#### `storm unreleased`

Manage `.changes` entries directly.

##### `add`

```text
storm unreleased add --type <kind> --summary <text> [--scope value]
```

| Flag                                                | Description                                 |
| --------------------------------------------------- | ------------------------------------------- |
| `--type <added\|changed\|fixed\|removed\|security>` | Entry category.                             |
| `--summary <text>`                                  | Short human readable note.                  |
| `--scope <value>`                                   | Optional component indicator (e.g., `cli`). |

##### `list`

```text
storm unreleased list [--json]
```

| Flag     | Description                                        |
| -------- | -------------------------------------------------- |
| `--json` | Emit machine-readable JSON instead of styled text. |

##### `partial`

```text
storm unreleased partial <commit-ref> [flags]
```

| Flag               | Description                                         |
| ------------------ | --------------------------------------------------- |
| `--type <value>`   | Override the inferred type from the commit message. |
| `--summary <text>` | Override the inferred summary.                      |
| `--scope <value>`  | Optional component indicator.                       |

##### `review`

```text
storm unreleased review
```

Launch a Bubble Tea TUI for editing and deleting entries before release.
Requires a TTY; fall back to `storm unreleased list` otherwise.

#### `storm version`

Print the current build’s version string.

## FILES

- `.changes/` — queue of unreleased entries created by `storm generate` or `storm unreleased add`.
- `CHANGELOG.md` — Keep a Changelog-compatible file updated by `storm release`.

## SEE ALSO

`CHANGELOG.md`, [Keep a Changelog](https://keepachangelog.com), semantic versioning at [semver.org](https://semver.org).
