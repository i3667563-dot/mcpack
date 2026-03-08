package datapack

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDatapackSpec(t *testing.T) {
	t.Run("Validate valid spec", func(t *testing.T) {
		spec := &DatapackSpec{
			Files: []FileSpec{
				{Path: "data/test/functions/load.mcfunction", Content: ""},
			},
			Namespace: "test",
		}
		if err := spec.Validate(); err != nil {
			t.Errorf("Validate() unexpected error: %v", err)
		}
	})

	t.Run("Validate no files", func(t *testing.T) {
		spec := &DatapackSpec{
			Files:     []FileSpec{},
			Namespace: "test",
		}
		if err := spec.Validate(); err != ErrNoFiles {
			t.Errorf("Validate() = %v, want %v", err, ErrNoFiles)
		}
	})

	t.Run("Validate no namespace", func(t *testing.T) {
		spec := &DatapackSpec{
			Files: []FileSpec{
				{Path: "data/test/functions/load.mcfunction", Content: ""},
			},
			Namespace: "",
		}
		if err := spec.Validate(); err != ErrNoNamespace {
			t.Errorf("Validate() = %v, want %v", err, ErrNoNamespace)
		}
	})
}

func TestWriter(t *testing.T) {
	t.Run("Write datapack", func(t *testing.T) {
		// Create temp directory
		tmpDir := t.TempDir()

		spec := &DatapackSpec{
			Files: []FileSpec{
				{
					Path:    "data/mydatapack/functions/load.mcfunction",
					Content: "# Load function\nscoreboard objectives add test dummy",
				},
				{
					Path:    "data/mydatapack/functions/tick.mcfunction",
					Content: "# Tick function",
				},
			},
			Namespace:     "mydatapack",
			Description:   "Test datapack",
			LoadFunctions: []string{"mydatapack:load"},
			TickFunctions: []string{"mydatapack:tick"},
		}

		writer := NewWriter(tmpDir, "mydatapack", false)
		if err := writer.Write(spec, 6); err != nil {
			t.Fatalf("Write() error: %v", err)
		}

		// Verify files exist
		files := []string{
			"pack.mcmeta",
			"data/mydatapack/functions/load.mcfunction",
			"data/mydatapack/functions/tick.mcfunction",
			"data/mydatapack/tags/function/load.json",
			"data/mydatapack/tags/function/tick.json",
		}

		for _, f := range files {
			path := filepath.Join(tmpDir, f)
			if _, err := os.Stat(path); os.IsNotExist(err) {
				t.Errorf("Expected file %s to exist", f)
			}
		}
	})
}
