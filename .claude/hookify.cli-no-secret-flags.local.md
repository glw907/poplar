---
name: cli-no-secret-flags
enabled: true
event: file
conditions:
  - field: file_path
    operator: regex_match
    pattern: cmd/.*\.go$
  - field: new_text
    operator: regex_match
    pattern: Flags\(\)\.\w+Var.*"(password|token|secret|api.key|api.token)"
---

**Never accept secrets via CLI flags** (clig.dev, security best practice)

Secrets passed as `--password=xyz` or `--token=xyz` are visible in `ps` output and shell history. Use environment variables, files (`--token-file`), or stdin instead.
