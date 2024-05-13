package config

// BotConfig contains the config for creating a new bot
type BotConfig struct {
	tkn string
}

// Token returns the bot token
func (b BotConfig) Token() string {
	return b.tkn
}
