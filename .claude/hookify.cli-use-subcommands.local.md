---
name: cli-use-subcommands
enabled: true
event: file
conditions:
  - field: file_path
    operator: regex_match
    pattern: cmd/.*root\.go$
  - field: new_text
    operator: regex_match
    pattern: cobra\.Command
---

**Use subcommands (verbs) for multi-action CLI tools** (clig.dev, cobra convention)

Structure CLIs as `tool verb [args]` (e.g., `tidytext fix file.txt`, `fastmail-cli rules add`). Each distinct action gets its own subcommand. The root command with no arguments shows help.

Single-action tools (grep, wc) are the exception — but if the tool has or may grow to have multiple actions, use subcommands from the start.
