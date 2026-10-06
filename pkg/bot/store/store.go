package store

import (
	"github.com/sid-sun/openwebui-bot/pkg/bot/contract"
)

var ChatStore map[int64]map[int]*contract.MessageLink
var SystemPromptStore map[int64]string
var BotUsername string
var ModelStore map[int64]string
var ShowThinkingStore map[int64]bool

// NewThreadStore marks chats whose next message should start a fresh thread
// (set by /new) instead of implicitly continuing the latest conversation.
var NewThreadStore map[int64]bool

func NewStore() {
	ChatStore = make(map[int64]map[int]*contract.MessageLink)
	SystemPromptStore = make(map[int64]string)
	ModelStore = make(map[int64]string)
	ShowThinkingStore = make(map[int64]bool)
	NewThreadStore = make(map[int64]bool)
}
