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
	userCfg := config.DefaultConfig()
	return RenderConfig{
		Width:      0,
		Theme:      theme.DefaultTheme(),
		Mode:       LayoutAuto,
		UserConfig: userCfg,
	}
}

type Option func(*RenderConfig)

func WithWidth(width int) Option {
	return func(c *RenderConfig) {
		c.Width = width
	}
}

func WithTheme(t theme.Theme) Option {
	return func(c *RenderConfig) {
		c.Theme = t
	}
}

func WithMode(mode LayoutMode) Option {
	return func(c *RenderConfig) {
		c.Mode = mode
	}
}

func WithUserConfig(uc *config.Config) Option {
	return func(c *RenderConfig) {
		if uc != nil {
			c.UserConfig = uc
			if uc.Theme.Palette != "" {
				c.Theme = theme.NamedTheme(uc.Theme.Palette)
			}
		}
	}
}

type Renderer struct {
	cfg RenderConfig
}

func NewRenderer(opts ...Option) *Renderer {
	cfg := DefaultRenderConfig()
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	return &Renderer{cfg: cfg}
}

var defaultRenderer = NewRenderer()

func Render(p *payload.StatePayload) string {
	return defaultRenderer.Render(p)
}

func RenderPreview(width int) string {
	if width <= 0 {
		width = 100
	}
	sample := &payload.StatePayload{
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

	r := NewRenderer(WithWidth(width))
	return r.Render(sample)
}

func (r *Renderer) Render(p *payload.StatePayload) string {
	if r == nil {
		r = defaultRenderer
	}

	if p == nil {
		p = payload.NewDefaultStatePayload()
	}

	uc := r.cfg.UserConfig
	if uc == nil {
		uc = config.DefaultConfig()
	}

	width := r.cfg.Width
	if width <= 0 {
		width = p.TerminalWidth
	}
	if width <= 0 {
		width = payload.DefaultTerminalWidth
	}

	wideBreak := uc.Layout.WideBreakpoint
	if wideBreak <= 0 {
		wideBreak = 120
	}
	stdBreak := uc.Layout.StandardBreakpoint
	if stdBreak <= 0 {
		stdBreak = 80
	}

	mode := r.cfg.Mode
	if mode == LayoutAuto {
		styleLower := strings.ToLower(strings.TrimSpace(uc.Layout.Style))
		switch styleLower {
		case "wide", "single-line":
			mode = LayoutWide
		case "compact":
			mode = LayoutCompact
		default:
			switch {
			case width >= wideBreak:
				mode = LayoutWide
			case width >= stdBreak:
				mode = LayoutStandard
			default:
				mode = LayoutCompact
			}
		}
	}

	palette := r.cfg.Theme.Palette

	switch mode {
	case LayoutWide:
		return r.renderWideLayout(p, uc, palette)
	case LayoutStandard:
		return r.renderStandardLayout(p, uc, palette)
	case LayoutCompact:
		return r.renderCompactLayout(p, uc, palette)
	default:
		return r.renderStandardLayout(p, uc, palette)
	}
}

func (r *Renderer) renderWideLayout(p *payload.StatePayload, uc *config.Config, palette theme.Palette) string {
	line1 := r.buildHeaderLine(p, uc, palette)
	line2 := r.buildMetricsLine(p, uc, palette, uc.ContextBar.Length)

	sepStr := "  " + theme.SepPipe + "  "
	sep := theme.Style(palette.Separator, sepStr)

	var b strings.Builder
	b.Grow(len(line1) + len(sep) + len(line2))
	b.WriteString(line1)
	b.WriteString(sep)
	b.WriteString(line2)
	return b.String()
}

func (r *Renderer) renderStandardLayout(p *payload.StatePayload, uc *config.Config, palette theme.Palette) string {
	line1 := r.buildHeaderLine(p, uc, palette)
	line2 := r.buildMetricsLine(p, uc, palette, uc.ContextBar.Length)

	topPrefix := "╭─ "
	if uc.Layout.FrameTop != "" {
		topPrefix = uc.Layout.FrameTop
	}
	bottomPrefix := "╰─ "
	if uc.Layout.FrameBottom != "" {
		bottomPrefix = uc.Layout.FrameBottom
	}

	borderTop := theme.Style(palette.Border, topPrefix)
	borderBottom := theme.Style(palette.Border, bottomPrefix)

	var b strings.Builder
	b.Grow(len(borderTop) + len(line1) + 1 + len(borderBottom) + len(line2))
	b.WriteString(borderTop)
	b.WriteString(line1)
	b.WriteByte('\n')
	b.WriteString(borderBottom)
	b.WriteString(line2)
	return b.String()
}

func (r *Renderer) renderCompactLayout(p *payload.StatePayload, uc *config.Config, palette theme.Palette) string {
	topItems := make([]string, 0, 4)

	if uc.Modules.VimMode {
		if vim := components.RenderVimBadge(p.Vim, palette, uc.Theme.Icons); vim != "" {
			topItems = append(topItems, vim)
		}
	}
	if uc.Modules.AgentState {
		topItems = append(topItems, components.RenderStateBadge(p.AgentState, palette, uc.Theme.Icons))
	}
	if uc.Modules.ModelName {
		if model := components.RenderModelBadge(p.Model, "", palette, uc.Theme.Icons); model != "" {
			topItems = append(topItems, model)
		}
	}

	dotSepStr := uc.Layout.MetricsSeparator
	if dotSepStr == "" {
		dotSepStr = " " + theme.SepDot + " "
	}
	dotSep := theme.Style(palette.Separator, dotSepStr)
	line1 := components.JoinWithSeparator(topItems, dotSep)

	compactBarConfig := uc.ContextBar
	compactBarConfig.Length = min(6, compactBarConfig.Length)

	line2Items := make([]string, 0, 4)
	if uc.Modules.ContextBar {
		ctxBar := components.RenderContextBar(p.ContextWindow, compactBarConfig, palette, uc.Theme.Icons)
		line2Items = append(line2Items, ctxBar)
	}
	if uc.Modules.TasksCount {
		tasks := components.RenderTasksBadge(p.TaskCount, palette, uc.Theme.Icons)
		line2Items = append(line2Items, tasks)
	}
	if uc.Modules.ArtifactsCount && p.ArtifactCount > 0 {
		line2Items = append(line2Items, components.RenderArtifactsBadge(p.ArtifactCount, palette, uc.Theme.Icons))
	}
	if uc.Modules.SubagentsCount && p.SubagentCount > 0 {
		line2Items = append(line2Items, components.RenderSubagentsBadge(p.SubagentCount, palette, uc.Theme.Icons))
	}

	line2 := components.JoinWithSeparator(line2Items, dotSep)

	var b strings.Builder
	b.Grow(len(line1) + 1 + len(line2))
	b.WriteString(line1)
	b.WriteByte('\n')
	b.WriteString(line2)
	return b.String()
}

func (r *Renderer) buildHeaderLine(p *payload.StatePayload, uc *config.Config, palette theme.Palette) string {
	headerItems := make([]string, 0, 6)

	if uc.Modules.VimMode {
		if vim := components.RenderVimBadge(p.Vim, palette, uc.Theme.Icons); vim != "" {
			headerItems = append(headerItems, vim)
		}
	}

	if uc.Modules.AgentState {
		headerItems = append(headerItems, components.RenderStateBadge(p.AgentState, palette, uc.Theme.Icons))
	}

	if uc.Modules.ModelName {
		if model := components.RenderModelBadge(p.Model, p.ExecutionMode, palette, uc.Theme.Icons); model != "" {
			headerItems = append(headerItems, model)
		}
	}

	if uc.Modules.GitBranch {
		if vcs := components.RenderVCSBadge(p.VCS, palette, uc.Theme.Icons); vcs != "" {
			headerItems = append(headerItems, vcs)
		}
	}

	sepStr := uc.Layout.Separator
	if sepStr == "" {
		sepStr = " " + theme.SepSlash + " "
	}
	slashSep := theme.Style(palette.Separator, sepStr)
	return components.JoinWithSeparator(headerItems, slashSep)
}

func (r *Renderer) buildMetricsLine(p *payload.StatePayload, uc *config.Config, palette theme.Palette, barLength int) string {
	barConfig := uc.ContextBar
	if barLength > 0 {
		barConfig.Length = barLength
	}

	allItems := make([]string, 0, 6)
	if uc.Modules.ContextBar {
		ctxBar := components.RenderContextBar(p.ContextWindow, barConfig, palette, uc.Theme.Icons)
		allItems = append(allItems, ctxBar)
	}

	statsItems := components.RenderStatsGroup(p, uc, palette)
	allItems = append(allItems, statsItems...)

	dotSepStr := uc.Layout.MetricsSeparator
	if dotSepStr == "" {
		dotSepStr = " " + theme.SepDot + " "
	}
	dotSep := theme.Style(palette.Separator, dotSepStr)
	return components.JoinWithSeparator(allItems, dotSep)
}
