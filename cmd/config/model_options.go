package config

type ModelOptions struct {
	Model           string
	modelTweakLevel string
}

func (m ModelOptions) UseMinimalTweaks() bool {
	return m.modelTweakLevel != "advanced"
}
