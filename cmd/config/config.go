package config

import (
	"github.com/spf13/viper"
)

var GlobalConfig Config

// Config contains all the neccessary configurations
type Config struct {
	Models     map[string]Model
	ModelNames []string
	OpenAIAPI  OpenAI
	Bot        BotConfig
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

	// Set default models
	viper.SetDefault("models", []Model{
		defaultModel, // set in model.go
	})

	modelList := make([]Model, 1)
	err := viper.UnmarshalKey("models", &modelList)
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
		models[model.Name] = model
	}
	if _, ok := models["default"]; !ok {
		panic("default model not found")
	}

	// print models
	// for name, model := range models {
	// 	fmt.Printf("%s: %+v\n", name, model)
	// }

	GlobalConfig = Config{
		Bot: BotConfig{
			tkn: viper.GetString("api_token"),
		},
		OpenAIAPI: OpenAI{
			Endpoint: viper.GetString("openai.endpoint"),
			APIKey:   viper.GetString("openai.api_key"),
		},
		Models:     models,
		ModelNames: modelNames,
	}
	return GlobalConfig
}
