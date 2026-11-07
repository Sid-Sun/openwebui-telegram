package thinking

import (
	"log/slog"
	"strings"

	"github.com/sid-sun/openwebui-telegram/pkg/bot/store"
	tele "gopkg.in/telebot.v3"
)

var logger = slog.Default().With(slog.String("package", "Thinking"))

// Handler toggles whether the model's reasoning/thinking content is shown
// alongside its reply. Usage: "/thinking", "/thinking on", "/thinking off".
func Handler(c tele.Context) error {
	logger.Info("[Thinking] [Attempt]")

	arg := strings.ToLower(strings.TrimSpace(c.Message().Text))
	chatID := c.Chat().ID

	switch arg {
	case "on":
		store.ShowThinkingStore[chatID] = true
	case "off":
		store.ShowThinkingStore[chatID] = false
	default:
		store.ShowThinkingStore[chatID] = !store.ShowThinkingStore[chatID]
	}

	state := "hidden"
	if store.ShowThinkingStore[chatID] {
		state = "visible"
	}
	err := c.Send("Model thinking is now " + state + " in replies")
	if err != nil {
		logger.Error("failed to send message", slog.Any("error", err))
		return err
	}

	logger.Info("[Thinking] [Success]")
	return nil
}
