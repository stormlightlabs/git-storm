# `storm` — Project Guide

> Internal operations and release workflow

This document outlines:

- The release lifecycle
- CLI integration into development workflows
- Distribution setup for major package managers

## Lifecycle

### Local Development

- Edit and test locally:

  ```bash
  go run ./cmd/storm unreleased add --type fixed --summary "Fix diff rendering"
  go run ./cmd/storm unreleased list
  go run ./cmd/storm unreleased review
  ```

- Each `.changes/*.md` entry represents one meaningful change.

### 2. Preparing a Release

1. **Generate and review**

   ```bash
   storm generate v1.2.0 HEAD --interactive
   ```

   - Opens a Bubble Tea UI for categorization.
   - Outputs candidate `.changes/` files if confirmed.

2. **Verify unreleased notes**

   ```bash
   storm unreleased list
   ```

3. **Promote to changelog**

   ```bash
   storm release --version 1.3.0
   ```

   This:

   - Merges `.changes/*.md` into `CHANGELOG.md`
   - Inserts the date
   - Clears `.changes/`
   - (Optional) creates an annotated Git tag

4. **Commit and tag**

   ```bash
   git add CHANGELOG.md
   git commit -m "Release 1.3.0"
   git tag -a v1.3.0 -m "Release 1.3.0"
   ```

### Validation

Before pushing:

```bash
go test ./...
go run ./cmd/storm release --dry-run
```

Once satisfied:

```bash
git push origin main --tags
```

## CLI Integration in Workflows

`storm` is meant to sit between `git log` and `CHANGELOG.md`.
Use it at the end of development cycles or integrate it into CI pipelines.

### Example integrations

#### Local Workflow

- Add a new `.changes` file per PR or major commit.
- Review unreleased entries before each merge to `main`.
- Run `storm release` when preparing a new version.

#### Automated Release Job (CI)

Example pseudo-pipeline (GitHub Actions / Drone / Woodpecker):

```yaml
steps:
  - name: Compute next version
    run: |
      NEXT=$(storm bump --bump patch)
      echo "::set-output name=version::$NEXT"
  - name: Generate changelog
    run: |
      go install ./cmd/storm
      storm release --version ${{ steps.bump.outputs.version }} --toolchain package.json
  - name: Tag and push
    run: |
      git add CHANGELOG.md
      git commit -m "Release ${{ steps.bump.outputs.version }}"
      git tag -a v${{ steps.bump.outputs.version }} -m "Release ${{ steps.bump.outputs.version }}"
      git push origin main --tags
```

#### Integration with other tools

- **Taskfile / Justfile:** add `release` recipe calling `storm release`.
- **GoReleaser:** run `storm release` in `before.hooks`.
- **Custom TUI tools:** embed `internal/changeset` and `changelog` packages directly.

## Packaging & Distribution

### Homebrew (macOS / Linux)

#### Create a tap repo

Make a repo: `github.com/stormlightlabs/homebrew-tools`.

#### Formula template (`storm.rb`)

```ruby
class Gostorm < Formula
  desc "Git-aware changelog manager with TUI review"
  homepage "https://github.com/stormlightlabs/git-storm"
  version "1.3.0"
  url "https://github.com/stormlightlabs/git-storm/archive/refs/tags/v1.3.0.tar.gz"
  sha256 "<insert_sha256_here>"
  license "MIT"

  depends_on "go" => :build

  def install
    system "go", "build", *std_go_args, "./cmd/storm"
  end

  test do
    system "#{bin}/storm", "--help"
  end
end
```

#### Update formula on each release

Automate with `goreleaser`:

```yaml
brews:
  - tap: stormlightlabs/homebrew-tools
    name: storm
    folder: Formula
    commit_author:
      name: Owais Jamil
      email: owais@example.com
```

### Chocolatey (Windows)

#### Create package skeleton

```sh
tools/
  chocolateyinstall.ps1
  VERIFICATION.txt
storm.nuspec
```

#### `chocolateyinstall.ps1`

```powershell
$ErrorActionPreference = 'Stop'
$toolsDir   = "$(Split-Path -Parent $MyInvocation.MyCommand.Definition)"
$url        = 'https://github.com/stormlightlabs/git-storm/releases/download/v1.3.0/storm_1.3.0_windows_amd64.zip'

Install-ChocolateyZipPackage 'storm' $url $toolsDir
```

#### `.nuspec`

```xml
<package>
  <metadata>
    <id>storm</id>
    <version>1.3.0</version>
    <authors>Owais Jamil</authors>
    <description>Git-aware changelog manager with TUI review</description>
    <licenseUrl>https://opensource.org/licenses/MIT</licenseUrl>
    <projectUrl>https://github.com/stormlightlabs/git-storm</projectUrl>
  </metadata>
</package>
```

#### Build & push

```bash
choco pack
choco push storm.1.3.0.nupkg --source https://push.chocolatey.org/
```

### AUR

#### Create PKGBUILD

```bash
pkgname=storm
pkgver=1.3.0
pkgrel=1
pkgdesc="Git-aware changelog manager with TUI review"
arch=('x86_64')
url="https://github.com/stormlightlabs/git-storm"
license=('MIT')
depends=('git' 'go')
source=("$url/archive/refs/tags/v${pkgver}.tar.gz")
sha256sums=('SKIP')

build() {
  cd "$srcdir/storm-$pkgver"
  go build -o storm ./cmd/storm
}

package() {
  install -Dm755 storm "$pkgdir/usr/bin/storm"
}
```

#### Submit

Clone AUR repo:

```bash
git clone ssh://aur@aur.archlinux.org/storm.git
cp PKGBUILD storm/
cd storm
makepkg --printsrcinfo > .SRCINFO
git add .
git commit -m "Add storm v1.3.0"
git push
```

## Summary

| Step | Action                                         | Output                      |
| ---- | ---------------------------------------------- | --------------------------- |
|  1   | `storm generate v1.2.0 HEAD --interactive`     | Draft unreleased notes      |
|  2   | `storm release --version 1.3.0`                | Updated `CHANGELOG.md`      |
|  3   | `git tag -a v1.3.0`                            | Annotated release tag       |
|  4   | `goreleaser release --clean`                   | Builds + publishes binaries |
|  5   | Update tap / choco / AUR formulas              | Public distribution         |

## Maintenance Notes

- Keep `.changes/` entries atomic and descriptive.
- Avoid retroactive changelog edits outside releases.
- Tag every release in Git with an exact semantic version (`vX.Y.Z`).
- Ensure TUI remains optional (disable automatically in CI).
- Treat changelog generation as a testable unit — not a side effect.
