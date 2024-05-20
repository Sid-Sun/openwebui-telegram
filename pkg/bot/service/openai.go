package service

import (
	"bufio"
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/sid-sun/openwebui-bot/cmd/config"
	"github.com/sid-sun/openwebui-bot/pkg/bot/contract"
	"github.com/sid-sun/openwebui-bot/pkg/bot/store"
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
			Model:  model.Model,
			Stream: true,
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
			Model:  model.Model,
			Stream: true,
		},
		Messages: generateMessages(chatID, promptID, []contract.ChatMessage{{
			Role:    "system",
			Content: getSystemPrompt(chatID),
		}}),
		ModelTweaks: model.GetAdvancedTweaks(),
	}
	return x
}

func GetChatResponseStream(chatID int64, promptID int, uc chan contract.CompletionUpdate) error {
	var payload any
	model := getModel(chatID)
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
	req, err := http.NewRequest(http.MethodPost, config.GlobalConfig.OpenAIAPI.Endpoint+"chat/completions", bytes.NewReader(payloadJSON))
	if err != nil {
		logger.Error("failed to create http request", slog.String("func", "GetChatResponseStream"), slog.Any("error", err))
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+config.GlobalConfig.OpenAIAPI.APIKey)

	res, err := client.Do(req)
	if err != nil {
		logger.Error("http request failed", slog.String("func", "GetChatResponseStream"), slog.Any("error", err))
		return err
	}
	defer res.Body.Close()

	resolveDeltaAndSendUpdates(res, uc)
	return nil
}

func resolveDeltaAndSendUpdates(res *http.Response, uc chan contract.CompletionUpdate) {
	var message string
	scanner := bufio.NewScanner(res.Body)
	for scanner.Scan() {
		data := scanner.Text()

		dataFields := strings.Split(data, "data: ")
		if len(dataFields) == 1 {
			// skip initial request without data
			continue
		}

		var resp contract.ChatCompletionResponse
		err := json.Unmarshal([]byte(dataFields[1]), &resp)
		if err != nil {
			logger.Error("failed to unmarshal payload body", slog.Any("error", err))
		}

		message += resp.Choices[0].Delta.Content
		if resp.Choices[0].FinishReason == "stop" {
			uc <- contract.CompletionUpdate{
				Message: message,
				IsLast:  true,
			}
			return
		}
		if resp.Choices[0].Delta.Content == "" {
			continue
		}
		uc <- contract.CompletionUpdate{
			Message: message,
			IsLast:  false,
		}
	}
}
