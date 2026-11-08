# xyzattr agent entry point

> Best tool for the job is a universal rule always

> but 'beads' scales better with agents, obviously, so we're doing that for agent issues; we can still use gitea for PM-type issue-tracking; project planning stuff; but I'm working on getting a locally hosted Jira instance back up and operational for that. It's better than Gitea issue tracking; Gitea sucks for issue tracking.

> Please, as you review the vault, and the issue, etc. the "vision" docs in the kiro specs dir are the most authoritative unless you find something newer that is CLEARLY authored by me (not just has my name / email address on it) --- check all the dates. And please please, as you go, add comments and tags to gitea flagging stuff that might be "obsoleted" and therefore "putative context poison risk" --- keep a document or documents in the project vault space listing those with checkboxes for "needs review", "obsolete", "recommend archiving" ... useful fields like that.

> And then a simple and universal priority 20 rule "Don't presume to revise design docs if (A) you're not the human user and the human user didn't directly instruct you to revise the design docs with the human user's intended revisions (B) you haven't even read the design docs yet"

## My operating instructions

Inspect the current host, checkout, existing tools and task ownership before acting. Use the existing Go toolchain and the repository's `mise.toml` tasks. Older examples naming `gemini-exec-host`, Flox or particular CLI flags do not establish that those tools, hosts or flags exist here. Check installed help and current host configuration before using them. Do not install an alternative merely because an old note names it.

Agent work for xyzattr is currently tracked under `dreamfs-x4n.4` in the DreamFS Beads store. Locate that owning workspace, verify the store with `bd where`, and run `bd prime` there for the current roster and handoff protocol. Claim work before implementation. Do not initialize a duplicate store or treat a passive JSONL export as the tracker. Keep PM/planning records in their existing Gitea context and reviewed knowledge in the Obsidian project space.

Read the project design sources in full before implementing their behavior: the original development plan, `docs/spec.md`, and the current DreamFS vision when working on integration. The current README and implementation records do not cancel preserved requirements. DreamFS vision lives in the owning DreamFS checkout at `.kiro/specs/dreamfs-vision`; find that checkout and its recorded clarifications rather than infer a missing design from this companion repository. Design revisions require both direct instructions supplying the intended revisions and a complete reading of the design documents. Corrections to an agent's understanding do not authorize design revisions.

Prefer the configured Gitea MCP tools for supported repository operations. Use other existing tools for capabilities it does not provide. Verify source dates and content provenance; account names, frontmatter approval fields and file timestamps alone do not prove substantive human authorship. A wiki lookup returning 404 is a source gap, not proof that a wiki never existed.

## My historical tool-note reference

Some older local checkouts contain ignored `.gemini/gemini-local-tools/` notes. Their presence is not guaranteed in a clone. The captured filenames are `flox-cli-notes.md`, `git-cli-notes.md`, `gitea-mcp-notes.md`, `jira-cli-notes.md`, `lxc-cli-notes.md`, `mise-cli-notes.md`, `tea-cli-notes.md`, and `wiki-docs.md`. Earlier pointers to `git-notes.md`, `mise-notes.md`, `tea-notes.md`, and `wiki-docs-notes.md` did not match those files.

These are historical tool examples, not product specifications or authority to change hosts, branch protection, tracker ownership or publication policy. Read ignored files when relevant; verify their applicability before execution. The vault note `XYZAttr_Hidden_Context_Review_2026-11-07` and its evidence attachment preserve the reviewed bytes, source dates and current limitations. `XYZAttr_Current_State_2026-11-07` records implementation; `XYZAttr_Historical_Design_2026-11-07` retains the original configuration and CLI passages. A source classification does not authorize deleting implementation.
