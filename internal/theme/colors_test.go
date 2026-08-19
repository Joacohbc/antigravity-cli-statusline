package theme_test

import (
	"os"
	"testing"

	"github.com/joaco/antigravity-statusline/internal/theme"
)

func TestColors_Codes(t *testing.T) {
	t.Parallel()

	if theme.Reset != "\033[0m" {
		t.Errorf("expected Reset to be \\033[0m, got %q", theme.Reset)
	}

	fg256 := theme.Fg256(196)
	if fg256 != "\033[38;5;196m" {
		t.Errorf("expected Fg256(196) to be \\033[38;5;196m, got %q", fg256)
	}

	bg256 := theme.Bg256(235)
	if bg256 != "\033[48;5;235m" {
		t.Errorf("expected Bg256(235) to be \\033[48;5;235m, got %q", bg256)
	}

	fgRGB := theme.FgRGB(255, 100, 50)
	if fgRGB != "\033[38;2;255;100;50m" {
		t.Errorf("expected FgRGB(255, 100, 50) to be \\033[38;2;255;100;50m, got %q", fgRGB)
	}

	bgRGB := theme.BgRGB(10, 20, 30)
	if bgRGB != "\033[48;2;10;20;30m" {
		t.Errorf("expected BgRGB(10, 20, 30) to be \\033[48;2;10;20;30m, got %q", bgRGB)
	}

	if theme.FgTrueColor(1, 2, 3) != theme.FgRGB(1, 2, 3) {
		t.Errorf("FgTrueColor mismatch with FgRGB")
	}
	if theme.BgTrueColor(1, 2, 3) != theme.BgRGB(1, 2, 3) {
		t.Errorf("BgTrueColor mismatch with BgRGB")
	}
}

func TestStyle_And_Paint(t *testing.T) {
	theme.SetColorEnabled(true)
	defer theme.ResetColorOverride()

	tests := []struct {
		name     string
		code     string
		input    string
		expected string
	}{
		{"empty text", theme.FgGreen, "", ""},
		{"simple styled text", theme.FgGreen, "READY", theme.FgGreen + "READY" + theme.Reset},
		{"compound style", theme.FgBrightYellow + theme.Bold, "WARN", theme.FgBrightYellow + theme.Bold + "WARN" + theme.Reset},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := theme.Style(tt.code, tt.input)
			if got != tt.expected {
				t.Errorf("Style(%q, %q) = %q; want %q", tt.code, tt.input, got, tt.expected)
			}
			paintGot := theme.Paint(tt.code, tt.input)
			if paintGot != tt.expected {
				t.Errorf("Paint(%q, %q) = %q; want %q", tt.code, tt.input, paintGot, tt.expected)
			}
		})
	}
}

func TestCapsule(t *testing.T) {
	theme.SetColorEnabled(true)
	defer theme.ResetColorOverride()

	if theme.Capsule(theme.FgWhite, theme.BgBlue, theme.FgBlue, "") != "" {
		t.Errorf("Capsule with empty text should return empty string")
	}

	cap := theme.Capsule(theme.FgWhite, theme.BgBlue, theme.FgBlue, "INFO")
	if cap == "" {
		t.Errorf("Capsule returned empty string for non-empty text")
	}
	if theme.VisibleWidth(cap) != theme.VisibleWidth(theme.PillLeft+" INFO "+theme.PillRight) {
		t.Errorf("Capsule visible width mismatch")
	}
}

func TestColorDisabled(t *testing.T) {
	theme.SetColorEnabled(false)
	defer theme.ResetColorOverride()

	styled := theme.Style(theme.FgRed, "ERROR")
	if styled != "ERROR" {
		t.Errorf("expected unstyled 'ERROR' when color disabled, got %q", styled)
	}

	capsule := theme.Capsule(theme.FgWhite, theme.BgBlue, theme.FgBlue, "TAG")
	if capsule != " TAG " {
		t.Errorf("expected ' TAG ' capsule when color disabled, got %q", capsule)
	}
}

func TestNoColor_And_TermDumb(t *testing.T) {
	theme.ResetColorOverride()

	origNoColor, hadNoColor := os.LookupEnv("NO_COLOR")
	origTerm, hadTerm := os.LookupEnv("TERM")
	defer func() {
		if hadNoColor {
			os.Setenv("NO_COLOR", origNoColor)
		} else {
			os.Unsetenv("NO_COLOR")
		}
		if hadTerm {
			os.Setenv("TERM", origTerm)
		} else {
			os.Unsetenv("TERM")
		}
		theme.ResetColorOverride()
	}()

	os.Setenv("NO_COLOR", "1")
	os.Unsetenv("TERM")
	if !theme.IsNoColor() {
		t.Errorf("expected IsNoColor() to be true when NO_COLOR=1")
	}
	if theme.ColorEnabled() {
		t.Errorf("expected ColorEnabled() to be false when NO_COLOR=1")
	}

	os.Unsetenv("NO_COLOR")
	os.Setenv("TERM", "dumb")
	if !theme.IsNoColor() {
		t.Errorf("expected IsNoColor() to be true when TERM=dumb")
	}
	if theme.ColorEnabled() {
		t.Errorf("expected ColorEnabled() to be false when TERM=dumb")
	}

	os.Unsetenv("NO_COLOR")
	os.Setenv("TERM", "xterm-256color")
	if theme.IsNoColor() {
		t.Errorf("expected IsNoColor() to be false when TERM=xterm-256color and NO_COLOR unset")
	}
	if !theme.ColorEnabled() {
		t.Errorf("expected ColorEnabled() to be true when TERM=xterm-256color and NO_COLOR unset")
	}
}

func TestClean(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty string", "", ""},
		{"plain ascii", "hello world", "hello world"},
		{"simple CSI color", "\033[31mhello\033[0m", "hello"},
		{"bold and compound CSI", "\033[1;32;40mstatus\033[0m", "status"},
		{"256 color CSI", "\033[38;5;196mred\033[0m", "red"},
		{"truecolor CSI", "\033[38;2;255;0;0mtruecolor\033[0m", "truecolor"},
		{"osc title sequence", "\033]0;title\007content", "content"},
		{"control characters", "line1\x00\x07line2", "line1line2"},
		{"mixed symbols and colors", "\033[34m main\033[0m \033[93m●\033[0m", " main ●"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := theme.Clean(tt.input)
			if got != tt.expected {
				t.Errorf("Clean(%q) = %q; want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestVisibleWidth(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"empty string", "", 0},
		{"ascii simple", "hello", 5},
		{"styled ascii", "\033[31mhello\033[0m", 5},
		{"compound styled text", "\033[1;32mREADY\033[0m ╱ \033[35mGemini\033[0m", 14},
		{"wide emoji", "🚀", 2},
		{"emoji and text", "🚀 launch", 9},
		{"cjk characters", "你好", 4},
		{"combining marks", "e\u0301", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := theme.VisibleWidth(tt.input)
			if got != tt.expected {
				t.Errorf("VisibleWidth(%q) = %d; want %d", tt.input, got, tt.expected)
			}
		})
	}
}

func TestTheme_Defaults(t *testing.T) {
	t.Parallel()

	def := theme.DefaultTheme()
	if def.Name != "default" || !def.ColorActive {
		t.Errorf("unexpected default theme: %+v", def)
	}

	plain := theme.PlainTheme()
	if plain.Name != "plain" || plain.ColorActive {
		t.Errorf("unexpected plain theme: %+v", plain)
	}
}
