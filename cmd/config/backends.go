package config

type Backend struct {
	Type     string
	Endpoint string `mapstructure:"endpoint"`
	APIKey   string `mapstructure:"api_key"`
}

func (b Backend) Validate(name string) Backend {
	if b.Type != "openai" && b.Type != "ollama" {
		panic("invalid backend type: " + b.Type + ", it must be \"openai\" or \"ollama\"")
	}
	if b.Endpoint == "" {
		panic("backend: " + name + "'s Endpoint is empty")
	}
	if b.Type == "openai" {
		if b.APIKey == "" {
			panic("backend: " + name + "'s API Key is empty")
		}
	}
	return b
}
