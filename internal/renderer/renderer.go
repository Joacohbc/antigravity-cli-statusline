package renderer

import (
	"strings"

	"github.com/joaco/antigravity-statusline/internal/components"
	"github.com/joaco/antigravity-statusline/internal/config"
	"github.com/joaco/antigravity-statusline/internal/payload"
	"github.com/joaco/antigravity-statusline/internal/theme"
)

type LayoutMode int

const (
	LayoutAuto LayoutMode = iota
	LayoutWide
	LayoutStandard
	LayoutCompact
)

type RenderConfig struct {
	Width      int
	Theme      theme.Theme
	Mode       LayoutMode
	UserConfig *config.Config
}

func DefaultRenderConfig() RenderConfig {
	userConfig := config.DefaultConfig()
	return RenderConfig{
		Width:      0,
		Theme:      theme.DefaultTheme(),
		Mode:       LayoutAuto,
		UserConfig: userConfig,
	}
}

type Option func(*RenderConfig)

func WithWidth(width int) Option {
	return func(renderConfig *RenderConfig) {
		renderConfig.Width = width
	}
}

func WithTheme(selectedTheme theme.Theme) Option {
	return func(renderConfig *RenderConfig) {
		renderConfig.Theme = selectedTheme
	}
}

func WithMode(mode LayoutMode) Option {
	return func(renderConfig *RenderConfig) {
		renderConfig.Mode = mode
	}
}

func WithUserConfig(userConfig *config.Config) Option {
	return func(renderConfig *RenderConfig) {
		if userConfig != nil {
			renderConfig.UserConfig = userConfig
			if userConfig.Theme.Palette != "" {
				renderConfig.Theme = theme.NamedTheme(userConfig.Theme.Palette)
			}
		}
	}
}

type Renderer struct {
	cfg RenderConfig
}

func NewRenderer(options ...Option) *Renderer {
	renderConfig := DefaultRenderConfig()
	for _, option := range options {
		if option != nil {
			option(&renderConfig)
		}
	}
	return &Renderer{cfg: renderConfig}
}

var defaultRenderer = NewRenderer()

func Render(statePayload *payload.StatePayload) string {
	return defaultRenderer.Render(statePayload)
}

func RenderPreview(width int) string {
	if width <= 0 {
		width = 100
	}
	samplePayload := &payload.StatePayload{
		AgentState:    "working",
		TerminalWidth: width,
		Model: payload.ModelInfo{
			DisplayName: "Gemini 3.7 Pro",
			ID:          "gemini-3.7-pro",
		},
		VCS: payload.VCSInfo{
			Type:   "git",
			Branch: "feature/statusline-v2",
			Dirty:  true,
		},
		ContextWindow: payload.ContextWindowInfo{
			UsedPercentage:      48.5,
			RemainingPercentage: 51.5,
			ContextWindowSize:   200000,
			TotalInputTokens:    85000,
			TotalOutputTokens:   12000,
		},
		ArtifactCount: 3,
		TaskCount:     2,
		SubagentCount: 1,
		Sandbox: payload.SandboxInfo{
			Enabled: true,
		},
		ExecutionMode: "PLAN",
		Vim: &payload.VimInfo{
			Mode: "NORMAL",
		},
	}

	activeRenderer := NewRenderer(WithWidth(width))
	return activeRenderer.Render(samplePayload)
}

func (rendererInstance *Renderer) Render(statePayload *payload.StatePayload) string {
	if rendererInstance == nil {
		rendererInstance = defaultRenderer
	}

	if statePayload == nil {
		statePayload = payload.NewDefaultStatePayload()
	}

	userConfig := rendererInstance.cfg.UserConfig
	if userConfig == nil {
		userConfig = config.DefaultConfig()
	}

	width := rendererInstance.cfg.Width
	if width <= 0 {
		width = statePayload.TerminalWidth
	}
	if width <= 0 {
		width = payload.DefaultTerminalWidth
	}

	wideBreakpoint := userConfig.Layout.WideBreakpoint
	if wideBreakpoint <= 0 {
		wideBreakpoint = 120
	}
	standardBreakpoint := userConfig.Layout.StandardBreakpoint
	if standardBreakpoint <= 0 {
		standardBreakpoint = 80
	}

	mode := rendererInstance.cfg.Mode
	if mode == LayoutAuto {
		styleLower := strings.ToLower(strings.TrimSpace(userConfig.Layout.Style))
		switch styleLower {
		case "wide", "single-line":
			mode = LayoutWide
		case "compact":
			mode = LayoutCompact
		default:
			switch {
			case width >= wideBreakpoint:
				mode = LayoutWide
			case width >= standardBreakpoint:
				mode = LayoutStandard
			default:
				mode = LayoutCompact
			}
		}
	}

	palette := rendererInstance.cfg.Theme.Palette

	switch mode {
	case LayoutWide:
		return rendererInstance.renderWideLayout(statePayload, userConfig, palette)
	case LayoutStandard:
		return rendererInstance.renderStandardLayout(statePayload, userConfig, palette)
	case LayoutCompact:
		return rendererInstance.renderCompactLayout(statePayload, userConfig, palette)
	default:
		return rendererInstance.renderStandardLayout(statePayload, userConfig, palette)
	}
}

func (rendererInstance *Renderer) renderWideLayout(statePayload *payload.StatePayload, userConfig *config.Config, palette theme.Palette) string {
	line1 := rendererInstance.buildHeaderLine(statePayload, userConfig, palette)
	line2 := rendererInstance.buildMetricsLine(statePayload, userConfig, palette, userConfig.ContextBar.Length)

	separatorString := "  " + theme.SepPipe + "  "
	separator := theme.Style(palette.Separator, separatorString)

	var builder strings.Builder
	builder.Grow(len(line1) + len(separator) + len(line2))
	builder.WriteString(line1)
	builder.WriteString(separator)
	builder.WriteString(line2)
	return builder.String()
}

func (rendererInstance *Renderer) renderStandardLayout(statePayload *payload.StatePayload, userConfig *config.Config, palette theme.Palette) string {
	line1 := rendererInstance.buildHeaderLine(statePayload, userConfig, palette)
	line2 := rendererInstance.buildMetricsLine(statePayload, userConfig, palette, userConfig.ContextBar.Length)

	topPrefix := "╭─ "
	if userConfig.Layout.FrameTop != "" {
		topPrefix = userConfig.Layout.FrameTop
	}
	bottomPrefix := "╰─ "
	if userConfig.Layout.FrameBottom != "" {
		bottomPrefix = userConfig.Layout.FrameBottom
	}

	borderTop := theme.Style(palette.Border, topPrefix)
	borderBottom := theme.Style(palette.Border, bottomPrefix)

	var builder strings.Builder
	builder.Grow(len(borderTop) + len(line1) + 1 + len(borderBottom) + len(line2))
	builder.WriteString(borderTop)
	builder.WriteString(line1)
	builder.WriteByte('\n')
	builder.WriteString(borderBottom)
	builder.WriteString(line2)
	return builder.String()
}

func (rendererInstance *Renderer) renderCompactLayout(statePayload *payload.StatePayload, userConfig *config.Config, palette theme.Palette) string {
	topItems := make([]string, 0, 4)

	if userConfig.Modules.VimMode {
		if vimBadge := components.RenderVimBadge(statePayload.Vim, palette, userConfig.Theme.Icons); vimBadge != "" {
			topItems = append(topItems, vimBadge)
		}
	}
	if userConfig.Modules.AgentState {
		topItems = append(topItems, components.RenderStateBadge(statePayload.AgentState, palette, userConfig.Theme.Icons))
	}
	if userConfig.Modules.ModelName {
		if modelBadge := components.RenderModelBadge(statePayload.Model, "", palette, userConfig.Theme.Icons); modelBadge != "" {
			topItems = append(topItems, modelBadge)
		}
	}

	dotSeparatorString := userConfig.Layout.MetricsSeparator
	if dotSeparatorString == "" {
		dotSeparatorString = " " + theme.SepDot + " "
	}
	dotSeparator := theme.Style(palette.Separator, dotSeparatorString)
	line1 := components.JoinWithSeparator(topItems, dotSeparator)

	compactBarConfig := userConfig.ContextBar
	compactBarConfig.Length = min(6, compactBarConfig.Length)

	line2Items := make([]string, 0, 5)
	if userConfig.Modules.ContextBar {
		ctxBar := components.RenderContextBar(statePayload.ContextWindow, compactBarConfig, palette, userConfig.Theme.Icons)
		line2Items = append(line2Items, ctxBar)
	}
	if userConfig.Modules.AttentionBadge {
		if attentionBadge := components.RenderAttentionBadge(statePayload.IsToolConfirmationPending(), statePayload.GetPendingInputCount(), palette, userConfig.Theme.Icons); attentionBadge != "" {
			line2Items = append(line2Items, attentionBadge)
		}
	}
	if userConfig.Modules.TasksCount {
		tasksBadge := components.RenderTasksBadge(statePayload.TaskCount, palette, userConfig.Theme.Icons)
		line2Items = append(line2Items, tasksBadge)
	}
	if userConfig.Modules.ArtifactsCount && statePayload.ArtifactCount > 0 {
		line2Items = append(line2Items, components.RenderArtifactsBadge(statePayload.ArtifactCount, palette, userConfig.Theme.Icons))
	}
	if userConfig.Modules.SubagentsCount && statePayload.SubagentCount > 0 {
		waitingCount := statePayload.GetWaitingSubagentsCount()
		if waitingCount > 0 {
			line2Items = append(line2Items, components.RenderSubagentsDetailedBadge(statePayload.SubagentCount, waitingCount, palette, userConfig.Theme.Icons))
		} else {
			line2Items = append(line2Items, components.RenderSubagentsBadge(statePayload.SubagentCount, palette, userConfig.Theme.Icons))
		}
	}

	line2 := components.JoinWithSeparator(line2Items, dotSeparator)

	var builder strings.Builder
	builder.Grow(len(line1) + 1 + len(line2))
	builder.WriteString(line1)
	builder.WriteByte('\n')
	builder.WriteString(line2)
	return builder.String()
}

func (rendererInstance *Renderer) buildHeaderLine(statePayload *payload.StatePayload, userConfig *config.Config, palette theme.Palette) string {
	headerItems := make([]string, 0, 6)

	if userConfig.Modules.VimMode {
		if vimBadge := components.RenderVimBadge(statePayload.Vim, palette, userConfig.Theme.Icons); vimBadge != "" {
			headerItems = append(headerItems, vimBadge)
		}
	}

	if userConfig.Modules.AgentState {
		headerItems = append(headerItems, components.RenderStateBadge(statePayload.AgentState, palette, userConfig.Theme.Icons))
	}

	if userConfig.Modules.ModelName {
		if modelBadge := components.RenderModelBadge(statePayload.Model, statePayload.ExecutionMode, palette, userConfig.Theme.Icons); modelBadge != "" {
			headerItems = append(headerItems, modelBadge)
		}
	}

	if userConfig.Modules.GitBranch {
		if vcsBadge := components.RenderVCSBadge(statePayload.VCS, palette, userConfig.Theme.Icons); vcsBadge != "" {
			headerItems = append(headerItems, vcsBadge)
		}
	}

	separatorString := userConfig.Layout.Separator
	if separatorString == "" {
		separatorString = " " + theme.SepSlash + " "
	}
	slashSeparator := theme.Style(palette.Separator, separatorString)
	return components.JoinWithSeparator(headerItems, slashSeparator)
}

func (rendererInstance *Renderer) buildMetricsLine(statePayload *payload.StatePayload, userConfig *config.Config, palette theme.Palette, barLength int) string {
	barConfig := userConfig.ContextBar
	if barLength > 0 {
		barConfig.Length = barLength
	}

	allItems := make([]string, 0, 6)
	if userConfig.Modules.ContextBar {
		ctxBar := components.RenderContextBar(statePayload.ContextWindow, barConfig, palette, userConfig.Theme.Icons)
		allItems = append(allItems, ctxBar)
	}

	statsItems := components.RenderStatsGroup(statePayload, userConfig, palette)
	allItems = append(allItems, statsItems...)

	dotSeparatorString := userConfig.Layout.MetricsSeparator
	if dotSeparatorString == "" {
		dotSeparatorString = " " + theme.SepDot + " "
	}
	dotSeparator := theme.Style(palette.Separator, dotSeparatorString)
	return components.JoinWithSeparator(allItems, dotSeparator)
}
