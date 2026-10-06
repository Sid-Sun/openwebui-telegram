package config

// Ollama holds the config for the native Ollama backend (direct daemon
// access via github.com/ollama/ollama/api, i.e. POST /api/chat).
type Ollama struct {
	// Endpoint is the base host of the Ollama daemon, e.g.
	// "http://localhost:11434". Unlike the OpenAI-compatible config, this is
	// NOT a /v1/ suffixed path.
	Endpoint string
}
