package e2e_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/joaco/antigravity-statusline/internal/theme"
)

var (
	binaryPath string
	projectDir string
)

func TestMain(m *testing.M) {
	// Discover project directory
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get working dir: %v\n", err)
		os.Exit(1)
	}

	// Navigate to antigravity-cli-statusline root (2 levels up from test/e2e)
	projectDir = filepath.Clean(filepath.Join(wd, "..", ".."))

	tmpDir, err := os.MkdirTemp("", "statusline-e2e-bin-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create temp dir for binary: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	binName := "statusline"
	if runtime.GOOS == "windows" {
		binName = "statusline.exe"
	}
	bin := filepath.Join(tmpDir, binName)

	buildCmd := exec.Command("go", "build", "-o", bin, "./cmd/statusline")
	buildCmd.Dir = projectDir
	buildCmd.Env = os.Environ()

	var buildErr bytes.Buffer
	buildCmd.Stderr = &buildErr

	if err := buildCmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to build statusline test binary: %v\n%s\n", err, buildErr.String())
		os.Exit(1)
	}

	binaryPath = bin

	exitCode := m.Run()
	os.Exit(exitCode)
}

// runCLI executes the compiled statusline binary as an isolated child process.
func runCLI(t *testing.T, args []string, stdin io.Reader, extraEnv ...string) (stdout string, stderr string, exitCode int) {
	t.Helper()

	cmd := exec.Command(binaryPath, args...)
	cmd.Dir = projectDir

	// Default clean environment inheriting PATH and system variables
	envMap := make(map[string]string)
	for _, envStr := range os.Environ() {
		if strings.HasPrefix(envStr, "NO_COLOR=") || strings.HasPrefix(envStr, "TERM=") {
			continue // filter out caller env that might interfere
		}
		parts := strings.SplitN(envStr, "=", 2)
		if len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		}
	}
	// Default TERM to xterm-256color for consistent test behavior
	envMap["TERM"] = "xterm-256color"

	// Apply extra custom environment overrides
	for _, envStr := range extraEnv {
		parts := strings.SplitN(envStr, "=", 2)
		if len(parts) == 2 {
			if parts[1] == "" {
				delete(envMap, parts[0])
			} else {
				envMap[parts[0]] = parts[1]
			}
		}
	}

	cmdEnv := make([]string, 0, len(envMap))
	for k, v := range envMap {
		cmdEnv = append(cmdEnv, k+"="+v)
	}
	cmd.Env = cmdEnv

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	if stdin != nil {
		cmd.Stdin = stdin
	}

	err := cmd.Run()
	stdout = stdoutBuf.String()
	stderr = stderrBuf.String()

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("failed to execute binary %q: %v", binaryPath, err)
		}
	} else {
		exitCode = 0
	}

	return stdout, stderr, exitCode
}

func getTestDataPath(t *testing.T, filename string) string {
	t.Helper()
	path := filepath.Join(projectDir, "test", "testdata", filename)
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("test fixture file not found: %s", path)
	}
	return path
}

var ansiPattern = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

func containsANSI(s string) bool {
	return ansiPattern.MatchString(s)
}

// =============================================================================
// TIER 1: FEATURE COVERAGE (ISOLATION)
// =============================================================================

func TestTier1_PipelineMode_ValidJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		fixtureFile  string
		expectedSubs []string
	}{
		{
			name:        "FullPayload",
			fixtureFile: "payload_full.json",
			expectedSubs: []string{
				"Gemini 3.7 Flash",
				"feat/statusline-modernize",
				"25.0%",
			},
		},
		{
			name:        "MinimalPayload",
			fixtureFile: "payload_minimal.json",
			expectedSubs: []string{
				"gemini-3.7-flash",
				"READY",
			},
		},
		{
			name:        "CompactPayload",
			fixtureFile: "payload_compact.json",
			expectedSubs: []string{
				"Gemini Flash",
				"NORMAL",
			},
		},
		{
			name:        "UnicodePayload",
			fixtureFile: "payload_unicode.json",
			expectedSubs: []string{
				"Gemini 3.7 Pro (日本語・UTF8 🚀✨)",
				"feat/🌿-feature-äöü-🚀-branch",
			},
		},
		{
			name:        "RealWorldAGYWorkload",
			fixtureFile: "payload_agy_workload.json",
			expectedSubs: []string{
				"Gemini 3.7 Pro",
				"feat/m4-e2e-testing",
				"81.5%",
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			fixturePath := getTestDataPath(t, tc.fixtureFile)
			content, err := os.ReadFile(fixturePath)
			if err != nil {
				t.Fatalf("failed reading fixture: %v", err)
			}

			stdout, stderr, exitCode := runCLI(t, nil, bytes.NewReader(content))

			if exitCode != 0 {
				t.Fatalf("expected exit code 0, got %d (stderr: %s)", exitCode, stderr)
			}

			for _, expected := range tc.expectedSubs {
				if !strings.Contains(stdout, expected) {
					t.Errorf("expected stdout to contain %q, but got:\n%s", expected, stdout)
				}
			}

			if stderr != "" {
				t.Errorf("expected empty stderr, got: %s", stderr)
			}
		})
	}
}

func TestTier1_Standalone_Help(t *testing.T) {
	t.Parallel()

	flags := []string{"-h", "--help", "-help"}
	for _, flag := range flags {
		flag := flag
		t.Run("Flag_"+flag, func(t *testing.T) {
			t.Parallel()

			stdout, stderr, exitCode := runCLI(t, []string{flag}, nil)

			if exitCode != 0 {
				t.Fatalf("expected exit code 0 for help flag %s, got %d", flag, exitCode)
			}

			requiredSections := []string{
				"antigravity-cli-statusline",
				"USAGE:",
				"FLAGS:",
				"ENVIRONMENT VARIABLES:",
				"LAYOUT MODES:",
				"EXAMPLES:",
				"EXIT CODES:",
			}

			for _, sec := range requiredSections {
				if !strings.Contains(stdout, sec) {
					t.Errorf("help output missing section %q, got:\n%s", sec, stdout)
				}
			}

			if stderr != "" {
				t.Errorf("expected empty stderr for help, got: %s", stderr)
			}
		})
	}
}

func TestTier1_Standalone_Version(t *testing.T) {
	t.Parallel()

	flags := []string{"-v", "--version", "-version"}
	for _, flag := range flags {
		flag := flag
		t.Run("Flag_"+flag, func(t *testing.T) {
			t.Parallel()

			stdout, stderr, exitCode := runCLI(t, []string{flag}, nil)

			if exitCode != 0 {
				t.Fatalf("expected exit code 0 for version flag %s, got %d", flag, exitCode)
			}

			expectedVersion := "antigravity-cli-statusline v1.0.0 (Go 1.22+)"
			if strings.TrimSpace(stdout) != expectedVersion {
				t.Errorf("expected version %q, got %q", expectedVersion, strings.TrimSpace(stdout))
			}

			if stderr != "" {
				t.Errorf("expected empty stderr for version, got: %s", stderr)
			}
		})
	}
}

func TestTier1_Standalone_Preview(t *testing.T) {
	t.Parallel()

	flags := []string{"-p", "--preview", "-preview"}
	for _, flag := range flags {
		flag := flag
		t.Run("Flag_"+flag, func(t *testing.T) {
			t.Parallel()

			stdout, stderr, exitCode := runCLI(t, []string{flag}, nil)

			if exitCode != 0 {
				t.Fatalf("expected exit code 0 for preview flag %s, got %d", flag, exitCode)
			}

			expectedKeywords := []string{
				"Gemini 3.7 Pro",
				"feature/statusline-v2",
				"WORKING",
				"PLAN",
			}

			for _, kw := range expectedKeywords {
				if !strings.Contains(stdout, kw) {
					t.Errorf("preview output missing keyword %q, got:\n%s", kw, stdout)
				}
			}

			if stderr != "" {
				t.Errorf("expected empty stderr for preview, got: %s", stderr)
			}
		})
	}
}

func TestTier1_Standalone_Config(t *testing.T) {
	t.Parallel()

	fullJSON := getTestDataPath(t, "payload_full.json")
	minJSON := getTestDataPath(t, "payload_minimal.json")

	tests := []struct {
		name         string
		args         []string
		expectedSubs []string
	}{
		{
			name:         "LongFlagConfigFull",
			args:         []string{"--config", fullJSON},
			expectedSubs: []string{"Gemini 3.7 Flash", "feat/statusline-modernize"},
		},
		{
			name:         "ShortFlagConfigMinimal",
			args:         []string{"-c", minJSON},
			expectedSubs: []string{"gemini-3.7-flash", "READY"},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			stdout, stderr, exitCode := runCLI(t, tc.args, nil)

			if exitCode != 0 {
				t.Fatalf("expected exit code 0, got %d (stderr: %s)", exitCode, stderr)
			}

			for _, sub := range tc.expectedSubs {
				if !strings.Contains(stdout, sub) {
					t.Errorf("expected stdout to contain %q, got:\n%s", sub, stdout)
				}
			}

			if stderr != "" {
				t.Errorf("expected empty stderr, got: %s", stderr)
			}
		})
	}
}

func TestTier1_WidthOverride_Layouts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		width        string
		expectSep    string
		expectFrame  string
		rejectFrame  string
		expectedMode string
	}{
		{
			name:         "Compact_Width60",
			width:        "60",
			rejectFrame:  "╭─",
			expectedMode: "Compact (<80)",
		},
		{
			name:         "CompactBoundary_Width79",
			width:        "79",
			rejectFrame:  "╭─",
			expectedMode: "Compact (<80)",
		},
		{
			name:         "StandardBoundary_Width80",
			width:        "80",
			expectFrame:  "╭─",
			expectedMode: "Standard (80-119)",
		},
		{
			name:         "Standard_Width100",
			width:        "100",
			expectFrame:  "╭─",
			expectedMode: "Standard (80-119)",
		},
		{
			name:         "StandardBoundary_Width119",
			width:        "119",
			expectFrame:  "╭─",
			expectedMode: "Standard (80-119)",
		},
		{
			name:         "WideBoundary_Width120",
			width:        "120",
			expectSep:    theme.SepPipe,
			rejectFrame:  "╭─",
			expectedMode: "Wide (>=120)",
		},
		{
			name:         "Wide_Width160",
			width:        "160",
			expectSep:    theme.SepPipe,
			rejectFrame:  "╭─",
			expectedMode: "Wide (>=120)",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			stdout, stderr, exitCode := runCLI(t, []string{"--preview", "-w", tc.width}, nil)

			if exitCode != 0 {
				t.Fatalf("expected exit code 0, got %d (stderr: %s)", exitCode, stderr)
			}

			if tc.expectFrame != "" && !strings.Contains(stdout, tc.expectFrame) {
				t.Errorf("[%s] expected frame %q in layout, got:\n%s", tc.expectedMode, tc.expectFrame, stdout)
			}
			if tc.rejectFrame != "" && strings.Contains(stdout, tc.rejectFrame) {
				t.Errorf("[%s] did NOT expect frame %q in layout, got:\n%s", tc.expectedMode, tc.rejectFrame, stdout)
			}
			if tc.expectSep != "" && !strings.Contains(stdout, tc.expectSep) {
				t.Errorf("[%s] expected separator %q in layout, got:\n%s", tc.expectedMode, tc.expectSep, stdout)
			}

			if stderr != "" {
				t.Errorf("expected empty stderr, got: %s", stderr)
			}
		})
	}
}

func TestTier1_Theme_Options(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		theme        string
		expectANSI   bool
		expectedExit int
	}{
		{name: "ThemeDefault", theme: "default", expectANSI: true, expectedExit: 0},
		{name: "ThemeNoColor", theme: "no-color", expectANSI: false, expectedExit: 0},
		{name: "ThemePlain", theme: "plain", expectANSI: false, expectedExit: 0},
		{name: "ThemeClassic", theme: "classic", expectANSI: false, expectedExit: 0},
		{name: "ThemeASCII", theme: "ascii", expectANSI: false, expectedExit: 0},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			stdout, stderr, exitCode := runCLI(t, []string{"--preview", "--theme", tc.theme}, nil)

			if exitCode != tc.expectedExit {
				t.Fatalf("expected exit code %d, got %d (stderr: %s)", tc.expectedExit, exitCode, stderr)
			}

			hasColor := containsANSI(stdout)
			if tc.expectANSI && !hasColor {
				t.Errorf("theme %s: expected ANSI colors, but found none in output:\n%s", tc.theme, stdout)
			}
			if !tc.expectANSI && hasColor {
				t.Errorf("theme %s: expected NO ANSI colors, but found ANSI escape codes:\n%s", tc.theme, stdout)
			}

			if stderr != "" {
				t.Errorf("expected empty stderr, got: %s", stderr)
			}
		})
	}
}

// =============================================================================
// TIER 2: BOUNDARY & CORNER CASES
// =============================================================================

func TestTier2_Boundary_EmptyInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
	}{
		{name: "EmptyZeroBytes", input: ""},
		{name: "SingleNewline", input: "\n"},
		{name: "MultipleWhitespaces", input: "   \t  \n  \r\n   "},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			stdout, stderr, exitCode := runCLI(t, nil, strings.NewReader(tc.input))

			if exitCode != 0 {
				t.Fatalf("expected exit code 0 for empty input, got %d (stderr: %s)", exitCode, stderr)
			}

			// Empty input should gracefully render default statusline
			if !strings.Contains(stdout, "READY") {
				t.Errorf("expected fallback default statusline to contain 'READY', got:\n%s", stdout)
			}

			if stderr != "" {
				t.Errorf("expected empty stderr, got: %s", stderr)
			}
		})
	}
}

func TestTier2_Boundary_MalformedAndCorruptedJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
	}{
		{name: "CorruptedFixture", input: `{"cwd": "/workspace", "model": {"id": "gemini", incomplete`},
		{name: "UnclosedObject", input: `{"model": {"id": "gemini"`},
		{name: "TruncatedString", input: `{"model": "gemini`},
		{name: "BinaryGarbage", input: string([]byte{0x00, 0xFF, 0xFE, 0x12, 0x88, 0x77})},
		{name: "ScalarJSONNumber", input: `12345`},
		{name: "ScalarJSONString", input: `"simple string"`},
		{name: "JSONArray", input: `[1, 2, 3]`},
		{name: "JSONBoolean", input: `true`},
		{name: "JSONNull", input: `null`},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			stdout, stderr, exitCode := runCLI(t, nil, strings.NewReader(tc.input))

			// In pipeline mode, invalid JSON must NOT crash or panic and gracefully fall back to ready state
			if exitCode != 0 {
				t.Fatalf("expected graceful exit code 0 for malformed input, got %d (stderr: %s)", exitCode, stderr)
			}

			if !strings.Contains(stdout, "READY") {
				t.Errorf("expected fallback statusline for malformed JSON, got:\n%s", stdout)
			}
		})
	}
}

func TestTier2_Boundary_OversizedInput(t *testing.T) {
	t.Parallel()

	t.Run("PipedOversizedPayload_BoundedGracefulFallback", func(t *testing.T) {
		t.Parallel()

		// Generate 2.5MB payload (exceeding 2MB bounded limit)
		const payloadSize = 2500 * 1024
		oversizedBytes := make([]byte, payloadSize)
		for i := range oversizedBytes {
			oversizedBytes[i] = 'A'
		}

		stdout, stderr, exitCode := runCLI(t, nil, bytes.NewReader(oversizedBytes))

		// In pipeline mode, bounded reader cuts off and falls back safely to default statusline
		if exitCode != 0 {
			t.Fatalf("expected exit code 0 on oversized piped input, got %d (stderr: %s)", exitCode, stderr)
		}

		if !strings.Contains(stdout, "READY") {
			t.Errorf("expected fallback statusline on oversized input, got:\n%s", stdout)
		}
	})

	t.Run("ConfigOversizedFile_ReturnsExitError", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		hugeFile := filepath.Join(tmpDir, "huge_payload.json")

		// Create a file > 2MB
		hugeData := bytes.Repeat([]byte(`{"model":{"id":"gemini"}},`), 100000)
		if err := os.WriteFile(hugeFile, hugeData, 0o600); err != nil {
			t.Fatalf("failed to write huge file: %v", err)
		}

		_, stderr, exitCode := runCLI(t, []string{"--config", hugeFile}, nil)

		if exitCode != 1 {
			t.Fatalf("expected exit code 1 for oversized config file, got %d", exitCode)
		}

		if !strings.Contains(stderr, "failed to parse config file") && !strings.Contains(stderr, "exceeds") {
			t.Errorf("expected error message mentioning parse failure/size limit, got: %s", stderr)
		}
	})
}

func TestTier2_Boundary_TerminalWidths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		widthArg      string
		expectedExit  int
		expectErrText string
	}{
		{name: "ZeroWidth", widthArg: "0", expectedExit: 0},
		{name: "NegativeOneWidth", widthArg: "-1", expectedExit: 1, expectErrText: "invalid width-override -1"},
		{name: "NegativeTenWidth", widthArg: "-10", expectedExit: 1, expectErrText: "invalid width-override -10"},
		{name: "ExtremeNarrowWidth", widthArg: "1", expectedExit: 0},
		{name: "ExtremeWideWidth", widthArg: "500", expectedExit: 0},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			stdout, stderr, exitCode := runCLI(t, []string{"--preview", "-w", tc.widthArg}, nil)

			if exitCode != tc.expectedExit {
				t.Fatalf("expected exit code %d, got %d (stderr: %s)", tc.expectedExit, exitCode, stderr)
			}

			if tc.expectErrText != "" {
				if !strings.Contains(stderr, tc.expectErrText) {
					t.Errorf("expected stderr to contain %q, got: %s", tc.expectErrText, stderr)
				}
			} else {
				if !strings.Contains(stdout, "Gemini 3.7 Pro") {
					t.Errorf("expected preview output on valid width, got:\n%s", stdout)
				}
			}
		})
	}
}

func TestTier2_Boundary_UnicodeAndEmojis(t *testing.T) {
	t.Parallel()

	unicodeJSON := getTestDataPath(t, "payload_unicode.json")
	content, err := os.ReadFile(unicodeJSON)
	if err != nil {
		t.Fatalf("failed reading unicode fixture: %v", err)
	}

	stdout, stderr, exitCode := runCLI(t, nil, bytes.NewReader(content))

	if exitCode != 0 {
		t.Fatalf("expected exit code 0 for unicode payload, got %d (stderr: %s)", exitCode, stderr)
	}

	expectedEntities := []string{
		"Gemini 3.7 Pro (日本語・UTF8 🚀✨)",
		"feat/🌿-feature-äöü-🚀-branch",
		"INSERT",
	}

	for _, entity := range expectedEntities {
		if !strings.Contains(stdout, entity) {
			t.Errorf("expected output to preserve unicode entity %q, got:\n%s", entity, stdout)
		}
	}

	if stderr != "" {
		t.Errorf("expected empty stderr, got: %s", stderr)
	}
}

func TestTier2_Boundary_ANSIInjectionSanitization(t *testing.T) {
	t.Parallel()

	maliciousJSON := getTestDataPath(t, "payload_malicious.json")
	content, err := os.ReadFile(maliciousJSON)
	if err != nil {
		t.Fatalf("failed reading malicious fixture: %v", err)
	}

	stdout, stderr, exitCode := runCLI(t, nil, bytes.NewReader(content))

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d (stderr: %s)", exitCode, stderr)
	}

	// Verify that injection strings inside payload fields were stripped
	dangerousSequences := []string{
		"\u001b[2J",            // Clear screen
		"\u001b[H",             // Cursor home
		"\u001b[?25l",          // Hide cursor
		"\u001b]0;TitleHack\a", // OSC terminal title hack
		"\u0000",               // Null byte
		"PwnedTerminal",
	}

	for _, seq := range dangerousSequences {
		if strings.Contains(stdout, seq) {
			t.Errorf("SECURITY: output contains unsanitized dangerous sequence %q in output:\n%s", seq, stdout)
		}
	}

	// Model name should be cleanly rendered without embedded control characters
	if !strings.Contains(stdout, "Gemini 3.7 Flash") {
		t.Errorf("expected sanitized model name 'Gemini 3.7 Flash', got:\n%s", stdout)
	}

	if stderr != "" {
		t.Errorf("expected empty stderr, got: %s", stderr)
	}
}

func TestTier2_Boundary_UnknownFlagsAndNonExistentConfig(t *testing.T) {
	t.Parallel()

	corruptedJSON := getTestDataPath(t, "payload_corrupted.json")

	tests := []struct {
		name          string
		args          []string
		expectedExit  int
		expectErrText string
	}{
		{
			name:          "UnknownFlag",
			args:          []string{"--invalid-flag-123"},
			expectedExit:  1,
			expectErrText: "unknown flag",
		},
		{
			name:          "UnknownTheme",
			args:          []string{"--preview", "--theme", "neon_cyberpunk"},
			expectedExit:  1,
			expectErrText: "unknown theme",
		},
		{
			name:          "NonExistentConfigFile",
			args:          []string{"--config", "/path/to/nonexistent/statusline_file_123.json"},
			expectedExit:  1,
			expectErrText: "failed to open config file",
		},
		{
			name:          "CorruptedConfigFile_FatalError",
			args:          []string{"--config", corruptedJSON},
			expectedExit:  1,
			expectErrText: "failed to parse config file",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, stderr, exitCode := runCLI(t, tc.args, nil)

			if exitCode != tc.expectedExit {
				t.Fatalf("expected exit code %d, got %d (stderr: %s)", tc.expectedExit, exitCode, stderr)
			}

			if !strings.Contains(stderr, tc.expectErrText) {
				t.Errorf("expected stderr to contain %q, got: %s", tc.expectErrText, stderr)
			}
		})
	}
}

// =============================================================================
// TIER 3: CROSS-FEATURE COMBINATIONS
// =============================================================================

func TestTier3_CrossFeature_NoColorEnvironment(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    []string
		envVar  string
		envVal  string
		payload string
	}{
		{
			name:   "NO_COLOR_WithPreview",
			args:   []string{"--preview"},
			envVar: "NO_COLOR",
			envVal: "1",
		},
		{
			name:   "NO_COLOR_WithFullConfig",
			args:   []string{"--config", getTestDataPath(t, "payload_full.json")},
			envVar: "NO_COLOR",
			envVal: "true",
		},
		{
			name:   "TERM_DUMB_WithPreview",
			args:   []string{"--preview"},
			envVar: "TERM",
			envVal: "dumb",
		},
		{
			name:    "TERM_DUMB_WithPipedJSON",
			args:    nil,
			envVar:  "TERM",
			envVal:  "dumb",
			payload: `{"agent_state":"working","model":{"display_name":"Gemini 3.7 Flash"}}`,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var stdin io.Reader
			if tc.payload != "" {
				stdin = strings.NewReader(tc.payload)
			}

			stdout, stderr, exitCode := runCLI(t, tc.args, stdin, tc.envVar+"="+tc.envVal)

			if exitCode != 0 {
				t.Fatalf("expected exit code 0, got %d (stderr: %s)", exitCode, stderr)
			}

			if containsANSI(stdout) {
				t.Errorf("environment %s=%s: output unexpectedly contains ANSI escape codes:\n%s", tc.envVar, tc.envVal, stdout)
			}

			if stderr != "" {
				t.Errorf("expected empty stderr, got: %s", stderr)
			}
		})
	}
}

func TestTier3_CrossFeature_ConfigWidthAndThemeCombination(t *testing.T) {
	t.Parallel()

	fullJSON := getTestDataPath(t, "payload_full.json")

	tests := []struct {
		name         string
		args         []string
		expectFrame  string
		rejectFrame  string
		expectSep    string
		expectANSI   bool
		expectedSubs []string
	}{
		{
			name:         "ConfigFull_Width60_ThemeNoColor",
			args:         []string{"--config", fullJSON, "-w", "60", "--theme", "no-color"},
			rejectFrame:  "╭─",
			expectANSI:   false,
			expectedSubs: []string{"Gemini 3.7 Flash", "NORMAL"},
		},
		{
			name:         "ConfigFull_Width100_ThemeClassic",
			args:         []string{"-c", fullJSON, "-w", "100", "--theme", "classic"},
			expectFrame:  "╭─",
			expectANSI:   false,
			expectedSubs: []string{"Gemini 3.7 Flash", "feat/statusline-modernize"},
		},
		{
			name:         "ConfigFull_Width140_ThemeDefault",
			args:         []string{"--config", fullJSON, "--width-override", "140", "--theme", "default"},
			expectSep:    theme.SepPipe,
			rejectFrame:  "╭─",
			expectANSI:   true,
			expectedSubs: []string{"Gemini 3.7 Flash", "25.0%"},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			stdout, stderr, exitCode := runCLI(t, tc.args, nil)

			if exitCode != 0 {
				t.Fatalf("expected exit code 0, got %d (stderr: %s)", exitCode, stderr)
			}

			for _, sub := range tc.expectedSubs {
				if !strings.Contains(stdout, sub) {
					t.Errorf("expected output to contain %q, got:\n%s", sub, stdout)
				}
			}

			if tc.expectFrame != "" && !strings.Contains(stdout, tc.expectFrame) {
				t.Errorf("expected frame %q in output, got:\n%s", tc.expectFrame, stdout)
			}
			if tc.rejectFrame != "" && strings.Contains(stdout, tc.rejectFrame) {
				t.Errorf("did not expect frame %q in output, got:\n%s", tc.rejectFrame, stdout)
			}
			if tc.expectSep != "" && !strings.Contains(stdout, tc.expectSep) {
				t.Errorf("expected separator %q in output, got:\n%s", tc.expectSep, stdout)
			}

			hasColor := containsANSI(stdout)
			if tc.expectANSI && !hasColor {
				t.Errorf("expected ANSI colors in output, got none")
			}
			if !tc.expectANSI && hasColor {
				t.Errorf("expected NO ANSI colors in output, got ANSI codes")
			}

			if stderr != "" {
				t.Errorf("expected empty stderr, got: %s", stderr)
			}
		})
	}
}

func TestTier3_CrossFeature_PipedStdinWithDebug(t *testing.T) {
	t.Parallel()

	corruptedData := `{"agent_state": "working", unclosed_key:`
	stdout, stderr, exitCode := runCLI(t, []string{"--debug"}, strings.NewReader(corruptedData))

	if exitCode != 0 {
		t.Fatalf("expected exit code 0 in pipeline mode, got %d", exitCode)
	}

	if !strings.Contains(stdout, "READY") {
		t.Errorf("expected fallback statusline in stdout, got:\n%s", stdout)
	}

	if !strings.Contains(stderr, "debug: stdin payload parse warning") {
		t.Errorf("expected debug log in stderr, got: %s", stderr)
	}
}

func TestTier3_CrossFeature_ConfigOverridesStdinPrecedence(t *testing.T) {
	t.Parallel()

	fullJSON := getTestDataPath(t, "payload_full.json")
	pipedStdin := `{"agent_state":"idle","model":{"display_name":"Conflicting Stdin Model"}}`

	// When --config is provided, the file should take precedence over stdin
	stdout, stderr, exitCode := runCLI(t, []string{"--config", fullJSON}, strings.NewReader(pipedStdin))

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d (stderr: %s)", exitCode, stderr)
	}

	if !strings.Contains(stdout, "Gemini 3.7 Flash") {
		t.Errorf("expected config file model 'Gemini 3.7 Flash' in output, got:\n%s", stdout)
	}

	if strings.Contains(stdout, "Conflicting Stdin Model") {
		t.Errorf("did NOT expect stdin model to override config file, but found in output:\n%s", stdout)
	}

	if stderr != "" {
		t.Errorf("expected empty stderr, got: %s", stderr)
	}
}

// =============================================================================
// TIER 4: REAL-WORLD APPLICATION WORKLOADS
// =============================================================================

func TestTier4_Workload_FullAGYSessionPayload(t *testing.T) {
	t.Parallel()

	fixturePath := getTestDataPath(t, "payload_agy_workload.json")
	content, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("failed to read workload fixture: %v", err)
	}

	layouts := []struct {
		name         string
		widthArg     string
		expectFrame  bool
		expectSep    bool
		expectedSubs []string
	}{
		{
			name:        "WideView",
			widthArg:    "140",
			expectFrame: false,
			expectSep:   true,
			expectedSubs: []string{
				"Gemini 3.7 Pro",
				"feat/m4-e2e-testing",
				"VISUAL",
				"81.5%",
			},
		},
		{
			name:        "StandardView",
			widthArg:    "100",
			expectFrame: true,
			expectSep:   false,
			expectedSubs: []string{
				"Gemini 3.7 Pro",
				"feat/m4-e2e-testing",
				"VISUAL",
				"81.5%",
			},
		},
		{
			name:        "CompactView",
			widthArg:    "65",
			expectFrame: false,
			expectSep:   false,
			expectedSubs: []string{
				"Gemini 3.7 Pro",
				"VISUAL",
				"WORKING",
				"81.5%",
			},
		},
	}

	for _, ly := range layouts {
		ly := ly
		t.Run(ly.name, func(t *testing.T) {
			t.Parallel()

			stdout, stderr, exitCode := runCLI(t, []string{"-w", ly.widthArg}, bytes.NewReader(content))

			if exitCode != 0 {
				t.Fatalf("expected exit code 0, got %d (stderr: %s)", exitCode, stderr)
			}

			for _, attr := range ly.expectedSubs {
				if !strings.Contains(stdout, attr) {
					t.Errorf("workload layout %s missing attribute %q, got:\n%s", ly.name, attr, stdout)
				}
			}

			if ly.expectFrame && !strings.Contains(stdout, "╭─") {
				t.Errorf("expected frame borders in %s, got:\n%s", ly.name, stdout)
			}
			if ly.expectSep && !strings.Contains(stdout, theme.SepPipe) {
				t.Errorf("expected pipe separator in %s, got:\n%s", ly.name, stdout)
			}

			if stderr != "" {
				t.Errorf("expected empty stderr, got: %s", stderr)
			}
		})
	}
}

func TestTier4_Workload_RapidStreamingPipelineSimulation(t *testing.T) {
	t.Parallel()

	// Simulate 100 sequential statusline renders with evolving token counts
	const iterations = 100
	start := time.Now()

	for i := 0; i < iterations; i++ {
		tokens := 1000 + i*500
		usedPct := float64(tokens) / 200000.0 * 100.0
		payload := fmt.Sprintf(`{
			"agent_state": "working",
			"model": {"display_name": "Gemini 3.7 Pro", "id": "gemini-3.7-pro"},
			"context_window": {
				"total_input_tokens": %d,
				"total_output_tokens": %d,
				"context_window_size": 200000,
				"used_percentage": %.1f
			},
			"terminal_width": 120
		}`, tokens, i*50, usedPct)

		stdout, stderr, exitCode := runCLI(t, nil, strings.NewReader(payload))

		if exitCode != 0 {
			t.Fatalf("iteration %d failed with exit code %d (stderr: %s)", i, exitCode, stderr)
		}

		if !strings.Contains(stdout, "Gemini 3.7 Pro") {
			t.Fatalf("iteration %d missing model in output: %s", i, stdout)
		}

		if stderr != "" {
			t.Fatalf("iteration %d had non-empty stderr: %s", i, stderr)
		}
	}

	elapsed := time.Since(start)
	t.Logf("completed %d sequential pipeline invocations in %v (avg %.2f ms/run)",
		iterations, elapsed, float64(elapsed.Milliseconds())/float64(iterations))
}

func TestTier4_Workload_ConcurrentMultiProcessExecution(t *testing.T) {
	t.Parallel()

	const concurrentWorkers = 20
	var wg sync.WaitGroup
	wg.Add(concurrentWorkers)

	fullJSON := getTestDataPath(t, "payload_full.json")
	fullData, err := os.ReadFile(fullJSON)
	if err != nil {
		t.Fatalf("failed reading full fixture: %v", err)
	}

	for i := 0; i < concurrentWorkers; i++ {
		workerID := i
		go func(id int) {
			defer wg.Done()

			var args []string
			var stdin io.Reader

			switch id % 4 {
			case 0:
				args = []string{"--preview", "-w", "140"}
			case 1:
				args = []string{"--config", fullJSON, "--theme", "no-color"}
			case 2:
				args = []string{"-w", "80"}
				stdin = bytes.NewReader(fullData)
			case 3:
				args = []string{"--version"}
			}

			stdout, stderr, exitCode := runCLI(t, args, stdin)

			if exitCode != 0 {
				t.Errorf("worker %d failed with exit code %d (stderr: %s)", id, exitCode, stderr)
			}
			if len(stdout) == 0 {
				t.Errorf("worker %d produced empty stdout", id)
			}
		}(workerID)
	}

	wg.Wait()
}
