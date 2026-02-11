# Wiki Sync Tool (`wiki-sync`)

The `wiki-sync` tool manages the synchronization of documentation between the local `docs/` directory and the Gitea Wiki repository.

## Overview

This tool facilitates a two-way sync workflow:
1.  **Push (Repo -> Wiki):** Flattens local documentation paths and publishes them to the Wiki for drafting and viewing.
2.  **Pull (Wiki -> Repo):** Retrieves updated content from the Wiki, filtering by version and status, and updates the local repository.

## Installation

The tool is located in `cmd/wiki-sync`. You can run it directly with `go run` or build it.

```bash
# Run directly
go run cmd/wiki-sync/main.go --help

# Build (Optional)
go build -o wiki-sync.exe ./cmd/wiki-sync
```

## Usage

### Configuration

The tool can be configured via flags or environment variables (Flags take precedence).

| Flag | Env Var | Description | Default |
| :--- | :--- | :--- | :--- |
| `--wiki-path` | `WIKI_PATH` | Path to local clone of wiki repo | `./wiki` |
| `--url` | `WIKI_URL` | URL to fetch Wiki content via HTTP | (None) |
| `--target-version` | | Filter files by `approved_versions` frontmatter | (None) |
| `--keep-attrs` | `WIKI_KEEP_ATTRS` | Comma-separated list of frontmatter attributes to preserve | (Defaults: Standard GitHub + Internal) |

### Push (Repo -> Wiki)

Copies all Markdown files from `docs/` to the Wiki directory, flattening the directory structure (e.g., `docs/agent/rules.md` -> `repo-root~docs~agent~rules.md`).

```bash
go run cmd/wiki-sync/main.go push
```

### Pull (Wiki -> Repo)

Retrieves changes from the Wiki.

**Interactive Mode (Default):**
Scans for changes and presents an interactive checklist to select files to update.

```bash
go run cmd/wiki-sync/main.go pull
```

**Non-Interactive Mode:**
Use `--force` to accept all changes without prompting.

```bash
go run cmd/wiki-sync/main.go pull --force
```

**Dry Run:**
Preview changes without modifying local files.

```bash
go run cmd/wiki-sync/main.go pull --dry-run
```

**CI/Check Mode:**
Exit with code 1 if changes are detected (useful for CI pipelines).

```bash
go run cmd/wiki-sync/main.go pull --check
```

**HTTP Fetch Mode:**
Fetch content directly from the Gitea Wiki URL (bypassing local wiki clone).
*Note: This mode iterates local files and checks for remote updates; it cannot discover new files created solely on the Wiki.*

```bash
go run cmd/wiki-sync/main.go pull --url https://gitea.gnomatix.com/gnomatix/dreamfs/wiki/
```

## Detailed Logic

### Path Flattening
- **Local:** `docs/sub/folder/file.md`
- **Wiki:** `repo-root~docs~sub~folder~file.md`

### YAML Frontmatter
### YAML Frontmatter
- **Push:** Preserves YAML frontmatter in `docs/` files when copying to Wiki.
- **Pull:** Reconstructs frontmatter using the whitelist defined by `--keep-attrs` (or defaults). Attributes not in the whitelist are stripped from file content to avoid conflicts.
  - **Defaults:** Includes standard GitHub Frontmatter (e.g., `title`, `nav_order`, `parent`) and internal attributes (`release_path`, `approved_versions`, `review_status`, etc.).

