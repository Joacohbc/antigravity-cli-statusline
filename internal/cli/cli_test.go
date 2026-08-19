package cli_test

import (
	"bytes"
	"io"
	"path/filepath"
	"strings"
	"testing"

	"github.com/joaco/antigravity-statusline/internal/cli"
	"github.com/joaco/antigravity-statusline/internal/theme"
)

func findTestData(t *testing.T, filename string) string {
	t.Helper()
	// Navigate relative to internal/cli directory
	path := filepath.Join("..", "..", "test", "testdata", filename)
	return path
}

func TestRun_HelpFlags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
	}{
		{"ShortHelp", []string{"-h"}},
		{"LongHelp", []string{"--help"}},
		{"HelpSingleDash", []string{"-help"}},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var stdout, stderr bytes.Buffer
			in := bytes.NewReader(nil)

			exitCode := cli.Run(tc.args, in, &stdout, &stderr)

			if exitCode != cli.ExitSuccess {
				t.Fatalf("expected exit code %d, got %d", cli.ExitSuccess, exitCode)
			}

			out := stdout.String()
			if !strings.Contains(out, "antigravity-cli-statusline") {
				t.Errorf("stdout missing tool title, got: %s", out)
			}
			if !strings.Contains(out, "FLAGS:") {
				t.Errorf("stdout missing FLAGS section, got: %s", out)
			}
			if !strings.Contains(out, "ENVIRONMENT VARIABLES:") {
				t.Errorf("stdout missing ENVIRONMENT VARIABLES section, got: %s", out)
			}
			if !strings.Contains(out, "LAYOUT MODES:") {
				t.Errorf("stdout missing LAYOUT MODES section, got: %s", out)
			}
			if !strings.Contains(out, "EXAMPLES:") {
				t.Errorf("stdout missing EXAMPLES section, got: %s", out)
			}
			if !strings.Contains(out, "EXIT CODES:") {
				t.Errorf("stdout missing EXIT CODES section, got: %s", out)
			}

			if stderr.Len() > 0 {
				t.Errorf("expected empty stderr, got: %s", stderr.String())
			}
		})
	}
}

func TestRun_VersionFlags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
	}{
		{"ShortVersion", []string{"-v"}},
		{"LongVersion", []string{"--version"}},
		{"VersionSingleDash", []string{"-version"}},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var stdout, stderr bytes.Buffer
			in := bytes.NewReader(nil)

			exitCode := cli.Run(tc.args, in, &stdout, &stderr)

			if exitCode != cli.ExitSuccess {
				t.Fatalf("expected exit code %d, got %d", cli.ExitSuccess, exitCode)
			}

			out := strings.TrimSpace(stdout.String())
			if out != cli.Version {
				t.Errorf("expected version %q, got %q", cli.Version, out)
			}
			if stderr.Len() > 0 {
				t.Errorf("expected empty stderr, got: %s", stderr.String())
			}
		})
	}
}

func TestRun_PreviewFlags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		args             []string
		expectedModel    string
		expectedSubstr   string
		unexpectedSubstr string
	}{
		{
			name:           "ShortPreview",
			args:           []string{"-p"},
			expectedModel:  "Gemini 3.7 Pro",
			expectedSubstr: "WORKING",
		},
		{
			name:           "LongPreview",
			args:           []string{"--preview"},
			expectedModel:  "Gemini 3.7 Pro",
			expectedSubstr: "WORKING",
		},
		{
			name:           "PreviewWideLayout",
			args:           []string{"--preview", "--width-override", "140"},
			expectedModel:  "Gemini 3.7 Pro",
			expectedSubstr: theme.SepPipe,
		},
		{
			name:           "PreviewStandardLayout",
			args:           []string{"-p", "-w", "100"},
			expectedModel:  "Gemini 3.7 Pro",
			expectedSubstr: "╭─",
		},
		{
			name:             "PreviewCompactLayout",
			args:             []string{"-p", "-w", "60"},
			expectedModel:    "Gemini 3.7 Pro",
			expectedSubstr:   "WORKING",
			unexpectedSubstr: "╭─",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var stdout, stderr bytes.Buffer
			in := bytes.NewReader(nil)

			exitCode := cli.Run(tc.args, in, &stdout, &stderr)

			if exitCode != cli.ExitSuccess {
				t.Fatalf("expected exit code %d, got %d (stderr: %s)", cli.ExitSuccess, exitCode, stderr.String())
			}

			out := stdout.String()
			if !strings.Contains(out, tc.expectedModel) {
				t.Errorf("expected model %q in output, got: %s", tc.expectedModel, out)
			}
			if tc.expectedSubstr != "" && !strings.Contains(out, tc.expectedSubstr) {
				t.Errorf("expected substring %q in output, got: %s", tc.expectedSubstr, out)
			}
			if tc.unexpectedSubstr != "" && strings.Contains(out, tc.unexpectedSubstr) {
				t.Errorf("did not expect substring %q in output, got: %s", tc.unexpectedSubstr, out)
			}
			if stderr.Len() > 0 {
				t.Errorf("expected empty stderr, got: %s", stderr.String())
			}
		})
	}
}

func TestRun_ThemeFlags(t *testing.T) {
	// Not running parallel due to global theme override setting and reset
	tests := []struct {
		name          string
		args          []string
		expectedExit  int
		expectNoANSI  bool
		expectErrText string
	}{
		{
			name:         "ThemeNoColor",
			args:         []string{"--preview", "--theme", "no-color"},
			expectedExit: cli.ExitSuccess,
			expectNoANSI: true,
		},
		{
			name:         "ThemePlain",
			args:         []string{"-p", "--theme", "plain"},
			expectedExit: cli.ExitSuccess,
			expectNoANSI: true,
		},
		{
			name:         "ThemeClassic",
			args:         []string{"-p", "--theme", "classic"},
			expectedExit: cli.ExitSuccess,
			expectNoANSI: true,
		},
		{
			name:         "ThemeASCII",
			args:         []string{"-p", "--theme", "ascii"},
			expectedExit: cli.ExitSuccess,
			expectNoANSI: true,
		},
		{
			name:         "ThemeDefault",
			args:         []string{"-p", "--theme", "default"},
			expectedExit: cli.ExitSuccess,
			expectNoANSI: false,
		},
		{
			name:          "ThemeInvalid",
			args:          []string{"--theme", "invalid_palette"},
			expectedExit:  cli.ExitError,
			expectErrText: "unknown theme",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			in := bytes.NewReader(nil)

			exitCode := cli.Run(tc.args, in, &stdout, &stderr)

			if exitCode != tc.expectedExit {
				t.Fatalf("expected exit code %d, got %d (stderr: %s)", tc.expectedExit, exitCode, stderr.String())
			}

			if tc.expectNoANSI {
				out := stdout.String()
				if strings.Contains(out, "\033[") {
					t.Errorf("expected no ANSI escape sequences, got: %s", out)
				}
			}

			if tc.expectErrText != "" {
				errStr := stderr.String()
				if !strings.Contains(errStr, tc.expectErrText) {
					t.Errorf("expected error message to contain %q, got: %s", tc.expectErrText, errStr)
				}
			}
		})
	}
}

func TestRun_InvalidFlagsAndArguments(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		args          []string
		expectedExit  int
		expectErrText string
	}{
		{
			name:          "UnknownFlag",
			args:          []string{"--unknown-flag"},
			expectedExit:  cli.ExitError,
			expectErrText: "unknown flag",
		},
		{
			name:          "NegativeWidthOverrideLong",
			args:          []string{"--width-override", "-5"},
			expectedExit:  cli.ExitError,
			expectErrText: "invalid width-override -5: width must be non-negative",
		},
		{
			name:          "NegativeWidthOverrideShort",
			args:          []string{"-w", "-1"},
			expectedExit:  cli.ExitError,
			expectErrText: "invalid width-override -1: width must be non-negative",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var stdout, stderr bytes.Buffer
			in := bytes.NewReader(nil)

			exitCode := cli.Run(tc.args, in, &stdout, &stderr)

			if exitCode != tc.expectedExit {
				t.Fatalf("expected exit code %d, got %d", tc.expectedExit, exitCode)
			}

			errStr := stderr.String()
			if !strings.Contains(errStr, tc.expectErrText) {
				t.Errorf("expected stderr to contain %q, got: %s", tc.expectErrText, errStr)
			}
		})
	}
}

func TestRun_ConfigFile(t *testing.T) {
	t.Parallel()

	fullJSONPath := findTestData(t, "payload_full.json")
	corruptedJSONPath := findTestData(t, "payload_corrupted.json")

	tests := []struct {
		name          string
		args          []string
		expectedExit  int
		expectSubstr  string
		expectErrText string
	}{
		{
			name:         "ValidFullPayloadLongFlag",
			args:         []string{"--config", fullJSONPath},
			expectedExit: cli.ExitSuccess,
			expectSubstr: "Gemini 3.7 Flash",
		},
		{
			name:         "ValidFullPayloadShortFlagWithWidth",
			args:         []string{"-c", fullJSONPath, "-w", "140"},
			expectedExit: cli.ExitSuccess,
			expectSubstr: theme.SepPipe,
		},
		{
			name:          "NonExistentConfigFile",
			args:          []string{"--config", "non_existent_file_path_12345.json"},
			expectedExit:  cli.ExitError,
			expectErrText: "failed to open config file",
		},
		{
			name:          "CorruptedConfigFile",
			args:          []string{"--config", corruptedJSONPath},
			expectedExit:  cli.ExitError,
			expectErrText: "failed to parse config file",
		},
		{
			name:          "CorruptedConfigFileWithDebug",
			args:          []string{"--config", corruptedJSONPath, "--debug"},
			expectedExit:  cli.ExitError,
			expectErrText: "debug: parse details",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var stdout, stderr bytes.Buffer
			in := bytes.NewReader(nil)

			exitCode := cli.Run(tc.args, in, &stdout, &stderr)

			if exitCode != tc.expectedExit {
				t.Fatalf("expected exit code %d, got %d (stderr: %s)", tc.expectedExit, exitCode, stderr.String())
			}

			if tc.expectSubstr != "" {
				out := stdout.String()
				if !strings.Contains(out, tc.expectSubstr) {
					t.Errorf("expected stdout to contain %q, got: %s", tc.expectSubstr, out)
				}
			}

			if tc.expectErrText != "" {
				errStr := stderr.String()
				if !strings.Contains(errStr, tc.expectErrText) {
					t.Errorf("expected stderr to contain %q, got: %s", tc.expectErrText, errStr)
				}
			}
		})
	}
}

func TestRun_PipedStdin(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		stdinContent  string
		args          []string
		expectedExit  int
		expectSubstr  string
		expectErrText string
	}{
		{
			name:         "ValidPipedJSON",
			stdinContent: `{"agent_state":"working","model":{"display_name":"Gemini 3.7 Pro"},"terminal_width":100}`,
			args:         nil,
			expectedExit: cli.ExitSuccess,
			expectSubstr: "Gemini 3.7 Pro",
		},
		{
			name:         "ValidPipedJSONWithWidthOverride",
			stdinContent: `{"agent_state":"working","model":{"display_name":"Gemini 3.7 Pro"},"terminal_width":60}`,
			args:         []string{"-w", "140"},
			expectedExit: cli.ExitSuccess,
			expectSubstr: theme.SepPipe,
		},
		{
			name:         "MalformedPipedJSON_GracefulFallback",
			stdinContent: `{"agent_state":"working", "incomplete_json:`,
			args:         nil,
			expectedExit: cli.ExitSuccess,
			expectSubstr: "READY",
		},
		{
			name:          "MalformedPipedJSON_WithDebug",
			stdinContent:  `{"agent_state":"working", "incomplete_json:`,
			args:          []string{"--debug"},
			expectedExit:  cli.ExitSuccess,
			expectSubstr:  "READY",
			expectErrText: "debug: stdin payload parse warning",
		},
		{
			name:         "EmptyPipedStdin_GracefulFallback",
			stdinContent: ``,
			args:         nil,
			expectedExit: cli.ExitSuccess,
			expectSubstr: "READY",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var stdout, stderr bytes.Buffer
			in := bytes.NewBufferString(tc.stdinContent)

			runner := cli.NewRunner(tc.args, in, &stdout, &stderr)
			runner.IsTerminal = func(io.Reader) bool { return false } // explicitly piped

			exitCode := runner.Execute()

			if exitCode != tc.expectedExit {
				t.Fatalf("expected exit code %d, got %d (stderr: %s)", tc.expectedExit, exitCode, stderr.String())
			}

			out := stdout.String()
			if !strings.Contains(out, tc.expectSubstr) {
				t.Errorf("expected stdout to contain %q, got: %s", tc.expectSubstr, out)
			}

			if tc.expectErrText != "" {
				errStr := stderr.String()
				if !strings.Contains(errStr, tc.expectErrText) {
					t.Errorf("expected stderr to contain %q, got: %s", tc.expectErrText, errStr)
				}
			}
		})
	}
}

func TestRun_InteractiveNonBlocking(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	// Create an empty input that would not be read because isTerminal is true
	in := bytes.NewReader(nil)

	runner := cli.NewRunner(nil, in, &stdout, &stderr)
	runner.IsTerminal = func(io.Reader) bool { return true } // simulate interactive TTY

	exitCode := runner.Execute()

	if exitCode != cli.ExitSuccess {
		t.Fatalf("expected exit code %d, got %d", cli.ExitSuccess, exitCode)
	}

	out := stdout.String()
	if !strings.Contains(out, "Gemini 3.7 Pro") {
		t.Errorf("expected interactive preview statusline to contain 'Gemini 3.7 Pro', got: %s", out)
	}
	if stderr.Len() > 0 {
		t.Errorf("expected empty stderr in interactive mode, got: %s", stderr.String())
	}
}

func TestNewRunner_Defaults(t *testing.T) {
	t.Parallel()

	r := cli.NewRunner(nil, nil, nil, nil)
	if r.In == nil {
		t.Errorf("expected non-nil default In")
	}
	if r.Out == nil {
		t.Errorf("expected non-nil default Out")
	}
	if r.Err == nil {
		t.Errorf("expected non-nil default Err")
	}
	if r.IsTerminal == nil {
		t.Errorf("expected non-nil default IsTerminal")
	}
}

func TestDefaultIsTerminal(t *testing.T) {
	t.Parallel()

	if cli.DefaultIsTerminal(nil) {
		t.Errorf("expected false for nil reader")
	}

	buf := bytes.NewBufferString("test")
	if cli.DefaultIsTerminal(buf) {
		t.Errorf("expected false for bytes.Buffer")
	}
}

func TestPreviewRenderers(t *testing.T) {
	t.Parallel()

	payload := cli.NewPreviewPayload(140)
	if payload == nil {
		t.Fatal("expected non-nil preview payload")
	}
	if payload.TerminalWidth != 140 {
		t.Errorf("expected terminal width 140, got %d", payload.TerminalWidth)
	}
	if payload.Model.DisplayName != "Gemini 3.7 Pro" {
		t.Errorf("expected model 'Gemini 3.7 Pro', got %s", payload.Model.DisplayName)
	}

	defaultWidthPayload := cli.NewPreviewPayload(0)
	if defaultWidthPayload.TerminalWidth != 100 {
		t.Errorf("expected fallback width 100, got %d", defaultWidthPayload.TerminalWidth)
	}

	rendered := cli.RenderPreviewStatusline(120, theme.DefaultTheme())
	if !strings.Contains(rendered, "Gemini 3.7 Pro") {
		t.Errorf("expected rendered preview to contain 'Gemini 3.7 Pro', got: %s", rendered)
	}
}
