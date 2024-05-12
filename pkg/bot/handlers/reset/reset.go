package reset

import (
	"log/slog"

	"github.com/sid-sun/openwebui-bot/pkg/bot/store"
	tele "gopkg.in/telebot.v3"
)

var logger = slog.Default().With(slog.String("package", "Reset"))

func Handler(c tele.Context) error {
	logger.Info("[Reset] [Attempt]")

	prompt := c.Message().Text
	store.SystemPromptStore[c.Chat().ID] = prompt

	err := c.Send("System prompt updated")
	if err != nil {
		logger.Error("failed to send message", slog.Any("error", err))
		return err
	}

	logger.Info("[Reset] [Success]")
	return nil
}
