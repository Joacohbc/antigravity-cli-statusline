package components

import (
	"strings"

	"github.com/joaco/antigravity-statusline/internal/config"
	"github.com/joaco/antigravity-statusline/internal/payload"
	"github.com/joaco/antigravity-statusline/internal/theme"
)

func RenderStateBadge(state string, palette theme.Palette, icons config.IconsConfig) string {
	normalizedState := strings.ToLower(strings.TrimSpace(payload.Sanitize(state)))

	switch normalizedState {
	case "idle":
		icon := theme.IconStateIdle
		if icons.Idle != "" {
			icon = icons.Idle
		}
		return theme.Style(palette.Ready, icon+" READY")
	case "thinking":
		icon := theme.IconStateThinking
		if icons.Thinking != "" {
			icon = icons.Thinking
		}
		return theme.Style(palette.Thinking, icon+" THINKING")
	case "working":
		icon := theme.IconStateWorking
		if icons.Working != "" {
			icon = icons.Working
		}
		return theme.Style(palette.Working, icon+" WORKING")
	case "tool_use":
		icon := theme.IconStateTool
		if icons.Tool != "" {
			icon = icons.Tool
		}
		return theme.Style(palette.Tool, icon+" TOOL")
	case "initializing":
		icon := theme.IconStateInit
		if icons.Init != "" {
			icon = icons.Init
		}
		return theme.Style(palette.Init, icon+" INIT")
	case "waiting_for_input", "waiting_input", "waiting-for-input", "waiting-input":
		icon := theme.IconStateWaitingInput
		if icons.WaitingInput != "" {
			icon = icons.WaitingInput
		}
		return theme.Style(palette.Attention, icon+" WAITING INPUT")
	case "waiting_for_approval", "waiting_approval", "waiting-for-approval", "waiting-approval":
		icon := theme.IconStateWaitingApproval
		if icons.WaitingApproval != "" {
			icon = icons.WaitingApproval
		}
		return theme.Style(palette.Attention, icon+" WAITING APPROVAL")
	case "waiting_for_dependents", "waiting_dependents", "waiting-for-dependents", "waiting-dependents":
		icon := theme.IconStateWaitingDependent
		if icons.WaitingDependent != "" {
			icon = icons.WaitingDependent
		}
		return theme.Style(palette.Thinking, icon+" WAITING TASK")
	case "streaming":
		icon := theme.IconStateStreaming
		if icons.Streaming != "" {
			icon = icons.Streaming
		}
		return theme.Style(palette.Working, icon+" STREAMING")
	case "compacting":
		icon := theme.IconStateCompacting
		if icons.Compacting != "" {
			icon = icons.Compacting
		}
		return theme.Style(palette.Thinking, icon+" COMPACTING")
	case "error", "errored", "failed":
		icon := theme.IconStateError
		if icons.Error != "" {
			icon = icons.Error
		}
		return theme.Style(palette.Error, icon+" ERROR")
	default:
		label := strings.ToUpper(normalizedState)
		if label == "" {
			label = "IDLE"
		}
		return theme.Style(palette.Unknown, "⏳ "+label)
	}
}

func RenderVimBadge(vim *payload.VimInfo, palette theme.Palette, icons config.IconsConfig) string {
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
		return theme.Style(palette.VimNormal, icon+" NORMAL")
	case "INSERT":
		return theme.Style(palette.VimInsert, icon+" INSERT")
	case "VISUAL", "VISUAL LINE":
		return theme.Style(palette.VimVisual, icon+" "+mode)
	default:
		return theme.Style(palette.VimVisual, icon+" "+mode)
	}
}
