package repeat

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

// Handler handles all repeat requests
func Handler(bot *tgbotapi.BotAPI, update tgbotapi.Update, logger *zap.Logger) {
	logger.Info("[Repeat] [Attempt]")

	msg := tgbotapi.NewCopyMessage(update.Message.Chat.ID, update.Message.Chat.ID, update.Message.MessageID)
	msg.ReplyToMessageID = update.Message.MessageID

	_, err := bot.Send(msg)
	if err != nil {
		logger.Sugar().Errorf("[%s] [%s] %s", handler, "Send", err.Error())
		return
	}

	logger.Info("[Repeat] [Success]")
}
