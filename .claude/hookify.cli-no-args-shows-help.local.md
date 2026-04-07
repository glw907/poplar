---
name: cli-no-args-shows-help
enabled: true
event: file
conditions:
  - field: file_path
    operator: regex_match
    pattern: cmd/.*\.go$
  - field: new_text
    operator: regex_match
    pattern: cobra\.Command
---

**CLI commands must show help when run with no arguments and no piped stdin** (clig.dev, POSIX)

Do not let a command hang waiting for stdin when invoked interactively with no arguments. If the command requires input (args or stdin), detect a TTY on stdin and show help instead of blocking.

For cobra root commands with subcommands, this is automatic (no `Run` set). For leaf commands that read stdin, check `os.Stdin` with `term.IsTerminal` and call `cmd.Help()` if it's a TTY with no file arguments.
