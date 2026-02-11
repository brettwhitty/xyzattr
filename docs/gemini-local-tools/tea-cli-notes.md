---
type: MANUAL
authority: PENDING MANUAL REVIEW
review_status: PENDING
version: 0.1.1
approved_versions: 0.1.*
generated_on: 2026-01-22 17:10
origin_persona: Rory Devlin (R&D Team Lead)
origin_session: d346cd76-699f-4ab0-b24a-53180588cb07
intent: Document common Tea (Gitea CLI) commands for non-interactive repository management.
primary_sources: [docs/gemini-local-tools/tea-cli-notes.md]
release_path: docs/gemini-local-tools/tea-cli-notes.md
related_issues: []
related_sops: []
tags: [tea, cli, gitea]
---

# Tea CLI Tool Notes

This document summarizes the key learnings and common usage patterns for the `tea` (Gitea CLI) tool.

## Basic Usage
The `tea` CLI tool follows a `tea <command> <subcommand> [global options] [command options]` structure.

### Non-Interactive Execution (MANDATORY JSON)
To prevent `tea` from hanging and to ensure 100% accurate data parsing:

1.  **Mandatory Flags:** Always include `--login <user>` and `--repo <owner/name>`.
2.  **Mandatory Output:** Always use `--output json`. This is the ONLY way to prevent the "dog ate my homework" truncation failures encountered in previous sessions.
3.  **Specific Flags:**
    *   **Issues/Comments:** Always use `--comments` or `--comments-all` to prevent prompts about comment visibility.
    *   **Creation:** Explicitly provide all required fields (`--title`, `--body`).

**Example Command Pattern:**
```bash
tea issue list --login admin --repo brett/dreamfs --comments --output json
```

## Getting Help
To get help for any command or subcommand, use the `--help` flag:
- `tea --help`
- `tea <command> --help`

## Authentication

### Managing Logins
`tea` supports multiple Gitea servers/accounts.
- **List Logins:** `tea logins list`
- **Add Login:** `tea login add` (Interactive)
- **Use Specific Login:** Append `--login <name>` to any command (e.g., `tea issues list --login admin`).

## Repositories

### Listing Repos
List repositories you have access to, with customizable output formats.
```bash
tea repo list --output table --fields name,owner,ssh,permission
```

### Searching Repos
Find a repository on the instance:
```bash
tea repo search <query>
```

### Limitations
`tea` is primarily for *interacting* with repo content (issues/PRs) and basic creation/deletion. It **cannot** currently modify advanced repository settings (like Branch Protection, Mirrors, or Default Branch) directly. These must be handled via the Web UI or API.

## Branch Management

### Unprotecting Branches (Admin)
If you have Admin/Owner permissions, you can disable branch protection (e.g., to allow Force Push):
```bash
tea branch unprotect <branch_name> --repo <owner>/<repo>
```

## Issues & Pull Requests

### Listing Issues
List issues with filters.
```bash
tea issues list --repo <owner>/<repo> --state open
```
*   Use `--output csv` or `--output json` for machine parsing.

### Creating Issues
Creating an issue non-interactively requires specific flags.
```bash
tea issue create --title "Issue Title" --body "Description here" --repo <owner>/<repo>
```

### Viewing Details
Detailed view of an issue or PR.
```bash
tea issue <index> --repo <owner>/<repo>
```

### Commenting
Add a comment to an issue.
```bash
tea comment <issue_index> "My specific comment" --repo <owner>/<repo>
```

## Releases

### Creating a Release
Create a new release and attach assets.
```bash
tea release create --tag <tag_name> --title "Release Title" --note "Release notes..." --asset <path/to/file> --repo <owner>/<repo>
```

## Environment Variables & Context
*   **Auto-Repo Detection:** `tea` tries to detect the repository context from the current `.git` directory. However, in scripts or agent actions, it is **Best Practice** to always explicit specify `--repo <owner>/<repo>` to ensure safety.
*   **Non-Interactive:** `tea` is generally CLI-friendly, but ensure necessary flags (like `--title` for creation) are provided to avoid interactive prompts.

## Wikis
The `tea` CLI does not currently have direct support for Gitea wikis. Wikis in Gitea are implemented as separate Git repositories.

### Interacting with the Wiki
To interact with a repository's wiki, clone its specific wiki Git URL:
- **Wiki URL Format:** `https://gitea.gnomatix.com/<owner>/<repo>.wiki.git`
- **Action:** Use standard `git` commands (`clone`, `pull`, `push`, `commit`) to manage wiki content as Markdown files.

## Troubleshooting
*   **Interactive Prompts:** If a command returns "user should be an owner" or prompts for Y/N:
    *   Verify the `--login` has the correct permissions.
    *   Check for missing mandatory flags (`--repo`).
    *   Ensure stdin is redirected (e.g., ` < NUL`).
