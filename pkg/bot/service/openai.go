package service

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/sid-sun/openwebui-telegram/cmd/config"
	"github.com/sid-sun/openwebui-telegram/pkg/bot/contract"
	"github.com/sid-sun/openwebui-telegram/pkg/bot/store"
)

var logger = slog.Default().With(slog.String("package", "Completion"))

func generateMessages(chatID int64, promptID int, messages []contract.ChatMessage) []contract.ChatMessage {
	parentID := store.ChatStore[chatID][promptID].Parent
	if parentID != 0 {
		messages = generateMessages(chatID, parentID, messages)
	}
	return append(messages, contract.ChatMessage{
		Role:    getRole(store.ChatStore[chatID][promptID].From),
		Content: store.ChatStore[chatID][promptID].Text,
	})
}

func generateAPIPayloadMinimal(chatID int64, promptID int) contract.ChatCompletionPayloadMinimal {
	model := getModel(chatID)
	x := contract.ChatCompletionPayloadMinimal{
		ModelOptions: contract.ModelOptions{
			Model:           model.Model,
			Stream:          true,
			ReasoningEffort: model.ReasoningEffort,
		},
		Messages: generateMessages(chatID, promptID, []contract.ChatMessage{{
			Role:    "system",
			Content: getSystemPrompt(chatID),
		}}),
		BasicModelTweaks: model.GetBasicTweaks(),
	}
	return x
}

func generateAPIPayload(chatID int64, promptID int) contract.ChatCompletionPayload {
	model := getModel(chatID)
	x := contract.ChatCompletionPayload{
		ModelOptions: contract.ModelOptions{
			Model:           model.Model,
			Stream:          true,
			ReasoningEffort: model.ReasoningEffort,
		},
		Messages: generateMessages(chatID, promptID, []contract.ChatMessage{{
			Role:    "system",
			Content: getSystemPrompt(chatID),
		}}),
		ModelTweaks: model.GetAdvancedTweaks(),
	}
	return x
}

// openAIBackend talks to an OpenAI-compatible chat/completions endpoint
// (OpenAI itself or a proxy such as OpenWebUI).
type openAIBackend struct{}

func (openAIBackend) StreamChat(ctx context.Context, chatID int64, promptID int, uc chan contract.CompletionUpdate) error {
	var payload any
	model := getModel(chatID)
	openaiBackend := config.GetConfig().Backends[model.Backend]
	if model.UseMinimalTweaks() {
		payload = generateAPIPayloadMinimal(chatID, promptID)
	} else {
		payload = generateAPIPayload(chatID, promptID)
	}

	payloadJSON, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		panic(err)
	}
	logger.Debug("api payload", slog.String("payload", string(payloadJSON)))

	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, openaiBackend.Endpoint+"chat/completions", bytes.NewReader(payloadJSON))
	if err != nil {
		logger.Error("failed to create http request", slog.String("func", "openAIBackend.StreamChat"), slog.Any("error", err))
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+openaiBackend.APIKey)

	res, err := client.Do(req)
	if err != nil {
		logger.Error("http request failed", slog.String("func", "openAIBackend.StreamChat"), slog.Any("error", err))
		return err
	}
	defer res.Body.Close()

	resolveDeltaAndSendUpdates(res, uc)
	return nil
}

func resolveDeltaAndSendUpdates(res *http.Response, uc chan contract.CompletionUpdate) {
	var content, reasoning string
	scanner := bufio.NewScanner(res.Body)
	for scanner.Scan() {
		data := scanner.Text()

		dataFields := strings.Split(data, "data: ")
		if len(dataFields) == 1 {
			// skip lines without a data payload (blank lines, SSE comments)
			continue
		}

		// the stream is terminated by a literal "data: [DONE]" line, which is
		// not JSON and carries no choices - stop before trying to parse it
		if dataFields[1] == "[DONE]" {
			return
		}

		var resp contract.ChatCompletionResponse
		err := json.Unmarshal([]byte(dataFields[1]), &resp)
		if err != nil {
			logger.Error("failed to unmarshal payload body", slog.Any("error", err))
			continue
		}
		if len(resp.Choices) == 0 {
			continue
		}

		choice := resp.Choices[0]
		content += choice.Delta.Content
		reasoning += choice.Delta.ReasoningContent

		// finish_reason is only set on the terminal chunk, but it isn't always
		// "stop" (e.g. "length" when max_tokens is hit, "tool_calls", etc.) -
		// any non-empty value means this is the last chunk
		isLast := choice.FinishReason != ""
		if !isLast && choice.Delta.Content == "" && choice.Delta.ReasoningContent == "" {
			continue
		}
		uc <- contract.CompletionUpdate{
			Content:   content,
			Reasoning: reasoning,
			IsLast:    isLast,
		}
		if isLast {
			return
		}
	}
	// stream ended without a terminal chunk (e.g. connection closed early) -
	// make sure the consumer isn't left blocked waiting on IsLast forever
	uc <- contract.CompletionUpdate{
		Content:   content,
		Reasoning: reasoning,
		IsLast:    true,
	}
}
