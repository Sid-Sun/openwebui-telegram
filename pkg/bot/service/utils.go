package service

import (
	"github.com/sid-sun/openwebui-bot/cmd/config"
	"github.com/sid-sun/openwebui-bot/pkg/bot/store"
)

func getRole(from string) string {
	if from == store.BotUsername {
		return "assistant"
	}
	return "user"
}

func getSystemPrompt(chatID int64) string {
	if store.SystemPromptStore[chatID] == "" {
		return "You are a friendly assistant"
	}
	return store.SystemPromptStore[chatID]
}

func getModel(chatID int64) config.Model {
	if store.ModelStore[chatID] == "" {
		return config.GetConfig().Models["default"]
	}
	return config.GetConfig().Models[store.ModelStore[chatID]]
}
