package compiler_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jaaanko/twitch-clip-compilation-tool/internal/compiler"
)

func TestRun(t *testing.T) {
	outputPath := t.TempDir()
	outputName := "compilation.mp4"
	compiler := compiler.New(outputPath, outputName)
	paths := []string{
		filepath.Join("testdata", "sample1.mp4"),
		filepath.Join("testdata", "sample2.mp4"),
	}

	err := compiler.Run(paths)
	if err != nil {
		t.Fatal(err)
	}

	dir, err := os.Open(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	defer dir.Close()

	fileNames, err := dir.Readdirnames(0)
	if err != nil {
		t.Fatal(err)
	}

	if len(fileNames) != 1 {
		t.Fatalf("expected %v files, got %v", 1, len(fileNames))
	}

	if fileNames[0] != outputName {
		t.Fatalf("expected output file to be called %v, got %v", outputName, fileNames[0])
	}
}
