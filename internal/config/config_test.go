package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/joaco/antigravity-statusline/internal/config"
)

func TestDefaultConfig(t *testing.T) {
	cfg := config.DefaultConfig()
	if cfg == nil {
		t.Fatal("expected non-nil default config")
	}

	if cfg.Theme.Palette != "default" {
		t.Errorf("expected default palette, got %s", cfg.Theme.Palette)
	}
	if !cfg.Modules.GitBranch {
		t.Error("expected GitBranch to be true by default")
	}
	if cfg.Layout.WideBreakpoint != 120 {
		t.Errorf("expected wide breakpoint 120, got %d", cfg.Layout.WideBreakpoint)
	}
}

func TestLoad_CustomConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "statusline.json")

	content := `{
		"theme": {
			"palette": "tokyonight"
		},
		"modules": {
			"sandbox_badge": false,
			"vim_mode": false
		},
		"layout": {
			"wide_breakpoint": 140
		}
	}`

	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("failed to load custom config: %v", err)
	}

	if cfg.Theme.Palette != "tokyonight" {
		t.Errorf("expected palette tokyonight, got %s", cfg.Theme.Palette)
	}
	if cfg.Modules.SandboxBadge {
		t.Error("expected SandboxBadge to be false")
	}
	if cfg.Modules.VimMode {
		t.Error("expected VimMode to be false")
	}
	if cfg.Layout.WideBreakpoint != 140 {
		t.Errorf("expected wide breakpoint 140, got %d", cfg.Layout.WideBreakpoint)
	}
}

func TestInitConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	dest := filepath.Join(tmpDir, "sub", "config.json")

	createdPath, err := config.InitConfigFile(dest)
	if err != nil {
		t.Fatalf("failed to init config file: %v", err)
	}

	if createdPath != dest {
		t.Errorf("expected created path %s, got %s", dest, createdPath)
	}

	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("failed to read created config file: %v", err)
	}

	if !strings.Contains(string(data), "nerd-fonts") {
		t.Error("expected created config to contain default theme settings")
	}
}
