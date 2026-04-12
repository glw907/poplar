---
title: Message list color: brightness, not hue
status: accepted
date: 2026-04-11  # Pass 2.5b-3
---

## Context

The earlier scheme used four hues per row (cursor blue
+ unread teal + answered purple + flagged orange) on a muted Nord
background. No single element won the eye, and the row read as
garish. This is Tufte's data-ink principle applied to color: spend
hue on the data that demands attention, withhold it everywhere else.
Apple Mail, Fastmail, Gmail, and Mutt all encode unread by brightness
or weight, not hue — the convention exists because it works. The
general rule is now codified in the `bubbletea-design` skill's "Hue
Budget" subsection so future TUI work picks it up automatically;
the poplar-specific application lives in `docs/poplar/styling.md`.

## Decision

Read state in the message list is encoded by brightness
(`FgBright` for unread, `FgDim` for read), not by hue. Glyphs carry
the flag/answered/unread distinction. Color hue is reserved for the
two states that genuinely demand attention: the cursor (`AccentPrimary`
on `▐`) and the unread+flagged row (`ColorWarning` on the `󰈻`
glyph). A read+flagged row dims the flag glyph along with the rest of
the row — read state always wins over flag state for color. Replaced
the per-flag-type hue scheme (`MsgListFlagUnread` teal +
`MsgListFlagAnswered` purple + `MsgListFlagFlagged` orange) with two
brightness-based icon styles (`MsgListIconUnread` /
`MsgListIconRead`) plus the narrowed `MsgListFlagFlagged` for the
single attention-worthy combination.

## Consequences

No follow-on notes recorded.
