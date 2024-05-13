package config

import (
	"github.com/sid-sun/openwebui-bot/pkg/bot/contract"
	"github.com/spf13/viper"
)

var GlobalConfig Config

// Config contains all the neccessary configurations
type Config struct {
	ModelTweaks contract.ModelTweaks
	ModelOpts   ModelOptions
	OpenAIAPI   OpenAI
	Bot         BotConfig
}

// Load reads all config from env to config
func Load() Config {
	viper.AutomaticEnv()

	// Model Tweaks
	viper.SetDefault("MAX_TOKENS", 1024)
	viper.SetDefault("TEMPERATURE", 0.8)
	viper.SetDefault("REPEAT_PENALTY", 1.2)
	viper.SetDefault("CONTEXT_LENGTH", 8192)
	viper.SetDefault("PRESENCE_PENALTY", 1.5)
	viper.SetDefault("FREQUENCY_PENALTY", 1.0)
	// Model Options
	viper.SetDefault("MODEL", "llama3:instruct")
	viper.SetDefault("MODEL_TWEAK_LEVEL", "minimal")

	GlobalConfig = Config{
		Bot: BotConfig{
			tkn: viper.GetString("API_TOKEN"),
		},
		OpenAIAPI: OpenAI{
			Endpoint: viper.GetString("OPENAI_ENDPOINT"),
			APIKey:   viper.GetString("OPENAI_API_KEY"),
		},
		ModelOpts: ModelOptions{
			Model:           viper.GetString("MODEL"),
			modelTweakLevel: viper.GetString("MODEL_TWEAK_LEVEL"),
		},
		ModelTweaks: contract.ModelTweaks{
			ContextLength:    viper.GetInt("CONTEXT_LENGTH"),
			MaxTokens:        viper.GetInt("MAX_TOKENS"),
			FrequencyPenalty: viper.GetFloat64("FREQUENCY_PENALTY"),
			PresencePenalty:  viper.GetFloat64("PRESENCE_PENALTY"),
			Temperature:      viper.GetFloat64("TEMPERATURE"),
			RepeatPenalty:    viper.GetFloat64("REPEAT_PENALTY"),
		},
	}
	return GlobalConfig
}
