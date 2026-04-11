# Poplar Status

**Current state:** bubbletea-design skill done. Chrome shell prototype
complete with tab bar, status bar, command footer, and focus cycling.
Ready for keybinding design (Pass 2.5b-keys).

## Passes

| Pass | Goal | Status |
|------|------|--------|
| 1 | Scaffold + Fork | done |
| 2 | Backend Adapter + Connect | done |
| 2.5-render | Lipgloss migration: block model + compiled themes | done |
| 2.5-fix | Fix first-level blockquote wrapping (BACKLOG #7) | done |
| 2.5a | Text wireframes for all screens | done |
| 2.5b-1 | Prototype: chrome shell | done |
| 2.5b-keys | Keybinding design: single-key scheme for all screens | pending |
| 2.5b-2 | Prototype: sidebar | pending |
| 2.5b-3 | Prototype: message list | pending |
| 2.5b-4 | Prototype: message viewer | pending |
| 2.5b-5 | Prototype: help popover | pending |
| 2.5b-6 | Prototype: status/toast system | pending |
| 2.5b-7 | Prototype: command mode | pending |
| 3 | Wire prototype to live backend | pending |
| 6 | Triage actions | pending |
| 7 | Command mode + search | pending |
| 8 | Gmail IMAP | pending |
| 9 | Compose + send | pending |
| 10 | Config | pending |
| 11 | Polish for daily use | pending |

## Plans

- [Design spec](../superpowers/specs/2026-04-09-poplar-design.md)
- [UI design spec](../superpowers/specs/2026-04-10-poplar-ui-wireframing-design.md)
- [Lipgloss migration spec](../superpowers/specs/2026-04-10-mailrender-lipgloss-design.md)
- [Lipgloss migration plan](../superpowers/plans/2026-04-10-mailrender-lipgloss.md)
- [Pass 1 plan](../superpowers/plans/2026-04-09-poplar-pass1-scaffold.md)
- [Pass 2 plan](../superpowers/plans/2026-04-09-poplar-pass2-backend-adapter.md)
- [Pass 2.5a wireframe plan](../superpowers/plans/2026-04-10-poplar-wireframes.md)
- [Pass 2.5b-1 chrome shell plan](../superpowers/plans/2026-04-10-poplar-chrome-shell.md)
- [Chrome shell design spec](../superpowers/specs/2026-04-10-poplar-chrome-shell-design.md)
- [Wireframes](../poplar/wireframes.md)
- [bubbletea-design skill spec](../superpowers/specs/2026-04-10-bubbletea-design-skill-design.md)
- [bubbletea-design skill plan](../superpowers/plans/2026-04-10-bubbletea-design-skill.md)
- [Sidebar plan](../superpowers/plans/2026-04-10-poplar-sidebar.md)

## Continuing Development

### Next steps

1. **Execute Pass 2.5b-keys** — keybinding design for all screens
2. **Execute Pass 2.5b-2** — sidebar prototype (uses keybinding decisions)

### Next starter prompt

> Start Pass 2.5b-keys: keybinding design. Bubbletea sends one
> KeyMsg per keypress — no multi-key chords. Design a complete
> single-key binding scheme covering sidebar navigation (folder
> jumps), message list, viewer, and global actions. Read the
> architecture doc at `docs/poplar/architecture.md` (especially
> "No multi-key sequences" and "Vim-first keybindings"), the
> wireframes at `docs/poplar/wireframes.md`, and BACKLOG #8.
> Brainstorm approaches, write a design spec, then update
> `keys.go` and the command footer.

### After keybinding design

> Start Pass 2.5b-2: sidebar prototype. Read the plan at
> `docs/superpowers/plans/2026-04-10-poplar-sidebar.md`, the
> wireframes at `docs/poplar/wireframes.md`, and the architecture
> doc at `docs/poplar/architecture.md`. Execute the plan — the
> keybinding scheme from Pass 2.5b-keys informs which keys the
> sidebar handles.

### Pass-end checklist

1. `/simplify` — code quality review
2. Update `docs/poplar/architecture.md` — design decisions
3. Update this file — mark pass done, next starter prompt
4. Update docs appropriate to the pass stage
5. Commit all changes
6. `git push`
