# Public Launch Documentation Overhaul

## Goal

Prepare beautiful-aerc for public release by rewriting all
user-facing documentation and config comments for a newcomer
audience. The target reader is someone curious about terminal email
— they may know what aerc is but aren't a power user. They may be
new to neovim. They should be able to clone this repo, follow the
README, and have a working setup without prior expertise.

## Audience Tiers

| Surface | Audience | Tone |
|---------|----------|------|
| README.md | Newcomers, curious terminal users | Friendly, explains *why* before *how* |
| Config file comments | Users reading configs to understand their setup | Educational, not terse |
| docs/nvim-mail.md | Potential neovim newcomers | Welcoming, step-by-step |
| docs/themes.md | Users who want to customize | Assumes README knowledge |
| docs/power-users.md | People who want to understand internals | Technical, thorough |
| docs/contributing.md | PR submitters | Expert-level |
| docs/styling.md | Contributors building UI elements | Expert-level |

## README.md

### Narrative Arc

The README tells a story before it becomes a reference:

1. **Why aerc?** — What makes aerc the right foundation for a
   terminal email client. Its strengths (fast, extensible filter
   protocol, good keyboard UX) and the gap it leaves (rendering).
2. **The problem: email is a mess.** Email on the internet is inconsistent HTML, broken
   formatting, tracking pixels, layout tables, inline styles.
   GUI clients (Outlook, Gmail, Apple Mail) do enormous work to
   clean this up and present a coherent reading experience. CLI
   clients show the raw mess, or at best pipe through basic
   text conversion.
3. **What beautiful-aerc gives you.** The same polished email
   experience GUI users get, but in your terminal. Specific
   deliverables: clean rendered HTML, numbered footnote links,
   interactive link picker, consistent theming, a proper compose editor.
4. **The markdown-forward design.** Markdown is the core
   abstraction throughout beautiful-aerc. Reading email: HTML is
   converted to clean markdown with ANSI styling. Writing email:
   you compose in markdown in neovim, and it converts to HTML
   multipart on send. Why markdown? It's readable as plain text
   in the terminal, gives you clean formatting options (headings,
   bold, lists, links) when composing, and converts losslessly
   to HTML for recipients who expect rich email.
5. **Screenshot placeholders** throughout, with descriptions of
   what to capture, recommended dimensions, and framing notes.

### Table of Contents

The README gets a markdown TOC near the top, after the narrative
intro. Includes an early, prominent link to docs/power-users.md
for readers who want to go deeper.

### Sections

1. **Why aerc** — 2-3 paragraphs on aerc's strengths and the gap
2. **The problem: email is a mess** — frame the problem broadly
   (all email, not just marketing), contrast with GUI clients
3. **What beautiful-aerc gives you** — concrete deliverables with
   screenshot placeholders
4. **The markdown-forward design** — why markdown is the core
   abstraction for both reading and writing
5. **Components overview** — what ships in the box, and why we
   replaced some aerc defaults. Each component gets a brief
   description of what it does and *why it's a Go binary* instead
   of a shell script. The general answer: Go gives us a single
   compiled binary with no runtime dependencies, proper error
   handling, and the ability to share code between tools (theme
   loading, header parsing, ANSI rendering). Shell scripts work
   for simple filters but break down when you need multi-stage
   pipelines, Unicode handling, or interactive TUIs.

   1. Go binaries (core)
      - **mailrender** — the filter pipeline. Replaces aerc's
        basic `colorize` and `plaintext` filters with a
        multi-stage pipeline that produces clean markdown from
        messy HTML. Why Go: the pipeline has 8+ stages (pre-clean,
        pandoc, artifact cleanup, footnotes, ANSI styling) that
        would be fragile as chained shell commands.
      - **compose-prep** — normalizes the compose buffer before
        nvim-mail opens it (unfold headers, strip brackets, reflow
        quoted text). Why Go: RFC 2822 header parsing and
        format-flowed text reflowing need proper string handling.
      - **pick-link** — interactive URL picker for the message
        viewer. Why Go: needs raw terminal mode, alternate screen
        buffer, and `/dev/tty` input — a real TUI application.

   2. Go binaries (optional)
      - **fastmail-cli** — Fastmail JMAP operations (mail rules,
        masked email, folder management). Why Go: JMAP is a
        JSON-over-HTTP API that benefits from typed structs and
        proper error handling.
      - **tidytext** — Claude-powered prose proofreader. Why Go:
        shares the theme system for styled output and needs
        structured API interaction with Anthropic.

   3. Config and scripts
      - aerc config and nvim-mail as core config — the settings
        that make everything work together
      - kitty profile, launcher scripts, desktop file as examples
        of how to make your CLI email experience feel more like
        a "regular app"

6. **Prerequisites** — what to install first, with brief
   explanations of each dependency and why it's needed
7. **Install** — step-by-step for newcomers. Rewrite the current
   6-step walkthrough with more explanation at each step.
   Cover: clone, build, stow, account setup, first launch.
8. **How email renders** — the 3 mailrender subcommands explained
   simply. What happens when you open an email.
9. **Footnote-style links** — with example output block
10. **Link picker** — how to open URLs from emails
11. **Theme system** — switching between the 3 built-in themes,
    brief mention of customization, link to docs/themes.md
12. **Composing email with nvim-mail** — subsections:
    - Why neovim for email? (programmable, modal editing fits
      the compose→review→send flow, plugins add real features)
    - Plugins and why (lazy.nvim for management, nord.nvim for
      theme consistency, telescope for contact picking, which-key
      for discoverability, treesitter for syntax highlighting)
    - How plugins install (lazy.nvim bootstraps automatically on
      first launch — no manual plugin setup needed)
    - Compose flow overview (open → headers reformatted → write
      reply → spell check → review → send)
    - Link to docs/nvim-mail.md for the full walkthrough
13. **Optional components** — subsections:
    - fastmail-cli (what it does, example usage)
    - tidytext (what it does, example usage)
    - Customizing your terminal for email (kitty-mail.conf as a
      teaching example: dedicated font, padding, colors, window
      size; the `mail` launcher script that launches kitty with
      the mail profile; the `aerc-mail.desktop` file for desktop
      integration so aerc appears in your app launcher. Explain
      *why* you might want a separate terminal profile for email.
      Show the actual contents of each file as examples. These
      are templates — adapt for your own terminal emulator.)
14. **Further reading** — links to all docs

### Screenshot Placeholders

Each placeholder should specify:
- What to capture (e.g., "message list with Nord theme")
- Recommended terminal size (columns x rows)
- What should be visible in the screenshot
- Any specific email/folder to show
- Suggested filename for the image

## docs/power-users.md — New

Absorbs all content from the current `filters.md` (which is then
deleted). Organized with a clear TOC.

### Sections

1. **aerc filter protocol** — how aerc calls filters (stdin/stdout,
   AERC_COLUMNS, the .headers/text-html/text-plain mapping)
2. **HTML filter pipeline** — all stages in order: pre-clean,
   pandoc + Lua filter, artifact cleanup, bold normalization,
   list normalization, whitespace normalization, footnote
   conversion, footnote styling, markdown highlighting
3. **Header filter** — why we replace aerc's built-in header
   display, what the filter does (reorder, colorize, wrap, strip
   brackets), the X-Collapse trick
4. **Plain text filter** — HTML-in-plain-text detection, routing
5. **Footnote system architecture** — how links are extracted,
   numbered, formatted with OSC 8 hyperlinks, truncated for
   terminal width
6. **pick-link architecture** — /dev/tty for input, runs HTML
   filter internally, alternate screen buffer, instant-select
   keys
7. **Theme token resolution** — how Go binaries read .toml at
   runtime, how tokens resolve to ANSI SGR sequences, the
   relationship between theme files and stylesets
8. **Known edge cases** — problem sender patterns, solved issues,
   regression test targets
9. **Troubleshooting** — common failure modes (from current
   filters.md)

## docs/nvim-mail.md — New

Full compose workflow documentation. Written for someone who may
be new to neovim.

### Sections

1. **Why neovim for composing email?** — programmable editor,
   modal editing maps naturally to the compose→review→send flow,
   real plugins (fuzzy contact search, spell check, prose
   tidying), consistent keybindings you already know if you use
   neovim elsewhere
2. **How it works** — aerc opens nvim-mail as the compose editor,
   nvim-mail is a dedicated neovim profile (NVIM_APPNAME), your
   regular neovim config is untouched
3. **Plugins** — for each plugin, explain what it does, why it's
   included, and what the user sees:
   - lazy.nvim: plugin manager, auto-bootstraps on first launch
   - nord.nvim: matches the aerc Nord theme
   - telescope.nvim: fuzzy finder, powers the contact picker
   - which-key.nvim: shows available keybindings when you press
     leader — helps discoverability
   - nvim-treesitter: syntax highlighting for the compose buffer
4. **The compose flow** — end-to-end walkthrough:
   - New message: cursor on To: line, ready to type or pick
     contacts
   - Reply: cursor in body between headers and quoted text
   - Headers are automatically reformatted (unfolded, brackets
     cleaned, addresses wrapped)
   - Write your message in markdown
   - `<space>q` to exit → spell check prompt → aerc review screen
   - `y` to convert to HTML and send
   - `<space>x` to abort at any time
5. **Keybindings reference** — full table
6. **Contact picker** — how it works (khard + telescope), how to
   set up khard, auto-comma behavior on header lines
7. **Signature** — `<leader>sig`, markdown bold for HTML
8. **Tidytext integration** — `<leader>t`, what it does, the
   highlight behavior
9. **Troubleshooting** — common issues (plugins not installing,
   khard not found, spell check language)

## docs/themes.md — Updated

Light rewrite. Assumes the reader has seen the README's theme
overview. Content stays largely the same:

- TOML file format and all 16 color slots
- Token definition syntax and all token categories
- Built-in themes table
- How to create a custom theme
- Styleset generation
- Keeping kitty and nvim-mail colors in sync (manual process)

Updates:
- Remove any assumption of prior knowledge beyond the README
- Add a brief intro linking back to the README theme section
- Ensure the "create a custom theme" steps are clear enough for
  someone following along for the first time

## docs/contributing.md — Light Update

- Remove references to internal planning docs
- Ensure project layout tree is current
- Verify code convention references are correct
- No tone change — stays expert-level

## docs/styling.md — No Change

Already well-targeted at contributors.

## Files to Remove

- `docs/filters.md` — content moves to power-users.md
- `docs/go-extraction-roadmap.md` — internal planning, not public
- `.config/aerc/mailrules.json` — personal mail rules

## Files to Add

- `docs/power-users.md`
- `docs/nvim-mail.md`
- `.config/aerc/mailrules.json.example` (empty rules array with
  a comment explaining what it's for)

## Config Comment Rewrites

All config files get comments rewritten for a newcomer audience.
The principle: explain *why* each setting exists, not just *what*
it does.

### aerc.conf

- Section headers with brief explanations of what the section
  controls
- Non-obvious settings get inline comments explaining why they're
  set this way (e.g., `header-layout=X-Collapse` — what it does
  and why, `alternatives=text/html,text/plain` — why HTML first)
- Filter entries explain what each filter does
- Compose settings explain what format-flowed means for recipients
- Group related settings visually with blank lines

### binds.conf

- Opening comment explaining aerc's keybinding model (sections
  map to contexts — what you can do depends on where you are)
- Each section gets a header explaining the context
- Triage bindings (d/D/a/A) explained as a workflow
- Optional integrations (fastmail-cli, aerc-save-email) clearly
  marked with what you need to enable them
- Link picker binding explained

### accounts.conf.example

- Expand with more guidance for newcomers
- Explain what JMAP is (one sentence)
- Note where to get credentials for common providers
- Explain each field
- Reference `aerc-accounts(5)` man page

### nvim-mail/init.lua

- Ensure each major block has a comment explaining its purpose
  in plain language
- Plugin specs should note what each plugin does
- The VimEnter autocmd should explain the buffer preparation
  pipeline in human terms
- Keybinding definitions should explain what they do

### kitty-mail.conf

- Header explaining what this file is: a dedicated terminal
  profile for email, used by the `mail` launcher script
- Explain the design choices: prose-optimized font, generous
  padding for comfortable reading, Nord colors to match aerc
  theme, hidden tab bar (aerc handles its own tabs), fixed
  window size for consistent column layout
- Note that this is an example — users can adapt for their own
  terminal emulator

## .gitignore Audit

Verify these are excluded:
- `accounts.conf` (credentials)
- `signature.md` (personal)
- `mailrules.json` (personal rules)
- `.config/aerc/generated/` (build artifacts)

## Screenshot Placeholder Format

Throughout the README, screenshot placeholders use this format:

```html
<!-- screenshot: {description}
     size: {columns}x{rows} or {width}px
     show: {what should be visible}
     file: docs/images/{filename}.png
-->
```

### Planned Placeholders

1. Hero shot — full aerc window showing a rendered HTML email
   with the Nord theme, message list visible on left
2. Before/after — stock aerc rendering vs beautiful-aerc (same
   email)
3. Footnote links — a rendered email showing numbered footnote
   references and the URL section at the bottom
4. Link picker — the interactive picker UI with a list of URLs
5. Theme comparison — the 3 built-in themes side by side (or
   stacked)
6. nvim-mail compose — the compose editor showing formatted
   headers, body text, and which-key popup
7. Message list — showing Nerd Font icons, thread prefixes, and
   the column layout
