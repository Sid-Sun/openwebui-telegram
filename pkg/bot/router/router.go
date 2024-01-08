package router

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sid-sun/sample-bot/cmd/config"
	"github.com/sid-sun/sample-bot/pkg/bot/handlers/repeat"
	"go.uber.org/zap"
)

type updates struct {
	ch     tgbotapi.UpdatesChannel
	bot    *tgbotapi.BotAPI
	logger *zap.Logger
}

// ListenAndServe starts listens on the update channel and handles routing the update to handlers
func (u updates) ListenAndServe() {
	u.logger.Info(fmt.Sprintf("[StartBot] Started Bot: %s", u.bot.Self.FirstName))
	for update := range u.ch {
		update := update
		go func() {
			if update.Message == nil {
				return
			}
			repeat.Handler(u.bot, update, u.logger)
		}()
	}
}

type bot struct {
	bot    *tgbotapi.BotAPI
	logger *zap.Logger
}

// NewUpdateChan creates a new channel to get update
func (b bot) NewUpdateChan() updates {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	ch := b.bot.GetUpdatesChan(u)
	return updates{ch: ch, bot: b.bot, logger: b.logger}
}

// New returns a new instance of the router
func New(cfg config.BotConfig, logger *zap.Logger) bot {
	b, err := tgbotapi.NewBotAPI(cfg.Token())
	if err != nil {
		panic(err)
	}
	return bot{
		bot:    b,
		logger: logger,
	}
}
