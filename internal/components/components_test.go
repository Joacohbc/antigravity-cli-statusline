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

	enabled := components.RenderSandboxBadge(payload.SandboxInfo{Enabled: true}, testPalette, testIcons)
	if !strings.Contains(enabled, "ON") {
		t.Errorf("expected ON for enabled sandbox, got %q", enabled)
	}

	disabled := components.RenderSandboxBadge(payload.SandboxInfo{Enabled: false}, testPalette, testIcons)
	if !strings.Contains(disabled, "OFF") {
		t.Errorf("expected OFF for disabled sandbox, got %q", disabled)
	}
}
