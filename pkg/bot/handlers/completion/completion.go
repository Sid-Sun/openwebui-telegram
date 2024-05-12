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
func Handler(b *tele.Bot, isResend bool) tele.HandlerFunc {
	return func(c tele.Context) error {
		logger.Info("[Completion] [Attempt]", slog.Bool("is_resend", isResend))

		// Check if this is a resend attempt
		promptID := c.Message().ID
		if isResend {
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
			promptID = c.Message().ReplyTo.ID
		}

		// notify user we are processing
		c.Notify(tele.Typing)

		if !isResend {
			addMessageToChain(c.Message())
		}

		updatesChan := make(chan contract.CompletionUpdate, 100)
		go func() {
			err := service.GetChatResponseStream(c.Chat().ID, promptID, updatesChan)
			if err != nil {
				logger.Error("could not generate completion", slog.String("conext", "GetChatResponseStream"), slog.Any("error", err))
			}
		}()

		firstCompletion := <-updatesChan

		replyTo := c.Message()
		if isResend {
			replyTo = c.Message().ReplyTo
		}

		botMessage, err := b.Send(c.Chat(), firstCompletion.Message, &tele.SendOptions{
			ReplyTo: replyTo,
		})
		if err != nil {
			logger.Error("failed to send message", slog.String("context", "new completion message"), slog.Any("error", err))
			return err
		}

		var finalMessage *string
		debounced := debounce.New(20 * time.Millisecond)
		for completion := range updatesChan {
			finalMessage = &completion.Message
			send := func() {
				_, err := b.Edit(botMessage, completion.Message)
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
		botMessage.Text = *finalMessage

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
	if m.ReplyTo != nil {
		parent = m.ReplyTo.ID
		if store.ChatStore[m.Chat.ID][parent] == nil {
			addMessageToChain(m.ReplyTo)
		}
		store.ChatStore[m.Chat.ID][parent].Children = append(store.ChatStore[m.Chat.ID][parent].Children, m.ID)
	}

	// fmt.Printf("Message: %+v\n", m)
	if m.Sender == nil {
		m.Sender = &tele.User{
			Username: "unknown",
		}
	}
	store.ChatStore[m.Chat.ID][m.ID] = &contract.MessageLink{
		Parent:   parent,
		Children: []int{},
		Text:     m.Text,
		From:     m.Sender.Username,
	}

	// x, err := json.MarshalIndent(store.ChatStore, "", "  ")
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// fmt.Println(string(x))

	return store.ChatStore[m.Chat.ID][m.ID]
}
