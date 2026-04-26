# Findings: Plan Shape (Forward Look) — 2026-04-25

Independent review of `STATUS.md`, `BACKLOG.md`, the next-starter
prompt, and ADR-0022 against current state. Method per
`docs/poplar/audits/2026-04-25-plan-shape.md`. Forward-looking only —
shipped passes are not re-litigated.

---

## Summary

| Pass | Verdict |
|---|---|
| 2.5b-4.5 (Audit-1+2 fixes) | **keep** |
| 2.5b-5 (help popover) | **re-scope** — answer "future bindings?" question first |
| 2.5b-6 (status/toast) | **split** — defer toast + undo to Pass 6; keep error banner + spinner consolidation pre-backend |
| 2.5b-train (mailrender training) | **keep** |
| 2.9 (library survey) | **re-scope + rename** — focus on BACKLOG #10 (emersion vs aerc fork), drop "SMTP/parser" framing |
| 3 (wire to live backend) | **keep** (contingent on 2.9 outcome) |
| 6 (triage actions) | **keep** |
| 8, 9, 9.5, 10, 11, 1.1 | out of scope per audit |
| ADR-0022 | **supersede** — actual sub-pass set has drifted from the original 7 |
| BACKLOG #8 | **close** — 2.5b-4.5 wires the folder jumps |

The two recommendations that matter most: split 2.5b-6 (the toast/undo
half is prototype-without-a-consumer until triage exists) and tighten
2.9 from "library survey" to "emersion vs aerc fork eval" (the
broader survey framing invites scope creep on a question that's
already focused).

---

## Pass 2.5b-4.5 — Audit-1+2 fixes

**Verdict: keep.**

Scope is mechanical and well-bounded: 5 invariant fixes + 3 filter-
package cleanups. All eight items are independently verified by their
respective audits and the starter prompt cites file/line numbers.
U14 (offline color) is the only design call — and the prompt
correctly flags it as such with two pre-resolved options. No
brainstorm needed; the "implement directly" framing is right.

One small note: the prompt's U6 wording ("`MessageList.SetThreaded
(bool)` that short-circuits `bucketByThreadID` to one bucket per
message when false") is a real implementation choice that warrants a
brief check before coding — `bucketByThreadID` may have other
consumers (search filtering?) that don't want the short-circuit.
Cheap to verify; not worth re-scoping the pass.

## Pass 2.5b-5 — Help popover

**Verdict: re-scope.**

The wireframe (wireframes.md §5) is clear about layout, modal
behavior, and the two contexts (account, viewer). What's not settled
is **which bindings the popover should advertise.** The current
wireframe shows `c compose`, `r reply`, `R all`, `f forward`,
`d delete`, `a archive`, `s star`, `v select`, `n/N` (search-next /
filtered-walk), and other keys that **don't exist yet** — they're
slated for Pass 6 (triage), Pass 9 (compose), or BACKLOG #9 (n/N
filtered walk).

Three choices, none neutral:

1. **Show only currently-wired keys.** Popover changes shape every
   pass. Users learn "what works today." Risk: looks sparse and
   incomplete pre-Pass 6.
2. **Show all eventual keys.** Stable popover from day one. Users
   hit "key does nothing" or worse, partial behavior. No way to
   signal "not yet" without a feedback channel — and the toast
   system that would carry that signal is itself in 2.5b-6.
3. **Show all eventual keys with a visual "future" marker** (e.g.
   `c compose ·` dimmed). Hybrid — discoverability without false
   promises. Costs a small styling decision.

This question wants a brainstorm before the plan doc. The starter
prompt for 2.5b-5 should mark it as **open**, not **settled**. Once
answered, the implementation pass is straightforward (modal
overlay, key routing, two context layouts).

Sequencing inside the pass is fine — the popover only needs the
overlay infrastructure, not any of the bindings it advertises.
2.5b-4.5 wiring the folder jumps doesn't gate 2.5b-5; it just
makes the popover honest about the `Go To` group.

## Pass 2.5b-6 — Status/toast system

**Verdict: split.**

Wireframes §6 lumps five distinct things into one prototype:

- **Status toast** (action feedback) — needs an action; triage =
  Pass 6.
- **Undo bar** (deferred destructive action) — needs a destructive
  action; triage = Pass 6.
- **Error banner** — currently triggerable. `xdg-open` and mark-read
  errors silently drop today (see invariants §A21, §U19); a banner
  could surface them now.
- **Loading spinner** — already partially built (msglist + viewer
  loading phases). Wireframes call for centralization into the
  transient-UI region.
- **Connection status** — already built in status bar; just needs
  the offline color settled (U14, in 2.5b-4.5).

Toast and undo are prototype-without-a-consumer. They demonstrate the
mechanism but exercise it with synthetic events. Building them now
means re-touching them when Pass 6 lands a real consumer — that's
exactly the rework the per-screen-prototype strategy is meant to
avoid.

Recommended split:

- **2.5b-6 (re-scoped)** — error banner + spinner consolidation +
  any cleanup the connection indicator still needs after 2.5b-4.5.
  Lands before Pass 3. Real consumers exist (mark-read errors,
  body-fetch errors, JMAP connection drop in Pass 3).
- **Pass 6.5 (new)** or **bundle into Pass 6** — toast + undo bar.
  Lands with triage actions, where the consumer is real on day one.

The wireframes don't need to change; the priority block ("error >
undo > toast > normal status") is a stable mental model that survives
the split. Only the implementation order moves.

## Pass 2.5b-train — Mailrender training capture

**Verdict: keep.**

Already explicitly deferred to "after Pass 3" in the STATUS table
(`pending (after Pass 3)`). That deferral is right — training
capture only earns its keep against real bodies arriving through
the live backend, not the mock. Spec exists at
`docs/superpowers/specs/2026-04-12-mailrender-training-design.md`.
No re-scope.

## Pass 2.9 — Library survey

**Verdict: re-scope + rename.**

Currently labeled "Research: JMAP/IMAP/SMTP/parser library survey."
The framing is a 2026-04-15 artifact that conflated several concerns
into one pass header. The actual driver is **BACKLOG #10** —
evaluating migration from the aerc fork to the emersion ecosystem.
That's a focused question with a clear blocker (no Go JMAP client in
emersion) and a well-scoped first step (WebFetch pkg.go.dev).

"SMTP/parser library survey" doesn't have a backlog entry, doesn't
have a question to answer, and reads like scope expansion that would
delay Pass 3 indefinitely. Drop it.

Recommended reshape:

- **Rename Pass 2.9** to "Emersion vs aerc fork evaluation."
- **Scope** to BACKLOG #10's three options: drop JMAP / hybrid /
  find-or-write Go JMAP client. Output is a recommendation
  document, not code. The recommendation either (a) confirms the
  current fork is right (no-op, Pass 3 proceeds unchanged), or
  (b) reverses ADR-0058 and reshapes Pass 3.
- **Keep the position** — between 2.5b-6 and Pass 3, since its
  output may reshape Pass 3.

If the survey ever grows back (SMTP for compose, parser libraries
for headers), each addition gets its own pass with its own driving
question.

## Pass 3 — Wire prototype to live backend

**Verdict: keep, contingent on 2.9 outcome.**

Scope is right: replace mock backend with real JMAP/IMAP, exercise
every prototype (sidebar, msglist, viewer, search, threading)
against real bodies. BACKLOG #11 (MIME-aware body fetch) is the
identified prereq and correctly tagged.

Two contingencies worth surfacing now:

- **If 2.9 recommends emersion migration**, Pass 3's implementation
  shape changes substantially — mailworker fork goes away, MIME
  handling moves into emersion's `go-message`, Pass 3 becomes
  "wire prototype to emersion stack." The pass goal (live backend
  validation) doesn't change; the surface area does.
- **BACKLOG #4 (JMAP blob preloading)** is currently tagged
  `#upstream` and parked as an aerc patch concern. If Pass 3
  surfaces real per-message latency >2s on Fastmail, this should
  promote to in-pass scope or to a fast-follow pass — slow opens
  break the optimistic-mark-read UX. Worth re-evaluating after 2.9
  decides on the worker substrate.

BACKLOG #9 (viewer n/N filtered walk) bundles correctly here —
prefetch semantics need real latency to be designed against.

## Pass 6 — Triage actions

**Verdict: keep.**

Position between Pass 3 (live backend) and Pass 8 (Gmail IMAP) is
correct: triage needs the live backend to exercise mark-read,
delete, archive, star against real folders. Adding Gmail before
triage would double the test matrix without adding design value.

If 2.5b-6 is split per the recommendation above, the toast + undo
half lands here as a single integrated pass rather than two
sequential ones — bundling reduces churn.

## ADR-0022 — Per-screen prototype sub-passes

**Verdict: supersede.**

Original ADR (dated 2026-04-10) lists 7 sub-passes: chrome shell,
sidebar, message list, viewer, help popover, status/toast, command
mode. The actual sequence has drifted:

- **Chrome shell** — done as 2.5b-1 ✓
- **Sidebar** — done as 2.5b-2 ✓
- **Message list** — done as 2.5b-3 ✓
- **Threading** — added as a 2.5b-3.x sub-pass (not in original ADR)
- **Search** — done as 2.5b-7 (replaced "command mode" slot)
- **Viewer** — done as 2.5b-4 ✓
- **Help popover** — pending as 2.5b-5 ✓
- **Status/toast** — pending as 2.5b-6 ✓
- **Command mode** — dropped (ADR-0024 / no `:` command mode)

The intent ("each screen is a learning opportunity, incremental
validation") still holds. The decision text is wrong about which
seven sub-passes and in what order. Either:

- Update ADR-0022 in place to list the actual sub-passes, **or**
- Mark ADR-0022 superseded and write a new ADR documenting the
  realized sequence and what changed (command mode dropped,
  threading + search added).

The supersede route is more honest — superseding records that
the design strategy survived but the specific sub-pass plan
adapted to what each screen taught.

## BACKLOG bundling review

| # | Title | Current bundling | Verdict |
|---|---|---|---|
| 4 | JMAP blob preloading | `#upstream`, parked | **revisit after 2.9** — may promote to Pass 3 follow-up |
| 5 | Built-in bubbletea compose editor | `#v2` | keep — post-1.0 |
| 6 | Neovim companion plugin | `#v2` | keep — Pass 1.1 |
| 8 | Folder jump keybinding design | "deserves its own pass" | **close** — 2.5b-4.5 wires I/D/S/A/X/T per the settled invariant |
| 9 | Viewer n/N walks filtered set | bundled with Pass 3 | keep |
| 10 | Emersion ecosystem migration | "medium" | promote — this is Pass 2.9 |
| 11 | MIME-aware body fetch | Pass 3 prereq | keep |
| 12 | Tidy/ collapse | Pass 9.5 prereq | keep |
| 13 | Drop blockKind/spanKind enums | "ride along next time content/ is touched" | keep |

#8 is the standout — it's described as deserving "its own pass" but
the design call has already been made (uppercase single-key
mnemonics, codified in ADR / invariants U5). 2.5b-4.5 is the wiring
pass; #8 has nothing left to do beyond what 2.5b-4.5 handles. Close
on Pass-2.5b-4.5 commit.

## STATUS.md edits (if accepted)

```diff
 | 2.5b-4.5 | Audit-1+2 fixes: ... | next |
-| 2.5b-5 | Prototype: help popover | pending |
-| 2.5b-6 | Prototype: status/toast system | pending |
+| 2.5b-5 | Prototype: help popover (open: future-binding policy) | pending |
+| 2.5b-6 | Prototype: error banner + spinner consolidation | pending |
 | 2.5b-train | Tooling: mailrender training capture system | pending (after Pass 3) |
-| 2.9 | Research: JMAP/IMAP/SMTP/parser library survey | pending |
+| 2.9 | Research: emersion vs aerc fork evaluation (BACKLOG #10) | pending |
 | 3 | Wire prototype to live backend | pending |
 | 6 | Triage actions | pending |
+| 6.5 | Toast + undo bar (or bundle into Pass 6) | pending |
 | 8 | Gmail IMAP | pending |
```

ADR-0022 needs either an in-place edit listing the realized sub-pass
set, or a supersede + new ADR documenting the drift.

BACKLOG.md: close #8 with a one-line resolution pointing at the
2.5b-4.5 commit.

## What this audit does NOT recommend

- **No code changes.** Recommendations are scope and ordering only.
- **No re-litigating shipped passes.** 2.5b-3.6 (threading) and
  2.5b-7 (search) are not in ADR-0022 but are done. They earned
  their place; the supersede is for accuracy, not blame.
- **No changes to Pass 9 / 9.5 / 10 / 11 / 1.1.** Too far out per
  the audit's own scope.
- **No changes to the "Settled" block in 2.5b-4.5's starter prompt.**
  The pass is queued and ready; this audit doesn't add work to it.
