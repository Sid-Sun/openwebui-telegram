package models

import (
	"fmt"

	"github.com/sid-sun/openwebui-bot/cmd/config"
	tele "gopkg.in/telebot.v3"
)

func getInlineKeyboardMarkup(currentModel string) (string, [][]tele.InlineButton) {
	var modelOptions [][]tele.InlineButton
	modelInfoMessage := "Here are the available models: \n"
	if currentModel == "" {
		currentModel = "default"
	}
	for _, modelName := range config.GlobalConfig.ModelNames {
		options := config.GlobalConfig.Models[modelName]
		text := fmt.Sprintf("%s (%s) - %d", modelName, options.Model, options.Tweaks.ContextLength)
		if currentModel == modelName {
			text = fmt.Sprintf("*%s* (%s) - %d", modelName, options.Model, options.Tweaks.ContextLength)
		}
		modelOptions = append(modelOptions, []tele.InlineButton{
			{
				Data: "model_" + modelName,
				Text: text,
			},
		})
	}
	return modelInfoMessage, modelOptions
}
