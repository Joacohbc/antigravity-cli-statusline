package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/joaco/antigravity-statusline/internal/config"
	"github.com/joaco/antigravity-statusline/internal/theme"
)

type InstallOptions struct {
	Dest         string
	SkipSettings bool
	InitConfig   bool
}

func defaultInstallDest() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not determine user home directory: %w", err)
	}

	destDir := filepath.Join(homeDir, ".gemini", "antigravity-cli")
	binName := "statusline"
	if runtime.GOOS == "windows" {
		binName = "statusline.exe"
	}
	return filepath.Join(destDir, binName), nil
}

func defaultSettingsPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not determine user home directory: %w", err)
	}
	return filepath.Join(homeDir, ".gemini", "antigravity-cli", "settings.json"), nil
}

func installBinary(srcPath, destPath string) error {
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory %s: %w", destDir, err)
	}

	srcFile, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("failed to open source executable %s: %w", srcPath, err)
	}
	defer srcFile.Close()

	tmpDest := destPath + ".tmp"
	destFile, err := os.OpenFile(tmpDest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("failed to create target binary %s: %w", tmpDest, err)
	}

	if _, err := io.Copy(destFile, srcFile); err != nil {
		destFile.Close()
		_ = os.Remove(tmpDest)
		return fmt.Errorf("failed to copy binary: %w", err)
	}
	destFile.Close()

	if err := os.Rename(tmpDest, destPath); err != nil {
		// On windows rename might fail if target exists
		_ = os.Remove(destPath)
		if renameErr := os.Rename(tmpDest, destPath); renameErr != nil {
			return fmt.Errorf("failed to finalize binary install: %w", renameErr)
		}
	}

	return nil
}

func updateSettingsFile(settingsPath, binaryPath string) error {
	dir := filepath.Dir(settingsPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory for settings: %w", err)
	}

	settings := make(map[string]any)

	if data, err := os.ReadFile(settingsPath); err == nil && len(data) > 0 {
		_ = json.Unmarshal(data, &settings)
	}

	settings["statusLine"] = map[string]any{
		"type":    "command",
		"command": binaryPath,
		"enabled": true,
	}

	formatted, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize settings: %w", err)
	}

	if err := os.WriteFile(settingsPath, formatted, 0644); err != nil {
		return fmt.Errorf("failed to write settings file %s: %w", settingsPath, err)
	}

	return nil
}

func NewInstallCmd(out, errWriter io.Writer) *cobra.Command {
	opts := InstallOptions{
		InitConfig: true,
	}

	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install statusline binary and configure Antigravity CLI",
		Long: `Installs the statusline binary into ~/.gemini/antigravity-cli/statusline
and automatically configures ~/.gemini/antigravity-cli/settings.json to enable it.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			srcPath, err := os.Executable()
			if err != nil {
				return fmt.Errorf("could not determine current executable path: %w", err)
			}
			srcPath, err = filepath.EvalSymlinks(srcPath)
			if err != nil {
				return fmt.Errorf("could not resolve executable symlinks: %w", err)
			}

			destPath := opts.Dest
			if destPath == "" {
				defaultDest, dErr := defaultInstallDest()
				if dErr != nil {
					return dErr
				}
				destPath = defaultDest
			}

			fmt.Fprintf(out, "%s Installing statusline to:\n   %s\n", theme.Style(theme.FgBrightCyan+theme.Bold, "==>"), destPath)
			if err := installBinary(srcPath, destPath); err != nil {
				return err
			}

			if !opts.SkipSettings {
				settingsPath, sErr := defaultSettingsPath()
				if sErr == nil {
					fmt.Fprintf(out, "%s Updating Antigravity CLI settings:\n   %s\n", theme.Style(theme.FgBrightGreen+theme.Bold, "==>"), settingsPath)
					if err := updateSettingsFile(settingsPath, destPath); err != nil {
						fmt.Fprintf(errWriter, "warning: could not update settings.json: %v\n", err)
					}
				}
			}

			if opts.InitConfig {
				for _, cfgCandidate := range config.DefaultConfigPaths() {
					if _, err := os.Stat(cfgCandidate); err == nil {
						// Already exists
						opts.InitConfig = false
						break
					}
				}
				if opts.InitConfig {
					createdCfg, cErr := config.InitConfigFile("")
					if cErr == nil {
						fmt.Fprintf(out, "%s Created default user configuration:\n   %s\n", theme.Style(theme.FgBrightYellow+theme.Bold, "==>"), createdCfg)
					}
				}
			}

			fmt.Fprintf(out, "\n%s Installation complete! Status line is active for Antigravity CLI.\n", theme.Style(theme.FgBrightGreen+theme.Bold, "✔"))
			fmt.Fprintf(out, "You can also activate it live with:\n   /statusline %s\n", destPath)
			return nil
		},
	}

	cmd.Flags().StringVarP(&opts.Dest, "dest", "d", "", "Custom binary destination path")
	cmd.Flags().BoolVar(&opts.SkipSettings, "skip-settings", false, "Do not modify settings.json")
	cmd.Flags().BoolVar(&opts.InitConfig, "create-config", true, "Create ~/.config/antigravity/statusline.json if not present")

	return cmd
}
