package help

import (
	"log/slog"

	tele "gopkg.in/telebot.v3"
)

var logger = slog.Default().With(slog.String("package", "Help"))

const helpText = `Here's what I can do:

Just keep sending messages — I continue the conversation automatically. Reply to an earlier message to branch off from that point instead.

/new - Start a new conversation thread. Your next message begins fresh.

/reset <system prompt> - Set the system prompt. Takes effect immediately, even in earlier conversations.
/resend - Reply to a message to regenerate the response from there.
/models - Pick which model to use, with an inline keyboard.
/thinking [on|off] - Toggle whether the model's reasoning/thinking is shown alongside its reply. No argument toggles the current state.
/help - Show this message.`

// Handler replies with a summary of the bot's commands.
func Handler(c tele.Context) error {
	logger.Info("[Help] [Attempt]")

	err := c.Send(helpText)
	if err != nil {
		logger.Error("failed to send message", slog.Any("error", err))
		return err
	}

	logger.Info("[Help] [Success]")
	return nil
}
