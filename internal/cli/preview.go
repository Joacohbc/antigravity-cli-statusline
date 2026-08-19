package cli

import (
	"github.com/joaco/antigravity-statusline/internal/payload"
	"github.com/joaco/antigravity-statusline/internal/renderer"
	"github.com/joaco/antigravity-statusline/internal/theme"
)

// NewPreviewPayload returns a rich sample StatePayload for interactive preview and debugging.
func NewPreviewPayload(widthOverride int) *payload.StatePayload {
	width := widthOverride
	if width <= 0 {
		width = 100
	}

	return &payload.StatePayload{
		CWD:            "/home/developer/workspace/antigravity-cli",
		SessionID:      "sess-preview-12345",
		ConversationID: "conv-preview-67890",
		TranscriptPath: "/home/developer/.antigravity/transcripts/preview.json",
		AgentState:     "working",
		TerminalWidth:  width,
		Model: payload.ModelInfo{
			ID:          "gemini-3.7-pro",
			DisplayName: "Gemini 3.7 Pro",
		},
		Workspace: payload.WorkspaceInfo{
			CurrentDir: "/home/developer/workspace/antigravity-cli",
			ProjectDir: "/home/developer/workspace",
		},
		Version: "v1.0.0",
		ContextWindow: payload.ContextWindowInfo{
			TotalInputTokens:    85000,
			TotalOutputTokens:   12000,
			ContextWindowSize:   200000,
			UsedPercentage:      48.5,
			RemainingPercentage: 51.5,
			CurrentUsage: payload.TokenUsage{
				InputTokens:              1540,
				OutputTokens:             320,
				CacheCreationInputTokens: 0,
				CacheReadInputTokens:     1200,
			},
		},
		VCS: payload.VCSInfo{
			Type:   "git",
			Branch: "feature/statusline-v2",
			Client: "git-cli",
			Dirty:  true,
		},
		Sandbox: payload.SandboxInfo{
			Enabled:      true,
			AllowNetwork: false,
		},
		ArtifactCount: 3,
		TaskCount:     2,
		SubagentCount: 1,
		PlanTier:      "pro",
		Email:         "developer@example.com",
		ExecutionMode: "PLAN",
		Vim: &payload.VimInfo{
			Mode: "NORMAL",
		},
		WeeklyUsagePercent:  float64Ptr(89.6),
		SessionUsagePercent: float64Ptr(75.6),
		Quota: map[string]payload.QuotaDetail{
			"gemini": {
				RemainingFraction: 0.85,
				ResetTime:         "12:00:00",
				ResetInSeconds:    3600,
			},
		},
	}
}

func float64Ptr(v float64) *float64 {
	return &v
}

// RenderPreviewStatusline generates a rendered statusline preview string using the given width and theme options.
func RenderPreviewStatusline(width int, t theme.Theme) string {
	sample := NewPreviewPayload(width)
	opts := []renderer.Option{
		renderer.WithTheme(t),
	}
	if width > 0 {
		opts = append(opts, renderer.WithWidth(width))
	}
	r := renderer.NewRenderer(opts...)
	return r.Render(sample)
}
