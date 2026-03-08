// mcpack - AI-powered Minecraft datapack generator
//
// Generates professional-quality Minecraft datapacks using Ollama AI.
// Supports both old format (<=1.20) and new format (>=1.21).
package main

import (
	"fmt"
	"os"

	"mcpack/cmd"
)

func main() {
	if err := cmd.Run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
