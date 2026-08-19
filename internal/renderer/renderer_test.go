package renderer_test

import (
	"strings"
	"testing"

	"github.com/joaco/antigravity-statusline/internal/payload"
	"github.com/joaco/antigravity-statusline/internal/renderer"
	"github.com/joaco/antigravity-statusline/internal/theme"
)

func TestRender_WideLayout(t *testing.T) {
	t.Parallel()

	p := &payload.StatePayload{
		AgentState:    "idle",
		TerminalWidth: 140,
		Model: payload.ModelInfo{
			DisplayName: "Gemini 3.5 Flash",
		},
		VCS: payload.VCSInfo{
			Branch: "main",
			Dirty:  false,
		},
		ContextWindow: payload.ContextWindowInfo{
			UsedPercentage: 14.2,
		},
		ArtifactCount: 2,
		TaskCount:     1,
		Sandbox: payload.SandboxInfo{
			Enabled: true,
		},
	}

	output := renderer.Render(p)

	if !strings.Contains(output, "READY") {
		t.Errorf("expected READY in output, got: %s", output)
	}
	if !strings.Contains(output, "Gemini 3.5 Flash") {
		t.Errorf("expected model name in output, got: %s", output)
	}
	if !strings.Contains(output, "main") {
		t.Errorf("expected branch in output, got: %s", output)
	}
	if !strings.Contains(output, "14.2%") {
		t.Errorf("expected context percentage, got: %s", output)
	}
	if !strings.Contains(output, theme.SepPipe) {
		t.Errorf("expected pipe separator for wide layout, got: %s", output)
	}
}

func TestRender_StandardLayout(t *testing.T) {
	t.Parallel()

	p := &payload.StatePayload{
		AgentState:    "working",
		TerminalWidth: 100,
		Model: payload.ModelInfo{
			DisplayName: "Gemini 3.7 Pro",
		},
		VCS: payload.VCSInfo{
			Branch: "feature/statusline",
			Dirty:  true,
		},
		ContextWindow: payload.ContextWindowInfo{
			UsedPercentage: 82.5,
		},
		TaskCount: 2,
		Vim: &payload.VimInfo{
			Mode: "NORMAL",
		},
	}

	output := renderer.Render(p)

	if !strings.Contains(output, "WORKING") {
		t.Errorf("expected WORKING in output, got: %s", output)
	}
	if !strings.Contains(output, "NORMAL") {
		t.Errorf("expected NORMAL vim mode in output, got: %s", output)
	}
	if !strings.Contains(output, "╭─") || !strings.Contains(output, "╰─") {
		t.Errorf("expected framed two-line borders, got: %s", output)
	}
}

func TestRender_CompactLayout(t *testing.T) {
	t.Parallel()

	p := &payload.StatePayload{
		AgentState:    "thinking",
		TerminalWidth: 60,
		Model: payload.ModelInfo{
			DisplayName: "Gemini 3.5",
		},
		ContextWindow: payload.ContextWindowInfo{
			UsedPercentage: 5.0,
		},
		TaskCount: 0,
	}

	output := renderer.Render(p)

	if !strings.Contains(output, "THINKING") {
		t.Errorf("expected THINKING in output, got: %s", output)
	}
	if strings.Contains(output, "╭─") {
		t.Errorf("expected no borders in compact mode, got: %s", output)
	}
}

func TestRender_NilSafe(t *testing.T) {
	t.Parallel()

	output := renderer.Render(nil)
	if !strings.Contains(output, "READY") {
		t.Errorf("expected fallback READY on nil payload, got: %s", output)
	}

	var r *renderer.Renderer
	nilReceiverOutput := r.Render(nil)
	if !strings.Contains(nilReceiverOutput, "READY") {
		t.Errorf("expected fallback READY on nil receiver and payload, got: %s", nilReceiverOutput)
	}
}

func TestRenderer_Options(t *testing.T) {
	t.Parallel()

	p := &payload.StatePayload{
		AgentState:    "working",
		TerminalWidth: 60,
		Model: payload.ModelInfo{
			DisplayName: "Gemini 3.7 Pro",
		},
	}

	// Override with WithWidth to Wide layout
	rWide := renderer.NewRenderer(renderer.WithWidth(140))
	outWide := rWide.Render(p)
	if !strings.Contains(outWide, theme.SepPipe) {
		t.Errorf("expected wide pipe separator when overridden with WithWidth(140)")
	}

	// Override with WithMode to Compact
	rCompact := renderer.NewRenderer(renderer.WithMode(renderer.LayoutCompact))
	outCompact := rCompact.Render(p)
	if strings.Contains(outCompact, "╭─") {
		t.Errorf("expected compact output without borders")
	}

	// WithTheme
	plainTheme := theme.PlainTheme()
	rPlain := renderer.NewRenderer(renderer.WithTheme(plainTheme))
	if rPlain == nil {
		t.Errorf("failed creating renderer with WithTheme")
	}
}

func TestRender_Thresholds(t *testing.T) {
	t.Parallel()

	p := &payload.StatePayload{
		AgentState: "idle",
	}

	// < 80: Compact
	p.TerminalWidth = 79
	out79 := renderer.Render(p)
	if strings.Contains(out79, "╭─") || strings.Contains(out79, theme.SepPipe) {
		t.Errorf("width 79 should be compact, got: %s", out79)
	}

	// 80: Standard
	p.TerminalWidth = 80
	out80 := renderer.Render(p)
	if !strings.Contains(out80, "╭─") {
		t.Errorf("width 80 should be standard with borders, got: %s", out80)
	}

	// 119: Standard
	p.TerminalWidth = 119
	out119 := renderer.Render(p)
	if !strings.Contains(out119, "╭─") {
		t.Errorf("width 119 should be standard with borders, got: %s", out119)
	}

	// 120: Wide
	p.TerminalWidth = 120
	out120 := renderer.Render(p)
	if !strings.Contains(out120, theme.SepPipe) {
		t.Errorf("width 120 should be wide with pipe separator, got: %s", out120)
	}
}

func TestRenderPreview(t *testing.T) {
	t.Parallel()

	previewWide := renderer.RenderPreview(140)
	if !strings.Contains(previewWide, "Gemini 3.7 Pro") || !strings.Contains(previewWide, theme.SepPipe) {
		t.Errorf("unexpected wide preview output: %s", previewWide)
	}

	previewStandard := renderer.RenderPreview(90)
	if !strings.Contains(previewStandard, "╭─") {
		t.Errorf("unexpected standard preview output: %s", previewStandard)
	}

	previewCompact := renderer.RenderPreview(50)
	if strings.Contains(previewCompact, "╭─") {
		t.Errorf("unexpected compact preview output: %s", previewCompact)
	}
}
