package config

import (
	"os"
)

var GlobalConfig Config

// Config contains all the neccessary configurations
type Config struct {
	OpenAIAPI   OpenAI
	Bot         BotConfig
	environment string
}

// GetEnv returns the current developemnt environment
func (c Config) GetEnv() string {
	return c.environment
}

// Load reads all config from env to config
func Load() Config {
	GlobalConfig = Config{
		environment: os.Getenv("APP_ENV"),
		Bot: BotConfig{
			tkn: os.Getenv("API_TOKEN"),
		},
		OpenAIAPI: OpenAI{
			Endpoint: os.Getenv("OPENAI_ENDPOINT"),
			APIKey:   os.Getenv("OPENAI_API_KEY"),
		},
	}
	return GlobalConfig
}
