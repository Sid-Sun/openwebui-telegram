package router

import (
	"log/slog"
	"time"

	"github.com/sid-sun/openwebui-bot/cmd/config"
	"github.com/sid-sun/openwebui-bot/pkg/bot/handlers/completion"
	"github.com/sid-sun/openwebui-bot/pkg/bot/handlers/help"
	"github.com/sid-sun/openwebui-bot/pkg/bot/handlers/models"
	"github.com/sid-sun/openwebui-bot/pkg/bot/handlers/newchat"
	"github.com/sid-sun/openwebui-bot/pkg/bot/handlers/reset"
	"github.com/sid-sun/openwebui-bot/pkg/bot/handlers/thinking"
	"github.com/sid-sun/openwebui-bot/pkg/bot/store"
	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/middleware"
)

type Bot struct {
	bot *tele.Bot
}

// StartupAnnouncement is broadcast to every previously-seen chat when the bot
// starts up.
const StartupAnnouncement = "I'm back, new and improved! 🎉"

// botCommands is the list of commands registered with Telegram so they show
// up in the client's command menu/autocomplete.
var botCommands = []tele.Command{
	{Text: "new", Description: "Start a new conversation thread"},
	{Text: "reset", Description: "Set the system prompt"},
	{Text: "resend", Description: "Regenerate a response (reply to the message)"},
	{Text: "models", Description: "Pick which model to use"},
	{Text: "thinking", Description: "Toggle showing the model's thinking"},
	{Text: "help", Description: "Show help"},
}

// ListenAndServe starts listens on the update channel and handles routing the update to handlers
func (b Bot) Start() {
	store.BotUsername = b.bot.Me.Username
	slog.Info("[StartBot] Started Bot", slog.String("bot_name", b.bot.Me.FirstName))
	// Keep Telegram's command menu in sync with what we actually support
	if err := b.bot.SetCommands(botCommands); err != nil {
		slog.Error("[StartBot] Failed to set bot commands", slog.Any("error", err))
	}
	// Register Special Commands
	// Start a new conversation thread
	b.bot.Handle("/new", newchat.Handler)
	resetGroup := b.bot.Group()
	resetGroup.Use(StripCommand("/reset"))
	resetGroup.Handle("/reset", reset.Handler)
	// Thinking display toggle
	thinkingGroup := b.bot.Group()
	thinkingGroup.Use(StripCommand("/thinking"))
	thinkingGroup.Handle("/thinking", thinking.Handler)
	// Help
	b.bot.Handle("/help", help.Handler)
	// Callbacks
	callbackGroup := b.bot.Group()
	callbackGroup.Use(middleware.AutoRespond())
	callbackGroup.Handle("/models", models.GetModelsHandler(b.bot))
	callbackGroup.Handle(tele.OnCallback, models.CallbackHandler(b.bot))
	// Add all other handlers
	b.bot.Handle("/resend", completion.Handler(b.bot, completion.RegenerateCompletion))
	b.bot.Handle(tele.OnEdited, completion.Handler(b.bot, completion.EditCompletion))
	b.bot.Handle(tele.OnText, completion.Handler(b.bot, completion.NewCompletion))
	b.bot.Start()
}

func (b Bot) Stop() {
	slog.Info("[StopBot] Stopping Bot")
	b.bot.Stop()
}

// Announce sends text to every chat the bot has previously talked to (the chats
// loaded into the store on startup). It is best-effort: recipients that have
// blocked or deleted the bot are logged and skipped so one failure doesn't
// abort the rest.
func (b Bot) Announce(text string) {
	ids := make([]int64, 0, len(store.ChatStore))
	for id := range store.ChatStore {
		ids = append(ids, id)
	}

	if len(ids) == 0 {
		return
	}

	slog.Info("[Announce] Broadcasting startup message", slog.Int("recipients", len(ids)))
	var sent int
	for _, id := range ids {
		if _, err := b.bot.Send(tele.ChatID(id), text); err != nil {
			slog.Warn("[Announce] Failed to send to chat", slog.Int64("chat_id", id), slog.Any("error", err))
			continue
		}
		sent++
	}
	slog.Info("[Announce] Broadcast complete", slog.Int("sent", sent), slog.Int("total", len(ids)))
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
