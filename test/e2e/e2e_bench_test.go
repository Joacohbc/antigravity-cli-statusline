package e2e_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func BenchmarkE2E_PipelineExecution(b *testing.B) {
	fixturePath := filepath.Join(projectDir, "test", "testdata", "payload_full.json")
	payloadData, err := os.ReadFile(fixturePath)
	if err != nil {
		b.Fatalf("failed reading test fixture: %v", err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		cmd := exec.Command(binaryPath)
		cmd.Dir = projectDir
		cmd.Stdin = bytes.NewReader(payloadData)

		var outBuf bytes.Buffer
		cmd.Stdout = &outBuf

		if err := cmd.Run(); err != nil {
			b.Fatalf("binary run failed: %v", err)
		}
	}
}

func BenchmarkE2E_PreviewExecution(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		cmd := exec.Command(binaryPath, "--preview", "-w", "120")
		cmd.Dir = projectDir

		var outBuf bytes.Buffer
		cmd.Stdout = &outBuf

		if err := cmd.Run(); err != nil {
			b.Fatalf("binary preview run failed: %v", err)
		}
	}
}

func BenchmarkE2E_ConfigFileExecution(b *testing.B) {
	fixturePath := filepath.Join(projectDir, "test", "testdata", "payload_full.json")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		cmd := exec.Command(binaryPath, "--config", fixturePath)
		cmd.Dir = projectDir

		var outBuf bytes.Buffer
		cmd.Stdout = &outBuf

		if err := cmd.Run(); err != nil {
			b.Fatalf("binary config run failed: %v", err)
		}
	}
}
