package models

import (
	"log/slog"
	"strings"

	"github.com/sid-sun/openwebui-bot/pkg/bot/store"
	tele "gopkg.in/telebot.v3"
)

var logger = slog.Default().With(slog.String("package", "Models"))

func GetModelsHandler(b *tele.Bot) tele.HandlerFunc {
	return func(c tele.Context) error {
		logger.Info("[Models] [GetModels] [Attempt]")

		modelInfoMessage, modelOptions := getInlineKeyboardMarkup(store.ModelStore[c.Chat().ID])

		mkp := b.NewMarkup()
		mkp.InlineKeyboard = modelOptions

		b.Send(c.Chat(), modelInfoMessage, mkp)

		logger.Info("[Models] [GetModels] [Success]")
		return nil
	}
}

func CallbackHandler(b *tele.Bot) tele.HandlerFunc {
	return func(c tele.Context) error {
		logger.Info("[Models] [Callback] [Success]", slog.String("scope", "callback"))

		model, found := strings.CutPrefix(c.Callback().Data, "model_")
		if !found {
			c.Send("requested model size not available")
		}

		store.ModelStore[c.Chat().ID] = model
		modelInfoMessage, modelOptions := getInlineKeyboardMarkup(store.ModelStore[c.Chat().ID])
		mkp := b.NewMarkup()
		mkp.InlineKeyboard = modelOptions
		_, err := b.Edit(c.Callback().Message, modelInfoMessage, mkp)
		if err != nil {
			logger.Error("[Models] [Callback] [Error]", slog.String("error", err.Error()))
		}

		logger.Info("[Models] [Callback] [Success]", slog.String("scope", "callback"))
		return nil
	}
}
