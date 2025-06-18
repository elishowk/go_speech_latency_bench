package providers

import (
	"context"
	"fmt"
	"io"

	"github.com/elishowk/speech_latency/pkg/providers/deepgram"
)

// Provider defines the interface that all speech recognition providers must implement
type Provider interface {
	// StreamAudio processes an audio stream and returns latency metrics
	StreamAudio(ctx context.Context, audioReader io.Reader) (*Result, error)
}

// Config holds common configuration for all providers
type Config struct {
	SampleRate  int
	Channels    int
	Language    string
	Interim     bool
	Punctuate   bool
	SmartFormat bool
}

// Result contains the benchmark results
type Result struct {
	Latency    float64 // in milliseconds
	Throughput float64 // words per second
}

// deepgramAdapter adapts deepgram.Provider to implement the Provider interface
type deepgramAdapter struct {
	provider *deepgram.Provider
}

func (a *deepgramAdapter) StreamAudio(ctx context.Context, audioReader io.Reader) (*Result, error) {
	result, err := a.provider.StreamAudio(ctx, audioReader)
	if err != nil {
		return nil, err
	}
	return &Result{
		Latency:    result.Latency,
		Throughput: result.Throughput,
	}, nil
}

// Factory creates provider instances
type Factory struct {
	providers map[string]func(*Config, string) (Provider, error)
}

// NewFactory creates a new provider factory
func NewFactory() *Factory {
	f := &Factory{
		providers: make(map[string]func(*Config, string) (Provider, error)),
	}
	
	// Register providers
	f.RegisterProvider("deepgram", func(config *Config, apiKey string) (Provider, error) {
		dgConfig := &deepgram.Config{
			SampleRate:  config.SampleRate,
			Channels:    config.Channels,
			Language:    config.Language,
			Interim:     config.Interim,
			Punctuate:   config.Punctuate,
			SmartFormat: config.SmartFormat,
		}
		dgProvider, err := deepgram.NewProvider(dgConfig, apiKey)
		if err != nil {
			return nil, err
		}
		return &deepgramAdapter{provider: dgProvider}, nil
	})
	
	return f
}

// RegisterProvider registers a new provider with the factory
func (f *Factory) RegisterProvider(name string, factory func(*Config, string) (Provider, error)) {
	f.providers[name] = factory
}

// CreateProvider creates a new provider instance
func (f *Factory) CreateProvider(name string, config *Config, apiKey string) (Provider, error) {
	factory, ok := f.providers[name]
	if !ok {
		return nil, fmt.Errorf("unknown provider: %s", name)
	}
	return factory(config, apiKey)
} 