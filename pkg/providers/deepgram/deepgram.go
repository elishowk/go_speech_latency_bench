package deepgram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Config holds the provider configuration
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

// Provider implements the speech recognition provider using Deepgram
type Provider struct {
	apiKey string
	config *Config
}

// NewProvider creates a new Deepgram provider
func NewProvider(config *Config, apiKey string) (*Provider, error) {
	return &Provider{
		apiKey: apiKey,
		config: config,
	}, nil
}

// DeepgramResponse represents the response from Deepgram API
type DeepgramResponse struct {
	Results struct {
		Channels []struct {
			Alternatives []struct {
				Transcript string `json:"transcript"`
				Words      []struct {
					Word       string  `json:"word"`
					Start      float64 `json:"start"`
					End        float64 `json:"end"`
					Confidence float64 `json:"confidence"`
				} `json:"words"`
			} `json:"alternatives"`
		} `json:"channels"`
		IsFinal bool `json:"is_final"`
	} `json:"results"`
}

// StreamAudio processes an audio stream and measures latency
func (p *Provider) StreamAudio(ctx context.Context, audioReader io.Reader) (*Result, error) {
	startTime := time.Now()
	
	// For REST API, we need the complete audio data
	// If it's a chunked reader, we need to read from the original file
	var audioData []byte
	var err error
	
	// Check if we can get the underlying file for faster reading
	if chunkedReader, ok := audioReader.(interface{ GetFile() io.Reader }); ok {
		audioData, err = io.ReadAll(chunkedReader.GetFile())
	} else {
		audioData, err = io.ReadAll(audioReader)
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to read audio data: %w", err)
	}

	// Create HTTP request
	url := "https://api.deepgram.com/v1/listen"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(audioData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Token "+p.apiKey)
	req.Header.Set("Content-Type", "audio/wav")
	
	// Add query parameters
	q := req.URL.Query()
	q.Add("model", "nova-3")
	q.Add("language", p.config.Language)
	if p.config.Punctuate {
		q.Add("punctuate", "true")
	}
	if p.config.SmartFormat {
		q.Add("smart_format", "true")
	}
	q.Add("words", "true") // Enable word timestamps
	q.Add("interim_results", "true") // Enable word timestamps
	req.URL.RawQuery = q.Encode()

	// Send request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var dgResp DeepgramResponse
	if err := json.NewDecoder(resp.Body).Decode(&dgResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Calculate latency (time to first response)
	latency := time.Since(startTime)

	// Calculate throughput
	var firstWord string
	var wordCount int
	var isFinal bool
	if len(dgResp.Results.Channels) > 0 && len(dgResp.Results.Channels[0].Alternatives) > 0 {
		firstWord = dgResp.Results.Channels[0].Alternatives[0].Words[0].Word
		wordCount = len(dgResp.Results.Channels[0].Alternatives[0].Words)
		isFinal = dgResp.Results.IsFinal
	}

	throughput := float64(wordCount) / latency.Seconds()

	fmt.Printf("Transcription: %s\n", dgResp.Results.Channels[0].Alternatives[0].Transcript)
	fmt.Printf("first Word: %s\n", firstWord)
	fmt.Printf("first Word is_final: %t\n", isFinal)

	return &Result{
		Latency:    float64(latency.Nanoseconds()) / 1e6, // Convert to milliseconds
		Throughput: throughput,
	}, nil
} 