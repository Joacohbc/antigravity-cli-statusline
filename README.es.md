# Antigravity CLI Status Line ⚡

[![Release](https://img.shields.io/github/v/release/Joacohbc/antigravity-cli-statusline)](https://github.com/Joacohbc/antigravity-cli-statusline/releases)

[English](README.md) | [Español](README.es.md)

Status line ultrarrápida y ligera para Google Antigravity CLI (`agy`) desarrollada en Go. Ofrece ejecución en submilisegundos, iconos Nerd Fonts, barra de contexto Unicode precisa, temas de colores, monitoreo de cuotas en tiempo real y compatibilidad completa con todas las capacidades del nuevo Antigravity CLI 2.x.

---

## 🚀 Capacidades Modernas de AGY

- **Ciclo de Vida Extendido del Agente**: `READY`, `THINKING`, `WORKING`, `TOOL`, `INIT`, `WAITING INPUT`, `WAITING APPROVAL`, `WAITING TASK`, `STREAMING`, `COMPACTING` y `ERROR`.
- **Insignia de Atención y Aprobaciones**: Notificación visual inmediata (`󱅝 APPROVAL REQ` / `󱅝 INPUT REQ (N)`) cuando el agente requiere confirmación o entrada del usuario.
- **Modos de Ejecución**: Identificación visual para modos como `[PLAN]`, `[BOOST]`, `[GOAL]`, `[CAVEMAN]`.
- **Subagentes con Estados de Espera**: Contador de subagentes activos con indicador de subagentes bloqueados o esperando (`󰮝 2 (1⏳)`).
- **Insignias del Sistema**: Contadores en vivo de Servidores MCP (`󱘖 N`), Skills (`󰑬 N`), Plugins (`󰏖 N`) y Terminales Persistentes (`󰆍 N`).
- **Aislamiento y Red del Sandbox**: Diferenciación transparente entre sandbox aislado (`󰌾 ON`) y modo con red permitida (`󰌾 ON (NET)`).
- **Alerta de Compactación de Contexto**: Indicador visual (`󰘚`) cuando el consumo del contexto alcanza o supera el 85%.

---

## ⚡ Instalación Rápida

```bash
curl -fsSL https://raw.githubusercontent.com/Joacohbc/antigravity-cli-statusline/main/install.sh | sh
```

Fijar una versión específica:
```bash
curl -fsSL https://raw.githubusercontent.com/Joacohbc/antigravity-cli-statusline/main/install.sh | VERSION=v1.1.0 sh
```

---

## 🔄 Actualizar statusline

Si ya tenías `statusline` instalado, la actualización es automática y conserva intacta toda tu configuración:

### Opción 1: Desde la propia CLI (Más rápido)
```bash
statusline update
# O fijar una versión específica:
statusline update --version v1.1.0
```

### Opción 2: Reejecutar el instalador
```bash
curl -fsSL https://raw.githubusercontent.com/Joacohbc/antigravity-cli-statusline/main/install.sh | sh
```
El instalador detecta automáticamente la versión instalada previa, verifica el checksum SHA256 del nuevo binario, respalda la versión anterior y no altera tu configuración personalizada en `settings.json`.

---

## 📸 Previsualización

### Modo Wide (1 sola línea)
```text
󱅝 WAITING APPROVAL ╱ 🤖 Gemini 3.7 Pro [BOOST] ╱  main  │  󰍛 ████████▎· 88.5% 󰘚 · 󱅝 APPROVAL REQ · 󰈙 5 · 󰮝 2 (1⏳) · 󰆍 2 · 󱘖 3 · 󰑬 6 · 󰥔 55.0% (5h) · 󰌾 ON (NET)
```

### Modo Estándar (2 líneas enmarcadas)
```text
╭─ 󰚩 READY ╱ 🤖 Gemini 3.7 Pro [PLAN] ╱  main
╰─ 󰍛 ████▊····· 48.5% · 󰈙 3 · 󰮝 2 · 󰆍 1 · 󱘖 2 · 󰑬 4 · 󰥔 98.2% (5h) · 󰃭 99.4% (7d) · 󰌾 ON
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
statusline config set modules.attention_badge true
statusline config set modules.mcp_count true
statusline config set modules.skills_count true
statusline config set modules.terminal_count true
statusline config set modules.sandbox_badge true
statusline config set modules.weekly_usage true

# Ver configuración activa
statusline config path
statusline config show
```

---

## ⚙️ Referencia de Configuración

Archivo de ejemplo `~/.config/antigravity/statusline.json`:

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
