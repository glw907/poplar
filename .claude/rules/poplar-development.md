---
description: Development workflow for poplar implementation passes
---

When the user says "continue development", "next pass", "start the
next pass", "finish pass", "ship pass", or "continue" in the context
of poplar work, invoke the `poplar-pass` skill. It handles both pass
start (read STATUS, invariants, plan, execute) and pass end (the
consolidation ritual: ADRs, invariants update, plan archival,
commit + push + install).
