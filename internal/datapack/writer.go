// Package datapack provides structures and functions for Minecraft datapack generation.
package datapack

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Writer handles writing datapack files to disk
type Writer struct {
	baseDir   string
	namespace string
	verbose   bool
}

// NewWriter creates a new datapack writer
func NewWriter(baseDir, namespace string, verbose bool) *Writer {
	return &Writer{
		baseDir:   baseDir,
		namespace: namespace,
		verbose:   verbose,
	}
}

// Write writes the datapack specification to disk
func (w *Writer) Write(spec *DatapackSpec, packFormat int) error {
	if err := spec.Validate(); err != nil {
		return fmt.Errorf("invalid datapack spec: %w", err)
	}

	// Create pack.mcmeta
	if err := w.writePackMcmeta(packFormat, spec.Description); err != nil {
		return fmt.Errorf("failed to write pack.mcmeta: %w", err)
	}

	// Create each file
	for _, file := range spec.Files {
		if err := w.writeFile(file); err != nil {
			return fmt.Errorf("failed to write file %s: %w", file.Path, err)
		}
	}

	// Create function tags if specified
	if err := w.createFunctionTags(spec); err != nil {
		return fmt.Errorf("failed to create function tags: %w", err)
	}

	return nil
}

// writePackMcmeta creates the pack.mcmeta file
func (w *Writer) writePackMcmeta(packFormat int, description string) error {
	if description == "" {
		description = "AI Generated Datapack"
	}

	packMcmeta := PackMcmeta{
		Pack: PackMeta{
			PackFormat:  packFormat,
			Description: description,
		},
	}

	data, err := json.MarshalIndent(packMcmeta, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal pack.mcmeta: %w", err)
	}

	path := filepath.Join(w.baseDir, "pack.mcmeta")
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write pack.mcmeta: %w", err)
	}

	if w.verbose {
		fmt.Printf("  [WRITE] pack.mcmeta\n")
	}

	return nil
}

// writeFile writes a single file to the datapack
func (w *Writer) writeFile(file FileSpec) error {
	// Skip pack.mcmeta from AI response as we generated it
	if filepath.Base(file.Path) == "pack.mcmeta" {
		return nil
	}

	fullPath := filepath.Join(w.baseDir, file.Path)

	// Create parent directories
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Write file content
	if err := os.WriteFile(fullPath, []byte(file.Content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	if w.verbose {
		fmt.Printf("  [WRITE] %s\n", file.Path)
	}

	return nil
}

// createFunctionTags creates the function tag files for load and tick
func (w *Writer) createFunctionTags(spec *DatapackSpec) error {
	tagsDir := filepath.Join(w.baseDir, "data", spec.Namespace, "tags", "function")

	// Create load.json if there are load functions
	if len(spec.LoadFunctions) > 0 {
		loadTag := FunctionTag{
			Values: spec.LoadFunctions,
		}
		if err := w.writeFunctionTag(tagsDir, "load.json", loadTag); err != nil {
			return err
		}
	}

	// Create tick.json if there are tick functions
	if len(spec.TickFunctions) > 0 {
		tickTag := FunctionTag{
			Values: spec.TickFunctions,
		}
		if err := w.writeFunctionTag(tagsDir, "tick.json", tickTag); err != nil {
			return err
		}
	}

	return nil
}

// writeFunctionTag writes a function tag file
func (w *Writer) writeFunctionTag(tagsDir, filename string, tag FunctionTag) error {
	if err := os.MkdirAll(tagsDir, 0755); err != nil {
		return fmt.Errorf("failed to create tags directory: %w", err)
	}

	data, err := json.MarshalIndent(tag, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal function tag: %w", err)
	}

	path := filepath.Join(tagsDir, filename)
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write function tag: %w", err)
	}

	if w.verbose {
		fmt.Printf("  [WRITE] data/%s/tags/function/%s\n", w.namespace, filename)
	}

	return nil
}

// GetDatapackPath returns the full path to the datapack directory
func (w *Writer) GetDatapackPath() string {
	return w.baseDir
}
