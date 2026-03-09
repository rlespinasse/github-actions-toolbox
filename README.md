# github-actions-toolbox

A CLI toolbox for GitHub Actions utilities.

## Installation

### Homebrew

```bash
brew install rlespinasse/tap/ghat
```

#### macOS Gatekeeper notice

On macOS, you may see a warning: _"Apple is not able to verify that it is free from malware that could harm your Mac or compromise your privacy."_

This is because the binary is ad-hoc signed but not notarized by Apple.

> [!NOTE]
> Apple notarization requires a paid Apple Developer account ($99/year). It consists of an automated malware scan — not a manual security review or code audit — and does not guarantee the software is safe. Open-source projects can be verified by reviewing the source code and build pipeline directly.

To allow it to run, either:

- **Via System Settings (UI):** Go to **System Settings > Privacy & Security**, scroll down, and click **Open Anyway** next to the blocked app message.
- **Via terminal:**

  ```bash
  xattr -d com.apple.quarantine $(which ghat)
  ```

### From source

```bash
go install github.com/rlespinasse/github-actions-toolbox@latest
```

### Binary releases

Download pre-built binaries from the [Releases](https://github.com/rlespinasse/github-actions-toolbox/releases) page. Available for Linux, macOS, and Windows (amd64/arm64).

## Usage

### `dependents` — Get GitHub dependents count

Fetch the number of repository dependents from GitHub's dependency graph.

```bash
ghat dependents owner/repo
```

Query multiple repositories at once:

```bash
ghat dependents owner/repo1 owner/repo2
```

Pipe repositories via stdin:

```bash
echo "owner/repo" | ghat dependents
```

```bash
cat repos.txt | ghat dependents
```

### Output format

Results are printed one per line:

```
owner/repo                          | deps | 42
```

The `deps` label is a clickable hyperlink (in supported terminals) pointing to the GitHub dependents page.

## Development

This project uses [just](https://github.com/casey/just) as a command runner.

```bash
just build    # Build the binary
just test     # Run all tests
just check    # Run fmt, vet, and test
just run dependents owner/repo  # Run without building
```

Releases are managed with [goreleaser](https://goreleaser.com/).
