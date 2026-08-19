package cli_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/joaco/antigravity-statusline/internal/cli"
)

func TestInstallCommand(t *testing.T) {
	tmpDir := t.TempDir()
	customDest := filepath.Join(tmpDir, "bin", "statusline")
	customSettings := filepath.Join(tmpDir, "settings.json")

	// Pre-create dummy settings
	initSettings := map[string]any{
		"theme": "dark",
	}
	setData, _ := json.Marshal(initSettings)
	if err := os.WriteFile(customSettings, setData, 0644); err != nil {
		t.Fatalf("failed to write dummy settings: %v", err)
	}

	var stdout, stderr bytes.Buffer
	in := bytes.NewReader(nil)

	args := []string{"install", "--dest", customDest, "--skip-settings"}
	exitCode := cli.Run(args, in, &stdout, &stderr)

	if exitCode != cli.ExitSuccess {
		t.Fatalf("expected exit code %d, got %d (stderr: %s)", cli.ExitSuccess, exitCode, stderr.String())
	}

	// Verify installed binary exists
	if _, err := os.Stat(customDest); err != nil {
		t.Errorf("expected installed binary to exist at %s: %v", customDest, err)
	}

	out := stdout.String()
	if !strings.Contains(out, "Installing statusline to:") {
		t.Errorf("expected install confirmation in output, got: %s", out)
	}
}
