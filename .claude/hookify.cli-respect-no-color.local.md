---
name: cli-respect-no-color
enabled: true
event: file
conditions:
  - field: file_path
    operator: regex_match
    pattern: \.go$
  - field: new_text
    operator: regex_match
    pattern: \\033\[|\\x1b\[|\\e\[
---

**Respect the NO_COLOR environment variable** (no-color.org)

If the code emits ANSI color codes, it must check `os.Getenv("NO_COLOR")` and disable color when the variable is set to any non-empty value. Also disable color when stdout is not a TTY.
