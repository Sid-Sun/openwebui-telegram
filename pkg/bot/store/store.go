package store

import (
	"github.com/sid-sun/openwebui-bot/pkg/bot/contract"
)

var ChatStore map[int64]map[int]*contract.MessageLink
var SystemPromptStore map[int64]string
var BotUsername string
var ModelStore map[int64]string

func NewStore() {
	ChatStore = make(map[int64]map[int]*contract.MessageLink)
	SystemPromptStore = make(map[int64]string)
	ModelStore = make(map[int64]string)
}
