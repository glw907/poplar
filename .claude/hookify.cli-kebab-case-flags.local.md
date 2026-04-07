---
name: cli-kebab-case-flags
enabled: true
event: file
conditions:
  - field: file_path
    operator: regex_match
    pattern: cmd/.*\.go$
  - field: new_text
    operator: regex_match
    pattern: Flags\(\)\.\w+Var.*"[a-z]+[A-Z]|Flags\(\)\.\w+Var.*"[a-z]+_[a-z]
---

**Multi-word flags must use kebab-case** (clig.dev, POSIX/GNU convention)

Use `--output-format`, not `--outputFormat` or `--output_format`. This is the universal convention used by kubectl, gh, docker, and all major Go CLIs.
