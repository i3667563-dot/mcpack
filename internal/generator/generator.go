// Package generator orchestrates the datapack generation process.
package generator

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"mcpack/internal/config"
	"mcpack/internal/datapack"
	"mcpack/internal/ollama"
	"mcpack/internal/progress"
)

// Generator handles the datapack generation workflow
type Generator struct {
	config  *config.Config
	client  *ollama.Client
	writer  *datapack.Writer
	verbose bool
}

// NewGenerator creates a new datapack generator
func NewGenerator(cfg *config.Config) *Generator {
	return &Generator{
		config:  cfg,
		client:  ollama.NewClient(cfg.OllamaURL, cfg.Model),
		verbose: cfg.Verbose,
	}
}

// Generate generates a datapack based on the configuration
func (g *Generator) Generate(ctx context.Context) error {
	if err := g.config.Validate(); err != nil {
		return &GenerateError{Op: "validate config", Err: err}
	}

	fmt.Printf("Generating datapack for Minecraft %s\n", g.config.Format.GetVersion())
	fmt.Printf("Pack format: %d\n", g.config.Format.GetPackFormat())
	fmt.Printf("Functions directory: %s\n", g.config.Format.GetFunctionsDir())
	fmt.Printf("Model: %s\n", g.config.Model)
	fmt.Printf("Output: %s\n\n", g.config.OutputDir)

	// Build the prompt
	systemPrompt := g.buildSystemPrompt()
	userPrompt := g.buildUserPrompt()
	fullPrompt := systemPrompt + "\n\n" + userPrompt

	if g.verbose {
		fmt.Println("=== Full Prompt ===")
		fmt.Println(fullPrompt)
		fmt.Println("===================")
	}

	// Call Ollama API
	fmt.Println("Contacting Ollama API...")
	spinner := progress.NewSpinner("Waiting for AI response (this may take a minute for large models)")
	spinner.Start()
	aiResponse, err := g.client.GenerateWithTimeout(fullPrompt, 5*time.Minute)
	if err != nil {
		spinner.StopWithError("Failed to get AI response")
		return &GenerateError{Op: "call Ollama API", Err: err}
	}
	spinner.StopWithMessage("AI response received")

	if g.verbose {
		fmt.Println()
		fmt.Println("=== AI Response ===")
		fmt.Println(aiResponse)
		fmt.Println("===================")
	}

	// Parse the JSON response
	parseSpinner := progress.NewSpinner("Processing AI response")
	parseSpinner.Start()
	spec, err := g.parseResponse(aiResponse)
	if err != nil {
		parseSpinner.StopWithError("Failed to parse AI response")
		return &GenerateError{Op: "parse AI response", Err: err}
	}
	parseSpinner.StopWithMessage("AI response processed")

	// Use namespace from config if not provided by AI
	if spec.Namespace == "" {
		spec.Namespace = "mydatapack"
		fmt.Println("Warning: No namespace provided by AI, using 'mydatapack'")
	}

	if g.verbose {
		fmt.Printf("\nParsed datapack spec:\n")
		fmt.Printf("  Namespace: %s\n", spec.Namespace)
		fmt.Printf("  Description: %s\n", spec.Description)
		fmt.Printf("  Files: %d\n", len(spec.Files))
		fmt.Printf("  Load functions: %v\n", spec.LoadFunctions)
		fmt.Printf("  Tick functions: %v\n", spec.TickFunctions)
	}

	// Create the datapack files
	if g.config.DryRun {
		fmt.Println("\n[DRY RUN] Would create datapack files in", g.config.OutputDir)
		return nil
	}

	fmt.Printf("\nCreating datapack files in %s...\n", g.config.OutputDir)
	writeSpinner := progress.NewSpinner("Writing datapack files")
	writeSpinner.Start()
	g.writer = datapack.NewWriter(g.config.OutputDir, spec.Namespace, g.verbose)
	if err := g.writer.Write(spec, g.config.Format.GetPackFormat()); err != nil {
		writeSpinner.StopWithError("Failed to write datapack files")
		return &GenerateError{Op: "write datapack files", Err: err}
	}
	writeSpinner.StopWithMessage("Datapack files created")

	fmt.Println("\n✓ Datapack generated successfully!")
	fmt.Printf("  Location: %s\n", g.config.OutputDir)
	fmt.Printf("  Namespace: %s\n", spec.Namespace)
	fmt.Printf("  Pack format: %d\n", g.config.Format.GetPackFormat())
	fmt.Printf("  Files created: %d\n", len(spec.Files)+1) // +1 for pack.mcmeta

	return nil
}

// buildSystemPrompt builds the system prompt with best practices
func (g *Generator) buildSystemPrompt() string {
	return fmt.Sprintf(`You are an expert Minecraft datapack developer. Generate professional-quality datapack code following these rules:

CRITICAL RULES:
1. Always use 'p' selector (e.g., @p, @a, @e, @s) NOT 's' - datapacks cannot execute commands on themselves
2. For old format (<=1.20): use "functions" folder (with 's')
3. For new format (>=1.21): use "function" folder (no 's')

SCOREBOARD BEST PRACTICES:
1. NEVER use random players for scoreboards - always create fake players
2. Always initialize scoreboards with 'dummy' criteria before using them
3. Create scoreboard objectives at the start of load function
4. Use descriptive names for fake players (e.g., "#datapack_name" or "__datapack_name__")
5. Example: scoreboard objectives add my_objective dummy "My Objective"
6. Example: scoreboard players set #counter my_objective 0
7. Always check if scoreboard objective exists or create it in load

ITEM AND NBT BEST PRACTICES:
1. Use /give command to give items to players, NOT item replace entity
2. NBT format for enchantments: {Enchantments:[{id:"minecraft:sharpness",lvl:255}]} - NO 's' suffix on lvl
3. For custom item names: use JSON text components
4. Effect durations are in ticks (20 ticks = 1 second, 60 seconds = 1200 ticks)
5. Use /effect give @a[tag=tag] effect_name duration_in_ticks amplifier
6. Do NOT invent custom inventory slots - use /give for items
7. Example: give @a[tag=hider] stick{Enchantments:[{id:"minecraft:sharpness",lvl:255}]}

COMMAND SYNTAX REFERENCE:
- Selectors: @a (all players), @p (nearest), @e (entities), @s (executor in context)
- NBT numbers: 1b (byte), 1s (short), 1l (long), but use plain ints for enchantment levels
- Give item: give @p minecraft:stick{Enchantments:[{id:"minecraft:sharpness",lvl:5}]}
- Effect: effect give @a minecraft:speed 60 1 true (60 seconds, amplifier 1, hide particles)
- Execute: execute as @a at @s run say hello
- Condition: execute if score @s health matches 1..10 run say Low health!

FUNCTION TAGS:
1. Use minecraft:load tag for initialization functions (runs once when datapack loads)
2. Use minecraft:tick tag for functions that run every game tick
3. Create proper function tags in tags/function/ directory
4. List function names in the tag's "values" array

GENERAL BEST PRACTICES:
1. Use namespaced IDs (e.g., mydatapack:myfunction, not just myfunction)
2. Create clean folder structure under your namespace
3. Use execute commands properly with correct syntax for the format version
4. For old format: use 'if block' / 'unless block', for new format: similar but check syntax
5. Always handle edge cases (e.g., check if scoreboard exists, handle empty selections)
6. Use storage commands when appropriate for complex data
7. Comment your code with # comments in .mcfunction files
8. Create a load.mcfunction that initializes everything
9. Use fake players prefixed with # for scoreboard operations
10. For effects: use 'true' as last parameter to hide particles if needed

RESPONSE FORMAT:
Return ONLY valid JSON with this structure:
{
  "files": [
    {"path": "data/mydatapack/functions/load.mcfunction", "content": "# Load function\nscoreboard objectives add my_objective dummy"},
    {"path": "data/mydatapack/functions/tick.mcfunction", "content": "# Tick function"},
    {"path": "data/mydatapack/tags/function/load.json", "content": "{\\\"values\\\":[\\\"mydatapack:load\\\"]}"},
    {"path": "data/mydatapack/tags/function/tick.json", "content": "{\\\"values\\\":[\\\"mydatapack:tick\\\"]}"}
  ],
  "namespace": "mydatapack",
  "description": "Brief description of the datapack",
  "load_functions": ["mydatapack:load"],
  "tick_functions": ["mydatapack:tick"]
}

IMPORTANT: The paths should use:
- "functions" folder for old format
- "function" folder for new format (1.21+)

Generate complete, working datapack files based on the user's request.
Return ONLY the JSON, no additional text.`)
}

// buildUserPrompt builds the user prompt
func (g *Generator) buildUserPrompt() string {
	return fmt.Sprintf(`Format version: Minecraft %s
Pack format number: %d
Functions folder name: %s

User request: %s

Generate a complete datapack following the best practices mentioned above. Return ONLY valid JSON.`,
		g.config.Format.GetVersion(),
		g.config.Format.GetPackFormat(),
		g.config.Format.GetFunctionsDir(),
		g.config.Prompt)
}

// parseResponse parses the AI response into a DatapackSpec
func (g *Generator) parseResponse(response string) (*datapack.DatapackSpec, error) {
	// Try to find JSON in the response
	startIdx := strings.Index(response, "{")
	endIdx := strings.LastIndex(response, "}")

	if startIdx == -1 || endIdx == -1 || startIdx >= endIdx {
		return nil, fmt.Errorf("no valid JSON found in response")
	}

	jsonStr := response[startIdx : endIdx+1]

	var spec datapack.DatapackSpec
	if err := json.Unmarshal([]byte(jsonStr), &spec); err != nil {
		// Log the invalid JSON for debugging
		fmt.Fprintf(os.Stderr, "Debug: Invalid JSON received:\n%s\n\n", jsonStr[:min(len(jsonStr), 500)])
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &spec, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// GenerateError represents a datapack generation error
type GenerateError struct {
	Op  string
	Err error
}

func (e *GenerateError) Error() string {
	return fmt.Sprintf("generator %s: %v", e.Op, e.Err)
}

func (e *GenerateError) Unwrap() error {
	return e.Err
}
