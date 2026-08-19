# Antigravity CLI Status Line ⚡

[![Release](https://img.shields.io/github/v/release/Joacohbc/antigravity-cli-statusline)](https://github.com/Joacohbc/antigravity-cli-statusline/releases)

[English](README.md) | [Español](README.es.md)

A lightweight, standalone status line for Google Antigravity CLI (`agy`) built with Go. Features sub-millisecond execution, rich Nerd Font glyphs, smooth Unicode context bars, theme palettes, and real-time quota tracking.

---

## ⚡ Quick Install

```bash
curl -fsSL https://raw.githubusercontent.com/Joacohbc/antigravity-cli-statusline/main/install.sh | sh
```

Pin a specific version:
```bash
curl -fsSL https://raw.githubusercontent.com/Joacohbc/antigravity-cli-statusline/main/install.sh | VERSION=v1.0.0 sh
```

---

## 📸 Preview

### Wide Mode (Single line)
```text
󰚩 READY ╱ 🤖 Gemini 3.7 Flash ╱  main  │  🧠 ·········· 2.7% · 󰈙 0 · 󰒋 0 · 󰥔 75.6% (5h) · 󰃭 89.6% (7d) · 󰌾 OFF
```

### Standard Mode (Framed 2-lines)
```text
╭─ 󰚩 READY ╱ 🤖 Claude Sonnet 4.6 ╱  main
╰─ 🧠 ████▊····· 48.5% · 󰈙 3 · 󰮝 1 · 󰒋 2 · 󰥔 98.2% (5h) · 󰃭 99.4% (7d) · 󰌾 ON
```

---

## 🛠️ Quick Commands

```bash
# Preview themes interactively
statusline preview --theme catppuccin
statusline --list-themes

# Change active theme
statusline config set theme.palette catppuccin
# Available: catppuccin, tokyonight, dracula, nord, gruvbox, default, ascii, plain, no-color

# Switch layout (wide or framed)
statusline config set layout.style wide

# Toggle modules
statusline config set modules.sandbox false
statusline config set modules.weekly_usage true

# Show active config file
statusline config path
statusline config show
```

---

## Antigravity CLI Integration

Add to `~/.gemini/antigravity-cli/settings.json`:

```json
{
  "statusLine": {
    "type": "command",
    "command": "/home/user/.gemini/antigravity-cli/statusline",
    "enabled": true
  }
}
```

Or toggle it live inside any active Antigravity CLI session:
```text
/statusline ~/.gemini/antigravity-cli/statusline
```
