# My preserved implementation candidates

These complete patches preserve unfinished work from redundant task checkouts. They are committed as review artifacts; this archive does not install the candidates into the runtime or establish that the missing platform checks passed. Apply each patch only to its recorded base in an isolated checkout. The manifest records every changed file and its SHA-256, plus the complete patch hash.

- `darwin-native.patch`: native Darwin candidate and tests; follow-up `dreamfs-x4n.4.3` remains dependent on native execution.
- `windows-boundaries.patch`: Windows boundary candidate and tests; follow-up `dreamfs-x4n.4.6` retains native UNC/network/mounted-volume validation.

The already published repairs in repository history remain the runtime implementation. Original investigation and validation records remain in the DreamFS project vault and evidence cache. No successful native validation is inferred from patch preservation.
