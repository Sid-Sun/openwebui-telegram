package completion

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/bep/debounce"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sid-sun/openwebui-bot/pkg/bot/contract"
	"github.com/sid-sun/openwebui-bot/pkg/bot/service"
	"github.com/sid-sun/openwebui-bot/pkg/bot/store"
)

var logger = slog.Default().With(slog.String("package", "Completion"))

// Handler handles all repeat requests
func Handler(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	logger.Info("[Completion] [Attempt]")

	// Check if this is a resend attempt
	promptID := update.Message.MessageID
	isResend := false
	if update.Message.IsCommand() {
		switch update.Message.Command() {
		case "resend":
			replyToMessage := store.ChatStore[update.FromChat().ID][update.Message.ReplyToMessage.MessageID]
			if update.Message.ReplyToMessage == nil || replyToMessage == nil {
				// drop invalid request
				em := tgbotapi.NewMessage(update.FromChat().ID, "Reply to the message you want to regenerate from - it can't be a previous /resend message")
				bot.Send(em)
				return
			}
			if replyToMessage.From != update.Message.From.UserName {
				// invalid request
				em := tgbotapi.NewMessage(update.FromChat().ID, "Last message for resend must be a user message")
				bot.Send(em)
			}
			promptID = update.Message.ReplyToMessage.MessageID
			isResend = true
		}
	}

	action := tgbotapi.NewChatAction(update.Message.Chat.ID, tgbotapi.ChatTyping)
	bot.Send(action)

	if !isResend {
		addMessageToChain(update.Message)
	}

	updatesChan := make(chan contract.CompletionUpdate, 100)
	go func() {
		err := service.GetChatResponseStream(update.Message.Chat.ID, promptID, updatesChan)
		if err != nil {
			logger.Error("could not generate completion", slog.String("conext", "GetChatResponseStream"), slog.Any("error", err))
		}
	}()

	firstCompletion := <-updatesChan

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, firstCompletion.Message)
	msg.ReplyToMessageID = promptID
	// msg.ReplyMarkup = tgbotapi.ForceReply{ForceReply: true}

	botMessage, err := bot.Send(msg)
	if err != nil {
		logger.Error("failed to send message", slog.String("context", "new completion message"), slog.Any("error", err))
		return
	}

	debounced := debounce.New(5 * time.Millisecond)
	for completion := range updatesChan {
		if completion.IsLast {
			break
		}
		edit := tgbotapi.NewEditMessageText(update.Message.Chat.ChatConfig().ChatID, botMessage.MessageID, completion.Message)
		send := func() {
			botMessage.Text = edit.Text
			_, err = bot.Send(edit)
			if err != nil {
				logger.Error("failed to send message", slog.String("context", "edit completion message"), slog.Any("error", err))
				return
			}
		}
		debounced(send)
	}

	time.Sleep(10 * time.Millisecond)
	botMessage.Chat = update.Message.Chat
	botMessage.ReplyToMessage = update.Message
	if isResend {
		botMessage.ReplyToMessage = update.Message.ReplyToMessage
	}
	addMessageToChain(&botMessage)
	// To print reply  messages, comment above and uncomment below
	// bm := addMessageToChain(&botMessage)
	// printReplyMessages(update.Message.Chat.ID, bm)

	slog.Info("[Completion] [Success]")
}

func printReplyMessages(chatID int64, m *contract.MessageLink) {
	if m.Parent != 0 {
		printReplyMessages(chatID, store.ChatStore[chatID][m.Parent])
	}
	fmt.Printf("Message: %s\n", m.Text)
}

func addMessageToChain(m *tgbotapi.Message) *contract.MessageLink {
	if store.ChatStore[m.Chat.ID] == nil {
		store.ChatStore[m.Chat.ID] = make(map[int]*contract.MessageLink)
	}

	var parent int
	if m.ReplyToMessage != nil {
		parent = m.ReplyToMessage.MessageID
		if store.ChatStore[m.Chat.ID][parent] == nil {
			addMessageToChain(m.ReplyToMessage)
		}
		store.ChatStore[m.Chat.ID][parent].Children = append(store.ChatStore[m.Chat.ID][parent].Children, m.MessageID)
	}

	// fmt.Printf("Message: %+v\n", m)
	if m.From == nil {
		m.From = &tgbotapi.User{UserName: "unknown"}
	}
	store.ChatStore[m.Chat.ID][m.MessageID] = &contract.MessageLink{
		Parent:   parent,
		Children: []int{},
		Text:     m.Text,
		From:     m.From.UserName,
	}

	// x, err := json.MarshalIndent(store.ChatStore, "", "  ")
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// fmt.Println(string(x))

	return store.ChatStore[m.Chat.ID][m.MessageID]
}
