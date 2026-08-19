package payload_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/joaco/antigravity-statusline/internal/payload"
)

func BenchmarkParse_Full(b *testing.B) {
	fixturePath := filepath.Join("..", "..", "test", "testdata", "payload_full.json")
	data, err := os.ReadFile(fixturePath)
	if err != nil {
		b.Fatalf("failed reading fixture: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		r := bytes.NewReader(data)
		_, err := payload.Parse(r)
		if err != nil {
			b.Fatalf("unexpected benchmark parse error: %v", err)
		}
	}
}

func BenchmarkParse_Minimal(b *testing.B) {
	fixturePath := filepath.Join("..", "..", "test", "testdata", "payload_minimal.json")
	data, err := os.ReadFile(fixturePath)
	if err != nil {
		b.Fatalf("failed reading fixture: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		r := bytes.NewReader(data)
		_, err := payload.Parse(r)
		if err != nil {
			b.Fatalf("unexpected benchmark parse error: %v", err)
		}
	}
}

func BenchmarkParseBytes_Full(b *testing.B) {
	fixturePath := filepath.Join("..", "..", "test", "testdata", "payload_full.json")
	data, err := os.ReadFile(fixturePath)
	if err != nil {
		b.Fatalf("failed reading fixture: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := payload.ParseBytes(data)
		if err != nil {
			b.Fatalf("unexpected benchmark parse error: %v", err)
		}
	}
}

func BenchmarkSanitize(b *testing.B) {
	dirtyString := "\x1b[31;1mfeat/statusline-v2\x1b[0m\x1b]0;TitleInjection\x07\r\n\t--extra"

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = payload.Sanitize(dirtyString)
	}
}

func BenchmarkSanitize_Clean(b *testing.B) {
	cleanString := "feat/statusline-v2-clean-branch-name"

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = payload.Sanitize(cleanString)
	}
}
