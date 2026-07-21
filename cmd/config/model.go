package config

import "github.com/sid-sun/openwebui-bot/pkg/bot/contract"

var defaultModel = Model{
	Name:            "default",
	Model:           "llama3:instruct",
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
	Name            string               `mapstructure:"name"`
	Model           string               `mapstructure:"model"`
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
