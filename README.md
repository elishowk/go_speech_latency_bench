# Speech Latency Benchmark

A command-line tool for measuring speech processing latency using various providers:
 - Deepgram nova-3 model.
 - TODO : more !

## Features

- Measure speech processing latency with Deepgram nova-3
- WAV file streaming simulation
- Configurable audio chunk processing
- Environment variable configuration
- Real-time transcription and metrics

## Installation

```bash
git clone https://github.com/elishowk/go_speech_latency_bench.git
cd go_speech_latency_bench
```

## Configuration

Create a `.env` file in the project root:

```env
DEEPGRAM_API_KEY=your_deepgram_api_key_here
DEFAULT_PROVIDER=deepgram
DEFAULT_LANGUAGE=en-US
DEFAULT_CHUNK_SIZE=4096
DEFAULT_CHUNK_INTERVAL=100
```

## Usage

```bash
# Run benchmark with default settings
go run cmd/speech_latency/main.go benchmark -a audio.wav

# Run benchmark with custom parameters
go run cmd/speech_latency/main.go benchmark -a audio.wav -p deepgram -l en-US -s 8192 -i 50

# Show version
go run cmd/speech_latency/main.go version
```

### Command Line Options

- `-a, --audio`: Path to the WAV audio file (required)
- `-p, --provider`: Speech recognition provider (default: deepgram)
- `-l, --language`: Language code (default: en-US)
- `-s, --chunk-size`: Size of audio chunks in bytes (default: 4096)
- `-i, --chunk-interval`: Interval between chunks in milliseconds (default: 100)
- `--interim`: Enable interim results (default: true)
- `--punctuate`: Enable punctuation (default: true)
- `--smart-format`: Enable smart formatting (default: true)

## Project Structure

```
.
├── cmd/
│   └── speech_latency/    # CLI application
├── pkg/
│   ├── audio/            # WAV file streaming
│   └── providers/        # Speech recognition providers
│       └── deepgram/     # Deepgram provider implementation
├── internal/
│   └── config/          # Environment configuration
├── audio.wav            # Sample audio file
└── .env                 # Environment variables
```

## Example Output

```
Audio format: 44100 Hz, 1 channels
Starting benchmark with deepgram provider...
Transcription: Split infinity. In a time when less is more...
First Word: split
Is Final : false
First word latency: 4599.87 ms
Throughput: 9.13 words/second
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details. 
