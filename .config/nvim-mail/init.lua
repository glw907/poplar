-- nvim-mail: neovim profile for aerc email composing.
-- Minimal visual style with markdown support and spell check.

-- Leader key
vim.g.mapleader = " "

-- Bootstrap lazy.nvim
local lazypath = vim.fn.stdpath("data") .. "/lazy/lazy.nvim"
if not vim.loop.fs_stat(lazypath) then
  vim.fn.system({
    "git", "clone", "--filter=blob:none",
    "https://github.com/folke/lazy.nvim.git",
    "--branch=stable", lazypath,
  })
end
vim.opt.rtp:prepend(lazypath)

-- Plugins
require("lazy").setup({
  {
    "shaunsingh/nord.nvim",
    priority = 1000,
    config = function()
      vim.cmd.colorscheme("nord")
    end,
  },
  {
    "nvim-treesitter/nvim-treesitter",
    build = ":TSUpdate",
    config = function()
      require("nvim-treesitter.configs").setup({
        ensure_installed = { "markdown", "markdown_inline" },
        highlight = { enable = true },
      })
    end,
  },
}, { ui = { border = "none" } })

-- Editor settings
vim.opt.wrap = true
vim.opt.linebreak = true
vim.opt.textwidth = 72
vim.opt.formatoptions = "tcrqwn"
vim.opt.breakat = " \t"
vim.opt.spell = true
vim.opt.spelllang = "en_us"
vim.opt.number = false
vim.opt.relativenumber = false
vim.opt.signcolumn = "no"
vim.opt.showmode = false
vim.opt.laststatus = 0
vim.opt.cursorline = false
vim.opt.autoindent = true
vim.opt.smartindent = true
vim.opt.breakindent = true
vim.opt.breakindentopt = "shift:2"
vim.opt.swapfile = false
vim.opt.termguicolors = true

-- Quote reflow: join consecutive quoted lines into paragraphs, re-wrap
local function get_quote_prefix(line)
  return line:match("^(>[ >]*%s?)")
end

-- Normalize prefix to just the ">" characters for comparison
local function normalize_prefix(prefix)
  if not prefix then return nil end
  return prefix:gsub("[^>]", "")
end

local function wrap_text(text, prefix, width)
  local result = {}
  local avail = width - #prefix
  while #text > avail do
    local break_at = nil
    for i = avail, 1, -1 do
      if text:sub(i, i) == " " then
        break_at = i
        break
      end
    end
    if not break_at then break_at = avail end
    result[#result + 1] = prefix .. text:sub(1, break_at):gsub("%s+$", "")
    text = text:sub(break_at + 1):gsub("^%s+", "")
  end
  if #text > 0 then
    result[#result + 1] = prefix .. text
  end
  return result
end

local function reflow_quoted(lines, width)
  local result = {}
  local i = 1
  while i <= #lines do
    local prefix = get_quote_prefix(lines[i])
    if prefix then
      local level = normalize_prefix(prefix)
      -- Build canonical prefix: "> " for single, "> > " for nested, etc.
      local canon = string.rep("> ", #level - 1) .. "> "
      local text = lines[i]:sub(#prefix + 1)
      -- Blank quoted line = paragraph break, emit and advance
      if text:match("^%s*$") then
        result[#result + 1] = canon:gsub("%s+$", "")
        i = i + 1
      else
        -- Join consecutive lines at the same quote level
        local j = i + 1
        while j <= #lines do
          local next_prefix = get_quote_prefix(lines[j])
          local next_level = normalize_prefix(next_prefix)
          if next_level ~= level then break end
          local next_text = lines[j]:sub(#next_prefix + 1)
          -- Blank quoted line = paragraph break
          if next_text:match("^%s*$") then break end
          text = text:gsub("%s+$", "") .. " " .. next_text:gsub("^%s+", "")
          j = j + 1
        end
        -- Re-wrap the joined paragraph
        text = text:gsub("^%s+", ""):gsub("%s+$", "")
        local wrapped = wrap_text(text, canon, width)
        for _, wl in ipairs(wrapped) do
          result[#result + 1] = wl
        end
        i = j
      end
    else
      result[#result + 1] = lines[i]
      i = i + 1
    end
  end
  return result
end



-- Set custom filetype for our syntax highlighting
vim.api.nvim_create_autocmd("VimEnter", {
  callback = function()
    vim.bo.filetype = "aercmail"
  end,
})

-- Prepare compose buffer: unfold headers, add separator, position cursor
vim.api.nvim_create_autocmd("VimEnter", {
  callback = function()
    local lines = vim.api.nvim_buf_get_lines(0, 0, -1, false)

    -- Unfold RFC 2822 continuation lines (leading whitespace) into single lines
    local unfolded = {}
    for _, line in ipairs(lines) do
      if line:match("^[ \t]") and #unfolded > 0 then
        unfolded[#unfolded] = unfolded[#unfolded] .. " " .. line:match("^%s*(.*)")
      else
        unfolded[#unfolded + 1] = line
      end
    end

    -- Strip angle brackets from bare <email> (no name) in address headers
    -- Bare = preceded by ":" or "," (plus whitespace), not by a name
    for i, line in ipairs(unfolded) do
      if line:match("^[A-Za-z-]+:") then
        unfolded[i] = line:gsub("([,:])(%s*)<([^>]+)>", "%1%2%3")
      end
    end

    -- Re-fold address headers at recipient boundaries (fill to ~72 cols)
    local result = {}
    local max_width = 72
    for _, line in ipairs(unfolded) do
      local key, value
      for _, k in ipairs({ "To", "Cc", "Bcc" }) do
        local pattern = "^(" .. k .. "):%s*(.*)"
        key, value = line:match(pattern)
        if key then break end
      end
      if key and value and value:find(",") then
        local indent = string.rep(" ", #key + 2)
        local recipients = {}
        for addr in (value .. ","):gmatch("(.-),%s*") do
          if addr ~= "" then
            recipients[#recipients + 1] = addr
          end
        end
        local cur = key .. ": " .. recipients[1]
        for j = 2, #recipients do
          local candidate = cur .. ", " .. recipients[j]
          if #candidate <= max_width then
            cur = candidate
          else
            result[#result + 1] = cur .. ","
            cur = indent .. recipients[j]
          end
        end
        result[#result + 1] = cur
      else
        result[#result + 1] = line
      end
    end

    -- Hard-wrap quoted text for display
    result = reflow_quoted(result, 72)

    -- Add blank line + separator line above headers
    table.insert(result, 1, "")  -- will be overlaid with separator
    table.insert(result, 1, "")  -- blank line at top

    vim.api.nvim_buf_set_lines(0, 0, -1, false, result)
    lines = result

    -- Find first blank line after headers (end of header block)
    local header_end = nil
    for i = 3, #lines do  -- skip the two lines we added
      if lines[i] == "" then
        header_end = i
        break
      end
    end

    if header_end then
      local ns = vim.api.nvim_create_namespace("mail_header_sep")
      local sep = string.rep("─", 72)

      -- Top separator (overlay on second line, after blank)
      vim.api.nvim_buf_set_extmark(0, ns, 1, 0, {
        virt_text = { { sep, "mailHeaderKey" } },
        virt_text_pos = "overlay",
      })

      -- Bottom separator (overlay on blank line after headers)
      vim.api.nvim_buf_set_extmark(0, ns, header_end - 1, 0, {
        virt_text = { { sep, "mailHeaderKey" } },
        virt_text_pos = "overlay",
      })

      -- Ensure: blank, cursor, blank, then body
      local body_start = header_end + 1
      if body_start <= #lines and lines[body_start] ~= "" then
        vim.api.nvim_buf_set_lines(0, header_end, header_end, false, { "", "", "" })
        vim.api.nvim_win_set_cursor(0, { header_end + 2, 0 })
      else
        vim.api.nvim_buf_set_lines(0, header_end, header_end, false, { "", "", "" })
        vim.api.nvim_win_set_cursor(0, { header_end + 2, 0 })
      end
    end

    vim.cmd("startinsert")
  end,
})

-- Strip decorative blank lines before headers on save so aerc sees
-- valid RFC 2822 (headers must start on line 1).
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

-- Keybindings
vim.keymap.set("n", "<leader>s", function()
  vim.opt.spell = not vim.opt.spell:get()
end, { desc = "Toggle spell check" })

vim.keymap.set("n", "<leader>q", function()
  local lines = vim.api.nvim_buf_get_lines(0, 0, -1, false)
  -- Find body start: first blank line after a header (Key: value) line
  local body_start = 1
  local in_headers = false
  for i, line in ipairs(lines) do
    if line:match("^[A-Za-z-]+:") then
      in_headers = true
    elseif in_headers and line == "" then
      body_start = i + 1
      break
    end
  end
  -- Check only non-quoted, non-empty body lines
  local misspelled = {}
  for i = body_start, #lines do
    local line = lines[i]
    if line ~= "" and not line:match("^>") and not line:match("^%-%-") then
      for _, entry in ipairs(vim.spell.check(line)) do
        if entry[2] == "bad" or entry[2] == "rare" then
          misspelled[#misspelled + 1] = entry[1]
        end
      end
    end
  end
  if #misspelled > 0 then
    local prompt = #misspelled .. " misspelled word"
    if #misspelled > 1 then prompt = prompt .. "s" end
    prompt = prompt .. " - (s)pellcheck, (y)es send, (n)o? "
    vim.ui.input({ prompt = prompt }, function(input)
      if not input then return end
      if input:lower() == "y" then
        vim.cmd("wq")
      elseif input:lower() == "s" then
        vim.cmd("normal! ]s")
      end
    end)
  else
    vim.cmd("wq")
  end
end, { desc = "Save and quit" })
vim.keymap.set("n", "<leader>x", "<cmd>cq<cr>", { desc = "Abort compose" })


vim.keymap.set("n", "<leader>sig", function()
  local sig = {
    "-- ",
    "**Your Name**",
    "your-email@example.com",
  }
  local row = vim.api.nvim_win_get_cursor(0)[1]
  vim.api.nvim_buf_set_lines(0, row, row, false, sig)
end, { desc = "Insert email signature" })
