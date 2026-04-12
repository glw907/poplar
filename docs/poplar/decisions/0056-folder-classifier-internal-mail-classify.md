---
title: Folder classifier in `internal/mail/classify.go`
status: accepted
date: 2026-04-12  # Pass 2.5b-3.5
---

## Context

Folder identity was previously scattered between
`sidebar.go:classifyGroup` (role-only) and `sidebar.go:sidebarIcon`
(role + name heuristic). Moving it to a pure function in the mail
package makes it testable in isolation, shareable between the
sidebar renderer and `poplar config init`, and backend-agnostic —
IMAP workers with `\Special-Use` flags set `Folder.Role` the same
way JMAP does. The alias table is the fallback for providers that
don't send role metadata (or send it inconsistently).

## Decision

A pure `Classify(folders []Folder) []ClassifiedFolder`
function in `internal/mail/` maps raw backend folders to canonical
identity + group (Primary / Disposal / Custom). Priority is role
attribute → alias table → Custom fallback. The alias table is
verified against Gmail IMAP, Fastmail JMAP, Outlook/M365, iCloud,
Yahoo/AOL, and Proton Mail Bridge.

## Consequences

No follow-on notes recorded.
