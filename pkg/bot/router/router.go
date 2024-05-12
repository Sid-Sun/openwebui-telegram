package router

import (
	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sid-sun/openwebui-bot/cmd/config"
	"github.com/sid-sun/openwebui-bot/pkg/bot/handlers/completion"
	"github.com/sid-sun/openwebui-bot/pkg/bot/handlers/reset"
	"github.com/sid-sun/openwebui-bot/pkg/bot/store"
)

type updates struct {
	ch  tgbotapi.UpdatesChannel
	bot *tgbotapi.BotAPI
}

// ListenAndServe starts listens on the update channel and handles routing the update to handlers
func (u updates) ListenAndServe() {
	store.BotUsername = u.bot.Self.UserName
	slog.Info("[StartBot] Started Bot", slog.String("bot_name", u.bot.Self.FirstName))
	for update := range u.ch {
		update := update
		go func() {
			if update.Message == nil {
				return
			}
			if update.Message.IsCommand() {
				switch update.Message.Command() {
				case "reset":
					reset.Handler(u.bot, update)
					return
				}
			}
			completion.Handler(u.bot, update)
		}()
	}
}

type bot struct {
	bot *tgbotapi.BotAPI
}

// NewUpdateChan creates a new channel to get update
func (b bot) NewUpdateChan() updates {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	ch := b.bot.GetUpdatesChan(u)
	return updates{ch: ch, bot: b.bot}
}

// New returns a new instance of the router
func New(cfg config.BotConfig) bot {
	b, err := tgbotapi.NewBotAPI(cfg.Token())
	if err != nil {
		panic(err)
	}
	return bot{
		bot: b,
	}
}
