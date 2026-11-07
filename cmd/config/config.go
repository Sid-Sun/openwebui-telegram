package config

import (
	"bytes"
	"fmt"
	"sync/atomic"

	"github.com/fitant/storage-engine-go/storageengine"
	"github.com/spf13/viper"
)

var globalConfig atomic.Pointer[Config]

func updateConfig(cfg Config) {
	globalConfig.Store(&cfg)
}

func GetConfig() *Config {
	return globalConfig.Load()
}

var seObject *storageengine.Object

// Config contains all the neccessary configurations
type Config struct {
	Models     map[string]Model
	ModelNames []string
	Backends   map[string]Backend
	Bot        BotConfig
}

func parseConfig(v *viper.Viper) (Config, error) {
	backends := make(map[string]Backend)
	err := v.UnmarshalKey("backends", &backends)
	if err != nil {
		return Config{}, err
	}

	for backendName, backend := range backends {
		backends[backendName] = backend.Validate(backendName)
	}

	var modelList []Model
	err = v.UnmarshalKey("models", &modelList)
	if err != nil {
		return Config{}, err
	}

	// Initialize modelNames and models from modelList
	modelNames := make([]string, len(modelList))
	models := make(map[string]Model)
	for i, model := range modelList {
		if _, ok := models[model.Name]; ok {
			return Config{}, fmt.Errorf("duplicate model: %s", model.Name)
		}
		err = model.Validate(backends)
		if err != nil {
			return Config{}, err
		}
		modelNames[i] = model.Name
		models[model.Name] = model
	}
	if _, ok := models["default"]; !ok {
		return Config{}, fmt.Errorf("default model not found")
	}

	return Config{
		Bot: BotConfig{
			tkn: v.GetString("api_token"),
		},
		Backends:   backends,
		Models:     models,
		ModelNames: modelNames,
	}, nil
}

// Load reads all config from env to config
func Load() (Config, error) {
	v := viper.New()
	v.SetConfigName("config")      // name of config file (without extension)
	v.SetConfigType("yaml")        // REQUIRED if the config file does not have the extension in the name
	v.AddConfigPath(".")           // optionally look for config in the working directory
	v.AddConfigPath("config")      // optionally look for config in the working directory
	v.AddConfigPath("data")        // optionally look for config in the working directory
	v.AddConfigPath("data/config") // optionally look for config in the working directory
	if err := v.ReadInConfig(); err != nil {
		return Config{}, fmt.Errorf("read local config: %w", err)
	}
	cfg, err := parseConfig(v)
	if err != nil {
		return Config{}, err
	}
	updateConfig(cfg)
	return cfg, nil
}

func LoadConfigFromSE() (Config, error) {
	if seObject == nil {
		obj, err := initSEObj()
		if err != nil {
			return Config{}, err
		}
		seObject = obj
	} else {
		if err := seObject.Refresh(); err != nil {
			return Config{}, err
		}
	}

	v := viper.New()
	v.SetConfigType("yaml")
	seCfgStr := seObject.GetData()
	if err := v.ReadConfig(bytes.NewReader([]byte(seCfgStr))); err != nil {
		return Config{}, fmt.Errorf("read local config: %w", err)
	}

	cfg, err := parseConfig(v)
	if err != nil {
		return Config{}, err
	}

	updateConfig(cfg)
	return cfg, nil
}
