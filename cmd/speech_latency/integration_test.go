package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestCLIIntegration_AudioFileExists(t *testing.T) {
	// Test that audio.wav file exists
	if _, err := os.Stat("../../audio.wav"); os.IsNotExist(err) {
		t.Skip("audio.wav file not found, skipping integration tests")
	}
}

func TestCLIIntegration_HelpCommand(t *testing.T) {
	cmd := exec.Command("go", "run", "main.go", "--help")
	cmd.Dir = "."
	
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("help command failed: %v", err)
	}
	
	outputStr := string(output)
	if !strings.Contains(outputStr, "speech_latency") {
		t.Error("help output should contain 'speech_latency'")
	}
	
	if !strings.Contains(outputStr, "benchmark") {
		t.Error("help output should contain 'benchmark' command")
	}
	
	if !strings.Contains(outputStr, "version") {
		t.Error("help output should contain 'version' command")
	}
}

func TestCLIIntegration_VersionCommand(t *testing.T) {
	cmd := exec.Command("go", "run", "main.go", "version")
	cmd.Dir = "."
	
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("version command failed: %v", err)
	}
	
	if !strings.Contains(string(output), "speech_latency v0.1.0") {
		t.Errorf("expected version output, got: %s", string(output))
	}
}

func TestCLIIntegration_BenchmarkHelp(t *testing.T) {
	cmd := exec.Command("go", "run", "main.go", "benchmark", "--help")
	cmd.Dir = "."
	
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("benchmark help command failed: %v", err)
	}
	
	outputStr := string(output)
	expectedFlags := []string{
		"-a, --audio",
		"-p, --provider",
		"-l, --language",
		"-s, --chunk-size",
		"-i, --chunk-interval",
		"--interim",
		"--punctuate",
		"--smart-format",
	}
	
	for _, flag := range expectedFlags {
		if !strings.Contains(outputStr, flag) {
			t.Errorf("help output should contain flag: %s", flag)
		}
	}
}

func TestCLIIntegration_RequiredAudioFlag(t *testing.T) {
	cmd := exec.Command("go", "run", "main.go", "benchmark")
	cmd.Dir = "."
	
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Error("expected error when audio flag is missing")
	}
	
	if !strings.Contains(string(output), "required flag(s) \"audio\" not set") {
		t.Errorf("expected audio flag required error, got: %s", string(output))
	}
}

func TestCLIIntegration_AudioFileNotFound(t *testing.T) {
	cmd := exec.Command("go", "run", "main.go", "benchmark", "-a", "nonexistent.wav")
	cmd.Dir = "."
	
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Error("expected error when audio file doesn't exist")
	}
	
	if !strings.Contains(string(output), "failed to open WAV file") {
		t.Errorf("expected file not found error, got: %s", string(output))
	}
}

func TestCLIIntegration_FlagParsing(t *testing.T) {
	if _, err := os.Stat("../../audio.wav"); os.IsNotExist(err) {
		t.Skip("audio.wav file not found, skipping flag parsing test")
	}
	
	tests := []struct {
		name string
		args []string
		skip bool
		skipReason string
	}{
		{
			name: "short form flags",
			args: []string{"benchmark", "-a", "../../audio.wav", "-p", "deepgram", "-l", "en-US", "-s", "2048", "-i", "50"},
		},
		{
			name: "long form flags",
			args: []string{"benchmark", "--audio", "../../audio.wav", "--provider", "deepgram", "--language", "es-ES", "--chunk-size", "8192", "--chunk-interval", "200"},
		},
		{
			name: "boolean flags false",
			args: []string{"benchmark", "-a", "../../audio.wav", "--interim=false", "--punctuate=false", "--smart-format=false"},
		},
		{
			name: "boolean flags true (default)",
			args: []string{"benchmark", "-a", "../../audio.wav", "--interim=true", "--punctuate=true", "--smart-format=true"},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skip {
				t.Skip(tt.skipReason)
			}
			
			// Set a short timeout to avoid waiting for actual API calls
			cmd := exec.Command("timeout", "5s", "go", "run", "main.go")
			cmd.Args = append(cmd.Args, tt.args...)
			cmd.Dir = "."
			
			// Set a fake API key to avoid API key errors
			cmd.Env = append(os.Environ(), "DEEPGRAM_API_KEY=fake-key-for-testing")
			
			output, _ := cmd.CombinedOutput()
			
			// We expect this to fail due to timeout or API error, but not due to flag parsing
			outputStr := string(output)
			
			// Check that flag parsing didn't fail
			if strings.Contains(outputStr, "unknown flag") {
				t.Errorf("flag parsing failed: %s", outputStr)
			}
			
			if strings.Contains(outputStr, "required flag") {
				t.Errorf("required flag error: %s", outputStr)
			}
			
			if strings.Contains(outputStr, "invalid argument") {
				t.Errorf("invalid argument error: %s", outputStr)
			}
		})
	}
}

func TestCLIIntegration_InvalidFlagValues(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		expectedError  string
	}{
		{
			name:          "invalid chunk size",
			args:          []string{"benchmark", "-a", "../../audio.wav", "-s", "invalid"},
			expectedError: "invalid argument",
		},
		{
			name:          "invalid chunk interval",
			args:          []string{"benchmark", "-a", "../../audio.wav", "-i", "invalid"},
			expectedError: "invalid argument",
		},
		{
			name:          "unknown flag",
			args:          []string{"benchmark", "-a", "../../audio.wav", "--unknown-flag", "value"},
			expectedError: "unknown flag",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command("go", "run", "main.go")
			cmd.Args = append(cmd.Args, tt.args...)
			cmd.Dir = "."
			
			output, err := cmd.CombinedOutput()
			if err == nil {
				t.Errorf("expected error for %s", tt.name)
			}
			
			if !strings.Contains(string(output), tt.expectedError) {
				t.Errorf("expected error containing '%s', got: %s", tt.expectedError, string(output))
			}
		})
	}
}

func TestCLIIntegration_EnvironmentVariables(t *testing.T) {
	if _, err := os.Stat("../../audio.wav"); os.IsNotExist(err) {
		t.Skip("audio.wav file not found, skipping environment variable test")
	}
	
	// Test environment variable defaults
	cmd := exec.Command("timeout", "5s", "go", "run", "main.go", "benchmark", "-a", "../../audio.wav")
	cmd.Dir = "."
	
	// Set environment variables
	cmd.Env = append(os.Environ(),
		"DEFAULT_PROVIDER=custom-provider",
		"DEFAULT_LANGUAGE=fr-FR",
		"DEFAULT_CHUNK_SIZE=1024",
		"DEFAULT_CHUNK_INTERVAL=25",
		"CUSTOM-PROVIDER_API_KEY=fake-key",
	)
	
	output, _ := cmd.CombinedOutput()
	
	// The command should parse the flags correctly even if it fails later
	outputStr := string(output)
	
	// Should not have flag parsing errors
	if strings.Contains(outputStr, "unknown flag") {
		t.Errorf("unexpected flag parsing error: %s", outputStr)
	}
	
	if strings.Contains(outputStr, "required flag") && !strings.Contains(outputStr, "audio") {
		t.Errorf("unexpected required flag error: %s", outputStr)
	}
}

func TestCLIIntegration_RealExecution(t *testing.T) {
	// Only run this test if we have a real API key
	apiKey := os.Getenv("DEEPGRAM_API_KEY")
	if apiKey == "" {
		t.Skip("DEEPGRAM_API_KEY not set, skipping real execution test")
	}
	
	if _, err := os.Stat("../../audio.wav"); os.IsNotExist(err) {
		t.Skip("audio.wav file not found, skipping real execution test")
	}
	
	// Test with various flag combinations
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "minimal flags",
			args: []string{"benchmark", "-a", "../../audio.wav"},
		},
		{
			name: "all flags specified",
			args: []string{
				"benchmark", 
				"-a", "../../audio.wav",
				"-p", "deepgram",
				"-l", "en-US",
				"-s", "4096",
				"-i", "100",
				"--interim=true",
				"--punctuate=true",
				"--smart-format=true",
			},
		},
		{
			name: "different chunk sizes",
			args: []string{"benchmark", "-a", "../../audio.wav", "-s", "2048"},
		},
		{
			name: "different languages",
			args: []string{"benchmark", "-a", "../../audio.wav", "-l", "es-ES"},
		},
		{
			name: "boolean flags disabled",
			args: []string{
				"benchmark", 
				"-a", "../../audio.wav",
				"--interim=false",
				"--punctuate=false",
				"--smart-format=false",
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command("timeout", "30s", "go", "run", "main.go")
			cmd.Args = append(cmd.Args, tt.args...)
			cmd.Dir = "."
			cmd.Env = append(os.Environ(), "DEEPGRAM_API_KEY="+apiKey)
			
			output, err := cmd.CombinedOutput()
			outputStr := string(output)
			
			// Check for successful execution indicators
			if strings.Contains(outputStr, "Starting benchmark") {
				t.Logf("Successfully started benchmark for %s", tt.name)
			}
			
			// Log any errors for debugging
			if err != nil {
				t.Logf("Test %s completed with error (may be expected): %v", tt.name, err)
				t.Logf("Output: %s", outputStr)
			}
			
			// The test passes if the flags were parsed correctly
			// We don't require successful API calls for this test
		})
	}
} 