# internal/mailauth

Small vendored snippets that fill gaps in the emersion mail stack.

| File | Origin | License | Why |
|---|---|---|---|
| `xoauth2.go` | aerc `auth/xoauth2.go` | MIT | XOAUTH2 SASL mech (`emersion/go-sasl` does not ship one). |
| `keepalive/keepalive.go` | aerc `lib/keepalive.go` (or similar) | MIT | TCP keepalive helper for long-lived IMAP/JMAP connections. |

Each file carries a provenance comment recording the source commit
and any modifications. Update those comments if upstream changes
land.
