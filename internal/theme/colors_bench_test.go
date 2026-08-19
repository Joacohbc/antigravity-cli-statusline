package theme_test

import (
	"testing"

	"github.com/joaco/antigravity-statusline/internal/theme"
)

func BenchmarkStyle(b *testing.B) {
	theme.SetColorEnabled(true)
	defer theme.ResetColorOverride()

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = theme.Style(theme.FgGreen+theme.Bold, "READY")
	}
}

func BenchmarkClean_NoEscapes(b *testing.B) {
	text := "standard text without escape codes"
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = theme.Clean(text)
	}
}

func BenchmarkClean_WithEscapes(b *testing.B) {
	text := "\033[1;32mREADY\033[0m ╱ \033[35mGemini 3.5\033[0m ╱ \033[34m main\033[0m"
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = theme.Clean(text)
	}
}

func BenchmarkVisibleWidth(b *testing.B) {
	text := "\033[1;32mREADY\033[0m ╱ \033[35mGemini 3.5\033[0m ╱ \033[34m main\033[0m"
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = theme.VisibleWidth(text)
	}
}

func BenchmarkCapsule(b *testing.B) {
	theme.SetColorEnabled(true)
	defer theme.ResetColorOverride()

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = theme.Capsule(theme.FgWhite, theme.BgBlue, theme.FgBlue, "INFO")
	}
}
