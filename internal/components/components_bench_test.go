package components_test

import (
	"testing"

	"github.com/joaco/antigravity-statusline/internal/components"
	"github.com/joaco/antigravity-statusline/internal/config"
	"github.com/joaco/antigravity-statusline/internal/payload"
	"github.com/joaco/antigravity-statusline/internal/theme"
)

var (
	benchPalette = theme.GetPalette("default")
	benchIcons   = config.DefaultConfig().Theme.Icons
	benchBarCfg  = config.ContextBarConfig{Length: 10, WarningThreshold: 75, CriticalThreshold: 90}
)

func BenchmarkRenderStateBadge(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = components.RenderStateBadge("working", benchPalette, benchIcons)
	}
}

func BenchmarkRenderVimBadge(b *testing.B) {
	vim := &payload.VimInfo{Mode: "NORMAL"}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = components.RenderVimBadge(vim, benchPalette, benchIcons)
	}
}

func BenchmarkRenderModelBadge(b *testing.B) {
	model := payload.ModelInfo{DisplayName: "Gemini 3.5 Flash", ID: "gemini-3.5-flash"}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = components.RenderModelBadge(model, "PLAN", benchPalette, benchIcons)
	}
}

func BenchmarkRenderVCSBadge(b *testing.B) {
	vcs := payload.VCSInfo{Branch: "feature/fast-renderer", Dirty: true}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = components.RenderVCSBadge(vcs, benchPalette, benchIcons)
	}
}

func BenchmarkRenderContextBar(b *testing.B) {
	ctx := payload.ContextWindowInfo{UsedPercentage: 68.4}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = components.RenderContextBar(ctx, benchBarCfg, benchPalette, benchIcons)
	}
}

func BenchmarkRenderSandboxBadge(b *testing.B) {
	sandbox := payload.SandboxInfo{Enabled: true}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = components.RenderSandboxBadge(sandbox, benchPalette, benchIcons)
	}
}

func BenchmarkRenderArtifactsBadge(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = components.RenderArtifactsBadge(7, benchPalette, benchIcons)
	}
}

func BenchmarkRenderSubagentsBadge(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = components.RenderSubagentsBadge(4, benchPalette, benchIcons)
	}
}

func BenchmarkRenderTasksBadge(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = components.RenderTasksBadge(2, benchPalette, benchIcons)
	}
}

func BenchmarkJoinWithSeparator(b *testing.B) {
	items := []string{"Ready", "Gemini 3.5 Flash", "main", "68.4%"}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = components.JoinWithSeparator(items, " ╱ ")
	}
}
