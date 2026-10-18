# Antigravity CLI Status Line ⚡

[![Release](https://img.shields.io/github/v/release/Joacohbc/antigravity-cli-statusline)](https://github.com/Joacohbc/antigravity-cli-statusline/releases)

[English](README.md) | [Español](README.es.md)

A lightweight, standalone status line for modern Google Antigravity CLI (`agy`) built with Go. Features sub-millisecond execution, rich Nerd Font glyphs, smooth Unicode context bars, theme palettes, real-time quota tracking, and full support for all modern Antigravity CLI 2.x capabilities.

---

## 🚀 Modern AGY Features

- **Rich Agent Lifecycle States**: `READY`, `THINKING`, `WORKING`, `TOOL`, `INIT`, `WAITING INPUT`, `WAITING APPROVAL`, `WAITING TASK`, `STREAMING`, `COMPACTING`, and `ERROR`.
- **Attention & Approval Alert**: Immediate badge (`󱅝 APPROVAL REQ` / `󱅝 INPUT REQ (N)`) when agent requires confirmation or human input.
- **Execution Modes**: Visual tags for modern Antigravity modes (`[PLAN]`, `[BOOST]`, `[GOAL]`, `[CAVEMAN]`).
- **Subagent Lifecycle & Waiting Indicator**: Tracks total active subagents and highlights waiting background subagents (`󰮝 2 (1⏳)`).
- **System Badges**: Live counters for MCP Servers (`󱘖 N`), Skills (`󰑬 N`), Plugins (`󰏖 N`), and Persistent Terminals (`󰆍 N`).
- **Network-Aware Sandbox Badge**: Distinguishes standard isolated sandbox (`󰌾 ON`) from network-enabled execution (`󰌾 ON (NET)`).
- **Context Compaction Warning**: Displays compaction alert icon (`󰘚`) when context window usage reaches $\ge 85\%$.

---

## ⚡ Quick Install

```bash
curl -fsSL https://raw.githubusercontent.com/Joacohbc/antigravity-cli-statusline/main/install.sh | sh
```

Pin a specific version:
```bash
curl -fsSL https://raw.githubusercontent.com/Joacohbc/antigravity-cli-statusline/main/install.sh | VERSION=v1.1.0 sh
```

---

## 📸 Preview

### Wide Mode (Single line)
```text
󱅝 WAITING APPROVAL ╱ 🤖 Gemini 3.7 Pro [BOOST] ╱  main  │  󰍛 ████████▎· 88.5% 󰘚 · 󱅝 APPROVAL REQ · 󰈙 5 · 󰮝 2 (1⏳) · 󰆍 2 · 󱘖 3 · 󰑬 6 · 󰥔 55.0% (5h) · 󰌾 ON (NET)
```

### Standard Mode (Framed 2-lines)
```text
╭─ 󰚩 READY ╱ 🤖 Gemini 3.7 Pro [PLAN] ╱  main
╰─ 󰍛 ████▊····· 48.5% · 󰈙 3 · 󰮝 2 · 󰆍 1 · 󱘖 2 · 󰑬 4 · 󰥔 98.2% (5h) · 󰃭 99.4% (7d) · 󰌾 ON
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
statusline config set modules.attention_badge true
statusline config set modules.mcp_count true
statusline config set modules.skills_count true
statusline config set modules.terminal_count true
statusline config set modules.sandbox_badge true
statusline config set modules.weekly_usage true

# Show active config file
statusline config path
statusline config show
```

---

## ⚙️ Configuration Reference

Example `~/.config/antigravity/statusline.json`:

```json
{
  "theme": {
    "palette": "catppuccin",
    "icon_set": "nerd_font"
  },
  "layout": {
    "style": "framed",
    "separator": " ╱ ",
    "metrics_separator": " · ",
    "wide_breakpoint": 120,
    "standard_breakpoint": 80
  },
  "modules": {
    "vim_mode": true,
    "agent_state": true,
    "model_name": true,
    "git_branch": true,
    "context_bar": true,
    "attention_badge": true,
    "artifacts_count": true,
    "subagents_count": true,
    "tasks_count": true,
    "terminal_count": true,
    "mcp_count": true,
    "skills_count": true,
    "plugins_count": true,
    "sandbox_badge": true,
    "session_usage": true,
    "weekly_usage": true
  },
  "context_bar": {
    "length": 10,
    "warning_threshold": 75.0,
    "critical_threshold": 90.0
  }
}
```

---

## 🔌 Antigravity CLI Integration

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
