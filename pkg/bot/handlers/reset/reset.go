package reset

import (
	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sid-sun/openwebui-bot/pkg/bot/store"
)

var logger = slog.Default().With(slog.String("package", "Reset"))

func Handler(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	logger.Info("[Reset] [Attempt]")

	prompt := update.Message.CommandArguments()
	store.SystemPromptStore[update.FromChat().ID] = prompt

	m := tgbotapi.NewMessage(update.FromChat().ID, "System prompt updated")
	_, err := bot.Send(m)
	if err != nil {
		logger.Error("failed to send message", slog.Any("error", err))
	}

	logger.Info("[Reset] [Success]")
}
