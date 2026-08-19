package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrConfigTooLarge = errors.New("configuration file exceeds maximum size limit (1MB)")
)

const MaxConfigFileSize = 1024 * 1024

type IconsConfig struct {
	Idle      string `json:"idle,omitempty"`
	Thinking  string `json:"thinking,omitempty"`
	Working   string `json:"working,omitempty"`
	Tool      string `json:"tool,omitempty"`
	Init      string `json:"init,omitempty"`
	Branch    string `json:"branch,omitempty"`
	Dirty     string `json:"dirty,omitempty"`
	Model     string `json:"model,omitempty"`
	Context   string `json:"context,omitempty"`
	Artifacts string `json:"artifacts,omitempty"`
	Subagents string `json:"subagents,omitempty"`
	Tasks     string `json:"tasks,omitempty"`
	Sandbox   string `json:"sandbox,omitempty"`
	Vim       string `json:"vim,omitempty"`
	Weekly    string `json:"weekly,omitempty"`
	Session   string `json:"session,omitempty"`
}

type ThemeConfig struct {
	Palette string      `json:"palette"`
	IconSet string      `json:"icon_set"`
	Icons   IconsConfig `json:"icons"`
}

type ModulesConfig struct {
	VimMode        bool `json:"vim_mode"`
	AgentState     bool `json:"agent_state"`
	ModelName      bool `json:"model_name"`
	GitBranch      bool `json:"git_branch"`
	ContextBar     bool `json:"context_bar"`
	ArtifactsCount bool `json:"artifacts_count"`
	SubagentsCount bool `json:"subagents_count"`
	TasksCount     bool `json:"tasks_count"`
	SandboxBadge   bool `json:"sandbox_badge"`
	SessionUsage   bool `json:"session_usage"`
	WeeklyUsage    bool `json:"weekly_usage"`
}

type LayoutConfig struct {
	Style              string `json:"style"`
	Separator          string `json:"separator"`
	MetricsSeparator   string `json:"metrics_separator"`
	FrameTop           string `json:"frame_top"`
	FrameBottom        string `json:"frame_bottom"`
	WideBreakpoint     int    `json:"wide_breakpoint"`
	StandardBreakpoint int    `json:"standard_breakpoint"`
}

type ContextBarConfig struct {
	Length            int     `json:"length"`
	Style             string  `json:"style"`
	WarningThreshold  float64 `json:"warning_threshold"`
	CriticalThreshold float64 `json:"critical_threshold"`
}

type Config struct {
	Theme      ThemeConfig      `json:"theme"`
	Modules    ModulesConfig    `json:"modules"`
	Layout     LayoutConfig     `json:"layout"`
	ContextBar ContextBarConfig `json:"context_bar"`
}

func DefaultConfig() *Config {
	return &Config{
		Theme: ThemeConfig{
			Palette: "default",
			IconSet: "nerd-fonts",
			Icons: IconsConfig{
				Idle:      "󰚩",
				Thinking:  "󰘦",
				Working:   "󱐋",
				Tool:      "󰘳",
				Init:      "󰑮",
				Branch:    "",
				Dirty:     "●",
				Model:     "󰧑",
				Context:   "󰍛",
				Artifacts: "󰈙",
				Subagents: "󰮝",
				Tasks:     "󰒋",
				Sandbox:   "󰌾",
				Vim:       "",
				Weekly:    "󰃭",
				Session:   "󰥔",
			},
		},
		Modules: ModulesConfig{
			VimMode:        true,
			AgentState:     true,
			ModelName:      true,
			GitBranch:      true,
			ContextBar:     true,
			ArtifactsCount: true,
			SubagentsCount: true,
			TasksCount:     true,
			SandboxBadge:   true,
			SessionUsage:   true,
			WeeklyUsage:    true,
		},
		Layout: LayoutConfig{
			Style:              "framed",
			Separator:          " ╱ ",
			MetricsSeparator:   " · ",
			FrameTop:           "╭─ ",
			FrameBottom:        "╰─ ",
			WideBreakpoint:     120,
			StandardBreakpoint: 80,
		},
		ContextBar: ContextBarConfig{
			Length:            10,
			Style:             "smooth",
			WarningThreshold:  75.0,
			CriticalThreshold: 90.0,
		},
	}
}

func DefaultConfigPaths() []string {
	var paths []string

	if envPath := os.Getenv("STATUSLINE_CONFIG"); envPath != "" {
		paths = append(paths, envPath)
	}

	if configDir, err := os.UserConfigDir(); err == nil {
		paths = append(paths, filepath.Join(configDir, "antigravity", "statusline.json"))
	}

	if homeDir, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(homeDir, ".gemini", "antigravity-cli", "statusline.json"))
	}

	return paths
}

func Load(explicitPath string) (*Config, error) {
	cfg := DefaultConfig()

	targetPath := explicitPath
	if targetPath == "" {
		for _, candidate := range DefaultConfigPaths() {
			if _, err := os.Stat(candidate); err == nil {
				targetPath = candidate
				break
			}
		}
	}

	if targetPath == "" {
		return cfg, nil
	}

	info, err := os.Stat(targetPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) && explicitPath == "" {
			return cfg, nil
		}
		return nil, err
	}

	if info.Size() > MaxConfigFileSize {
		return nil, ErrConfigTooLarge
	}

	data, err := os.ReadFile(targetPath)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse configuration file %s: %w", targetPath, err)
	}

	cfg.validateAndSanitize()
	return cfg, nil
}

func (c *Config) validateAndSanitize() {
	if c.Layout.WideBreakpoint <= 0 {
		c.Layout.WideBreakpoint = 120
	}
	if c.Layout.StandardBreakpoint <= 0 {
		c.Layout.StandardBreakpoint = 80
	}
	if c.ContextBar.Length <= 0 {
		c.ContextBar.Length = 10
	}
	if c.ContextBar.WarningThreshold <= 0 {
		c.ContextBar.WarningThreshold = 75.0
	}
	if c.ContextBar.CriticalThreshold <= 0 {
		c.ContextBar.CriticalThreshold = 90.0
	}
}

func InitConfigFile(destPath string) (string, error) {
	if destPath == "" {
		destPath = FindActiveConfigPath()
	}
	cfg := DefaultConfig()
	if err := Save(destPath, cfg); err != nil {
		return "", err
	}
	return destPath, nil
}

func Save(destPath string, cfg *Config) error {
	if destPath == "" {
		destPath = FindActiveConfigPath()
	}
	dir := filepath.Dir(destPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory %s: %w", dir, err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(destPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file %s: %w", destPath, err)
	}
	return nil
}

func FindActiveConfigPath() string {
	for _, candidate := range DefaultConfigPaths() {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	configDir, err := os.UserConfigDir()
	if err != nil {
		homeDir, hErr := os.UserHomeDir()
		if hErr != nil {
			return filepath.Join(".config", "antigravity", "statusline.json")
		}
		configDir = filepath.Join(homeDir, ".config")
	}
	return filepath.Join(configDir, "antigravity", "statusline.json")
}

func SetProperty(targetPath string, key string, val string) (string, error) {
	if targetPath == "" {
		targetPath = FindActiveConfigPath()
	}

	cfg, err := Load(targetPath)
	if err != nil {
		return "", fmt.Errorf("failed to load existing config: %w", err)
	}

	keyLower := strings.ToLower(strings.TrimSpace(key))
	switch keyLower {
	case "theme.palette", "palette":
		cfg.Theme.Palette = val
	case "theme.icon_set", "icon_set":
		cfg.Theme.IconSet = val
	case "theme.icons.branch", "icons.branch":
		cfg.Theme.Icons.Branch = val
	case "theme.icons.dirty", "icons.dirty":
		cfg.Theme.Icons.Dirty = val
	case "theme.icons.idle", "icons.idle":
		cfg.Theme.Icons.Idle = val
	case "theme.icons.thinking", "icons.thinking":
		cfg.Theme.Icons.Thinking = val
	case "theme.icons.working", "icons.working":
		cfg.Theme.Icons.Working = val
	case "theme.icons.tool", "icons.tool":
		cfg.Theme.Icons.Tool = val
	case "theme.icons.model", "icons.model":
		cfg.Theme.Icons.Model = val
	case "theme.icons.context", "icons.context":
		cfg.Theme.Icons.Context = val
	case "theme.icons.artifacts", "icons.artifacts":
		cfg.Theme.Icons.Artifacts = val
	case "theme.icons.subagents", "icons.subagents":
		cfg.Theme.Icons.Subagents = val
	case "theme.icons.tasks", "icons.tasks":
		cfg.Theme.Icons.Tasks = val
	case "theme.icons.sandbox", "icons.sandbox":
		cfg.Theme.Icons.Sandbox = val
	case "theme.icons.vim", "icons.vim":
		cfg.Theme.Icons.Vim = val
	case "theme.icons.weekly", "icons.weekly":
		cfg.Theme.Icons.Weekly = val
	case "theme.icons.session", "icons.session":
		cfg.Theme.Icons.Session = val

	case "modules.vim_mode", "vim_mode":
		cfg.Modules.VimMode = parseBool(val)
	case "modules.agent_state", "agent_state":
		cfg.Modules.AgentState = parseBool(val)
	case "modules.model_name", "model_name":
		cfg.Modules.ModelName = parseBool(val)
	case "modules.git_branch", "git_branch":
		cfg.Modules.GitBranch = parseBool(val)
	case "modules.context_bar", "context_bar":
		cfg.Modules.ContextBar = parseBool(val)
	case "modules.artifacts_count", "artifacts_count":
		cfg.Modules.ArtifactsCount = parseBool(val)
	case "modules.subagents_count", "subagents_count":
		cfg.Modules.SubagentsCount = parseBool(val)
	case "modules.tasks_count", "tasks_count":
		cfg.Modules.TasksCount = parseBool(val)
	case "modules.sandbox_badge", "sandbox_badge":
		cfg.Modules.SandboxBadge = parseBool(val)
	case "modules.session_usage", "session_usage":
		cfg.Modules.SessionUsage = parseBool(val)
	case "modules.weekly_usage", "weekly_usage":
		cfg.Modules.WeeklyUsage = parseBool(val)

	case "layout.style", "style":
		cfg.Layout.Style = val
	case "layout.separator", "separator":
		cfg.Layout.Separator = val
	case "layout.metrics_separator", "metrics_separator":
		cfg.Layout.MetricsSeparator = val
	case "layout.frame_top", "frame_top":
		cfg.Layout.FrameTop = val
	case "layout.frame_bottom", "frame_bottom":
		cfg.Layout.FrameBottom = val
	case "layout.wide_breakpoint", "wide_breakpoint":
		fmt.Sscanf(val, "%d", &cfg.Layout.WideBreakpoint)
	case "layout.standard_breakpoint", "standard_breakpoint":
		fmt.Sscanf(val, "%d", &cfg.Layout.StandardBreakpoint)

	case "context_bar.length":
		fmt.Sscanf(val, "%d", &cfg.ContextBar.Length)
	case "context_bar.style":
		cfg.ContextBar.Style = val
	case "context_bar.warning_threshold":
		fmt.Sscanf(val, "%f", &cfg.ContextBar.WarningThreshold)
	case "context_bar.critical_threshold":
		fmt.Sscanf(val, "%f", &cfg.ContextBar.CriticalThreshold)

	default:
		return "", fmt.Errorf("unknown configuration key %q", key)
	}

	cfg.validateAndSanitize()
	if err := Save(targetPath, cfg); err != nil {
		return "", err
	}
	return targetPath, nil
}

func parseBool(val string) bool {
	v := strings.ToLower(strings.TrimSpace(val))
	return v == "true" || v == "1" || v == "yes" || v == "on" || v == "enable" || v == "enabled"
}
