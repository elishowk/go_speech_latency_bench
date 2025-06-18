package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/elishowk/speech_latency/internal/config"
	"github.com/elishowk/speech_latency/pkg/audio"
	"github.com/elishowk/speech_latency/pkg/providers"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "speech_latency",
	Short: "A CLI tool for measuring speech processing latency",
	Long: `speech_latency is a command-line tool designed to measure and benchmark
speech processing latency across different models and configurations.`,
}

func init() {
	// Load environment variables from .env file
	if err := config.LoadEnv(); err != nil {
		fmt.Printf("Warning: Failed to load .env file: %v\n", err)
	}

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(benchmarkCmd)

	// Add flags for the benchmark command
	benchmarkCmd.Flags().StringP("provider", "p", config.GetEnvWithDefault("DEFAULT_PROVIDER", "deepgram"), "Speech recognition provider (deepgram, etc.)")
	benchmarkCmd.Flags().StringP("audio", "a", "", "Path to the WAV audio file")
	benchmarkCmd.Flags().IntP("chunk-size", "s", getEnvInt("DEFAULT_CHUNK_SIZE", audio.DefaultChunkSize), "Size of audio chunks in bytes")
	benchmarkCmd.Flags().IntP("chunk-interval", "i", getEnvInt("DEFAULT_CHUNK_INTERVAL", int(audio.DefaultChunkInterval/time.Millisecond)), "Interval between chunks in milliseconds")
	benchmarkCmd.Flags().StringP("language", "l", config.GetEnvWithDefault("DEFAULT_LANGUAGE", "en-US"), "Language code")
	benchmarkCmd.Flags().Bool("interim", true, "Enable interim results")
	benchmarkCmd.Flags().Bool("punctuate", true, "Enable punctuation")
	benchmarkCmd.Flags().Bool("smart-format", true, "Enable smart formatting")
	benchmarkCmd.MarkFlagRequired("audio")
}

// getEnvInt gets an integer environment variable with a default value
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("speech_latency v0.1.0")
	},
}

var benchmarkCmd = &cobra.Command{
	Use:   "benchmark",
	Short: "Run a speech latency benchmark",
	Run: func(cmd *cobra.Command, args []string) {
		// Get command line flags
		providerName, _ := cmd.Flags().GetString("provider")
		audioPath, _ := cmd.Flags().GetString("audio")
		chunkSize, _ := cmd.Flags().GetInt("chunk-size")
		chunkInterval, _ := cmd.Flags().GetInt("chunk-interval")
		language, _ := cmd.Flags().GetString("language")
		interim, _ := cmd.Flags().GetBool("interim")
		punctuate, _ := cmd.Flags().GetBool("punctuate")
		smartFormat, _ := cmd.Flags().GetBool("smart-format")

		// Create WAV streamer
		streamer, err := audio.NewWAVStreamer(audioPath, chunkSize, time.Duration(chunkInterval)*time.Millisecond)
		if err != nil {
			fmt.Printf("Error creating WAV streamer: %v\n", err)
			os.Exit(1)
		}
		defer streamer.Close()

		// Get audio format details
		sampleRate, channels, _ := streamer.GetAudioFormat()
		fmt.Printf("Audio format: %d Hz, %d channels\n", sampleRate, channels)

		// Create provider configuration
		providerConfig := &providers.Config{
			SampleRate:  sampleRate,
			Channels:    channels,
			Language:    language,
			Interim:     interim,
			Punctuate:   punctuate,
			SmartFormat: smartFormat,
		}

		// Get provider API key
		apiKey, err := config.GetProviderAPIKey(providerName)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		// Create provider factory and get provider
		factory := providers.NewFactory()
		provider, err := factory.CreateProvider(providerName, providerConfig, apiKey)
		if err != nil {
			fmt.Printf("Error creating provider: %v\n", err)
			os.Exit(1)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Get audio stream
		audioStream, err := streamer.Stream()
		if err != nil {
			fmt.Printf("Error getting audio stream: %v\n", err)
			os.Exit(1)
		}

		// Run benchmark
		fmt.Printf("Starting benchmark with %s provider...\n", providerName)
		result, err := provider.StreamAudio(ctx, audioStream)
		if err != nil {
			fmt.Printf("Error streaming audio: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("First word latency: %.2f ms\n", result.Latency)
		fmt.Printf("Throughput: %.2f words/second\n", result.Throughput)
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
} 