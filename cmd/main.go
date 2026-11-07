package main

import (
	"os"
	"strings"

	"github.com/sid-sun/openwebui-bot/cmd/config"
	"github.com/sid-sun/openwebui-bot/pkg/bot"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	useSEConfig := strings.ToUpper(os.Getenv("SE_CONFIG_ENABLED"))
	if useSEConfig == "TRUE" {
		cfg, err = config.LoadConfigFromSE()
		if err != nil {
			panic(err)
		}
	}

	bot.StartBot(cfg)
}
