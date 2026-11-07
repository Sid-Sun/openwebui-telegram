package completion

import (
	"fmt"
	"log/slog"
	"strings"
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
		chatID := c.Chat().ID

		// notify user we are processing
		c.Notify(tele.Typing)

		addMessageToChain(message)

		updatesChan := make(chan contract.CompletionUpdate, 100)
		go func() {
			err := service.GetChatResponseStream(chatID, promptID, updatesChan)
			if err != nil {
				logger.Error("could not generate completion", slog.String("conext", "GetChatResponseStream"),
					slog.Any("error", err))
			}
		}()

		replyTo := message
		showThinking := store.ShowThinkingStore[chatID]

		var botMessage *tele.Message
		var lastContent, lastDisplay string
		nextMessageRequiredLen := 0
		debounced := debounce.New(200 * time.Millisecond)
		for completion := range updatesChan {
			lastContent = completion.Content
			// Trim: reasoning/content deltas are frequently whitespace-only
			// (e.g. "\n\n", "  "). Telegram's API rejects an all-whitespace
			// message body as empty, so a leading whitespace-only chunk must
			// not be sent as-is.
			display := strings.TrimSpace(formatDisplay(showThinking, completion))
			if display == "" {
				if completion.IsLast {
					display = "(no response)"
				} else {
					// still thinking and nothing worth showing yet
					continue
				}
			}

			if botMessage == nil {
				sent, err := b.Send(c.Chat(), display, &tele.SendOptions{ReplyTo: replyTo})
				if err != nil {
					logger.Error("failed to send message", slog.String("context", "new completion message"), slog.Any("error", err))
					return err
				}
				botMessage = sent
				lastDisplay = display
				nextMessageRequiredLen = len(display) * 110 / 100
				if completion.IsLast {
					break
				}
				continue
			}

			// Only update message if the new message is at least 10% longer than last message
			// To avoid Telegram rate limiting
			if len(display) <= nextMessageRequiredLen && !completion.IsLast {
				continue
			}
			lastDisplay = display
			nextMessageRequiredLen = len(lastDisplay) * 110 / 100
			send := func() {
				_, err := b.Edit(botMessage, lastDisplay)
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

		if botMessage == nil {
			// the stream produced nothing at all (e.g. request failed before any chunk arrived)
			logger.Error("no response received from model", slog.String("context", "completion loop"))
			return nil
		}

		// store only the actual reply content (not the thinking display) so
		// future turns don't feed reasoning text back into the model as context
		if lastContent == "" {
			lastContent = "(no response)"
		}
		botMessage.Text = lastContent

		addMessageToChain(botMessage)
		// To print reply  messages, comment above and uncomment below
		// bm := addMessageToChain(botMessage)
		// printReplyMessages(c.Chat().ID, bm)

		slog.Info("[Completion] [Success]")
		return nil
	}
}

// formatDisplay builds the text shown in Telegram for a given update. When
// showThinking is disabled (the default), only the actual reply content is
// shown, matching pre-reasoning-model behavior. When enabled, the model's
// reasoning is shown above the content, labeled, so it isn't mistaken for
// the answer itself.
func formatDisplay(showThinking bool, u contract.CompletionUpdate) string {
	if !showThinking || u.Reasoning == "" {
		return u.Content
	}
	if u.Content == "" {
		return "[Thinking...]\n" + u.Reasoning
	}
	return "[Thinking]\n" + u.Reasoning + "\n\n[Answer]\n" + u.Content
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
	chat := store.ChatStore[m.Chat.ID]

	var parent int
	switch {
	case chat[m.ID] != nil:
		// Re-storing an existing message (e.g. an edit): keep its original
		// parent so the thread structure stays stable and we never relink it
		// to a message that came after it.
		parent = chat[m.ID].Parent
	case m.ReplyTo != nil && chat[m.ReplyTo.ID] != nil:
		// Explicit reply: branch/continue from the replied-to message. This
		// lets users pick up an older point in the conversation on purpose,
		// and is what /resend relies on.
		parent = m.ReplyTo.ID
	case store.NewThreadStore[m.Chat.ID]:
		// The user ran /new: start a fresh thread. parent stays 0, and we
		// consume the flag so only this message begins anew.
		delete(store.NewThreadStore, m.Chat.ID)
	default:
		// Implicit follow: with no explicit reply, continue the most recent
		// message in this chat so users don't have to reply to keep a thread
		// going. Telegram message IDs increase monotonically per chat, so the
		// highest stored ID is the current tip of the conversation.
		for id := range chat {
			if id > parent {
				parent = id
			}
		}
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
