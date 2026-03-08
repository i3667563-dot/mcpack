// Package cmd provides the CLI command handling for mcpack.
package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"

	"mcpack/internal/config"
	"mcpack/internal/generator"
	"mcpack/internal/ollama"
)

// Run executes the mcpack CLI
func Run(args []string) error {
	cmd := &Command{}
	fs := flag.NewFlagSet("mcpack", flag.ExitOnError)

	// Define flags - long versions
	fs.BoolVar(&cmd.oldFormat, "old", false, "Use old datapack format (<=1.20, default 1.16.5)")
	fs.BoolVar(&cmd.newFormat, "new", false, "Use new datapack format (>=1.21, default 1.21.1)")
	fs.StringVar(&cmd.prompt, "prompt", "", "Description of the datapack to generate")
	fs.StringVar(&cmd.model, "model", config.DefaultModel, "Ollama model to use")
	fs.StringVar(&cmd.dir, "dir", config.DefaultOutputDir, "Output directory for datapack files")
	fs.StringVar(&cmd.ollamaURL, "ollama", config.DefaultOllamaURL, "Ollama API URL")
	fs.BoolVar(&cmd.verbose, "verbose", false, "Enable verbose output")
	fs.BoolVar(&cmd.dryRun, "dry-run", false, "Generate datapack structure but don't write files")
	fs.BoolVar(&cmd.help, "help", false, "Show help")

	// Define short versions as aliases (avoiding conflicts)
	fs.StringVar(&cmd.prompt, "p", "", "Shorthand for --prompt")
	fs.StringVar(&cmd.model, "m", config.DefaultModel, "Shorthand for --model")
	fs.StringVar(&cmd.dir, "d", config.DefaultOutputDir, "Shorthand for --dir")
	fs.BoolVar(&cmd.verbose, "v", false, "Shorthand for --verbose")
	fs.BoolVar(&cmd.help, "h", false, "Shorthand for --help")
	// Note: -o conflicts with --old/--ollama, -n conflicts with --new/--dry-run

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, `mcpack - AI-powered Minecraft datapack generator

Usage:
  mcpack [flags]

Flags:
  --old                 Use old datapack format (<=1.20, default 1.16.5)
  --new                 Use new datapack format (>=1.21, default 1.21.1)
  --prompt, -p string   Description of the datapack to generate (required)
  --model, -m string    Ollama model to use (default: %s)
  --dir, -d string      Output directory for datapack files (default: current directory)
  --ollama string       Ollama API URL (default: %s)
  --verbose, -v         Enable verbose output
  --dry-run             Generate datapack structure but don't write files
  --help, -h            Show help

Examples:
  mcpack -p "create a scoreboard system for tracking player deaths"
  mcpack --new -p "custom enchantment system" -m llama3.2
  mcpack --old -d ./mydatapack -p "teleportation commands"

Notes:
  - --old and --new are mutually exclusive
  - Default format is old (1.16.5, pack format 6)
  - New format uses pack format 41 (1.21.1)
  - Old format uses "functions" folder, new uses "function"
`,
			config.DefaultModel, config.DefaultOllamaURL)
	}

	if err := fs.Parse(args); err != nil {
		return &CLIError{Op: "parse flags", Err: err}
	}

	return cmd.Execute()
}

// Command represents the mcpack CLI command
type Command struct {
	oldFormat bool
	newFormat bool
	prompt    string
	model     string
	dir       string
	ollamaURL string
	verbose   bool
	dryRun    bool
	help      bool
}

// Execute executes the command
func (c *Command) Execute() error {
	// Show help if requested
	if c.help {
		fmt.Fprintf(os.Stderr, `mcpack - AI-powered Minecraft datapack generator

Usage:
  mcpack [flags]

Flags:
  --old                 Use old datapack format (<=1.20, default 1.16.5)
  --new                 Use new datapack format (>=1.21, default 1.21.1)
  --prompt, -p string   Description of the datapack to generate (required)
  --model, -m string    Ollama model to use (default: %s)
  --dir, -d string      Output directory for datapack files (default: current directory)
  --ollama string       Ollama API URL (default: %s)
  --verbose, -v         Enable verbose output
  --dry-run             Generate datapack structure but don't write files
  --help, -h            Show help

Examples:
  mcpack -p "create a scoreboard system for tracking player deaths"
  mcpack --new -p "custom enchantment system" -m llama3.2
  mcpack --old -d ./mydatapack -p "teleportation commands"

Notes:
  - --old and --new are mutually exclusive
  - Default format is old (1.16.5, pack format 6)
  - New format uses pack format 41 (1.21.1)
  - Old format uses "functions" folder, new uses "function"
`,
			config.DefaultModel, config.DefaultOllamaURL)
		os.Exit(0)
	}

	// Validate flags
	if c.oldFormat && c.newFormat {
		return &CLIError{Op: "validate flags", Err: fmt.Errorf("--old and --new are mutually exclusive")}
	}

	if c.prompt == "" {
		return &CLIError{Op: "validate flags", Err: fmt.Errorf("--prompt (-p) is required")}
	}

	// Determine format
	var format config.FormatType
	if c.newFormat {
		format = config.FormatNew
	} else if c.oldFormat {
		format = config.FormatOld
	} else {
		format = config.FormatAuto
	}

	// Create configuration
	cfg := config.NewConfig(
		config.WithFormat(format),
		config.WithPrompt(c.prompt),
		config.WithModel(c.model),
		config.WithOutputDir(c.dir),
		config.WithOllamaURL(c.ollamaURL),
		config.WithVerbose(c.verbose),
		config.WithDryRun(c.dryRun),
	)

	// Create generator and run
	gen := generator.NewGenerator(cfg)
	ctx := context.Background()

	if err := gen.Generate(ctx); err != nil {
		return c.handleError(err)
	}

	return nil
}

// handleError handles and formats errors
func (c *Command) handleError(err error) error {
	if apiErr, ok := err.(*ollama.APIError); ok {
		if apiErr.IsConnectionError() {
			return &CLIError{
				Op:  "connect to Ollama",
				Err: fmt.Errorf("make sure Ollama is running at %s", c.ollamaURL),
			}
		}
		if apiErr.IsNotFound() {
			return &CLIError{
				Op:  "find model",
				Err: fmt.Errorf("model '%s' not found, run 'ollama pull %s'", c.model, c.model),
			}
		}
	}

	return err
}

// CLIError represents a CLI error
type CLIError struct {
	Op  string
	Err error
}

func (e *CLIError) Error() string {
	return fmt.Sprintf("mcpack %s: %v", e.Op, e.Err)
}

func (e *CLIError) Unwrap() error {
	return e.Err
}
