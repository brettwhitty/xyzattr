---
type: MANUAL
authority: Brett Whitty
review_status: APPROVED
version: 0.1.0
approved_versions: 0.1.0
generated_on: 2026-02-10
origin_persona: Brett Whitty
origin_session: 3e4870c0-37e1-40ad-a20e-e181db348cc4
intent: Document usage of the Gitea Model Context Protocol (MCP) server integration.
primary_sources: [user request]
release_path: docs/gemini-local-tools/gitea-mcp-notes.md
related_issues: []
related_sops: []
tags: [mcp, gitea, tools, agent]
---

# Gitea MCP Integration Notes

The Gitea MCP server allows the agent to interact directly with the configured Gitea instance (`gitea.gnomatix.com`) without relying solely on the legacy `tea` CLI tool or manual API calls.

## Key Capabilities

*   **Repository Access:** Search, list, and inspect repositories (`gnomatix/*` and user repos).
*   **Issue Management:** Create, read, update, and comment on issues directly.
*   **Content Management:** Read file contents, directories, and branches.
*   **Pull Requests:** Interact with PRs and reviews.

## Common Tools

| Tool Name | Description | Example Usage |
| :--- | :--- | :--- |
| `mcp_gitea_search_repos` | Search for repositories by keyword | `{ "keyword": "dreamfs", "page": 1 }` |
| `mcp_gitea_get_issue_by_index` | Retrieve detailed issue context | `{ "owner": "gnomatix", "repo": "dreamfs", "index": 7 }` |
| `mcp_gitea_create_issue_comment` | Post a comment to an issue | `{ "owner": "gnomatix", "repo": "dreamfs", "index": 7, "body": "Fixed in commit ABC." }` |
| `mcp_gitea_get_file_content` | Read a file from the repo | `{ "owner": "gnomatix", "repo": "dreamfs", "filePath": "README.md", "ref": "develop" }` |

## Usage Policies

1.  **Prioritize MCP over `tea`:** Use MCP tools for standard interactions (issues, comments, reading files) as they provide structured JSON output and better error handling than parsing `tea` CLI stdout.
2.  **Fallback to `tea`:** Use the `tea` CLI only for operations not yet supported by the MCP server or for complex batch scripts.
3.  **Context Awareness:** Always specify the correct `owner` (e.g., `gnomatix`) and `repo` context.
