---
title: Tea.Cmd-based backend I/O before Pass 3
status: accepted
date: 2026-04-12  # Pass 2.5b-3.5
---

## Context

Pass 3 wires real JMAP/IMAP backends. Their
`ListFolders` and `FetchHeaders` calls take 200–500ms. Running
them in `Update` or constructors would freeze the UI on every
keypress and on startup. Fixing the pattern now — while the mock
backend is instant and the regression surface is small — is
cheaper than landing JMAP latency on top of a blocking Update
loop. This also matches the Elm architecture the `CLAUDE.md`
elm-conventions file mandates.

## Decision

`AccountTab.Init` returns a `loadFoldersCmd`. J/K
navigation dispatches `loadFolderCmd(name)`. Results come back as
`foldersLoadedMsg` and `folderLoadedMsg` handled in `Update`. The
synchronous `ListFolders` call in `NewAccountTab` and the
`loadSelectedFolder` helper in `handleKey` are both gone.
`AccountTab` emits `FolderChangedMsg` when selection moves; `App`
consumes it to update the status bar instead of reaching through
`m.acct.sidebar.SelectedFolderInfo()`. The dead `case ":":` stub
in `App.Update` is deleted, and the `: cmd` rank-0 footer hint
is removed alongside it.

## Consequences

No follow-on notes recorded.
