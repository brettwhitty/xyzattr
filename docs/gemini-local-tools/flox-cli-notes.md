---
type: MANUAL
authority: PENDING MANUAL REVIEW
review_status: PENDING
version: 0.1.1
approved_versions: 0.1.*
generated_on: 2026-01-22 17:10
origin_persona: Rory Devlin (R&D Team Lead)
origin_session: d346cd76-699f-4ab0-b24a-53180588cb07
intent: Document usage and features of the Flox CLI tool.
primary_sources: [docs/gemini-local-tools/flox-cli-notes.md]
release_path: docs/gemini-local-tools/flox-cli-notes.md
related_issues: []
related_sops: []
tags: [flox, cli, packages]
---

# Flox CLI Tool Notes

This document records specific `flox` commands used and their purposes, for quick reference.

## Searching for Packages

- **Command:** `lxc exec gemini-exec-host -- flox activate -- flox search <keyword>`
- **Purpose:** To search for available software packages within the `gemini-exec-host` environment before installation. This ensures I can identify the correct package name.

## Installing Packages

- **Command:** `lxc exec gemini-exec-host -- flox activate -- flox install <package-name>`
- **Purpose:** To install a specified software package within the `gemini-exec-host` environment. This is the mandatory method for software installation.
