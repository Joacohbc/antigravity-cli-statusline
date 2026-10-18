package components_test

import (
	"strings"
	"testing"

	"github.com/joaco/antigravity-statusline/internal/components"
	"github.com/joaco/antigravity-statusline/internal/config"
	"github.com/joaco/antigravity-statusline/internal/payload"
	"github.com/joaco/antigravity-statusline/internal/theme"
)

var (
	testPalette = theme.GetPalette("default")
	testIcons   = config.DefaultConfig().Theme.Icons
	testBarCfg  = config.ContextBarConfig{Length: 10, WarningThreshold: 75, CriticalThreshold: 90}
)

func TestRenderStateBadge(t *testing.T) {
	t.Parallel()

	tests := []struct {
		state    string
		expected string
	}{
		{"idle", "READY"},
		{"IDLE", "READY"},
		{"thinking", "THINKING"},
		{"working", "WORKING"},
		{"tool_use", "TOOL"},
		{"initializing", "INIT"},
		{"waiting_for_input", "WAITING INPUT"},
		{"waiting_for_approval", "WAITING APPROVAL"},
		{"waiting_for_dependents", "WAITING TASK"},
		{"streaming", "STREAMING"},
		{"compacting", "COMPACTING"},
		{"error", "ERROR"},
		{"unknown", "UNKNOWN"},
		{"", "IDLE"},
	}

	for _, tt := range tests {
		t.Run(tt.state, func(t *testing.T) {
			badge := components.RenderStateBadge(tt.state, testPalette, testIcons)
			if !strings.Contains(badge, tt.expected) {
				t.Errorf("state %q: expected badge to contain %q, got %q", tt.state, tt.expected, badge)
			}
		})
	}
}

func TestRenderVimBadge(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		vim      *payload.VimInfo
		expected string
	}{
		{"nil vim", nil, ""},
		{"empty mode", &payload.VimInfo{Mode: ""}, ""},
		{"normal mode", &payload.VimInfo{Mode: "NORMAL"}, "NORMAL"},
		{"insert mode", &payload.VimInfo{Mode: "insert"}, "INSERT"},
		{"visual mode", &payload.VimInfo{Mode: "VISUAL"}, "VISUAL"},
		{"visual line mode", &payload.VimInfo{Mode: "VISUAL LINE"}, "VISUAL LINE"},
		{"custom mode", &payload.VimInfo{Mode: "REPLACE"}, "REPLACE"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			badge := components.RenderVimBadge(tt.vim, testPalette, testIcons)
			if tt.expected == "" {
				if badge != "" {
					t.Errorf("expected empty badge, got: %q", badge)
				}
			} else {
				if !strings.Contains(badge, tt.expected) {
					t.Errorf("expected badge to contain %q, got: %q", tt.expected, badge)
				}
			}
		})
	}
}

func TestRenderModelBadge(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		model    payload.ModelInfo
		mode     string
		expected string
	}{
		{"empty", payload.ModelInfo{}, "", ""},
		{"display name", payload.ModelInfo{DisplayName: "Gemini 3.7 Pro"}, "", "Gemini 3.7 Pro"},
		{"id fallback", payload.ModelInfo{ID: "gemini-flash"}, "", "gemini-flash"},
		{"with mode", payload.ModelInfo{DisplayName: "Gemini"}, "PLAN", "[PLAN]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			badge := components.RenderModelBadge(tt.model, tt.mode, testPalette, testIcons)
			if tt.expected == "" {
				if badge != "" {
					t.Errorf("expected empty badge, got: %q", badge)
				}
			} else {
				if !strings.Contains(badge, tt.expected) {
					t.Errorf("expected badge to contain %q, got: %q", tt.expected, badge)
				}
			}
		})
	}
}

func TestRenderVCSBadge(t *testing.T) {
	t.Parallel()

	cleanVCS := payload.VCSInfo{Branch: "main", Dirty: false}
	cleanBadge := components.RenderVCSBadge(cleanVCS, testPalette, testIcons)
	if !strings.Contains(cleanBadge, "main") || strings.Contains(cleanBadge, theme.IconGitDirty) {
		t.Errorf("clean VCS badge incorrect: %q", cleanBadge)
	}

	dirtyVCS := payload.VCSInfo{Branch: "dev", Dirty: true}
	dirtyBadge := components.RenderVCSBadge(dirtyVCS, testPalette, testIcons)
	if !strings.Contains(dirtyBadge, "dev") || !strings.Contains(dirtyBadge, theme.IconGitDirty) {
		t.Errorf("dirty VCS badge incorrect: %q", dirtyBadge)
	}

	emptyVCS := payload.VCSInfo{Branch: ""}
	if emptyBadge := components.RenderVCSBadge(emptyVCS, testPalette, testIcons); emptyBadge != "" {
		t.Errorf("expected empty string for empty branch, got %q", emptyBadge)
	}
}

func TestRenderContextBar(t *testing.T) {
	t.Parallel()

	ctx := payload.ContextWindowInfo{UsedPercentage: 75.0}
	bar := components.RenderContextBar(ctx, testBarCfg, testPalette, testIcons)
	if !strings.Contains(bar, "75.0%") {
		t.Errorf("expected 75.0%% in bar, got %q", bar)
	}
	if !strings.Contains(bar, theme.IconContext) {
		t.Errorf("expected context icon in bar, got %q", bar)
	}
}

func TestRenderSandboxBadge(t *testing.T) {
	t.Parallel()

	enabled := components.RenderSandboxBadge(payload.SandboxInfo{Enabled: true, AllowNetwork: false}, testPalette, testIcons)
	if !strings.Contains(enabled, "ON") {
		t.Errorf("expected ON for enabled sandbox, got %q", enabled)
	}

	enabledNet := components.RenderSandboxBadge(payload.SandboxInfo{Enabled: true, AllowNetwork: true}, testPalette, testIcons)
	if !strings.Contains(enabledNet, "ON (NET)") {
		t.Errorf("expected ON (NET) for network enabled sandbox, got %q", enabledNet)
	}

	disabled := components.RenderSandboxBadge(payload.SandboxInfo{Enabled: false}, testPalette, testIcons)
	if !strings.Contains(disabled, "OFF") {
		t.Errorf("expected OFF for disabled sandbox, got %q", disabled)
	}
}

func TestRenderAttentionBadge(t *testing.T) {
	t.Parallel()

	none := components.RenderAttentionBadge(false, 0, testPalette, testIcons)
	if none != "" {
		t.Errorf("expected empty badge when no attention needed, got %q", none)
	}

	approval := components.RenderAttentionBadge(true, 0, testPalette, testIcons)
	if !strings.Contains(approval, "APPROVAL REQ") {
		t.Errorf("expected APPROVAL REQ in badge, got %q", approval)
	}

	input := components.RenderAttentionBadge(false, 3, testPalette, testIcons)
	if !strings.Contains(input, "INPUT REQ (3)") {
		t.Errorf("expected INPUT REQ (3) in badge, got %q", input)
	}
}

func TestRenderSystemBadges(t *testing.T) {
	t.Parallel()

	mcpBadge := components.RenderMCPBadge(3, testPalette, testIcons)
	if !strings.Contains(mcpBadge, "3") {
		t.Errorf("expected 3 in MCP badge, got %q", mcpBadge)
	}

	skillsBadge := components.RenderSkillsBadge(5, testPalette, testIcons)
	if !strings.Contains(skillsBadge, "5") {
		t.Errorf("expected 5 in skills badge, got %q", skillsBadge)
	}

	terminalsBadge := components.RenderTerminalsBadge(2, testPalette, testIcons)
	if !strings.Contains(terminalsBadge, "2") {
		t.Errorf("expected 2 in terminals badge, got %q", terminalsBadge)
	}

	pluginsBadge := components.RenderPluginsBadge(4, testPalette, testIcons)
	if !strings.Contains(pluginsBadge, "4") {
		t.Errorf("expected 4 in plugins badge, got %q", pluginsBadge)
	}
}

func TestRenderSubagentsDetailedBadge(t *testing.T) {
	t.Parallel()

	normal := components.RenderSubagentsDetailedBadge(3, 0, testPalette, testIcons)
	if !strings.Contains(normal, "3") || strings.Contains(normal, "⏳") {
		t.Errorf("expected 3 without waiting indicator, got %q", normal)
	}

	waiting := components.RenderSubagentsDetailedBadge(3, 1, testPalette, testIcons)
	if !strings.Contains(waiting, "3") || !strings.Contains(waiting, "1⏳") {
		t.Errorf("expected 3 and (1⏳), got %q", waiting)
	}
}
