package audio

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"time"
)

const (
	// DefaultChunkSize is the default size of audio chunks in bytes
	DefaultChunkSize = 4096
	// DefaultChunkInterval is the default interval between chunks in milliseconds
	DefaultChunkInterval = 100
)

// WAVStreamer streams a WAV file in chunks to simulate real-time audio capture
type WAVStreamer struct {
	file           *os.File
	chunkSize      int
	chunkInterval  time.Duration
	headerSize     int
	sampleRate     int
	bytesPerSample int
	channels       int
}

// NewWAVStreamer creates a new WAV file streamer
func NewWAVStreamer(filePath string, chunkSize int, chunkInterval time.Duration) (*WAVStreamer, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open WAV file: %w", err)
	}

	// Read WAV header
	header := make([]byte, 44)
	if _, err := file.Read(header); err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to read WAV header: %w", err)
	}

	// Verify WAV format
	if string(header[0:4]) != "RIFF" || string(header[8:12]) != "WAVE" {
		file.Close()
		return nil, fmt.Errorf("invalid WAV file format")
	}

	// Get audio format details
	channels := int(binary.LittleEndian.Uint16(header[22:24]))
	sampleRate := int(binary.LittleEndian.Uint32(header[24:28]))
	bitsPerSample := int(binary.LittleEndian.Uint16(header[34:36]))
	bytesPerSample := bitsPerSample / 8

	return &WAVStreamer{
		file:           file,
		chunkSize:      chunkSize,
		chunkInterval:  chunkInterval,
		headerSize:     44,
		sampleRate:     sampleRate,
		bytesPerSample: bytesPerSample,
		channels:       channels,
	}, nil
}

// Stream streams the WAV file in chunks
func (w *WAVStreamer) Stream() (io.Reader, error) {
	// Seek to the start of audio data
	if _, err := w.file.Seek(int64(w.headerSize), io.SeekStart); err != nil {
		return nil, fmt.Errorf("failed to seek to audio data: %w", err)
	}

	return &wavChunkReader{
		file:          w.file,
		chunkSize:     w.chunkSize,
		chunkInterval: w.chunkInterval,
	}, nil
}

// Close closes the WAV file
func (w *WAVStreamer) Close() error {
	return w.file.Close()
}

// GetAudioFormat returns the audio format details
func (w *WAVStreamer) GetAudioFormat() (sampleRate, channels, bytesPerSample int) {
	return w.sampleRate, w.channels, w.bytesPerSample
}

// wavChunkReader implements io.Reader to stream WAV data in chunks
type wavChunkReader struct {
	file          *os.File
	chunkSize     int
	chunkInterval time.Duration
}

func (r *wavChunkReader) Read(p []byte) (n int, err error) {
	// Read a chunk of audio data, but don't exceed the buffer size
	chunkSize := r.chunkSize
	if chunkSize > len(p) {
		chunkSize = len(p)
	}
	
	n, err = r.file.Read(p[:chunkSize])
	if err != nil && err != io.EOF {
		return n, err
	}

	// If we read any data, wait for the chunk interval
	if n > 0 {
		time.Sleep(r.chunkInterval)
	}

	return n, err
}

// GetFile returns the underlying file reader for direct access
func (r *wavChunkReader) GetFile() io.Reader {
	// Reset file to beginning and return it
	r.file.Seek(0, 0)
	return r.file
} 