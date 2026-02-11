---
type: MANUAL
authority: PENDING MANUAL REVIEW
review_status: PENDING
version: 0.1.1
approved_versions: 0.1.*
generated_on: 2026-01-22 17:15
origin_persona: Rory Devlin (R&D Team Lead)
origin_session: d346cd76-699f-4ab0-b24a-53180588cb07
intent: Document Mise CLI commands for runtime and tool management.
primary_sources: [docs/gemini-local-tools/mise-cli-notes.md]
release_path: docs/gemini-local-tools/mise-cli-notes.md
related_issues: []
related_sops: []
tags: [mise, cli, runtimes]
---

# Mise CLI Tool Notes

This document records specific `mise` usage and its purpose, for quick reference.

## Managing Language Runtimes

- **Usage Context:** `mise` must be used from within an activated `flox` environment.
- **Purpose:** To manage language runtimes (e.g., Node.js, Python, Go) within the `gemini-exec-host` environment. This ensures consistent and controlled language environments for development tasks.
- **Example Pattern:** `lxc exec gemini-exec-host -- flox activate -- mise <mise-command>`
