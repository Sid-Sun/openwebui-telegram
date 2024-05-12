package main

import (
	"github.com/sid-sun/openwebui-bot/cmd/config"
	"github.com/sid-sun/openwebui-bot/pkg/bot"
)

func main() {
	cfg := config.Load()
	bot.StartBot(cfg)
}
