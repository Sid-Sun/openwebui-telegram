package completion

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/bep/debounce"
	"github.com/sid-sun/openwebui-bot/pkg/bot/contract"
	"github.com/sid-sun/openwebui-bot/pkg/bot/service"
	"github.com/sid-sun/openwebui-bot/pkg/bot/store"
	tele "gopkg.in/telebot.v3"
)

var logger = slog.Default().With(slog.String("package", "Completion"))

// Handler handles all repeat requests
func Handler(b *tele.Bot, handlerMode HandlerMode) tele.HandlerFunc {
	return func(c tele.Context) error {
		logger.Info("[Completion] [Attempt]", slog.Any("handlerMode", handlerMode))

		// Check if this is a resend attempt
		message := c.Message()
		if handlerMode == RegenerateCompletion {
			if c.Message().ReplyTo == nil {
				c.Send("Reply to the message you want to regenerate from")
				return nil
			}
			replyToMessage := store.ChatStore[c.Chat().ID][c.Message().ReplyTo.ID]
			if replyToMessage == nil {
				c.Send("Reply to message can't be a previous /resend message or previous conversation is lost")
				return nil
			}
			if replyToMessage.From != c.Chat().Username {
				c.Send("Last message for resend must be a user message")
				return nil
			}
			message = c.Message().ReplyTo
		}
		promptID := message.ID

		// notify user we are processing
		c.Notify(tele.Typing)

		addMessageToChain(message)

		updatesChan := make(chan contract.CompletionUpdate, 100)
		go func() {
			err := service.GetChatResponseStream(c.Chat().ID, promptID, updatesChan)
			if err != nil {
				logger.Error("could not generate completion", slog.String("conext", "GetChatResponseStream"), slog.Any("error", err))
			}
		}()

		firstCompletion := <-updatesChan

		replyTo := message

		botMessage, err := b.Send(c.Chat(), firstCompletion.Message, &tele.SendOptions{
			ReplyTo: replyTo,
		})
		if err != nil {
			logger.Error("failed to send message", slog.String("context", "new completion message"), slog.Any("error", err))
			return err
		}

		var lastMessage string
		nextMessageRequiredLen := 0
		debounced := debounce.New(200 * time.Millisecond)
		for completion := range updatesChan {
			// Only update message if the new message is at least 10% longer than last message
			// To avoid Telegram rate limiting
			if len(completion.Message) <= nextMessageRequiredLen && !completion.IsLast {
				continue
			}
			lastMessage = completion.Message
			nextMessageRequiredLen = len(lastMessage) * 110 / 100
			send := func() {
				_, err := b.Edit(botMessage, lastMessage)
				if err != nil {
					logger.Error("failed to send message", slog.String("context", "edit completion message"), slog.Any("error", err))
					return
				}
			}
			debounced(send)
			if completion.IsLast {
				break
			}
		}

		// the sleep may not actually be necessary
		botMessage.Text = lastMessage

		addMessageToChain(botMessage)
		// To print reply  messages, comment above and uncomment below
		// bm := addMessageToChain(botMessage)
		// printReplyMessages(c.Chat().ID, bm)

		slog.Info("[Completion] [Success]")
		return nil
	}
}

func printReplyMessages(chatID int64, m *contract.MessageLink) {
	if m.Parent != 0 {
		printReplyMessages(chatID, store.ChatStore[chatID][m.Parent])
	}
	fmt.Printf("Message: %s\n", m.Text)
}

func addMessageToChain(m *tele.Message) *contract.MessageLink {
	if store.ChatStore[m.Chat.ID] == nil {
		store.ChatStore[m.Chat.ID] = make(map[int]*contract.MessageLink)
	}

	var parent int
	// If this is a reply message, set the parent ID if it exists in the store
	if m.ReplyTo != nil && store.ChatStore[m.Chat.ID][m.ReplyTo.ID] != nil {
		parent = m.ReplyTo.ID
	}

	// fmt.Printf("Message: %+v\n", m)
	store.ChatStore[m.Chat.ID][m.ID] = &contract.MessageLink{
		Parent: parent,
		Text:   m.Text,
		From:   m.Sender.Username,
	}

	// x, err := json.MarshalIndent(store.ChatStore, "", "  ")
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// fmt.Println(string(x))

	return store.ChatStore[m.Chat.ID][m.ID]
}
