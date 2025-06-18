package main

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/elishowk/speech_latency/internal/config"
	"github.com/spf13/cobra"
)

// Test helper functions
func executeCommand(cmd *cobra.Command, args ...string) (string, error) {
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs(args)
	
	err := cmd.Execute()
	return strings.TrimSpace(buf.String()), err
}

func TestBenchmarkCommand_RequiredAudioFlag(t *testing.T) {
	// Test that audio flag is required
	_, err := executeCommand(rootCmd, "benchmark")
	if err == nil {
		t.Error("expected error when audio flag is missing")
	}
	
	if !strings.Contains(err.Error(), "required flag(s) \"audio\" not set") {
		t.Errorf("expected audio flag required error, got: %v", err)
	}
}

func TestEnvironmentVariableDefaults(t *testing.T) {
	// Test that environment variables are used as defaults
	originalProvider := os.Getenv("DEFAULT_PROVIDER")
	originalLanguage := os.Getenv("DEFAULT_LANGUAGE")
	originalChunkSize := os.Getenv("DEFAULT_CHUNK_SIZE")
	originalChunkInterval := os.Getenv("DEFAULT_CHUNK_INTERVAL")
	
	// Set test environment variables
	os.Setenv("DEFAULT_PROVIDER", "test-provider")
	os.Setenv("DEFAULT_LANGUAGE", "test-lang")
	os.Setenv("DEFAULT_CHUNK_SIZE", "1024")
	os.Setenv("DEFAULT_CHUNK_INTERVAL", "50")
	
	// Test getEnvInt function
	if result := getEnvInt("DEFAULT_CHUNK_SIZE", 4096); result != 1024 {
		t.Errorf("expected getEnvInt to return 1024, got %d", result)
	}
	
	if result := getEnvInt("NONEXISTENT_VAR", 4096); result != 4096 {
		t.Errorf("expected getEnvInt to return default 4096, got %d", result)
	}
	
	// Test config.GetEnvWithDefault function
	if result := config.GetEnvWithDefault("DEFAULT_PROVIDER", "deepgram"); result != "test-provider" {
		t.Errorf("expected GetEnvWithDefault to return test-provider, got %s", result)
	}
	
	if result := config.GetEnvWithDefault("NONEXISTENT_VAR", "default"); result != "default" {
		t.Errorf("expected GetEnvWithDefault to return default, got %s", result)
	}
	
	// Restore original environment variables
	if originalProvider != "" {
		os.Setenv("DEFAULT_PROVIDER", originalProvider)
	} else {
		os.Unsetenv("DEFAULT_PROVIDER")
	}
	if originalLanguage != "" {
		os.Setenv("DEFAULT_LANGUAGE", originalLanguage)
	} else {
		os.Unsetenv("DEFAULT_LANGUAGE")
	}
	if originalChunkSize != "" {
		os.Setenv("DEFAULT_CHUNK_SIZE", originalChunkSize)
	} else {
		os.Unsetenv("DEFAULT_CHUNK_SIZE")
	}
	if originalChunkInterval != "" {
		os.Setenv("DEFAULT_CHUNK_INTERVAL", originalChunkInterval)
	} else {
		os.Unsetenv("DEFAULT_CHUNK_INTERVAL")
	}
}

func TestInvalidFlagValues(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "invalid chunk size",
			args: []string{"benchmark", "-a", "../../audio.wav", "-s", "invalid"},
		},
		{
			name: "invalid chunk interval",
			args: []string{"benchmark", "-a", "../../audio.wav", "-i", "invalid"},
		},
		{
			name: "negative chunk size",
			args: []string{"benchmark", "-a", "../../audio.wav", "-s", "-1"},
		},
		{
			name: "negative chunk interval",
			args: []string{"benchmark", "-a", "../../audio.wav", "-i", "-1"},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := executeCommand(rootCmd, tt.args...)
			if err == nil {
				t.Errorf("expected error for invalid flag value")
			}
		})
	}
} 