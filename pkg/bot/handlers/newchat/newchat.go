package newchat

import (
	"log/slog"

	"github.com/sid-sun/openwebui-telegram/pkg/bot/store"
	tele "gopkg.in/telebot.v3"
)

var logger = slog.Default().With(slog.String("package", "NewChat"))

// Handler marks the chat so the next message starts a fresh conversation thread
// instead of implicitly continuing the most recent one. Existing threads are
// left intact and can still be resumed by replying to one of their messages.
func Handler(c tele.Context) error {
	logger.Info("[NewChat] [Attempt]")

	store.NewThreadStore[c.Chat().ID] = true

	err := c.Send("Started a new conversation. Your next message begins a fresh thread.")
	if err != nil {
		logger.Error("failed to send message", slog.Any("error", err))
		return err
	}

	logger.Info("[NewChat] [Success]")
	return nil
}
