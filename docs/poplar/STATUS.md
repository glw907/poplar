# Poplar Status

**Current state:** Pass 2 complete. Ready to start Pass 3.

## Passes

| Pass | Goal | Status |
|------|------|--------|
| 1 | Scaffold + Fork | done |
| 2 | Backend Adapter + Connect | done |
| 3 | Bubbletea Shell | pending |
| 4 | Message List | pending |
| 5 | Message Viewer | pending |
| 6 | Triage Actions | pending |
| 7 | Command Mode + Search | pending |
| 8 | Gmail IMAP | pending |
| 9 | Compose + Send | pending |
| 10 | Keybindings + Config | pending |
| 11 | Polish for Daily Use | pending |

## Plans

- [Design spec](../superpowers/specs/2026-04-09-poplar-design.md)
- [Pass 1 plan](../superpowers/plans/2026-04-09-poplar-pass1-scaffold.md)
- [Pass 2 plan](../superpowers/plans/2026-04-09-poplar-pass2-backend-adapter.md)

## Continuing Development

### Next starter prompt

> Start Pass 3: Bubbletea Shell. Add bubbletea + lipgloss + bubbles
> dependencies. Build the root model (`ui/app.go`) with tab bar, tab
> interface, sidebar component showing live folder list from the
> backend, and theme-to-lipgloss bridge reading theme TOML. See
> `docs/poplar/STATUS.md` for context and
> `docs/superpowers/specs/2026-04-09-poplar-design.md` for the full
> spec.

### Pass-end checklist

1. `/simplify` — code quality review
2. Update `docs/poplar/architecture.md` — add design decisions made
3. Update this file — mark pass done, update current state, set next
   starter prompt
4. Commit all changes
5. `git push`
