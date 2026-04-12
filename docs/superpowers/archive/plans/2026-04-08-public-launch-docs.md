# Public Launch Documentation Overhaul — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

> **IMPORTANT: All document authoring (README, docs, config comments) must be done by opus. Subagents (sonnet) may research, explore files, and make targeted edits, but opus writes the prose.**

**Goal:** Rewrite all user-facing documentation and config comments for a newcomer audience, preparing the project for public release.

**Architecture:** The work splits into independent document authoring tasks plus config cleanup. Each task produces a self-contained file that can be committed independently. Config onboarding fixes go first since they affect what the docs reference.

**Tone:** Friendly and educational. We want clear, well-organized documentation, but not dry or formal. The reader is a curious nerd exploring terminal email — write like you're showing a friend around a project you're proud of. Explain *why* before *how*. Be direct but warm.

**Tech Stack:** Markdown, INI (aerc config), Lua (nvim-mail), TOML (kitty)

**Spec:** `docs/superpowers/specs/2026-04-08-public-launch-docs-design.md`

---

## Task 1: Config Onboarding Fixes

**Files:**
- Modify: `.gitignore`
- Remove from git: `.config/aerc/mailrules.json` (if tracked)
- Create: `.config/aerc/mailrules.json.example`

- [ ] **Step 1: Verify mailrules.json git state**

mailrules.json is already in `.gitignore` and not tracked by git (confirmed during planning). No removal needed. Skip to creating the example.

- [ ] **Step 2: Create mailrules.json.example**

Create `.config/aerc/mailrules.json.example`:

```json
{
  "rules": []
}
```

Add a comment header. JSON doesn't support comments natively, so use a top-level `_comment` field:

```json
{
  "_comment": "Mail filter rules for fastmail-cli. See: fastmail-cli rules --help",
  "rules": []
}
```

- [ ] **Step 3: Add generated/ to .gitignore**

Check if `.config/aerc/generated/` is already ignored. Currently it's not in `.gitignore` but also not tracked. Add it to prevent accidental commits:

Append to `.gitignore`:
```
.config/aerc/generated/
```

- [ ] **Step 4: Verify .gitignore coverage**

Run `git status` and confirm none of these are tracked:
- `accounts.conf`
- `signature.md`
- `mailrules.json`
- `.config/aerc/generated/`

Expected: all absent from tracked files.

- [ ] **Step 5: Commit**

```bash
git add .config/aerc/mailrules.json.example .gitignore
git commit -m "Add mailrules.json.example and gitignore generated/"
```

---

## Task 2: Rewrite aerc.conf Comments

**Files:**
- Modify: `.config/aerc/aerc.conf`
- Modify: `~/.dotfiles/beautiful-aerc/.config/aerc/aerc.conf` (personal dotfiles copy)

The config values stay the same. Only comments change. Every section and non-obvious setting gets an educational comment explaining *why* it exists.

- [ ] **Step 1: Rewrite .config/aerc/aerc.conf comments**

Replace the entire comment structure. The new version should have:

```ini
#
# aerc main configuration — beautiful-aerc
#
# This config works out of the box with the beautiful-aerc filter suite.
# Only non-default values are set here. For all available options and
# their defaults, see the aerc-config(5) man page: `man aerc-config`
#

[ui]
# --- Message list columns ---
# The message list is the main view: one row per email. These settings
# control what columns appear and how wide they are.
#
# Layout: [padding] [flags] [sender] [subject] [date] [padding]
# The start/end columns add 1-char padding at the edges.
# The subject column (<*) fills remaining space.
index-columns=start>0,flags<2,name<22,subject<*,date>12,end>1
column-separator="  "
column-date={{.DateAutoFormat .Date.Local}}
column-start=
column-end=
column-flags={{.Flags | join ""}}
column-name={{.Peer | names | join ", "}}

# --- Sidebar ---
# The sidebar shows your mailbox folders on the left side of the screen.
# dirlist-tree=false keeps the list flat — aerc's tree mode sorts
# children alphabetically and ignores your preferred folder order.
sidebar-width=30
dirlist-tree=false

# Folder icons use Nerd Font symbols. If you don't have a Nerd Font
# installed, you'll see placeholder squares — remove this line or
# install a Nerd Font (https://www.nerdfonts.com/).
dirlist-left={{if eq .Folder "Inbox"}} 󰇰{{else if eq .Folder "Notifications"}} 󰂚{{else if eq .Folder "Drafts"}} 󰏫{{else if eq .Folder "Sent"}} 󰑚{{else if eq .Folder "Archive"}} 󰀼{{else if eq .Folder "Spam"}} 󰍷{{else if eq .Folder "Trash"}} 󰩺{{else if eq .Folder "Remind"}} 󰑴{{else}} 󰡡{{end}}  {{.Folder}}
dirlist-right= {{if .Unread}}{{humanReadable .Unread}} {{end}}

# --- Sort & threading ---
# Default: newest first. Inbox and Notifications override below to
# oldest first, so conversations read chronologically in your primary
# folders.
sort=-r date
threading-enabled=true

# --- Appearance ---
# styleset-name must match a theme filename in themes/ (without .toml).
# After changing, run: mailrender themes generate
# Then restart aerc to pick up the new colors.
styleset-name=nord
border-char-vertical=│
fuzzy-complete=true
mouse-enabled=true

# --- Tabs ---
# tab-title-composer shows only the subject line. The default includes
# "to:recipient" which wastes tab bar space when you have multiple tabs.
tab-title-account=Mail
tab-title-composer={{.Subject}}

# --- Status icons (Nerd Font Material Design) ---
# These appear in the flags column of the message list.
# Requires a Nerd Font — see sidebar note above.
icon-new=󰇮
icon-old=󰇮
icon-replied=󰑚
icon-forwarded=󰒊
icon-flagged=󰈻
icon-marked=󰄬
icon-draft=󰏫
icon-deleted=󰆴
icon-attachment=󰏢

# --- Thread display ---
# Box-drawing characters for clean thread visualization in the message list.
thread-prefix-tip = "›"
thread-prefix-indent = " "
thread-prefix-stem = "│"
thread-prefix-limb = "─"
thread-prefix-has-siblings = "├─"
thread-prefix-last-sibling = "└─"

# --- Per-folder sort overrides ---
# Inbox and Notifications use oldest-first so conversations read
# chronologically (you see the oldest unread message at the top).
[ui:folder=Inbox]
sort=date

[ui:folder=Notifications]
sort=date

[statusline]
column-left={{.StatusInfo}}

[viewer]
# --- MIME part preference ---
# When an email has both HTML and plain text parts, prefer HTML.
# Marketing emails and newsletters have better structure (paragraphs,
# headings, lists) in their HTML part. Their plain text parts are often
# a wall of text with no formatting at all.
alternatives=text/html,text/plain

# --- Custom header rendering ---
# beautiful-aerc replaces aerc's built-in header display with the
# mailrender headers filter, which gives us colored fields, address
# wrapping, and a consistent layout.
#
# show-headers=true tells aerc to pipe raw headers through the .headers
# filter. header-layout=X-Collapse is a trick: it tells aerc to display
# only the "X-Collapse" header in its built-in area — but no email has
# that header, so the built-in area collapses to nothing. Only the
# filter output is shown.
show-headers=true
header-layout=X-Collapse

[compose]
# nvim-mail is a dedicated Neovim profile for composing email.
# See docs/nvim-mail.md for the full walkthrough.
editor=nvim-mail

# edit-headers=true means header fields (To, Cc, Subject, etc.) appear
# as editable text at the top of the compose buffer, not in a separate
# aerc prompt. nvim-mail reformats them for readability.
edit-headers=true

# Contact completion: press Ctrl-o in header fields to search contacts.
# Requires khard (https://github.com/lucc/khard) with contacts synced
# via vdirsyncer.
address-book-cmd=khard email --parsable --remove-first-line %s

empty-subject-warning=true

# Warn if the message body mentions "attach" (not in quoted text) but
# has no attachments. The ^[^>]* prefix skips quoted lines.
no-attachment-warning=^[^>]*attach

# format-flowed (RFC 3676): the editor hard-wraps at 72 characters,
# and aerc adds reflow markers on send. Recipients' email clients can
# then reflow the text to fit their display width, rather than seeing
# fixed 72-character lines. This is the standard for polite plain text
# email.
format-flowed=true

[multipart-converters]
# When you press 'y' in the compose review screen, aerc converts your
# markdown body to HTML using pandoc, then sends both as a multipart
# message. Recipients see the HTML version in GUI clients and the
# plain text version in CLI clients.
text/html=pandoc -f markdown -t html --standalone

[filters]
# These map MIME types to filter commands. Each filter receives the
# message part on stdin and writes styled text to stdout.
#
# mailrender is the beautiful-aerc filter binary — see the README for
# what each subcommand does, or docs/power-users.md for internals.
text/plain=mailrender plain
text/html=mailrender html

# Binary attachments can't be displayed in the terminal. These filters
# show a helpful message instead of the "no filter configured" error.
application/zip=echo "ZIP archive - use :open or :save to download"
application/pdf=echo "PDF document - use :open or :save to download"
application/*=echo "Binary attachment - use :open or :save to download"

# Built-in aerc filters for calendar invites and delivery status.
text/calendar=calendar
message/delivery-status=colorize
message/rfc822=colorize

# .headers is a special aerc filter that runs on the raw RFC 2822
# headers of every message. mailrender headers reformats and colorizes
# them — see [viewer] section above for how the built-in header area
# is suppressed.
.headers=mailrender headers

[openers]

[templates]
```

- [ ] **Step 2: Apply to personal dotfiles**

Apply the same comment changes to `~/.dotfiles/beautiful-aerc/.config/aerc/aerc.conf`. The personal copy may have additional settings (e.g., cache-state, cache-blobs) — preserve those, only update the comments on shared settings.

- [ ] **Step 3: Commit**

```bash
git add .config/aerc/aerc.conf
git commit -m "Rewrite aerc.conf comments for newcomer audience"
```

---

## Task 3: Rewrite binds.conf Comments

**Files:**
- Modify: `.config/aerc/binds.conf`
- Modify: `~/.dotfiles/beautiful-aerc/.config/aerc/binds.conf` (personal dotfiles copy)

Same principle: values stay the same, comments become educational.

- [ ] **Step 1: Rewrite .config/aerc/binds.conf comments**

Replace the comment structure. The new version should have:

```ini
#
# aerc keybindings — beautiful-aerc
#
# Keybindings map key sequences to aerc commands. Each [section] below
# corresponds to a different context in aerc — the bindings that are
# active depend on what you're doing (reading the message list, viewing
# a message, composing, etc.).
#
# Syntax: <key sequence> = <command>
# Use "Eq" for '=' in key sequences: "<Ctrl+Eq>"
# Wrap '#' in quotes to use as a key: "#" = quit
#
# See aerc-binds(5) for all options: `man aerc-binds`
#

# =====================================================================
# Global bindings — active everywhere in aerc
# =====================================================================
<C-p> = :prev-tab<Enter>
<C-PgUp> = :prev-tab<Enter>
<C-n> = :next-tab<Enter>
<C-PgDn> = :next-tab<Enter>
\[t = :prev-tab<Enter>
\]t = :next-tab<Enter>
<C-t> = :term<Enter>
? = :help keys<Enter>
<C-c> = :prompt 'Quit?' quit<Enter>
<C-q> = :prompt 'Quit?' quit<Enter>
<C-z> = :suspend<Enter>

# =====================================================================
# Message list — browsing your mailbox
# =====================================================================
# This is the main screen: your folder sidebar on the left, message
# list on the right. Navigation is vim-style (j/k/g/G) with page
# movement on Ctrl-d/u/f/b.
[messages]
q = :prompt 'Quit?' quit<Enter>

# --- Navigation ---
j = :next<Enter>
<Down> = :next<Enter>
<C-d> = :next 50%<Enter>
<C-f> = :next 100%<Enter>
<PgDn> = :next 100%<Enter>

k = :prev<Enter>
<Up> = :prev<Enter>
<C-u> = :prev 50%<Enter>
<C-b> = :prev 100%<Enter>
<PgUp> = :prev 100%<Enter>
g = :select 0<Enter>
G = :select -1<Enter>

# --- Folder navigation ---
# J/K move between folders in the sidebar.
# H/L collapse/expand folder trees.
J = :next-folder<Enter>
<C-Down> = :next-folder<Enter>
K = :prev-folder<Enter>
<C-Up> = :prev-folder<Enter>
H = :collapse-folder<Enter>
<C-Left> = :collapse-folder<Enter>
L = :expand-folder<Enter>
<C-Right> = :expand-folder<Enter>

# --- Selection ---
# v toggles selection on the current message. Space selects and moves
# to the next message (handy for batch operations). V selects all.
v = :mark -t<Enter>
<Space> = :mark -t<Enter>:next<Enter>
V = :mark -v<Enter>

# --- Threading ---
# T toggles thread view. zc/zo/za fold/unfold threads (vim-style).
# Tab toggles fold on the current thread.
T = :toggle-threads<Enter>
zc = :fold<Enter>
zo = :unfold<Enter>
za = :fold -t<Enter>
zM = :fold -a<Enter>
zR = :unfold -a<Enter>
<tab> = :fold -t<Enter>

# --- Actions ---
<Enter> = :view<Enter>
d = :prompt 'Really delete this message?' 'delete-message'<Enter>
D = :delete<Enter>
a = :archive flat<Enter>
A = :unmark -a<Enter>:mark -T<Enter>:archive flat<Enter>

# --- Compose ---
C = :compose<Enter>
m = :compose<Enter>

# --- Reply ---
# rr = reply-all (quoted), rq = reply-all (quoted, no attribution)
# Rr/Rq = reply to sender only
rr = :reply -a<Enter>
rq = :reply -aq<Enter>
Rr = :reply<Enter>
Rq = :reply -q<Enter>

c = :cf<space>
$ = :term<space>
! = :term<space>
| = :pipe<space>

# --- Search & filter ---
/ = :search<space>
\ = :filter<space>
n = :next-result<Enter>
N = :prev-result<Enter>
<Esc> = :clear<Enter>

# --- Split views ---
s = :split<Enter>
S = :vsplit<Enter>

# --- Fastmail integration (optional) ---
# These bindings create mail filter rules and manage masked email
# addresses directly from the message list. Uncomment to enable.
#
# Requires:
#   1. fastmail-cli binary installed (make install)
#   2. FASTMAIL_API_TOKEN environment variable set
#   3. A Fastmail account
#
# ff = create filter rule from sender (From header)
# fs = create filter rule from subject
# ft = create filter rule from recipient (To header)
# md = delete the masked email address used in this message
#
# ff = :pipe -m fastmail-cli rules interactive from<Enter>
# fs = :pipe -m fastmail-cli rules interactive subject<Enter>
# ft = :pipe -m fastmail-cli rules interactive to<Enter>
# md = :pipe -m fastmail-cli masked delete<Enter>:delete<Enter>

# --- Patches (git-email) ---
# For working with git patches sent via email.
pl = :patch list<Enter>
pa = :patch apply <Tab>
pd = :patch drop <Tab>
pb = :patch rebase<Enter>
pt = :patch term<Enter>
ps = :patch switch <Tab>

[messages:folder=Drafts]
<Enter> = :recall<Enter>

# =====================================================================
# Message viewer — reading a single email
# =====================================================================
# When you press Enter on a message, you enter the viewer. The email
# body is rendered through mailrender (see [filters] in aerc.conf).
[view]
/ = :toggle-key-passthrough<Enter>/
q = :close<Enter>
O = :open<Enter>
o = :open<Enter>
S = :save<space>
| = :pipe<space>

# --- Save to corpus (optional, for development) ---
# Saves the raw email to the test corpus for filter development.
# Requires: aerc-save-email script on PATH.
# b = :pipe -m aerc-save-email<Enter>

# --- Triage workflow ---
# These bindings support a fast triage flow: process a message and
# immediately see the next one, or close the viewer after acting.
#
# d = delete this message (stays in viewer, shows next)
# D = delete and close viewer (returns to message list)
# a = archive this message (stays in viewer, shows next)
# A = archive and close viewer (returns to message list)
d = :delete<Enter>
D = :close<Enter>:delete<Enter>
a = :archive flat<Enter>
A = :close<Enter>:archive flat<Enter>

# --- Link picker ---
# Tab opens an interactive URL picker showing all links in the message.
# See README for picker controls (1-9 for instant select, j/k, Enter).
# Ctrl-l lets you manually type a URL to open.
<C-l> = :open-link <space>
<Tab> = :pipe pick-link<Enter>

f = :forward<Enter>
rr = :reply -a<Enter>
rq = :reply -aq<Enter>
Rr = :reply<Enter>
Rq = :reply -q<Enter>

# --- Fastmail integration (optional) ---
# Same as [messages] section. Uncomment to enable in the viewer.
# Capital F prefix avoids conflict with 'f' for forward.
#
# Ff = :pipe -m fastmail-cli rules interactive from<Enter>
# Fs = :pipe -m fastmail-cli rules interactive subject<Enter>
# Ft = :pipe -m fastmail-cli rules interactive to<Enter>
# md = :pipe -m fastmail-cli masked delete<Enter>:delete<Enter>

H = :toggle-headers<Enter>
<C-k> = :prev-part<Enter>
<C-Up> = :prev-part<Enter>
<C-j> = :next-part<Enter>
<C-Down> = :next-part<Enter>
J = :next<Enter>
<C-Right> = :next<Enter>
K = :prev<Enter>
<C-Left> = :prev<Enter>

# --- Passthrough mode ---
# When key passthrough is enabled (via / above), all keys go to the
# terminal pager instead of aerc. Ctrl-x returns to aerc command mode,
# Escape exits passthrough.
[view::passthrough]
$noinherit = true
$ex = <C-x>
<Esc> = :toggle-key-passthrough<Enter>

# =====================================================================
# Compose — writing an email
# =====================================================================
# When you compose, reply, or forward, aerc opens the compose tab.
# The editor (nvim-mail) handles most interaction. These bindings
# control the aerc compose wrapper around the editor.
[compose]
$noinherit = true
$ex = <C-x>
$complete = <C-o>
<A-p> = :switch-account -p<Enter>
<C-Left> = :switch-account -p<Enter>
<A-n> = :switch-account -n<Enter>
<C-Right> = :switch-account -n<Enter>
<C-p> = :prev-tab<Enter>
<C-PgUp> = :prev-tab<Enter>
<C-n> = :next-tab<Enter>
<C-PgDn> = :next-tab<Enter>

[compose::editor]
$noinherit = true
$ex = <C-x>
<C-p> = :prev-tab<Enter>
<C-PgUp> = :prev-tab<Enter>
<C-n> = :next-tab<Enter>
<C-PgDn> = :next-tab<Enter>

# --- Compose review screen ---
# After exiting the editor normally (:wq or <space>q in nvim-mail),
# aerc shows this review screen where you decide what to do with the
# message.
#
# y = convert body to HTML multipart and send (recipients see rich text)
# n = abort — discard the message cleanly
# v = preview — see what the message will look like
# p = postpone — save as a draft for later
# q = choose between discard or postpone
# e = re-edit — go back to nvim-mail
# a = attach a file
# d = detach (remove) an attachment
[compose::review]
y = :multipart text/html<Enter>:send<Enter>
n = :abort<Enter>
v = :preview<Enter>
p = :postpone<Enter>
q = :choose -o d discard abort -o p postpone postpone<Enter>
e = :edit<Enter>
a = :attach<space>
d = :detach<space>

[terminal]
$noinherit = true
$ex = <C-x>

<C-p> = :prev-tab<Enter>
<C-n> = :next-tab<Enter>
<C-PgUp> = :prev-tab<Enter>
<C-PgDn> = :next-tab<Enter>
```

- [ ] **Step 2: Apply to personal dotfiles**

Apply the same comment changes to `~/.dotfiles/beautiful-aerc/.config/aerc/binds.conf`. The personal copy has the fastmail-cli bindings uncommented — preserve those, only update comments.

- [ ] **Step 3: Commit**

```bash
git add .config/aerc/binds.conf
git commit -m "Rewrite binds.conf comments for newcomer audience"
```

---

## Task 4: Rewrite accounts.conf.example

**Files:**
- Modify: `.config/aerc/accounts.conf.example`

- [ ] **Step 1: Expand accounts.conf.example**

```ini
#
# aerc account configuration — beautiful-aerc
#
# Copy this file to accounts.conf and fill in your details:
#   cp accounts.conf.example accounts.conf
#
# This file contains your mail server credentials and is excluded from
# git (see .gitignore). Never commit accounts.conf.
#
# aerc supports multiple mail protocols. The example below uses JMAP
# (a modern, fast mail protocol). IMAP works too — see aerc-accounts(5)
# for IMAP examples.
#
# Common providers:
#   Fastmail — JMAP:  jmap://you@fastmail.com (recommended)
#              IMAP:  imaps://you@fastmail.com
#   Gmail    — IMAP:  imaps://you@gmail.com (requires app password)
#   Other    — Check your provider's IMAP/SMTP settings
#
# For full options, see: `man aerc-accounts` or aerc-accounts(5)
#

[Mail]
# source: where to fetch mail from
source = jmap://you@example.com

# Credential helper (optional but recommended):
# Instead of storing your password in this file, use a helper that
# retrieves it from your system keychain or a password manager.
# Example: source-cred-cmd = secret-tool lookup email you@example.com
# source-cred-cmd = your-credential-helper

# outgoing: where to send mail through
outgoing = jmap://you@example.com
# outgoing-cred-cmd = your-credential-helper

# default: which folder to show when aerc starts
default = Inbox

# from: your display name and email address for sent messages
from = Your Name <you@example.com>

# copy-to: save sent messages to this folder
copy-to = Sent

# cache-headers: cache message headers locally for faster browsing
cache-headers = true
```

- [ ] **Step 2: Apply to personal dotfiles**

Update `~/.dotfiles/beautiful-aerc/.config/aerc/accounts.conf.example` with the same expanded comments.

- [ ] **Step 3: Commit**

```bash
git add .config/aerc/accounts.conf.example
git commit -m "Expand accounts.conf.example with newcomer guidance"
```

---

## Task 5: Rewrite kitty-mail.conf Comments

**Files:**
- Modify: `.config/kitty/kitty-mail.conf`

- [ ] **Step 1: Rewrite kitty-mail.conf comments**

```conf
# kitty-mail.conf — Dedicated terminal profile for email
#
# This is an example of how you can customize your terminal emulator
# for a CLI email client. The mail launcher script uses this profile:
#   kitty --config ~/.config/kitty/kitty-mail.conf --title Mail aerc
#
# Design choices:
#   - Prose-optimized font (iA Writer Mono) for comfortable reading
#   - Generous padding so email doesn't feel cramped
#   - Nord color scheme matching the aerc theme
#   - Hidden tab bar (aerc has its own tab system)
#   - Fixed window size for consistent column layout
#
# This file is specific to kitty (https://sw.kovidgoyal.net/kitty/).
# If you use a different terminal, adapt these ideas to your own
# emulator's config format.

# -- Font --
# iA Writer Mono is designed for long-form reading and writing.
# You can use any font you prefer. If your font doesn't include
# Nerd Font symbols (used for folder icons and message flags in aerc),
# the symbol_map lines below pull those glyphs from a Nerd Font.
font_family      postscript_name=iAWriterMonoS-Regular
bold_font        postscript_name=iAWriterMonoS-Bold
italic_font      postscript_name=iAWriterMonoS-Italic
bold_italic_font postscript_name=iAWriterMonoS-BoldItalic
font_size        11.5

# Nerd Font symbol fallback — only needed if your primary font
# doesn't include Nerd Font glyphs.
symbol_map U+E000-U+F8FF   JetBrainsMonoNL Nerd Font
symbol_map U+F0000-U+FFFFF JetBrainsMonoNL Nerd Font
symbol_map U+100000-U+10FFFF JetBrainsMonoNL Nerd Font

# -- Window --
# Generous padding makes email feel less like a code terminal.
# The fixed window size (140 columns x 48 rows) gives mailrender
# a consistent width for text wrapping and link truncation.
background           #2e3440
background_opacity   1.0
window_padding_width 15 30 15 30
remember_window_size no
initial_window_width  140c
initial_window_height 48c

# -- Chrome --
# Hide kitty's tab bar — aerc has its own tab system and showing
# both is confusing. stack layout means only one pane is visible.
tab_bar_min_tabs     999
enabled_layouts      stack

# -- Shell integration --
shell_integration enabled
copy_on_select          yes
strip_trailing_spaces   smart

# -- Scrollback --
scrollback_lines 10000

# -- Colors: Nord --
# These match the Nord theme in .config/aerc/themes/nord.toml.
# If you switch aerc themes, update these colors to match.
foreground            #d8dee9
cursor                #88c0d0
cursor_blink_interval 0
cursor_shape          block
cursor_shape_unfocused unchanged
selection_background  #4c566a

color0   #3b4252
color1   #bf616a
color2   #a3be8c
color3   #ebcb8b
color4   #81a1c1
color5   #b48ead
color6   #88c0d0
color7   #e5e9f0
color8   #4c566a
color9   #bf616a
color10  #a3be8c
color11  #ebcb8b
color12  #81a1c1
color13  #b48ead
color14  #8fbcbb
color15  #eceff4
```

- [ ] **Step 2: Commit**

```bash
git add .config/kitty/kitty-mail.conf
git commit -m "Rewrite kitty-mail.conf comments for newcomer audience"
```

---

## Task 6: Rewrite nvim-mail/init.lua Comments

**Files:**
- Modify: `.config/nvim-mail/init.lua`

Comments only — no code changes. Add educational comments explaining what each section does and why, for someone who may be reading neovim Lua config for the first time.

- [ ] **Step 1: Rewrite init.lua comments**

The new comment structure (interleaved with the existing code, which stays unchanged):

```lua
-- nvim-mail: Neovim profile for composing email in aerc
--
-- This is a dedicated Neovim configuration (isolated via NVIM_APPNAME)
-- that does not affect your regular Neovim setup. It provides:
--   - Custom syntax highlighting for email headers and quoted text
--   - Hard-wrap at 72 characters with format-flowed support
--   - Spell check on body text (skips headers and quoted lines)
--   - Fuzzy contact picker via khard + Telescope
--   - Prose tidying via tidytext (optional, requires Anthropic API key)
--
-- See docs/nvim-mail.md for the full compose workflow.

vim.g.mapleader = " "

-- Plugin management: lazy.nvim
-- lazy.nvim is a modern Neovim plugin manager. It auto-installs itself
-- on first launch (the bootstrap block below), then installs all
-- plugins listed in the setup() call. You don't need to install
-- anything manually — just launch nvim-mail and wait a few seconds.
local lazypath = vim.fn.stdpath("data") .. "/lazy/lazy.nvim"
if not vim.loop.fs_stat(lazypath) then
  vim.fn.system({
    "git", "clone", "--filter=blob:none",
    "https://github.com/folke/lazy.nvim.git",
    "--branch=stable", lazypath,
  })
end
vim.opt.rtp:prepend(lazypath)

require("lazy").setup({
  -- nord.nvim: Nord color scheme, matching the aerc theme.
  -- Loads first (priority=1000) so all UI elements use Nord colors.
  {
    "shaunsingh/nord.nvim",
    priority = 1000,
    config = function()
      vim.cmd.colorscheme("nord")
    end,
  },
  -- telescope.nvim: Fuzzy finder framework. Used here for the
  -- contact picker (search your address book by typing a few letters).
  -- plenary.nvim is a required dependency (utility library).
  {
    "nvim-telescope/telescope.nvim",
    dependencies = { "nvim-lua/plenary.nvim" },
  },
  -- which-key.nvim: Shows available keybindings when you press the
  -- leader key (Space). Helps you discover what's available without
  -- memorizing everything. Loads lazily to avoid slowing startup.
  {
    "folke/which-key.nvim",
    event = "VeryLazy",
  },
}, { ui = { border = "none" } })

-- Editor settings: optimized for prose composition, not code.
-- These create a comfortable writing environment with automatic
-- line wrapping, spell check, and a clean distraction-free display.
vim.opt.wrap = true
vim.opt.linebreak = true
vim.opt.textwidth = 72          -- hard-wrap at 72 characters (email standard)
vim.opt.formatoptions = "tcrqwn" -- auto-wrap as you type, continue lists
vim.opt.breakat = " \t"         -- only break at spaces/tabs, not punctuation
vim.opt.spell = true
vim.opt.spelllang = "en_us"
vim.opt.number = false          -- no line numbers (this is prose, not code)
vim.opt.relativenumber = false
vim.opt.signcolumn = "no"       -- no gutter
vim.opt.showmode = false         -- hide --INSERT-- indicator
vim.opt.laststatus = 0           -- hide status line
vim.opt.cursorline = false
vim.opt.autoindent = true
vim.opt.smartindent = true
vim.opt.breakindent = true
vim.opt.breakindentopt = "shift:2" -- wrapped lines indent for visual alignment
vim.opt.swapfile = false
vim.opt.termguicolors = true

-- Custom filetype: "aercmail"
-- We use a custom filetype instead of the built-in "mail" type because
-- mail's syntax definitions conflict with our custom highlighting.
vim.api.nvim_create_autocmd("VimEnter", {
  callback = function()
    vim.bo.filetype = "aercmail"
  end,
})

-- Buffer preparation on open
-- When aerc opens the compose editor, the raw email buffer has RFC 2822
-- formatted headers (folded continuations, bare angle brackets, etc.).
-- This autocmd pipes the buffer through compose-prep to normalize it,
-- then adds visual separator lines and positions the cursor.
vim.api.nvim_create_autocmd("VimEnter", {
  callback = function()
    -- Normalize headers and reflow quoted text via compose-prep binary
    local raw_lines = vim.api.nvim_buf_get_lines(0, 0, -1, false)
    local result = vim.fn.systemlist("compose-prep", raw_lines)
    if vim.v.shell_error == 0 and #result > 0 then
      vim.api.nvim_buf_set_lines(0, 0, -1, false, result)
    else
      result = raw_lines
    end

    -- Insert blank lines at top for decorative separator extmarks
    table.insert(result, 1, "")
    table.insert(result, 1, "")
    vim.api.nvim_buf_set_lines(0, 0, -1, false, result)

    -- Find the blank line that separates headers from body
    local header_end = nil
    for i = 3, #result do
      if result[i] == "" then
        header_end = i
        break
      end
    end

    if header_end then
      -- Draw decorative separator lines above and below the headers
      local ns = vim.api.nvim_create_namespace("mail_header_sep")
      local sep = string.rep("─", 72)

      vim.api.nvim_buf_set_extmark(0, ns, 1, 0, {
        virt_text = { { sep, "mailHeaderKey" } },
        virt_text_pos = "overlay",
      })
      vim.api.nvim_buf_set_extmark(0, ns, header_end - 1, 0, {
        virt_text = { { sep, "mailHeaderKey" } },
        virt_text_pos = "overlay",
      })

      -- Add blank lines after headers for the cursor landing zone
      vim.api.nvim_buf_set_lines(0, header_end, header_end, false, { "", "", "" })

      -- Cursor placement:
      -- Empty To: (new compose / forward) → land on To: line for recipient entry
      -- Populated To: (reply) → land in body, ready to type your response
      local to_line_nr = nil
      local to_empty = false
      for i = 3, header_end - 1 do
        local l = vim.api.nvim_buf_get_lines(0, i - 1, i, false)[1]
        if l:match("^To:") then
          to_line_nr = i
          to_empty = not l:match("^To:%s*%S")
          break
        end
      end

      if to_empty and to_line_nr then
        vim.api.nvim_buf_set_lines(0, to_line_nr - 1, to_line_nr, false, { "To: " })
        vim.api.nvim_win_set_cursor(0, { to_line_nr, 3 })
        vim.cmd("startinsert!")
      else
        vim.api.nvim_win_set_cursor(0, { header_end + 2, 0 })
        vim.cmd("startinsert")
      end
    else
      vim.cmd("startinsert")
    end
  end,
})

-- Save cleanup: strip decorative blank lines before headers
-- The buffer preparation above inserts blank lines for the separator
-- extmarks. This BufWritePre handler removes them before saving so
-- aerc sees valid RFC 2822 headers starting on line 1. Without this,
-- aerc fails with "no valid From: address found".
vim.api.nvim_create_autocmd("BufWritePre", {
  callback = function()
    local lines = vim.api.nvim_buf_get_lines(0, 0, -1, false)
    local first_header = nil
    for i, line in ipairs(lines) do
      if line:match("^[A-Za-z-]+:") then
        first_header = i
        break
      end
    end
    if first_header and first_header > 1 then
      vim.api.nvim_buf_set_lines(0, 0, first_header - 1, false, {})
    end
  end,
})

-- Tidytext integration (optional)
-- Runs Claude-powered prose tidying on the email body. Requires:
--   1. tidytext binary installed (make install)
--   2. ANTHROPIC_API_KEY environment variable set
-- Changed words are highlighted with teal undercurl marks that clear
-- on the next edit.
vim.api.nvim_set_hl(0, "EmailTidyChange", { undercurl = true, sp = "#8fbcbb" })

-- [run_tidy function — no changes to code, keep existing comments]

vim.keymap.set("n", "<leader>t", run_tidy, { desc = "Tidy prose (tidytext)" })

-- Keybindings
-- Space is the leader key (set at the top of this file).
-- which-key.nvim will show these when you press Space and wait.

vim.keymap.set("n", "<leader>s", function()
  vim.opt.spell = not vim.opt.spell:get()
end, { desc = "Toggle spell check" })

-- Save and quit with spellcheck prompt
-- Before exiting, checks for misspelled words in the body (skips
-- headers, quoted lines, and signature). If found, offers three
-- options: (s)pellcheck to jump to the first error, (y)es to send
-- anyway, (n)o to stay in the editor.

-- [<leader>q function — no changes to code]

-- Abort compose immediately — exits with non-zero code so aerc
-- closes the compose tab without sending.
vim.keymap.set("n", "<leader>x", "<cmd>cq<cr>", { desc = "Abort compose" })

-- Insert your email signature from ~/.config/aerc/signature.md
-- Copy signature.md.example to signature.md and edit it with your info.

-- [<leader>sig function — no changes to code]

-- Undo breakpoints: pressing . , ! ? : creates an undo point so
-- pressing 'u' undoes smaller chunks instead of the whole paragraph.
for _, ch in ipairs({ ".", ",", "!", "?", ":" }) do
  vim.keymap.set("i", ch, ch .. "<C-g>u", { desc = "Undo breakpoint at " .. ch })
end

-- Spell navigation
vim.keymap.set("n", "<leader>]", "]s", { desc = "Next misspelled word" })
vim.keymap.set("n", "<leader>[", "[s", { desc = "Prev misspelled word" })
vim.keymap.set("n", "<leader>z", "z=", { desc = "Spelling suggestions" })
vim.keymap.set("n", "<leader>r", "gqip", { desc = "Reflow paragraph" })

-- Contact picker: search your address book with fuzzy matching
-- Requires: khard (https://github.com/lucc/khard) with CardDAV
-- contacts synced via vdirsyncer.
--
-- <leader>k in normal mode, Ctrl-k in insert mode.
-- Type to filter contacts, Enter to insert, Escape to cancel.
-- On To:/Cc:/Bcc: lines, automatically prepends ", " when the line
-- already has a recipient.

-- [khard_pick function — no changes to code]
```

Note: The above shows the comment structure. The actual edit replaces comments only, preserving all code exactly. The `[no changes to code]` markers indicate where the existing code stays in place — do NOT add those markers to the file.

- [ ] **Step 2: Apply to personal dotfiles**

Apply the same comment changes to `~/.dotfiles/beautiful-aerc/.config/nvim-mail/init.lua`.

- [ ] **Step 3: Commit**

```bash
git add .config/nvim-mail/init.lua
git commit -m "Rewrite nvim-mail init.lua comments for newcomer audience"
```

---

## Task 7: Write README.md

**Files:**
- Rewrite: `README.md`

This is the largest task. The full README follows the narrative arc and sections defined in the spec. Write the complete file.

- [ ] **Step 1: Read current code snippets for inclusion**

Read the exact code from these locations for the "under the hood" snippets:
- `internal/filter/html.go` lines 39-42 (Unicode regex cluster)
- `internal/filter/html.go` lines 69-76 (tracking pixel + hidden div regexes)
- `internal/filter/html.go` lines 100-135 (stripHiddenElements function)

- [ ] **Step 2: Write README.md**

Complete rewrite. The structure:

```markdown
# beautiful-aerc

A themeable, productive email environment for the
[aerc](https://aerc-mail.org/) email client.

<!-- screenshot: Hero shot — full aerc window with Nord theme
     size: 140x48 terminal (match kitty-mail.conf dimensions)
     show: Message list on left with Nerd Font icons and thread
           prefixes. Rendered HTML email on right showing colored
           headers, markdown body with footnote links, and the
           URL reference section at the bottom.
     folder: Inbox, with a mix of read/unread messages
     file: docs/images/hero.png
-->

## Table of contents

- [Why aerc?](#why-aerc)
- [The problem: email is a mess](#the-problem-email-is-a-mess)
- [What beautiful-aerc gives you](#what-beautiful-aerc-gives-you)
- [The markdown-forward design](#the-markdown-forward-design)
- [Components](#components)
- [Prerequisites](#prerequisites)
- [Install](#install)
- [How email renders](#how-email-renders)
- [Footnote-style links](#footnote-style-links)
- [Link picker](#link-picker)
- [Theme system](#theme-system)
- [Composing email with nvim-mail](#composing-email-with-nvim-mail)
- [Optional components](#optional-components)
- [Further reading](#further-reading)

> **Want to go deeper?** See [docs/power-users.md](docs/power-users.md)
> for filter pipeline internals, theme token resolution, and
> architectural details.

## Why aerc?

[2-3 paragraphs. Key points:
- aerc is fast, keyboard-driven, runs in a terminal
- Designed for extensibility: a simple filter protocol where any
  program that reads stdin and writes ANSI-styled stdout can be a
  filter
- Written in Go, actively maintained, good community
- The gap: aerc provides the engine, but the out-of-box rendering
  for HTML email is basic (shells out to w3m or lynx). Theming is
  manual. The compose experience is bare.]

## The problem: email is a mess

[2-3 paragraphs. Key points:
- Not just marketing spam — ALL email HTML is messy
- Every sender generates HTML differently: layout tables, tracking
  pixels, invisible divs, Unicode abuse, broken nesting
- GUI clients (Gmail, Outlook, Apple Mail) do enormous hidden work
  to clean this up and present a coherent reading experience
- CLI clients show the raw mess, or at best pipe through basic
  text conversion (w3m, lynx) that loses structure and theming
- The real-world examples that drove this project: Apple receipts
  with nested tables, Bank of America emails with tracking pixels
  inside URLs, Thunderbird's moz-* attributes]

## What beautiful-aerc gives you

[Brief intro paragraph, then bullet list of deliverables with
screenshot placeholders.]

- Clean, readable rendering of HTML emails — even the messy ones

<!-- screenshot: Before/after comparison
     size: Two panels side by side, each ~70 columns wide
     show: The same HTML email rendered by stock aerc (w3m/lynx)
           on the left, and beautiful-aerc (mailrender html) on
           the right. Pick a marketing email or newsletter that
           shows the difference dramatically.
     file: docs/images/before-after.png
-->

- Numbered footnote-style links that keep body text clean

<!-- screenshot: Footnote links
     size: 80x30 terminal
     show: A rendered email with colored link text, dimmed [^N]
           markers in the body, and the numbered URL reference
           section at the bottom with the separator line.
     file: docs/images/footnote-links.png
-->

- An interactive link picker for opening URLs

<!-- screenshot: Link picker
     size: 80x30 terminal
     show: The pick-link UI in an alternate screen buffer,
           showing numbered URLs with the selection highlight
           on one of them.
     file: docs/images/link-picker.png
-->

- A semantic theme system with three built-in themes

<!-- screenshot: Theme comparison
     size: Three panels stacked or side by side
     show: The same email rendered in Nord, Solarized Dark,
           and Gruvbox Dark. Same message, same layout, different
           color palettes.
     file: docs/images/themes.png
-->

- A proper compose editor in Neovim with spell check, contact
  picker, and prose tidying
- Consistent visual design across the entire experience

This pipeline was built by processing real personal email over many
hours of iteration. Every edge case fix came from an actual broken
email. The project is actively maintained — check back for ongoing
improvements as new sender patterns surface.

## The markdown-forward design

[2-3 paragraphs. Key points:
- Markdown is the core abstraction throughout beautiful-aerc
- Reading: HTML emails are converted to clean markdown, then
  ANSI-styled for the terminal
- Writing: you compose in markdown in the editor, and aerc
  converts it to HTML multipart on send
- Why markdown? Readable as plain text (no markup noise in the
  terminal), gives you clean formatting options (headings, bold,
  lists, links) when composing, and converts losslessly to HTML
  for recipients who expect rich email
- This means your reading and writing experiences are consistent
  — the same formatting language in both directions]

## Components

[Intro paragraph framing aerc's extensibility model — see spec
for the full framing text about built-in filters, the filter
protocol, why Go instead of shell scripts.]

### Go binaries (core)

**mailrender** — the filter pipeline, and the heart of the project.

[Description from spec, including "this is where the hard work
happens" and the email HTML mess framing. Then 2-3 annotated code
snippets.]

Here's a taste of what email actually looks like under the hood —
and what mailrender handles for you:

```go
// Every email client on the internet embeds invisible Unicode
// characters: soft hyphens, zero-width joiners, Mongolian vowel
// separators, byte order marks. These are invisible to you but
// wreak havoc on terminal rendering and text processing.
reNBSP      = regexp.MustCompile(`[\x{a0}\x{2000}-\x{200a}]+`)
reZeroWidth = regexp.MustCompile(`[\x{ad}\x{34f}\x{180e}\x{200b}-\x{200d}\x{2060}-\x{2064}\x{feff}]`)
```

```go
// Bank of America (and others) embed 1×1 tracking pixel <img>
// tags literally inside hyperlink text. This causes pandoc to
// split a single URL across multiple disconnected paragraphs.
reZeroImg = regexp.MustCompile(
    `(?i)<img[^>]*(?:width:\s*0|height:\s*0|width="0"|height="0")[^>]*/?>`)
```

```go
// Apple receipts and responsive emails embed a hidden copy of the
// entire email body inside a display:none div — often deeply nested.
// A simple regex would close at the first inner </div>, so this
// function hand-tracks nesting depth to find the real closing tag.
func stripHiddenElements(body string) string {
    for {
        loc := reHiddenDivOpen.FindStringIndex(body)
        if loc == nil { break }
        start := loc[0]
        rest := body[loc[1]:]
        depth := 1
        pos := 0
        for depth > 0 && pos < len(rest) {
            nextOpen := strings.Index(rest[pos:], "<div")
            nextClose := strings.Index(rest[pos:], "</div>")
            if nextClose < 0 { pos = len(rest); break }
            if nextOpen >= 0 && nextOpen < nextClose {
                depth++
                pos += nextOpen + len("<div")
            } else {
                depth--
                pos += nextClose + len("</div>")
            }
        }
        body = body[:start] + body[loc[1]+pos:]
    }
    return body
}
```

**compose-prep** — [description from spec]

**pick-link** — [description from spec]

### Go binaries (optional)

**fastmail-cli** — [brief description]

**tidytext** — [brief description]

### Configuration and scripts

[aerc config and nvim-mail as core config. kitty profile, launcher
script, and desktop file as examples.]

## Prerequisites

[Rewrite with brief explanations of why each dependency is needed]

## Install

[Rewrite the 6-step walkthrough with more explanation at each step.
Same steps as today but with context for newcomers.]

## How email renders

[Same content as today but with slightly more explanation for
newcomers. The 3 subcommands, the aerc.conf filter mapping.]

## Footnote-style links

[Same content as today — the example block, explanation of
self-referencing links.]

## Link picker

[Same content as today — Tab to open, controls list.]

## Theme system

[Brief overview: 3 built-in themes, how to switch. Link to
docs/themes.md for customization.]

## Composing email with nvim-mail

### Why Neovim for email?

[Brief explanation: programmable editor, modal editing maps
naturally to the compose→review→send flow, real plugins add
features no other compose editor has (fuzzy contact search,
inline spell check, prose tidying)]

### Plugins

[For each plugin, one paragraph: what it does, why it's here,
what you see as a user]

- **lazy.nvim** — plugin manager...
- **nord.nvim** — color scheme...
- **telescope.nvim** — fuzzy finder...
- **which-key.nvim** — keybinding discovery...

Plugins install automatically on first launch via lazy.nvim.
No manual plugin setup needed.

### The compose flow

[Brief end-to-end walkthrough:
1. Open compose/reply/forward
2. Headers are automatically reformatted
3. Write your message in markdown
4. <space>q → spell check prompt → aerc review screen
5. y to convert to HTML and send
6. <space>x to abort at any time]

For the full walkthrough, keybindings reference, contact picker
setup, and more, see [docs/nvim-mail.md](docs/nvim-mail.md).

## Optional components

### fastmail-cli

[Same as today — brief description, example keybindings,
example CLI usage]

### tidytext

[Same as today — brief description, example usage]

### Customizing your terminal for email

[New section. Explain why you might want a dedicated terminal
profile for email: different font optimized for reading, generous
padding, matching colors, fixed window size. Show the actual
contents of the mail launcher script and desktop file as
examples. Note these are templates to adapt.]

The `mail` launcher script:

```bash
#!/usr/bin/env bash
exec kitty --class aerc-mail --config ~/.config/kitty/kitty-mail.conf --title Mail aerc
```

The `aerc-mail.desktop` file for your application launcher:

```ini
[Desktop Entry]
Name=Mail (aerc)
Comment=Terminal email client
Exec=mail
Icon=mutt
Terminal=false
Type=Application
Categories=Network;Email;
StartupWMClass=aerc-mail
Keywords=mail;email;aerc;
```

See `.config/kitty/kitty-mail.conf` for the full terminal profile
with annotated design choices.

## Further reading

- [docs/themes.md](docs/themes.md) — ...
- [docs/nvim-mail.md](docs/nvim-mail.md) — ...
- [docs/power-users.md](docs/power-users.md) — ...
- [docs/contributing.md](docs/contributing.md) — ...
- [docs/styling.md](docs/styling.md) — ...
```

The `[bracketed text]` above are authoring instructions, not literal content. The actual README will contain fully written prose in those sections.

- [ ] **Step 3: Commit**

```bash
git add README.md
git commit -m "Rewrite README for public launch

Narrative arc: why aerc, the email mess problem, markdown-forward
design, components with Go-vs-shell rationale. Annotated code
snippets from the filter pipeline. Screenshot placeholders for
future captures."
```

---

## Task 8: Write docs/power-users.md

**Files:**
- Create: `docs/power-users.md`
- Read (source material): `docs/filters.md`

This document absorbs all content from `docs/filters.md` and adds new material. Write the full document following the spec's section list.

- [ ] **Step 1: Write docs/power-users.md**

Structure with a TOC at the top:

```markdown
# beautiful-aerc for Power Users

Deep technical reference for the filter pipeline, theme system,
and architectural decisions behind beautiful-aerc.

## Table of contents

- [aerc filter protocol](#aerc-filter-protocol)
- [HTML filter pipeline](#html-filter-pipeline)
- [Header filter](#header-filter)
- [Plain text filter](#plain-text-filter)
- [Footnote system](#footnote-system)
- [Link picker architecture](#link-picker-architecture)
- [Theme token resolution](#theme-token-resolution)
- [Known edge cases](#known-edge-cases)
- [Troubleshooting](#troubleshooting)
```

Sections follow the spec's section list (Task 8 in spec: sections 1-9). Content comes primarily from `docs/filters.md` with additional material from the aerc-setup.md claude doc.

Key content for each section:

1. **aerc filter protocol** — from contributing.md's "How aerc calls the filter binary" section
2. **HTML filter pipeline** — from filters.md's pipeline description, expanded with the full stage list from the spec
3. **Header filter** — from filters.md's "Header formatting" section, including the X-Collapse trick explanation
4. **Plain text filter** — from filters.md's "Plain text handling" section
5. **Footnote system** — from filters.md's footnote sections, plus OSC 8 hyperlink detail from aerc-setup.md
6. **Link picker architecture** — from filters.md's link picker section, expanded with /dev/tty and alternate screen buffer details
7. **Theme token resolution** — from filters.md's "How theme tokens map to visual output" section
8. **Known edge cases** — from aerc-setup.md's "Known Edge Cases (Solved)" and "Problem Sender Patterns" sections
9. **Troubleshooting** — from filters.md's troubleshooting section

- [ ] **Step 2: Commit**

```bash
git add docs/power-users.md
git commit -m "Add power-users.md — deep technical reference

Absorbs content from docs/filters.md with additional material on
edge cases, problem sender patterns, and architecture."
```

---

## Task 9: Write docs/nvim-mail.md

**Files:**
- Create: `docs/nvim-mail.md`

Full compose workflow documentation for potential neovim newcomers.

- [ ] **Step 1: Write docs/nvim-mail.md**

Structure following the spec's section list:

```markdown
# Composing Email with nvim-mail

A guide to the nvim-mail compose editor — from first launch to
sending your first email.

## Table of contents

- [Why Neovim for email?](#why-neovim-for-email)
- [How it works](#how-it-works)
- [Plugins](#plugins)
- [The compose flow](#the-compose-flow)
- [Keybindings reference](#keybindings-reference)
- [Contact picker](#contact-picker)
- [Signature](#signature)
- [Tidytext integration](#tidytext-integration)
- [Troubleshooting](#troubleshooting)
```

Content for each section comes from the spec and aerc-setup.md / neovim-setup.md claude docs. Key details:

1. **Why Neovim for email?** — programmable, modal editing, real plugins, consistent keybindings
2. **How it works** — NVIM_APPNAME isolation, aerc calls nvim-mail as editor, regular neovim untouched
3. **Plugins** — lazy.nvim (auto-bootstraps), nord.nvim, telescope.nvim, which-key.nvim, nvim-treesitter. Each gets a paragraph.
4. **The compose flow** — end-to-end walkthrough: open → headers reformatted → write → spell check → review → send. Include the `<space>q` flow and `<space>x` abort.
5. **Keybindings reference** — full table from aerc-setup.md
6. **Contact picker** — khard + telescope, how to set up khard, auto-comma behavior
7. **Signature** — `<leader>sig`, copy from example, markdown bold
8. **Tidytext integration** — `<leader>t`, what it does, highlight behavior
9. **Troubleshooting** — plugins not installing, khard not found, spell check language, compose-prep missing

- [ ] **Step 2: Commit**

```bash
git add docs/nvim-mail.md
git commit -m "Add nvim-mail.md — compose workflow for newcomers"
```

---

## Task 10: Update docs/themes.md

**Files:**
- Modify: `docs/themes.md`

Light rewrite: add intro linking to README, ensure "create a custom theme" is clear for first-timers. Content stays largely the same.

- [ ] **Step 1: Update docs/themes.md**

Changes:
- Add intro paragraph: "This document covers the theme system in detail. For a quick overview of switching themes, see the [Theme system](../README.md#theme-system) section of the README."
- Ensure the "Creating a custom theme" section has clear numbered steps
- Verify the "Keeping kitty and nvim-mail in sync" section mentions that this is a manual process (not automatic)
- No structural changes

- [ ] **Step 2: Commit**

```bash
git add docs/themes.md
git commit -m "Add README cross-reference to themes.md"
```

---

## Task 11: Update docs/contributing.md

**Files:**
- Modify: `docs/contributing.md`

Light update: remove references to internal planning docs, verify layout tree is current.

- [ ] **Step 1: Update docs/contributing.md**

Changes:
- Verify the project layout tree matches the current file structure (check for compose-prep, any new files)
- Remove any references to `docs/go-extraction-roadmap.md` or other internal docs
- Verify the code convention reference path is correct
- Update "Further reading" or cross-references if they point to `docs/filters.md` (now `docs/power-users.md`)

- [ ] **Step 2: Commit**

```bash
git add docs/contributing.md
git commit -m "Update contributing.md for public launch"
```

---

## Task 12: Remove Internal Files and Update Cross-References

**Files:**
- Remove: `docs/filters.md`
- Remove: `docs/go-extraction-roadmap.md`
- Modify: any file referencing these

- [ ] **Step 1: Find all references to removed files**

```bash
grep -r "filters\.md\|go-extraction-roadmap" --include="*.md" --include="*.lua" --include="*.conf"
```

Known references:
- `README.md` links to `docs/filters.md` (already rewritten in Task 7)
- `.config/nvim-mail/init.lua` line 2 references `docs/filters.md`
- `docs/contributing.md` may reference these

- [ ] **Step 2: Update remaining references**

Update any references found in Step 1 that weren't already handled by earlier tasks. Change `docs/filters.md` references to `docs/power-users.md`.

- [ ] **Step 3: Remove files and commit**

```bash
git rm docs/filters.md docs/go-extraction-roadmap.md
git add -u  # stage reference updates
git commit -m "Remove internal docs, update cross-references

docs/filters.md content moved to docs/power-users.md.
docs/go-extraction-roadmap.md was internal planning."
```

---

## Task 13: Final Verification

- [ ] **Step 1: Verify all links in README**

Check that every `[text](path)` link in README.md points to a file that exists.

- [ ] **Step 2: Verify all links in docs/**

Check that cross-references between docs files are valid.

- [ ] **Step 3: Verify .gitignore**

```bash
git status
```

Confirm no sensitive files are tracked.

- [ ] **Step 4: Build and test**

```bash
make check
```

Ensure no code was accidentally changed.

- [ ] **Step 5: Review screenshot placeholders**

List all `<!-- screenshot:` comments in README.md and confirm each has:
- Description of what to capture
- Recommended size
- What should be visible
- Suggested filename

- [ ] **Step 6: Final commit (if any fixups needed)**

```bash
git add <fixed files>
git commit -m "Fix cross-references and verify docs for public launch"
```
