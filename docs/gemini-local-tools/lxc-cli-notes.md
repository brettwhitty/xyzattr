---
type: MANUAL
authority: PENDING MANUAL REVIEW
review_status: PENDING
version: 0.1.1
approved_versions: 0.1.*
generated_on: 2026-01-22 17:15
origin_persona: Rory Devlin (R&D Team Lead)
origin_session: d346cd76-699f-4ab0-b24a-53180588cb07
intent: Document LXC CLI commands for container management.
primary_sources: [docs/gemini-local-tools/lxc-cli-notes.md]
release_path: docs/gemini-local-tools/lxc-cli-notes.md
related_issues: []
related_sops: []
tags: [lxc, cli, containers]
---

# LXC CLI Tool Notes

This document records specific `lxc` commands used and their purposes, for quick reference.

## Executing Commands in `gemini-exec-host`

- **Command:** `lxc exec gemini-exec-host -- flox activate -- <COMMAND>`
- **Purpose:** To execute a specific command within the `gemini-exec-host` LXD container, ensuring it runs in a `flox` activated environment. This is the primary method for me to interact with the development virtual host.

## Listing LXD Containers

- **Command:** `lxc list`
- **Purpose:** To view a list of all LXD containers and their current status, which is useful for verifying the `gemini-exec-host` is running.
