package renderer_test

import (
	"testing"

	"github.com/joaco/antigravity-statusline/internal/payload"
	"github.com/joaco/antigravity-statusline/internal/renderer"
)

func sampleStatePayload(width int) *payload.StatePayload {
	return &payload.StatePayload{
		AgentState:    "working",
		TerminalWidth: width,
		Model: payload.ModelInfo{
			DisplayName: "Gemini 3.7 Pro",
			ID:          "gemini-3.7-pro",
		},
		VCS: payload.VCSInfo{
			Type:   "git",
			Branch: "feature/fast-renderer",
			Dirty:  true,
		},
		ContextWindow: payload.ContextWindowInfo{
			UsedPercentage:      45.8,
			RemainingPercentage: 54.2,
			ContextWindowSize:   200000,
			TotalInputTokens:    91600,
			TotalOutputTokens:   14200,
		},
		ArtifactCount: 4,
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
}

func BenchmarkRender_Wide(b *testing.B) {
	p := sampleStatePayload(140)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = renderer.Render(p)
	}
}

func BenchmarkRender_Standard(b *testing.B) {
	p := sampleStatePayload(100)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = renderer.Render(p)
	}
}

func BenchmarkRender_Compact(b *testing.B) {
	p := sampleStatePayload(60)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = renderer.Render(p)
	}
}

func BenchmarkRender_Nil(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = renderer.Render(nil)
	}
}

func BenchmarkRenderPreview(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = renderer.RenderPreview(100)
	}
}
