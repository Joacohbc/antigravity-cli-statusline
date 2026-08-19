# Antigravity CLI Status Line (Go Edition) ⚡

A high-performance, layout-adaptive status line for Google Antigravity CLI (`agy`) written in Go. Features rich **Nerd Fonts v3** glyphs, fine-grained Unicode token usage progress bars, theme palettes, and zero-dependency execution.

---

## Features

- ⚡ **Zero-compile installation**: Download precompiled, statically linked binaries directly.
- 🚀 **Ultra-low latency**: Executes in ~1-2 ms in-memory with pure Go stdlib and Cobra.
- 󰚩 **State Badges**: Live color-coded indicators for `READY`, `THINKING`, `WORKING`, `TOOL`, and `INIT`.
-  **Git / VCS Awareness**: Shows branch name with live dirty state marker (`●`).
- 󰍛 **Fine-Grained Context Bar**: 8-level smooth Unicode sub-block progress bar with warning color thresholds.
- 🎨 **Named Color Palettes**: Built-in themes (`tokyonight`, `catppuccin`, `dracula`, `nord`, `gruvbox`, `default`, `ascii`, `plain`, `no-color`).
- 🛠️ **Full User Configurability**: Configure visible modules, custom icons, layout borders, and thresholds via `statusline.json`.
-  **Vim Mode Display**: Supports `NORMAL`, `INSERT`, `VISUAL`, and `VISUAL LINE` modes.
- 📐 **Adaptive Width Rendering**:
  - **Wide ($\ge 120$ cols)**: Single line connected with modern separators.
  - **Standard ($\ge 80$ cols)**: Clean 2-line layout with rounded frame borders (`╭─` / `╰─`).
  - **Compact ($< 80$ cols)**: Streamlined 2-line minimal layout for small splits.

---

## Visual Preview

### Standard View ($\ge 80$ cols)
```text
╭─ 󰚩 READY ╱ 󰧑 Gemini 3.7 Pro ╱  main
╰─ 󰍛 █▍········ 14.2% · 󰈙 2 · 󰒋 1 · 󰌾 ON
```

### Wide View ($\ge 120$ cols)
```text
󱐋 WORKING ╱ 󰧑 Gemini 3.5 Flash ╱  feat/statusline●  │  󰍛 ██████▍········ 42.5% · 󰈙 3 · 󰒋 1 · 󰌾 ON
```

### Compact View ($< 80$ cols)
```text
 NORMAL · 󰘦 THINKING · 󰧑 Gemini 3.5
󰍛 █████▌ 92.0% · 󰒋 2
```

---

## Installation (No Go or build tools required)

### One-line automatic install
```bash
curl -fsSL https://raw.githubusercontent.com/Joacohbc/antigravity-cli-statusline/main/install.sh | bash
```

This script detects your OS (`linux`, `darwin`, `windows`) and architecture (`amd64`, `arm64`), downloads the precompiled binary to `~/.gemini/antigravity-cli/statusline`, and enables it in `settings.json`.

---

## CLI Commands & Subcommands (Cobra)

```bash
# Install binary to ~/.gemini/antigravity-cli and configure settings.json automatically
statusline install

# Install to a custom destination without modifying settings.json
statusline install --dest /custom/path/statusline --skip-settings

# List all available themes with live visual samples
statusline --list-themes
# or: statusline themes

# Render interactive preview with default theme
statusline preview

# Render preview with Tokyo Night theme and custom width
statusline preview --theme tokyonight --width-override 120

# Initialize default configuration file in ~/.config/antigravity/statusline.json
statusline config init

# Set or modify a specific configuration property directly
statusline config set theme.palette tokyonight
statusline config set modules.vim_mode false
statusline config set layout.wide_breakpoint 140

# Display resolved active configuration
statusline config show

# Print path of the active configuration file
statusline config path

# Generate shell autocompletion
statusline completion bash > /etc/bash_completion.d/statusline
```

---

## User Configuration (`statusline.json`)

The binary automatically searches for configuration in:
1. `--config <path>` flag or `$STATUSLINE_CONFIG`
2. `~/.config/antigravity/statusline.json` (XDG standard)
3. `~/.gemini/antigravity-cli/statusline.json`

Example `~/.config/antigravity/statusline.json`:

```json
{
  "theme": {
    "palette": "tokyonight",
    "icon_set": "nerd-fonts",
    "icons": {
      "idle": "󰚩",
      "thinking": "󰘦",
      "working": "󱐋",
      "branch": "",
      "context": "󰍛"
    }
  },
  "modules": {
    "vim_mode": true,
    "agent_state": true,
    "model_name": true,
    "git_branch": true,
    "context_bar": true,
    "artifacts_count": true,
    "subagents_count": true,
    "tasks_count": true,
    "session_usage": true,
    "weekly_usage": true,
    "sandbox_badge": true
  },
  "layout": {
    "style": "framed",
    "separator": " ╱ ",
    "metrics_separator": " · ",
    "frame_top": "╭─ ",
    "frame_bottom": "╰─ ",
    "wide_breakpoint": 120,
    "standard_breakpoint": 80
  },
  "context_bar": {
    "length": 10,
    "style": "smooth",
    "warning_threshold": 75,
    "critical_threshold": 90
  }
}
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
