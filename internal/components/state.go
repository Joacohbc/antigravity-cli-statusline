package components

import (
	"strings"

	"github.com/joaco/antigravity-statusline/internal/config"
	"github.com/joaco/antigravity-statusline/internal/payload"
	"github.com/joaco/antigravity-statusline/internal/theme"
)

func RenderStateBadge(state string, p theme.Palette, icons config.IconsConfig) string {
	normalized := strings.ToLower(strings.TrimSpace(payload.Sanitize(state)))

	switch normalized {
	case "idle":
		icon := theme.IconStateIdle
		if icons.Idle != "" {
			icon = icons.Idle
		}
		return theme.Style(p.Ready, icon+" READY")
	case "thinking":
		icon := theme.IconStateThinking
		if icons.Thinking != "" {
			icon = icons.Thinking
		}
		return theme.Style(p.Thinking, icon+" THINKING")
	case "working":
		icon := theme.IconStateWorking
		if icons.Working != "" {
			icon = icons.Working
		}
		return theme.Style(p.Working, icon+" WORKING")
	case "tool_use":
		icon := theme.IconStateTool
		if icons.Tool != "" {
			icon = icons.Tool
		}
		return theme.Style(p.Tool, icon+" TOOL")
	case "initializing":
		icon := theme.IconStateInit
		if icons.Init != "" {
			icon = icons.Init
		}
		return theme.Style(p.Init, icon+" INIT")
	default:
		label := strings.ToUpper(normalized)
		if label == "" {
			label = "IDLE"
		}
		return theme.Style(p.Unknown, "⏳ "+label)
	}
}

func RenderVimBadge(vim *payload.VimInfo, p theme.Palette, icons config.IconsConfig) string {
	if vim == nil || vim.Mode == "" {
		return ""
	}

	mode := strings.ToUpper(strings.TrimSpace(payload.Sanitize(vim.Mode)))
	if mode == "" {
		return ""
	}

	icon := theme.IconVim
	if icons.Vim != "" {
		icon = icons.Vim
	}

	switch mode {
	case "NORMAL":
		return theme.Style(p.VimNormal, icon+" NORMAL")
	case "INSERT":
		return theme.Style(p.VimInsert, icon+" INSERT")
	case "VISUAL", "VISUAL LINE":
		return theme.Style(p.VimVisual, icon+" "+mode)
	default:
		return theme.Style(p.VimVisual, icon+" "+mode)
	}
}
