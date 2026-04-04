" aercmail syntax: custom mail highlighting for aerc compose

if exists("b:current_syntax")
  finish
endif

" Header keys (To:, From:, Subject:, etc.)
syntax match aercmailHeaderKey /^[A-Za-z-]\+:/

" Email addresses in angle brackets
syntax match aercmailAngleBracket /<[^>]\+>/

" Quoted text (more specific match first)
syntax match aercmailQuote2 /^> > .*$/
syntax match aercmailQuote1 /^> .*$/

highlight aercmailHeaderKey guifg=#81A1C1 gui=bold ctermfg=110 cterm=bold
highlight aercmailAngleBracket guifg=#616E88 ctermfg=60
highlight aercmailQuote1 guifg=#8FBCBB ctermfg=108
highlight aercmailQuote2 guifg=#616E88 ctermfg=60

" Exclude from spell check
syntax cluster Spell remove=aercmailHeaderKey,aercmailAngleBracket,aercmailQuote1,aercmailQuote2

let b:current_syntax = "aercmail"
