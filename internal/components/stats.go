package components

import (
	"strconv"
	"strings"

	"github.com/joaco/antigravity-statusline/internal/config"
	"github.com/joaco/antigravity-statusline/internal/payload"
	"github.com/joaco/antigravity-statusline/internal/theme"
)

func RenderArtifactsBadge(count int, p theme.Palette, icons config.IconsConfig) string {
	count = max(0, count)
	num := theme.Style(p.Highlight, strconv.Itoa(count))
	icon := theme.IconArtifact
	if icons.Artifacts != "" {
		icon = icons.Artifacts
	}
	return icon + " " + num
}

func RenderSubagentsBadge(count int, p theme.Palette, icons config.IconsConfig) string {
	if count <= 0 {
		return ""
	}
	num := theme.Style(p.Subagent, strconv.Itoa(count))
	icon := theme.IconSubagent
	if icons.Subagents != "" {
		icon = icons.Subagents
	}
	return icon + " " + num
}

func RenderTasksBadge(count int, p theme.Palette, icons config.IconsConfig) string {
	count = max(0, count)
	num := theme.Style(p.Highlight, strconv.Itoa(count))
	icon := theme.IconTask
	if icons.Tasks != "" {
		icon = icons.Tasks
	}
	return icon + " " + num
}

func RenderSandboxBadge(sandbox payload.SandboxInfo, p theme.Palette, icons config.IconsConfig) string {
	icon := theme.IconSandbox
	if icons.Sandbox != "" {
		icon = icons.Sandbox
	}
	if sandbox.Enabled {
		return theme.Style(p.SandboxON, icon+" ON")
	}
	return theme.Style(p.SandboxOFF, icon+" OFF")
}

func RenderWeeklyUsageBadge(remPercent float64, p theme.Palette, icons config.IconsConfig) string {
	remPercent = min(100.0, max(0.0, remPercent))
	icon := theme.IconWeekly
	if icons.Weekly != "" {
		icon = icons.Weekly
	}
	color := p.ContextOK
	if remPercent < 15.0 {
		color = p.ContextCrit
	} else if remPercent < 30.0 {
		color = p.ContextWarn
	}

	var numBuf [16]byte
	numSlice := strconv.AppendFloat(numBuf[:0], remPercent, 'f', 1, 64)
	numSlice = append(numSlice, '%')
	pctFormatted := string(numSlice)

	val := theme.Style(color, pctFormatted)
	tag := theme.Style(p.Dim, "(7d)")
	return icon + " " + val + " " + tag
}

func RenderSessionUsageBadge(remPercent float64, p theme.Palette, icons config.IconsConfig) string {
	remPercent = min(100.0, max(0.0, remPercent))
	icon := theme.IconSession
	if icons.Session != "" {
		icon = icons.Session
	}
	color := p.ContextOK
	if remPercent < 15.0 {
		color = p.ContextCrit
	} else if remPercent < 30.0 {
		color = p.ContextWarn
	}

	var numBuf [16]byte
	numSlice := strconv.AppendFloat(numBuf[:0], remPercent, 'f', 1, 64)
	numSlice = append(numSlice, '%')
	pctFormatted := string(numSlice)

	val := theme.Style(color, pctFormatted)
	tag := theme.Style(p.Dim, "(5h)")
	return icon + " " + val + " " + tag
}

func RenderStatsGroup(payload *payload.StatePayload, cfg *config.Config, p theme.Palette) []string {
	if payload == nil {
		return []string{}
	}
	if cfg == nil {
		cfg = config.DefaultConfig()
	}

	items := make([]string, 0, 6)

	if cfg.Modules.ArtifactsCount {
		items = append(items, RenderArtifactsBadge(payload.ArtifactCount, p, cfg.Theme.Icons))
	}

	if cfg.Modules.SubagentsCount {
		if sub := RenderSubagentsBadge(payload.SubagentCount, p, cfg.Theme.Icons); sub != "" {
			items = append(items, sub)
		}
	}

	if cfg.Modules.TasksCount {
		items = append(items, RenderTasksBadge(payload.TaskCount, p, cfg.Theme.Icons))
	}

	if cfg.Modules.SessionUsage {
		if sessionUsage, ok := payload.GetSessionUsage(); ok {
			items = append(items, RenderSessionUsageBadge(sessionUsage, p, cfg.Theme.Icons))
		}
	}

	if cfg.Modules.WeeklyUsage {
		if weeklyUsage, ok := payload.GetWeeklyUsage(); ok {
			items = append(items, RenderWeeklyUsageBadge(weeklyUsage, p, cfg.Theme.Icons))
		}
	}

	if cfg.Modules.SandboxBadge {
		items = append(items, RenderSandboxBadge(payload.Sandbox, p, cfg.Theme.Icons))
	}

	return items
}

func JoinWithSeparator(items []string, separator string) string {
	if len(items) == 0 {
		return ""
	}

	nonEmpty := make([]string, 0, len(items))
	totalLen := 0
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed != "" {
			nonEmpty = append(nonEmpty, item)
			totalLen += len(item)
		}
	}

	if len(nonEmpty) == 0 {
		return ""
	}
	if len(nonEmpty) == 1 {
		return nonEmpty[0]
	}

	totalLen += len(separator) * (len(nonEmpty) - 1)
	var b strings.Builder
	b.Grow(totalLen)
	b.WriteString(nonEmpty[0])
	for i := 1; i < len(nonEmpty); i++ {
		b.WriteString(separator)
		b.WriteString(nonEmpty[i])
	}
	return b.String()
}
