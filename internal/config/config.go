// Package config provides configuration structures and constants for the datapack generator.
package config

import (
	"fmt"
)

// Datapack format versions: https://minecraft.wiki/w/Data_pack
const (
	// Old format (<= 1.20)
	OldPackFormat   = 6  // 1.16.5 default
	OldFunctionsDir = "functions"

	// New format (>= 1.21)
	NewPackFormat   = 41 // 1.21.1 default
	NewFunctionsDir = "function"

	// Default values
	DefaultModel    = "llama3.2"
	DefaultOllamaURL = "http://localhost:11434"
	DefaultOutputDir = "."
)

// FormatType represents the datapack format version
type FormatType int

const (
	FormatAuto FormatType = iota
	FormatOld
	FormatNew
)

// String returns the string representation of the format type
func (f FormatType) String() string {
	switch f {
	case FormatOld:
		return "old (<=1.20)"
	case FormatNew:
		return "new (>=1.21)"
	default:
		return "auto (default to old)"
	}
}

// GetPackFormat returns the pack format number for the given format type
func (f FormatType) GetPackFormat() int {
	if f == FormatNew {
		return NewPackFormat
	}
	return OldPackFormat
}

// GetFunctionsDir returns the functions directory name for the given format type
func (f FormatType) GetFunctionsDir() string {
	if f == FormatNew {
		return NewFunctionsDir
	}
	return OldFunctionsDir
}

// GetVersion returns the Minecraft version string for the given format type
func (f FormatType) GetVersion() string {
	switch f {
	case FormatOld:
		return "1.16.5"
	case FormatNew:
		return "1.21.1"
	default:
		return "1.16.5"
	}
}

// Config holds all configuration for the datapack generator
type Config struct {
	// Format specifies old or new datapack format
	Format FormatType

	// Prompt is the user's description of the datapack to generate
	Prompt string

	// Model is the Ollama model to use
	Model string

	// OutputDir is the directory where datapack files will be created
	OutputDir string

	// OllamaURL is the URL of the Ollama API
	OllamaURL string

	// Verbose enables verbose output
	Verbose bool

	// DryRun generates the datapack structure but doesn't write files
	DryRun bool
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Prompt == "" {
		return fmt.Errorf("prompt is required")
	}

	if c.Model == "" {
		c.Model = DefaultModel
	}

	if c.OutputDir == "" {
		c.OutputDir = DefaultOutputDir
	}

	if c.OllamaURL == "" {
		c.OllamaURL = DefaultOllamaURL
	}

	return nil
}

// NewConfig creates a new configuration with the given options
func NewConfig(opts ...Option) *Config {
	cfg := &Config{
		Format:    FormatAuto,
		Model:     DefaultModel,
		OutputDir: DefaultOutputDir,
		OllamaURL: DefaultOllamaURL,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return cfg
}

// Option is a function that modifies a Config
type Option func(*Config)

// WithFormat sets the datapack format
func WithFormat(f FormatType) Option {
	return func(c *Config) {
		c.Format = f
	}
}

// WithPrompt sets the prompt
func WithPrompt(p string) Option {
	return func(c *Config) {
		c.Prompt = p
	}
}

// WithModel sets the Ollama model
func WithModel(m string) Option {
	return func(c *Config) {
		c.Model = m
	}
}

// WithOutputDir sets the output directory
func WithOutputDir(d string) Option {
	return func(c *Config) {
		c.OutputDir = d
	}
}

// WithOllamaURL sets the Ollama API URL
func WithOllamaURL(u string) Option {
	return func(c *Config) {
		c.OllamaURL = u
	}
}

// WithVerbose enables verbose output
func WithVerbose(v bool) Option {
	return func(c *Config) {
		c.Verbose = v
	}
}

// WithDryRun enables dry run mode
func WithDryRun(v bool) Option {
	return func(c *Config) {
		c.DryRun = v
	}
}
