package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/joaco/antigravity-statusline/internal/config"
	"github.com/joaco/antigravity-statusline/internal/payload"
	"github.com/joaco/antigravity-statusline/internal/renderer"
	"github.com/joaco/antigravity-statusline/internal/theme"
)

const (
	Version = "antigravity-cli-statusline v1.0.0 (Go 1.22+)"

	ExitSuccess = 0
	ExitError   = 1
)

func HelpText() string {
	return `antigravity-cli-statusline - High-performance statusline renderer for Antigravity CLI

USAGE:
  statusline [flags]
  cat payload.json | statusline [flags]
  statusline [command]

AVAILABLE COMMANDS:
  preview     Render interactive preview statusline
  themes      List all available color theme palettes
  config      Manage statusline user configuration (init, show)
  install     Install binary to ~/.gemini/antigravity-cli and configure settings
  version     Print version information
  completion  Generate shell autocompletion scripts (bash, zsh, fish, powershell)

FLAGS:
  -h, --help            Show this help message and exit
  -v, --version         Print version information and exit
  -p, --preview         Render interactive preview statusline
  -l, --list-themes     List all available color themes with visual samples
  -w, --width-override  Override terminal width (columns, e.g. 120)
  -c, --config          Path to custom JSON config file or state payload fixture
      --init-config     Create a default configuration file in ~/.config/antigravity/statusline.json
      --theme           Theme palette: default, tokyonight, catppuccin, dracula, nord, gruvbox, ascii, plain, no-color
      --debug           Enable verbose diagnostic logging to stderr

ENVIRONMENT VARIABLES:
  STATUSLINE_CONFIG     Path to custom configuration JSON file
  NO_COLOR              Disable ANSI color styling (https://no-color.org)
  TERM                  Terminal type (TERM=dumb disables color styling)

LAYOUT MODES:
  Wide (>=120 cols)     Single-line full layout with header and metrics
  Standard (80-119)     Two-line framed layout with bracket borders
  Compact (<80 cols)    Two-line dense minimal layout

CONFIGURATION FILE:
  Searched automatically in:
    1. $STATUSLINE_CONFIG
    2. ~/.config/antigravity/statusline.json
    3. ~/.gemini/antigravity-cli/statusline.json

EXAMPLES:
  # List all available themes
  statusline --list-themes

  # Initialize user configuration file
  statusline --init-config

  # Standalone interactive preview with Tokyo Night theme
  statusline preview --theme tokyonight

  # Pipeline mode (AGY CLI integration)
  echo '{"agent_state":"working","model":{"display_name":"Gemini 3.7 Pro"}}' | statusline

EXIT CODES:
  0                     Success
  1                     Fatal error
`
}

type Runner struct {
	Args       []string
	In         io.Reader
	Out        io.Writer
	Err        io.Writer
	IsTerminal func(io.Reader) bool
}

func DefaultIsTerminal(r io.Reader) bool {
	if r == nil {
		return false
	}
	if f, ok := r.(*os.File); ok {
		stat, err := f.Stat()
		if err == nil {
			return (stat.Mode() & os.ModeCharDevice) != 0
		}
	}
	return false
}

func NewRunner(args []string, in io.Reader, out, errWriter io.Writer) *Runner {
	if in == nil {
		in = os.Stdin
	}
	if out == nil {
		out = os.Stdout
	}
	if errWriter == nil {
		errWriter = os.Stderr
	}

	return &Runner{
		Args:       args,
		In:         in,
		Out:        out,
		Err:        errWriter,
		IsTerminal: DefaultIsTerminal,
	}
}

func Execute() int {
	return Run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr)
}

func Run(args []string, in io.Reader, out, errWriter io.Writer) int {
	return NewRunner(args, in, out, errWriter).Execute()
}

func ShowThemes(out io.Writer, width int) {
	if width <= 0 {
		width = 90
	}
	sample := NewPreviewPayload(width)

	fmt.Fprintln(out, theme.Style(theme.FgBrightCyan+theme.Bold, "Available Theme Palettes:")+"\n")

	for _, name := range theme.AvailableThemes {
		t := theme.NamedTheme(name)
		if name == "ascii" || name == "plain" || name == "no-color" {
			theme.SetColorEnabled(false)
		} else {
			theme.SetColorEnabled(true)
		}
		rndr := renderer.NewRenderer(
			renderer.WithTheme(t),
			renderer.WithWidth(width),
		)
		header := theme.Style(t.Palette.Ready, "● "+name)
		rendered := rndr.Render(sample)
		fmt.Fprintf(out, "%s:\n%s\n\n", header, rendered)
	}
	theme.ResetColorOverride()
}

func (r *Runner) Execute() int {
	var (
		previewFlag    bool
		listThemesFlag bool
		widthOverride  int
		configPath     string
		initConfigFlag bool
		themeName      string
		debug          bool
		versionFlag    bool
	)

	rootCmd := &cobra.Command{
		Use:           "statusline",
		Short:         "High-performance statusline renderer for Antigravity CLI",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if versionFlag {
				fmt.Fprintln(r.Out, Version)
				return nil
			}

			if listThemesFlag {
				ShowThemes(r.Out, widthOverride)
				return nil
			}

			if initConfigFlag {
				path, err := config.InitConfigFile("")
				if err != nil {
					return fmt.Errorf("failed to initialize configuration: %w", err)
				}
				fmt.Fprintf(r.Out, "Successfully created default configuration at:\n  %s\n", path)
				return nil
			}

			if widthOverride < 0 {
				return fmt.Errorf("invalid width-override %d: width must be non-negative", widthOverride)
			}

			userCfg, _ := config.Load(configPath)
			if userCfg == nil {
				userCfg = config.DefaultConfig()
			}

			if themeName != "" {
				userCfg.Theme.Palette = themeName
			}

			themeLower := strings.ToLower(strings.TrimSpace(userCfg.Theme.Palette))
			switch themeLower {
			case "no-color", "plain", "classic", "ascii":
				theme.SetColorEnabled(false)
				defer theme.ResetColorOverride()
			case "default", "tokyonight", "catppuccin", "dracula", "nord", "gruvbox", "":
				if !theme.IsNoColor() {
					theme.SetColorEnabled(true)
					defer theme.ResetColorOverride()
				}
			default:
				return fmt.Errorf("unknown theme %q (valid options: default, tokyonight, catppuccin, dracula, nord, gruvbox, ascii, no-color, plain)", themeName)
			}

			selectedTheme := theme.NamedTheme(themeLower)
			rendererOpts := []renderer.Option{
				renderer.WithUserConfig(userCfg),
				renderer.WithTheme(selectedTheme),
			}
			if widthOverride > 0 {
				rendererOpts = append(rendererOpts, renderer.WithWidth(widthOverride))
			}
			activeRenderer := renderer.NewRenderer(rendererOpts...)

			// Check if configPath was a file passed
			if configPath != "" {
				data, readErr := os.ReadFile(configPath)
				if readErr != nil {
					return fmt.Errorf("failed to open config file %q: %w", configPath, readErr)
				}

				if len(data) > payload.MaxPayloadSize {
					return fmt.Errorf("failed to parse config file %q: exceeds maximum allowed size", configPath)
				}

				var probe map[string]json.RawMessage
				if jsonErr := json.Unmarshal(data, &probe); jsonErr != nil {
					if debug {
						fmt.Fprintf(r.Err, "debug: parse details: %v\n", jsonErr)
					}
					return fmt.Errorf("failed to parse config file %q: %w", configPath, jsonErr)
				}

				if _, hasState := probe["agent_state"]; hasState || probe["context_window"] != nil || probe["model"] != nil {
					state, parseErr := payload.Parse(strings.NewReader(string(data)))
					if parseErr != nil {
						if debug {
							fmt.Fprintf(r.Err, "debug: parse details: %v\n", parseErr)
						}
						return fmt.Errorf("failed to parse config file %q: %w", configPath, parseErr)
					}
					if widthOverride > 0 {
						state.TerminalWidth = widthOverride
					}
					fmt.Fprintln(r.Out, activeRenderer.Render(state))
					return nil
				}
			}

			if previewFlag {
				previewState := NewPreviewPayload(widthOverride)
				fmt.Fprintln(r.Out, activeRenderer.Render(previewState))
				return nil
			}

			isInteractive := r.IsTerminal != nil && r.IsTerminal(r.In)
			if isInteractive {
				previewState := NewPreviewPayload(widthOverride)
				fmt.Fprintln(r.Out, activeRenderer.Render(previewState))
				return nil
			}

			// Piped stdin
			stdinData, _ := io.ReadAll(r.In)
			if len(stdinData) > 0 {
				_ = os.WriteFile("/tmp/agy_statusline_last_payload.json", stdinData, 0644)
			}
			state, parseErr := payload.Parse(bytes.NewReader(stdinData))
			if parseErr != nil && debug {
				fmt.Fprintf(r.Err, "debug: stdin payload parse warning: %v\n", parseErr)
			}

			if widthOverride > 0 {
				state.TerminalWidth = widthOverride
			}

			fmt.Fprintln(r.Out, activeRenderer.Render(state))
			return nil
		},
	}

	rootCmd.SetIn(r.In)
	rootCmd.SetOut(r.Out)
	rootCmd.SetErr(r.Err)
	rootCmd.SetHelpTemplate(HelpText())

	pflags := rootCmd.PersistentFlags()
	pflags.BoolVarP(&previewFlag, "preview", "p", false, "Render interactive preview statusline")
	pflags.BoolVarP(&listThemesFlag, "list-themes", "l", false, "List all available color themes with visual samples")
	pflags.IntVarP(&widthOverride, "width-override", "w", 0, "Terminal width override")
	pflags.StringVarP(&configPath, "config", "c", "", "Path to custom JSON config or payload file")
	pflags.BoolVar(&initConfigFlag, "init-config", false, "Initialize default configuration file")
	pflags.StringVar(&themeName, "theme", "", "Theme option (default, tokyonight, catppuccin, dracula, nord, gruvbox, ascii, no-color, plain)")
	pflags.BoolVar(&debug, "debug", false, "Print error details and diagnostics to stderr")
	pflags.BoolVarP(&versionFlag, "version", "v", false, "Print version information")

	// Subcommand: preview
	previewCmd := &cobra.Command{
		Use:   "preview",
		Short: "Render interactive preview statusline",
		RunE: func(cmd *cobra.Command, args []string) error {
			previewState := NewPreviewPayload(widthOverride)
			userCfg, _ := config.Load(configPath)
			if userCfg == nil {
				userCfg = config.DefaultConfig()
			}
			if themeName != "" {
				userCfg.Theme.Palette = themeName
			}
			rndr := renderer.NewRenderer(
				renderer.WithUserConfig(userCfg),
				renderer.WithTheme(theme.NamedTheme(userCfg.Theme.Palette)),
				renderer.WithWidth(widthOverride),
			)
			fmt.Fprintln(r.Out, rndr.Render(previewState))
			return nil
		},
	}

	// Subcommand: themes
	themesCmd := &cobra.Command{
		Use:   "themes",
		Short: "List all available color theme palettes",
		RunE: func(cmd *cobra.Command, args []string) error {
			ShowThemes(r.Out, widthOverride)
			return nil
		},
	}

	// Subcommand: version
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(r.Out, Version)
			return nil
		},
	}

	// Subcommand: config
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Manage statusline user configuration",
	}

	configInitCmd := &cobra.Command{
		Use:   "init [path]",
		Short: "Create a default configuration file",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target := ""
			if len(args) > 0 {
				target = args[0]
			}
			path, err := config.InitConfigFile(target)
			if err != nil {
				return fmt.Errorf("failed to initialize configuration: %w", err)
			}
			fmt.Fprintf(r.Out, "Successfully created default configuration at:\n  %s\n", path)
			return nil
		},
	}

	configShowCmd := &cobra.Command{
		Use:   "show",
		Short: "Display active resolved configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(configPath)
			if err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}
			data, err := json.MarshalIndent(cfg, "", "  ")
			if err != nil {
				return err
			}
			fmt.Fprintln(r.Out, string(data))
			return nil
		},
	}

	configSetCmd := &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a configuration property (e.g. theme.palette tokyonight)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key, val := args[0], args[1]
			path, err := config.SetProperty(configPath, key, val)
			if err != nil {
				return err
			}
			fmt.Fprintf(r.Out, "Updated %s = %s in %s\n", key, val, path)
			return nil
		},
	}

	configPathCmd := &cobra.Command{
		Use:   "path",
		Short: "Print the path to the active configuration file",
		RunE: func(cmd *cobra.Command, args []string) error {
			path := configPath
			if path == "" {
				path = config.FindActiveConfigPath()
			}
			fmt.Fprintln(r.Out, path)
			return nil
		},
	}

	configCmd.AddCommand(configInitCmd, configShowCmd, configSetCmd, configPathCmd)
	rootCmd.AddCommand(previewCmd, themesCmd, versionCmd, configCmd, NewInstallCmd(r.Out, r.Err))

	// Pre-process single-dash long flags (e.g. -help, -version, -preview) for Unix tool compatibility
	normalizedArgs := make([]string, len(r.Args))
	for i, arg := range r.Args {
		switch arg {
		case "-help":
			normalizedArgs[i] = "--help"
		case "-version":
			normalizedArgs[i] = "--version"
		case "-preview":
			normalizedArgs[i] = "--preview"
		default:
			normalizedArgs[i] = arg
		}
	}

	rootCmd.SetArgs(normalizedArgs)

	if err := rootCmd.Execute(); err != nil {
		if errors.Is(err, pflag.ErrHelp) {
			return ExitSuccess
		}
		fmt.Fprintf(r.Err, "error: %v\n", err)
		return ExitError
	}

	return ExitSuccess
}
