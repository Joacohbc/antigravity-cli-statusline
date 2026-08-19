# Antigravity CLI Status Line ⚡

[![Release](https://img.shields.io/github/v/release/Joacohbc/antigravity-cli-statusline)](https://github.com/Joacohbc/antigravity-cli-statusline/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

[English](README.md) | [Español](README.es.md)

Status line ultrarrápida y ligera para Google Antigravity CLI (`agy`) desarrollada en Go. Ofrece ejecución en submilisegundos, iconos Nerd Fonts, barra de contexto Unicode precisa, temas de colores y monitoreo de cuotas en tiempo real.

---

## ⚡ Instalación Rápida

```bash
curl -fsSL https://raw.githubusercontent.com/Joacohbc/antigravity-cli-statusline/main/install.sh | sh
```

Fijar una versión específica:
```bash
curl -fsSL https://raw.githubusercontent.com/Joacohbc/antigravity-cli-statusline/main/install.sh | VERSION=v1.0.0 sh
```

---

## 📸 Previsualización

### Modo Wide (1 sola línea)
```text
󰚩 READY ╱ 🤖 Gemini 3.7 Flash ╱  main  │  🧠 ·········· 2.7% · 󰈙 0 · 󰒋 0 · 󰥔 75.6% (5h) · 󰃭 89.6% (7d) · 󰌾 OFF
```

### Modo Estándar (2 líneas enmarcadas)
```text
╭─ 󰚩 READY ╱ 🤖 Claude Sonnet 4.6 ╱  main
╰─ 🧠 ████▊····· 48.5% · 󰈙 3 · 󰮝 1 · 󰒋 2 · 󰥔 98.2% (5h) · 󰃭 99.4% (7d) · 󰌾 ON
```

---

## 🛠️ Comandos Frecuentes

```bash
# Previsualizar temas en la terminal
statusline preview --theme catppuccin
statusline --list-themes

# Cambiar el tema de color
statusline config set theme.palette catppuccin
# Temas disponibles: catppuccin, tokyonight, dracula, nord, gruvbox, default, ascii, plain, no-color

# Cambiar diseño a 1 sola línea (wide)
statusline config set layout.style wide

# Activar o desactivar módulos
statusline config set modules.sandbox false
statusline config set modules.weekly_usage true

# Ver configuración activa
statusline config path
statusline config show
```

---

## 🔌 Integración con Antigravity CLI

Agrega a `~/.gemini/antigravity-cli/settings.json`:

```json
{
  "statusLine": {
    "type": "command",
    "command": "~/.gemini/antigravity-cli/statusline",
    "enabled": true
  }
}
```

O actívalo en vivo dentro de cualquier sesión activa de Antigravity CLI:
```text
/statusline ~/.gemini/antigravity-cli/statusline
```

---

## 📄 Licencia

MIT © [Joacohbc](https://github.com/Joacohbc)
