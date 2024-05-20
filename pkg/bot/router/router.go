package router

import (
	"log/slog"
	"time"

	"github.com/sid-sun/openwebui-bot/cmd/config"
	"github.com/sid-sun/openwebui-bot/pkg/bot/handlers/completion"
	"github.com/sid-sun/openwebui-bot/pkg/bot/handlers/models"
	"github.com/sid-sun/openwebui-bot/pkg/bot/handlers/reset"
	"github.com/sid-sun/openwebui-bot/pkg/bot/store"
	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/middleware"
)

type Bot struct {
	bot *tele.Bot
}

// ListenAndServe starts listens on the update channel and handles routing the update to handlers
func (b Bot) Start() {
	store.BotUsername = b.bot.Me.Username
	slog.Info("[StartBot] Started Bot", slog.String("bot_name", b.bot.Me.FirstName))
	// Register Special Commands
	resetGroup := b.bot.Group()
	resetGroup.Use(StripCommand("/reset"))
	resetGroup.Handle("/reset", reset.Handler)
	// Callbacks
	callbackGroup := b.bot.Group()
	callbackGroup.Use(middleware.AutoRespond())
	callbackGroup.Handle("/models", models.GetModelsHandler(b.bot))
	callbackGroup.Handle(tele.OnCallback, models.CallbackHandler(b.bot))
	// Add all other handlers
	b.bot.Handle("/resend", completion.Handler(b.bot, true))
	b.bot.Handle(tele.OnText, completion.Handler(b.bot, false))
	b.bot.Start()
}

func (b Bot) Stop() {
	slog.Info("[StopBot] Stopping Bot")
	b.bot.Stop()
}

// New returns a new instance of the router
func New(cfg config.BotConfig) *Bot {
	b, err := tele.NewBot(tele.Settings{
		Token:  cfg.Token(),
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		panic(err)
	}
	return &Bot{
		bot: b,
	}
}
