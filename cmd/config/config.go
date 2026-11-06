package config

import (
	"github.com/spf13/viper"
)

var GlobalConfig Config

// Config contains all the neccessary configurations
type Config struct {
	Models     map[string]Model
	ModelNames []string
	Backends   map[string]Backend
	// OpenAIModels map[string]OpenAI
	// OllamaModels map[string]Ollama
	Bot BotConfig
}

// Load reads all config from env to config
func Load() Config {
	viper.AutomaticEnv()

	viper.SetConfigName("config")      // name of config file (without extension)
	viper.SetConfigType("yaml")        // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath(".")           // optionally look for config in the working directory
	viper.AddConfigPath("config")      // optionally look for config in the working directory
	viper.AddConfigPath("data")        // optionally look for config in the working directory
	viper.AddConfigPath("data/config") // optionally look for config in the working directory
	viper.ReadInConfig()               // Find and read the config file

	backends := make(map[string]Backend)
	err := viper.UnmarshalKey("backends", &backends)
	if err != nil {
		panic(err)
	}

	for backendName, backend := range backends {
		backends[backendName] = backend.Validate(backendName)
	}

	modelList := make([]Model, 1)
	err = viper.UnmarshalKey("models", &modelList)
	if err != nil {
		panic(err)
	}

	// Initialize modelNames and models from modelList
	modelNames := make([]string, len(modelList))
	models := make(map[string]Model)
	for i, model := range modelList {
		modelNames[i] = model.Name
		if _, ok := models[model.Name]; ok {
			panic("duplicate model name")
		}
		models[model.Name] = model.Validate(backends)
	}
	if _, ok := models["default"]; !ok {
		panic("default model not found")
	}

	GlobalConfig = Config{
		Bot: BotConfig{
			tkn: viper.GetString("api_token"),
		},
		Backends:   backends,
		Models:     models,
		ModelNames: modelNames,
	}
	return GlobalConfig
}
