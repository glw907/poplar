---
paths:
  - "cmd/tidytext/**"
  - "internal/tidy/**"
  - "e2e-tidytext/**"
---

# tidytext

Claude-powered prose tidier. Fixes spelling, grammar, and punctuation
without altering meaning or the author's style.

## Command Structure

    tidytext
      fix           Fix spelling, grammar, and punctuation
      config        Show effective configuration
        init          Create default config file

## Usage

    tidytext                              # show help
    echo "text" | tidytext fix            # piped stdin
    tidytext fix message.txt              # read file
    tidytext fix --in-place message.txt   # modify file directly
    tidytext fix --no-config              # skip config, use defaults
    tidytext fix --rule spelling=false    # override a rule
    tidytext fix --style em_dash_spaces=true
    tidytext fix --style time_format=uppercase
    tidytext config                       # show current config
    tidytext config init                  # create default config

## Config

Location: `~/.config/tidytext/config.toml`

If the file does not exist, all rules are enabled with defaults.
Users only need to specify what they want to change.

## Environment Variables

    ANTHROPIC_API_KEY    Claude API key (required for fix command)
    TIDYTEXT_API_URL     Override API endpoint (testing only)

## nvim-mail Integration

`<leader>t` runs tidytext on the compose buffer body (excluding
headers and signature). Changed words are highlighted with teal
undercurl extmarks (`EmailTidyChange` highlight group) that clear
on next edit.
