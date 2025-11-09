# storm

> Local-first changelog manager with TUIs for review and release.

## Highlights

- **Keep a Changelog native:** unreleased notes live in `.changes/*.md` until you promote them.
- **Toolchain aware:** `storm bump`/`storm release` can update Cargo, npm, Python, and Deno manifests.
- **TUI friendly:** commit selectors, diff viewers, and toolchain pickers reuse the same palette and key bindings.
- **Scriptable CLI:** every subcommand prints concise status messages suitable for CI logs.

## Install

### Homebrew (macOS / Linux)

```sh
brew install stormlightlabs/tap/storm
```

The goreleaser workflow keeps the [`stormlightlabs/homebrew-tap`](https://github.com/stormlightlabs/homebrew-tap)
formula up to date.

### Go toolchain

```sh
go install github.com/stormlightlabs/git-storm/cmd/storm@latest
```

## Quick Start

```sh
storm generate --since v1.2.0 --interactive
storm unreleased review
storm release --bump patch --toolchain package.json --tag
```

## Documentation

- [Introduction](docs/introduction.md)
- [Quickstart](docs/quickstart.md)
- [Manual](docs/manual.md)
- [Development Guide](docs/development.md)

For a deeper dive into release automation, see `PROJECT.md`.

## Contributing

Run the full test suite before opening a PR:

```sh
go test ./...
```

Issues and feature ideas are welcome—Storm is intentionally modular so new
commands and TUIs can be added without touching the entire codebase.
