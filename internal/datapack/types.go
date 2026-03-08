// Package datapack provides structures and functions for Minecraft datapack generation.
package datapack

// FileSpec represents a file to be created in the datapack
type FileSpec struct {
	// Path is the relative path within the datapack (e.g., "data/my_namespace/functions/load.mcfunction")
	Path string `json:"path"`

	// Content is the file content
	Content string `json:"content"`

	// Description is an optional description of what this file does
	Description string `json:"-"`
}

// DatapackSpec represents the complete datapack structure returned by the AI
type DatapackSpec struct {
	// Files is the list of files to create
	Files []FileSpec `json:"files"`

	// Namespace is the datapack namespace (e.g., "mydatapack")
	Namespace string `json:"namespace"`

	// Description is a brief description of the datapack
	Description string `json:"description,omitempty"`

	// LoadFunctions lists functions that should be in the load tag
	LoadFunctions []string `json:"load_functions,omitempty"`

	// TickFunctions lists functions that should be in the tick tag
	TickFunctions []string `json:"tick_functions,omitempty"`
}

// Validate validates the datapack specification
func (d *DatapackSpec) Validate() error {
	if len(d.Files) == 0 {
		return ErrNoFiles
	}

	if d.Namespace == "" {
		return ErrNoNamespace
	}

	return nil
}

// PackMcmeta represents the pack.mcmeta file structure
type PackMcmeta struct {
	Pack PackMeta `json:"pack"`
}

// PackMeta represents the pack section of pack.mcmeta
type PackMeta struct {
	PackFormat  int    `json:"pack_format"`
	Description string `json:"description"`
}

// FunctionTag represents a function tag JSON structure
type FunctionTag struct {
	Values []string `json:"values"`
	Replace bool    `json:"replace,omitempty"`
}

// Errors for datapack validation
var (
	ErrNoFiles    = &ValidationError{"no files specified in datapack"}
	ErrNoNamespace = &ValidationError{"no namespace specified in datapack"}
)

// ValidationError represents a datapack validation error
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}
