package config

import "github.com/sid-sun/openwebui-bot/pkg/bot/contract"

var _ = Model{
	Name:            "default",
	Model:           "llama3:instruct",
	Backend:         "openai",
	modelTweakLevel: "basic",
	Tweaks: contract.ModelTweaks{
		ContextLength:    8192,
		MaxTokens:        1024,
		Temperature:      0.8,
		RepeatPenalty:    1.2,
		PresencePenalty:  1.5,
		FrequencyPenalty: 1.0,
	},
}

type Model struct {
	Name  string `mapstructure:"name"`
	Model string `mapstructure:"model"`
	// Backend selects which completion backend serves this model: "openai"
	// (OpenAI-compatible chat/completions endpoint) or "ollama" (native
	// Ollama daemon via POST /api/chat). Empty defaults to "openai".
	Backend         string               `mapstructure:"backend"`
	modelTweakLevel string               `mapstructure:"tweak_level"`
	Tweaks          contract.ModelTweaks `mapstructure:"tweaks"`
	// ReasoningEffort is passed through verbatim as "reasoning_effort" to the
	// backend for models that support toggling/tuning extended thinking (e.g.
	// "none" to disable thinking, or "low"/"medium"/"high"). Leave empty to
	// omit the field and let the backend/model use its own default.
	ReasoningEffort string `mapstructure:"reasoning_effort"`
}

func (m Model) UseMinimalTweaks() bool {
	return m.modelTweakLevel != "advanced"
}

func (m Model) Validate(backends map[string]Backend) Model {
	if _, ok := backends[m.Backend]; !ok {
		panic("Invalid backend: " + m.Backend + " for model: " + m.Name + ", backend is not defined")
	}
	return m
}

func (m Model) GetAdvancedTweaks() contract.ModelTweaks {
	return m.Tweaks
}

func (m Model) GetBasicTweaks() contract.BasicModelTweaks {
	return contract.BasicModelTweaks{
		ContextLength: m.Tweaks.ContextLength,
		MaxTokens:     m.Tweaks.MaxTokens,
		Temperature:   m.Tweaks.Temperature,
	}
}
