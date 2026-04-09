---
paths:
  - ".config/aerc/**"
  - ".config/nvim-mail/**"
  - ".config/kitty/**"
  - ".local/bin/**"
---

# aerc Config Changes

**MANDATORY: When changing any file under `.config/aerc/`, apply the
same change to both locations:**
- **Project repo**: `.config/aerc/` (this repo)
- **Personal dotfiles**: `~/.dotfiles/beautiful-aerc/.config/aerc/`

The live config is symlinked from `~/.dotfiles/`, so project-only
changes don't take effect. The dotfiles copy may have additional
personal customizations -- apply the same logical fix without
overwriting personal additions.

## Personal Config

This project ships working defaults that any user can stow directly
from their clone. The author's personal configs live in
`~/.dotfiles/beautiful-aerc/` (workstation repo) as a real stow
package -- not a symlink to this project. Personal differences from
repo defaults:

- `binds.conf` -- all optional bindings enabled (fastmail-cli,
  aerc-save-email)
- `signature.md` -- real signature (repo ships `.example` only)
- `mailrules.json` -- personal mail rules (not in this repo)
- `accounts.conf` -- personal credentials (repo ships `.example` only)
