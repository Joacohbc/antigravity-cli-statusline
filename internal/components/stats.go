package components

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/joaco/antigravity-statusline/internal/config"
	"github.com/joaco/antigravity-statusline/internal/payload"
	"github.com/joaco/antigravity-statusline/internal/theme"
)

func RenderArtifactsBadge(count int, palette theme.Palette, icons config.IconsConfig) string {
	safeCount := max(0, count)
	numberText := theme.Style(palette.Highlight, strconv.Itoa(safeCount))
	icon := theme.IconArtifact
	if icons.Artifacts != "" {
		icon = icons.Artifacts
	}
	return icon + " " + numberText
}

func RenderSubagentsBadge(count int, palette theme.Palette, icons config.IconsConfig) string {
	if count <= 0 {
		return ""
	}
	numberText := theme.Style(palette.Subagent, strconv.Itoa(count))
	icon := theme.IconSubagent
	if icons.Subagents != "" {
		icon = icons.Subagents
	}
	return icon + " " + numberText
}

func RenderTasksBadge(count int, palette theme.Palette, icons config.IconsConfig) string {
	safeCount := max(0, count)
	numberText := theme.Style(palette.Highlight, strconv.Itoa(safeCount))
	icon := theme.IconTask
	if icons.Tasks != "" {
		icon = icons.Tasks
	}
	return icon + " " + numberText
}

func RenderSubagentsDetailedBadge(count int, waitingCount int, palette theme.Palette, icons config.IconsConfig) string {
	if count <= 0 {
		return ""
	}
	icon := theme.IconSubagent
	if icons.Subagents != "" {
		icon = icons.Subagents
	}
	if waitingCount > 0 {
		countText := theme.Style(palette.Subagent, strconv.Itoa(count))
		waitingText := theme.Style(palette.Attention, fmt.Sprintf("(%d⏳)", waitingCount))
		return icon + " " + countText + " " + waitingText
	}
	return icon + " " + theme.Style(palette.Subagent, strconv.Itoa(count))
}

func RenderTerminalsBadge(count int, palette theme.Palette, icons config.IconsConfig) string {
	if count <= 0 {
		return ""
	}
	numberText := theme.Style(palette.Terminal, strconv.Itoa(count))
	icon := theme.IconTerminal
	if icons.Terminal != "" {
		icon = icons.Terminal
	}
	return icon + " " + numberText
}

func RenderMCPBadge(count int, palette theme.Palette, icons config.IconsConfig) string {
	if count <= 0 {
		return ""
	}
	numberText := theme.Style(palette.MCP, strconv.Itoa(count))
	icon := theme.IconMCP
	if icons.MCP != "" {
		icon = icons.MCP
	}
	return icon + " " + numberText
}

func RenderSkillsBadge(count int, palette theme.Palette, icons config.IconsConfig) string {
	if count <= 0 {
		return ""
	}
	numberText := theme.Style(palette.Skill, strconv.Itoa(count))
	icon := theme.IconSkill
	if icons.Skill != "" {
		icon = icons.Skill
	}
	return icon + " " + numberText
}

func RenderPluginsBadge(count int, palette theme.Palette, icons config.IconsConfig) string {
	if count <= 0 {
		return ""
	}
	numberText := theme.Style(palette.Highlight, strconv.Itoa(count))
	icon := theme.IconPlugin
	if icons.Plugin != "" {
		icon = icons.Plugin
	}
	return icon + " " + numberText
}

func RenderAttentionBadge(toolConfirmationPending bool, pendingInputCount int, palette theme.Palette, icons config.IconsConfig) string {
	if !toolConfirmationPending && pendingInputCount <= 0 {
		return ""
	}
	icon := theme.IconAttention
	if icons.Attention != "" {
		icon = icons.Attention
	}
	label := "ATTENTION"
	if toolConfirmationPending {
		label = "APPROVAL REQ"
	} else if pendingInputCount > 0 {
		label = fmt.Sprintf("INPUT REQ (%d)", pendingInputCount)
	}
	return theme.Style(palette.Attention, icon+" "+label)
}

func RenderSandboxBadge(sandbox payload.SandboxInfo, palette theme.Palette, icons config.IconsConfig) string {
	icon := theme.IconSandbox
	if icons.Sandbox != "" {
		icon = icons.Sandbox
	}
	if sandbox.Enabled {
		if sandbox.AllowNetwork {
			return theme.Style(palette.SandboxON, icon+" ON (NET)")
		}
		return theme.Style(palette.SandboxON, icon+" ON")
	}
	return theme.Style(palette.SandboxOFF, icon+" OFF")
}

func renderUsageBadge(remainingPercent float64, icon string, durationTag string, palette theme.Palette) string {
	clampedPercent := min(100.0, max(0.0, remainingPercent))
	textColor := palette.ContextOK
	if clampedPercent < 15.0 {
		textColor = palette.ContextCrit
	} else if clampedPercent < 30.0 {
		textColor = palette.ContextWarn
	}

	percentString := fmt.Sprintf("%.1f%%", clampedPercent)
	valueText := theme.Style(textColor, percentString)
	tagText := theme.Style(palette.Dim, durationTag)
	return icon + " " + valueText + " " + tagText
}

func RenderWeeklyUsageBadge(remainingPercent float64, palette theme.Palette, icons config.IconsConfig) string {
	icon := theme.IconWeekly
	if icons.Weekly != "" {
		icon = icons.Weekly
	}
	return renderUsageBadge(remainingPercent, icon, "(7d)", palette)
}

func RenderSessionUsageBadge(remainingPercent float64, palette theme.Palette, icons config.IconsConfig) string {
	icon := theme.IconSession
	if icons.Session != "" {
		icon = icons.Session
	}
	return renderUsageBadge(remainingPercent, icon, "(5h)", palette)
}

func RenderStatsGroup(statePayload *payload.StatePayload, userConfig *config.Config, palette theme.Palette) []string {
	if statePayload == nil {
		return []string{}
	}
	if userConfig == nil {
		userConfig = config.DefaultConfig()
	}

	items := make([]string, 0, 10)

	if userConfig.Modules.AttentionBadge {
		if attentionBadge := RenderAttentionBadge(statePayload.IsToolConfirmationPending(), statePayload.GetPendingInputCount(), palette, userConfig.Theme.Icons); attentionBadge != "" {
			items = append(items, attentionBadge)
		}
	}

	if userConfig.Modules.ArtifactsCount {
		items = append(items, RenderArtifactsBadge(statePayload.ArtifactCount, palette, userConfig.Theme.Icons))
	}

	if userConfig.Modules.SubagentsCount {
		waitingCount := statePayload.GetWaitingSubagentsCount()
		if waitingCount > 0 {
			if subagentsBadge := RenderSubagentsDetailedBadge(statePayload.SubagentCount, waitingCount, palette, userConfig.Theme.Icons); subagentsBadge != "" {
				items = append(items, subagentsBadge)
			}
		} else if subagentsBadge := RenderSubagentsBadge(statePayload.SubagentCount, palette, userConfig.Theme.Icons); subagentsBadge != "" {
			items = append(items, subagentsBadge)
		}
	}

	if userConfig.Modules.TasksCount {
		items = append(items, RenderTasksBadge(statePayload.TaskCount, palette, userConfig.Theme.Icons))
	}

	if userConfig.Modules.TerminalCount {
		if terminalBadge := RenderTerminalsBadge(statePayload.GetTerminalCount(), palette, userConfig.Theme.Icons); terminalBadge != "" {
			items = append(items, terminalBadge)
		}
	}

	if userConfig.Modules.MCPCount {
		if mcpBadge := RenderMCPBadge(statePayload.GetMCPCount(), palette, userConfig.Theme.Icons); mcpBadge != "" {
			items = append(items, mcpBadge)
		}
	}

	if userConfig.Modules.SkillsCount {
		if skillsBadge := RenderSkillsBadge(statePayload.GetSkillsCount(), palette, userConfig.Theme.Icons); skillsBadge != "" {
			items = append(items, skillsBadge)
		}
	}

	if userConfig.Modules.PluginsCount {
		if pluginsBadge := RenderPluginsBadge(statePayload.GetPluginsCount(), palette, userConfig.Theme.Icons); pluginsBadge != "" {
			items = append(items, pluginsBadge)
		}
	}

	if userConfig.Modules.SessionUsage {
		if sessionUsage, hasSessionUsage := statePayload.GetSessionUsage(); hasSessionUsage {
			items = append(items, RenderSessionUsageBadge(sessionUsage, palette, userConfig.Theme.Icons))
		}
	}

	if userConfig.Modules.WeeklyUsage {
		if weeklyUsage, hasWeeklyUsage := statePayload.GetWeeklyUsage(); hasWeeklyUsage {
			items = append(items, RenderWeeklyUsageBadge(weeklyUsage, palette, userConfig.Theme.Icons))
		}
	}

	if userConfig.Modules.SandboxBadge {
		items = append(items, RenderSandboxBadge(statePayload.Sandbox, palette, userConfig.Theme.Icons))
	}

	return items
}

func JoinWithSeparator(items []string, separator string) string {
	if len(items) == 0 {
		return ""
	}

	nonEmptyItems := make([]string, 0, len(items))
	for _, item := range items {
		if strings.TrimSpace(item) != "" {
			nonEmptyItems = append(nonEmptyItems, item)
		}
	}

	return strings.Join(nonEmptyItems, separator)
}
