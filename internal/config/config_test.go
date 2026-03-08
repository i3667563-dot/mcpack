package config

import (
	"testing"
)

func TestFormatType(t *testing.T) {
	tests := []struct {
		name          string
		format        FormatType
		wantFormat    int
		wantDir       string
		wantVersion   string
		wantString    string
	}{
		{
			name:        "old format",
			format:      FormatOld,
			wantFormat:  OldPackFormat,
			wantDir:     OldFunctionsDir,
			wantVersion: "1.16.5",
			wantString:  "old (<=1.20)",
		},
		{
			name:        "new format",
			format:      FormatNew,
			wantFormat:  NewPackFormat,
			wantDir:     NewFunctionsDir,
			wantVersion: "1.21.1",
			wantString:  "new (>=1.21)",
		},
		{
			name:        "auto format",
			format:      FormatAuto,
			wantFormat:  OldPackFormat,
			wantDir:     OldFunctionsDir,
			wantVersion: "1.16.5",
			wantString:  "auto (default to old)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.format.GetPackFormat(); got != tt.wantFormat {
				t.Errorf("GetPackFormat() = %v, want %v", got, tt.wantFormat)
			}
			if got := tt.format.GetFunctionsDir(); got != tt.wantDir {
				t.Errorf("GetFunctionsDir() = %v, want %v", got, tt.wantDir)
			}
			if got := tt.format.GetVersion(); got != tt.wantVersion {
				t.Errorf("GetVersion() = %v, want %v", got, tt.wantVersion)
			}
			if got := tt.format.String(); got != tt.wantString {
				t.Errorf("String() = %v, want %v", got, tt.wantString)
			}
		})
	}
}

func TestConfig(t *testing.T) {
	t.Run("NewConfig with options", func(t *testing.T) {
		cfg := NewConfig(
			WithFormat(FormatNew),
			WithPrompt("test prompt"),
			WithModel("test-model"),
			WithOutputDir("/test"),
			WithOllamaURL("http://test:11434"),
			WithVerbose(true),
			WithDryRun(true),
		)

		if cfg.Format != FormatNew {
			t.Errorf("Format = %v, want %v", cfg.Format, FormatNew)
		}
		if cfg.Prompt != "test prompt" {
			t.Errorf("Prompt = %v, want %v", cfg.Prompt, "test prompt")
		}
		if cfg.Model != "test-model" {
			t.Errorf("Model = %v, want %v", cfg.Model, "test-model")
		}
		if cfg.OutputDir != "/test" {
			t.Errorf("OutputDir = %v, want %v", cfg.OutputDir, "/test")
		}
		if cfg.OllamaURL != "http://test:11434" {
			t.Errorf("OllamaURL = %v, want %v", cfg.OllamaURL, "http://test:11434")
		}
		if !cfg.Verbose {
			t.Error("Verbose = false, want true")
		}
		if !cfg.DryRun {
			t.Error("DryRun = false, want true")
		}
	})

	t.Run("Validate empty prompt", func(t *testing.T) {
		cfg := NewConfig()
		if err := cfg.Validate(); err == nil {
			t.Error("Validate() expected error for empty prompt")
		}
	})

	t.Run("Validate with prompt", func(t *testing.T) {
		cfg := NewConfig(WithPrompt("test"))
		if err := cfg.Validate(); err != nil {
			t.Errorf("Validate() unexpected error: %v", err)
		}
	})

	t.Run("Validate sets defaults", func(t *testing.T) {
		cfg := NewConfig(WithPrompt("test"))
		_ = cfg.Validate()

		if cfg.Model != DefaultModel {
			t.Errorf("Model = %v, want %v", cfg.Model, DefaultModel)
		}
		if cfg.OutputDir != DefaultOutputDir {
			t.Errorf("OutputDir = %v, want %v", cfg.OutputDir, DefaultOutputDir)
		}
		if cfg.OllamaURL != DefaultOllamaURL {
			t.Errorf("OllamaURL = %v, want %v", cfg.OllamaURL, DefaultOllamaURL)
		}
	})
}
