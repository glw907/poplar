---
paths:
  - "cmd/fastmail-cli/**"
  - "internal/jmap/**"
  - "internal/rules/**"
  - "e2e-fastmail/**"
---

# fastmail-cli

Fastmail JMAP CLI, built as a binary from the beautiful-aerc module.

## Command Structure

    fastmail-cli
      rules         Manage mail filter rules
        interactive   Full interactive filter creation flow
        add           Add a filter rule
        sweep         Move matching messages
        count         Count matching messages
        export        Copy rules to export destination
        export-check  Check if export is needed
        extract       Extract header fields from a message
      masked        Manage masked email addresses
        delete        Delete a masked email address
      folders       List custom mailboxes
      version       Print version

## Environment Variables

    FASTMAIL_API_TOKEN       Fastmail API token (required for JMAP commands)
    AERC_RULES_FILE          Path to rules file (default: ~/.config/aerc/mailrules.json)
    AERC_RULES_EXPORT_DEST   Export destination (default: ~/Documents/mailrules.json)
