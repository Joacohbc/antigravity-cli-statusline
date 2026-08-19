package payload_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/joaco/antigravity-statusline/internal/payload"
)

func TestParse_TableDriven(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		input             string
		fixtureFile       string
		wantErr           error
		checkState        string
		checkWidth        int
		checkModel        string
		checkBranch       string
		checkDirty        bool
		checkSubagents    int
		checkArtifacts    int
		checkTasks        int
		expectEmptyReturn bool
	}{
		{
			name:        "Full payload from fixture file",
			fixtureFile: filepath.Join("..", "..", "test", "testdata", "payload_full.json"),
			wantErr:     nil,
			checkState:  "working",
			checkWidth:  140,
			checkModel:  "Gemini 3.7 Flash",
			checkBranch: "feat/statusline-modernize",
			checkDirty:  true,
			checkSubagents: 3,
			checkArtifacts: 5,
			checkTasks:     4,
		},
		{
			name:        "Minimal payload from fixture file",
			fixtureFile: filepath.Join("..", "..", "test", "testdata", "payload_minimal.json"),
			wantErr:     nil,
			checkState:  "idle",
			checkWidth:  80,
			checkModel:  "gemini-3.7-flash",
		},
		{
			name:        "Corrupted JSON fixture returns ErrInvalidJSON",
			fixtureFile: filepath.Join("..", "..", "test", "testdata", "payload_corrupted.json"),
			wantErr:     payload.ErrInvalidJSON,
			checkState:  "idle",
			checkWidth:  80,
		},
		{
			name:        "Malicious payload with ANSI & control char injection",
			fixtureFile: filepath.Join("..", "..", "test", "testdata", "payload_malicious.json"),
			wantErr:     nil,
			checkState:  "working",
			checkWidth:  100,
			checkModel:  "Gemini 3.7 FlashNew line",
			checkBranch: "feat/red-branch--evil",
			checkDirty:  true,
		},
		{
			name:              "Empty reader returns ErrEmptyPayload",
			input:             "",
			wantErr:           payload.ErrEmptyPayload,
			checkState:        "idle",
			checkWidth:        80,
			expectEmptyReturn: true,
		},
		{
			name:              "Whitespace-only reader returns ErrEmptyPayload",
			input:             "   \t\r\n   ",
			wantErr:           payload.ErrEmptyPayload,
			checkState:        "idle",
			checkWidth:        80,
			expectEmptyReturn: true,
		},
		{
			name:           "Subagents as integer count",
			input:          `{"subagents": 4, "agent_state": "working"}`,
			wantErr:        nil,
			checkState:     "working",
			checkWidth:     80,
			checkSubagents: 4,
		},
		{
			name:           "Subagents as array of strings",
			input:          `{"subagents": ["worker-a", "worker-b"], "agent_state": "working"}`,
			wantErr:        nil,
			checkState:     "working",
			checkWidth:     80,
			checkSubagents: 2,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var inputData string
			if tt.fixtureFile != "" {
				b, err := os.ReadFile(tt.fixtureFile)
				if err != nil {
					t.Fatalf("failed reading fixture file %s: %v", tt.fixtureFile, err)
				}
				inputData = string(b)
			} else {
				inputData = tt.input
			}

			parsed, err := payload.Parse(strings.NewReader(inputData))

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error matching %v, got nil", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("expected error %v, got %v", tt.wantErr, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error parsing payload: %v", err)
				}
			}

			if parsed == nil {
				t.Fatal("parsed payload must never be nil")
			}

			if tt.checkState != "" && parsed.GetAgentState() != tt.checkState {
				t.Errorf("expected state %q, got %q", tt.checkState, parsed.GetAgentState())
			}

			if tt.checkWidth != 0 && parsed.GetTerminalWidth() != tt.checkWidth {
				t.Errorf("expected width %d, got %d", tt.checkWidth, parsed.GetTerminalWidth())
			}

			if tt.checkModel != "" && parsed.GetModel().GetDisplayName() != tt.checkModel {
				t.Errorf("expected model %q, got %q", tt.checkModel, parsed.GetModel().GetDisplayName())
			}

			if tt.checkBranch != "" && parsed.GetVCS().GetBranch() != tt.checkBranch {
				t.Errorf("expected branch %q, got %q", tt.checkBranch, parsed.GetVCS().GetBranch())
			}

			if tt.checkDirty != parsed.GetVCS().IsDirty() {
				t.Errorf("expected dirty=%v, got %v", tt.checkDirty, parsed.GetVCS().IsDirty())
			}

			if tt.checkSubagents != 0 && parsed.GetSubagentCount() != tt.checkSubagents {
				t.Errorf("expected %d subagents, got %d", tt.checkSubagents, parsed.GetSubagentCount())
			}

			if tt.checkArtifacts != 0 && parsed.GetArtifactCount() != tt.checkArtifacts {
				t.Errorf("expected %d artifacts, got %d", tt.checkArtifacts, parsed.GetArtifactCount())
			}

			if tt.checkTasks != 0 && parsed.GetTaskCount() != tt.checkTasks {
				t.Errorf("expected %d tasks, got %d", tt.checkTasks, parsed.GetTaskCount())
			}
		})
	}
}

func TestParse_NilReader(t *testing.T) {
	t.Parallel()

	parsed, err := payload.Parse(nil)
	if err == nil {
		t.Fatal("expected error on nil reader, got nil")
	}
	if !errors.Is(err, payload.ErrEmptyPayload) {
		t.Errorf("expected ErrEmptyPayload on nil reader, got %v", err)
	}
	if parsed == nil {
		t.Fatal("parsed payload on nil reader must return safe default, not nil")
	}
	if parsed.GetAgentState() != "idle" {
		t.Errorf("expected default state 'idle', got %s", parsed.GetAgentState())
	}
	if parsed.GetTerminalWidth() != 80 {
		t.Errorf("expected default width 80, got %d", parsed.GetTerminalWidth())
	}
}

func TestParse_OversizedPayload(t *testing.T) {
	t.Parallel()

	// 2MB + 100 bytes payload
	oversizedBytes := make([]byte, payload.MaxPayloadSize+100)
	for i := range oversizedBytes {
		oversizedBytes[i] = ' '
	}
	// Add some valid json at start
	copy(oversizedBytes, []byte(`{"agent_state":"working"}`))

	parsed, err := payload.Parse(bytes.NewReader(oversizedBytes))
	if err == nil {
		t.Fatal("expected ErrPayloadOversized for payload > 2MB, got nil error")
	}
	if !errors.Is(err, payload.ErrPayloadOversized) {
		t.Errorf("expected ErrPayloadOversized, got %v", err)
	}
	if parsed == nil {
		t.Fatal("expected fallback default payload on oversized error, got nil")
	}

	// Also test ParseBytes with oversized slice
	parsedBytes, errBytes := payload.ParseBytes(oversizedBytes)
	if errBytes == nil || !errors.Is(errBytes, payload.ErrPayloadOversized) {
		t.Errorf("expected ParseBytes to fail with ErrPayloadOversized, got %v", errBytes)
	}
	if parsedBytes == nil {
		t.Fatal("expected fallback payload from ParseBytes, got nil")
	}
}

func TestNilSafety_Accessors(t *testing.T) {
	t.Parallel()

	t.Run("Nil StatePayload receiver", func(t *testing.T) {
		var p *payload.StatePayload

		if p.GetCWD() != "" {
			t.Errorf("expected empty CWD, got %q", p.GetCWD())
		}
		if p.GetSessionID() != "" {
			t.Errorf("expected empty session ID, got %q", p.GetSessionID())
		}
		if p.GetConversationID() != "" {
			t.Errorf("expected empty conversation ID, got %q", p.GetConversationID())
		}
		if p.GetTranscriptPath() != "" {
			t.Errorf("expected empty transcript path, got %q", p.GetTranscriptPath())
		}
		if p.GetVersion() != "" {
			t.Errorf("expected empty version, got %q", p.GetVersion())
		}
		if p.GetProduct() != "" {
			t.Errorf("expected empty product, got %q", p.GetProduct())
		}
		if p.GetPlanTier() != "" {
			t.Errorf("expected empty plan tier, got %q", p.GetPlanTier())
		}
		if p.GetEmail() != "" {
			t.Errorf("expected empty email, got %q", p.GetEmail())
		}
		if p.GetExecutionMode() != "" {
			t.Errorf("expected empty execution mode, got %q", p.GetExecutionMode())
		}
		if p.GetAgentState() != "idle" {
			t.Errorf("expected 'idle', got %q", p.GetAgentState())
		}
		if p.GetTerminalWidth() != 80 {
			t.Errorf("expected 80, got %d", p.GetTerminalWidth())
		}
		if p.GetArtifactCount() != 0 {
			t.Errorf("expected 0, got %d", p.GetArtifactCount())
		}
		if p.GetTaskCount() != 0 {
			t.Errorf("expected 0, got %d", p.GetTaskCount())
		}
		if p.GetSubagentCount() != 0 {
			t.Errorf("expected 0, got %d", p.GetSubagentCount())
		}
		if p.GetPendingInputCount() != 0 {
			t.Errorf("expected 0, got %d", p.GetPendingInputCount())
		}
		if p.IsToolConfirmationPending() {
			t.Error("expected false for ToolConfirmationPending")
		}
		if p.GetVim() != nil {
			t.Error("expected nil Vim")
		}
		if p.GetQuota() == nil {
			t.Error("expected non-nil empty quota map")
		}

		model := p.GetModel()
		if model.GetID() != "unknown" {
			t.Errorf("expected 'unknown', got %q", model.GetID())
		}

		vcs := p.GetVCS()
		if vcs.GetType() != "git" {
			t.Errorf("expected 'git', got %q", vcs.GetType())
		}

		stats := p.GetStats()
		if stats.GetArtifactCount() != 0 || stats.GetTaskCount() != 0 {
			t.Errorf("expected zero stats, got %+v", stats)
		}

		// Mutators on nil receiver should not panic
		p.SanitizeFields()
	})

	t.Run("Zero ModelInfo value", func(t *testing.T) {
		var m payload.ModelInfo
		if m.GetID() != "unknown" {
			t.Errorf("expected unknown, got %s", m.GetID())
		}
		if m.GetDisplayName() != "unknown" {
			t.Errorf("expected unknown, got %s", m.GetDisplayName())
		}
		if m.Name() != "unknown" {
			t.Errorf("expected unknown, got %s", m.Name())
		}
	})

	t.Run("Zero WorkspaceInfo value", func(t *testing.T) {
		var w payload.WorkspaceInfo
		if w.GetCurrentDir() != "" || w.GetProjectDir() != "" {
			t.Errorf("expected empty strings, got %+v", w)
		}
	})

	t.Run("Zero TokenUsage value", func(t *testing.T) {
		var u payload.TokenUsage
		if u.GetInputTokens() != 0 || u.GetOutputTokens() != 0 || u.TotalTokens() != 0 {
			t.Errorf("expected zero tokens, got %+v", u)
		}
		if u.GetCacheCreationInputTokens() != 0 || u.GetCacheReadInputTokens() != 0 {
			t.Errorf("expected zero cache tokens, got %+v", u)
		}
	})

	t.Run("Zero ContextWindowInfo value", func(t *testing.T) {
		var c payload.ContextWindowInfo
		if c.GetTotalInputTokens() != 0 || c.GetTotalOutputTokens() != 0 || c.GetContextWindowSize() != 0 {
			t.Errorf("expected zero token limits, got %+v", c)
		}
		if c.GetUsedPercentage() != 0.0 {
			t.Errorf("expected 0.0 used, got %f", c.GetUsedPercentage())
		}
		if c.GetRemainingPercentage() != 100.0 {
			t.Errorf("expected 100.0 remaining, got %f", c.GetRemainingPercentage())
		}
	})

	t.Run("Zero VCSInfo value", func(t *testing.T) {
		var v payload.VCSInfo
		if v.GetType() != "git" {
			t.Errorf("expected 'git', got %q", v.GetType())
		}
		if v.GetBranch() != "" || v.GetClient() != "" || v.IsDirty() {
			t.Errorf("expected default vcs, got %+v", v)
		}
	})

	t.Run("Zero SandboxInfo value", func(t *testing.T) {
		var s payload.SandboxInfo
		if s.IsEnabled() || s.IsAllowNetwork() {
			t.Errorf("expected false, got %+v", s)
		}
	})

	t.Run("Zero QuotaDetail value", func(t *testing.T) {
		var q payload.QuotaDetail
		if q.GetRemainingFraction() != 1.0 || q.GetResetTime() != "" || q.GetResetInSeconds() != 0 {
			t.Errorf("expected default quota detail, got %+v", q)
		}
	})

	t.Run("Nil VimInfo receiver", func(t *testing.T) {
		var v *payload.VimInfo
		if v.GetMode() != "" {
			t.Errorf("expected empty mode, got %q", v.GetMode())
		}
	})

	t.Run("Zero StatsInfo value", func(t *testing.T) {
		var s payload.StatsInfo
		if s.GetArtifactCount() != 0 || s.GetTaskCount() != 0 || s.GetSubagentCount() != 0 {
			t.Errorf("expected zero stats, got %+v", s)
		}
	})
}

func TestSanitize_Vectors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Plain text unchanged",
			input:    "feat/statusline-v2",
			expected: "feat/statusline-v2",
		},
		{
			name:     "ANSI SGR color code stripped",
			input:    "\x1b[31mfeat/red-branch\x1b[0m",
			expected: "feat/red-branch",
		},
		{
			name:     "ANSI bold, italic, 256 colors stripped",
			input:    "\x1b[1;3;38;5;196mGemini 3.7 Flash\x1b[0m",
			expected: "Gemini 3.7 Flash",
		},
		{
			name:     "OSC title injection stripped",
			input:    "\x1b]0;Evil Terminal Title\x07feat/clean-branch",
			expected: "feat/clean-branch",
		},
		{
			name:     "OSC 8 hyperlink stripped",
			input:    "\x1b]8;;http://example.com\x1b\\Click Here\x1b]8;;\x1b\\",
			expected: "Click Here",
		},
		{
			name:     "Control characters stripped",
			input:    "line1\r\nline2\t\x00\x07end",
			expected: "line1line2end",
		},
		{
			name:     "Unicode characters preserved",
			input:    "✨ 🚀 feat/español-日本語-한국어",
			expected: "✨ 🚀 feat/español-日本語-한국어",
		},
		{
			name:     "Multiple mixed escape codes and controls",
			input:    "\x1b[2J\x1b[H\x1b[?25l\x1b]2;Hacked\x1b\\statusline\r\n",
			expected: "statusline",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := payload.Sanitize(tt.input)
			if got != tt.expected {
				t.Errorf("Sanitize(%q) = %q, expected %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestErrors_Hierarchy(t *testing.T) {
	t.Parallel()

	// Test errors.Is
	if !errors.Is(payload.ErrEmptyPayload, payload.ErrEmptyPayload) {
		t.Error("expected errors.Is(ErrEmptyPayload, ErrEmptyPayload) to be true")
	}
	if !errors.Is(payload.ErrPayloadOversized, payload.ErrPayloadOversized) {
		t.Error("expected errors.Is(ErrPayloadOversized, ErrPayloadOversized) to be true")
	}
	if !errors.Is(payload.ErrInvalidJSON, payload.ErrInvalidJSON) {
		t.Error("expected errors.Is(ErrInvalidJSON, ErrInvalidJSON) to be true")
	}

	// Test ParseError struct and errors.As
	parseErr := &payload.ParseError{
		Op:  "unmarshal",
		Err: payload.ErrInvalidJSON,
		Msg: "unexpected end of JSON input",
	}

	if !errors.Is(parseErr, payload.ErrInvalidJSON) {
		t.Error("expected errors.Is(parseErr, ErrInvalidJSON) to be true via Unwrap")
	}

	var target *payload.ParseError
	if !errors.As(parseErr, &target) {
		t.Error("expected errors.As to succeed for *ParseError")
	}
	if target.Op != "unmarshal" {
		t.Errorf("expected Op 'unmarshal', got %q", target.Op)
	}

	errStr := parseErr.Error()
	if !strings.Contains(errStr, "unmarshal") || !strings.Contains(errStr, "unexpected end") {
		t.Errorf("unexpected error string formatting: %s", errStr)
	}
}
