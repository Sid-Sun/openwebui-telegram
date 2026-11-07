package service

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/shared"
	"github.com/sid-sun/openwebui-telegram/cmd/config"
	"github.com/sid-sun/openwebui-telegram/pkg/bot/contract"
)

// openAISDKBackend talks to an OpenAI-compatible chat/completions endpoint
// using the official github.com/openai/openai-go/v3 SDK.
type openAISDKBackend struct{}

var (
	openAISDKClientsMu sync.Mutex
	openAISDKClients   = make(map[string]openai.Client)
)

func getOpenAISDKClient(model config.Model) openai.Client {
	backend := config.GetConfig().Backends[model.Backend]
	openAISDKClientsMu.Lock()
	defer openAISDKClientsMu.Unlock()
	if client, ok := openAISDKClients[model.Backend]; ok {
		return client
	}
	client := openai.NewClient(
		option.WithBaseURL(backend.Endpoint),
		option.WithAPIKey(backend.APIKey),
	)
	openAISDKClients[model.Backend] = client
	return client
}

func mapOpenAISDKMessages(history []contract.ChatMessage) []openai.ChatCompletionMessageParamUnion {
	messages := make([]openai.ChatCompletionMessageParamUnion, len(history))
	for i, m := range history {
		switch m.Role {
		case "system":
			messages[i] = openai.SystemMessage(m.Content)
		case "assistant":
			messages[i] = openai.AssistantMessage(m.Content)
		default:
			messages[i] = openai.UserMessage(m.Content)
		}
	}
	return messages
}

func buildOpenAISDKParams(chatID int64, promptID int, model config.Model) openai.ChatCompletionNewParams {
	history := generateMessages(chatID, promptID, []contract.ChatMessage{{
		Role:    "system",
		Content: getSystemPrompt(chatID),
	}})

	params := openai.ChatCompletionNewParams{
		Model:    shared.ChatModel(model.Model),
		Messages: mapOpenAISDKMessages(history),
	}

	if model.ReasoningEffort != "" {
		params.ReasoningEffort = shared.ReasoningEffort(model.ReasoningEffort)
	}

	t := model.Tweaks
	if t.MaxTokens != 0 {
		params.MaxTokens = openai.Int(int64(t.MaxTokens))
	}
	if t.Temperature != 0 {
		params.Temperature = openai.Float(t.Temperature)
	}
	if !model.UseMinimalTweaks() {
		if t.FrequencyPenalty != 0 {
			params.FrequencyPenalty = openai.Float(t.FrequencyPenalty)
		}
		if t.PresencePenalty != 0 {
			params.PresencePenalty = openai.Float(t.PresencePenalty)
		}
	}

	extra := make(map[string]any)
	if t.ContextLength != 0 {
		extra["context_length"] = t.ContextLength
	}
	if !model.UseMinimalTweaks() && t.RepeatPenalty != 0 {
		extra["repeat_penalty"] = t.RepeatPenalty
	}
	if len(extra) > 0 {
		params.SetExtraFields(extra)
	}

	return params
}

func extractReasoningDelta(delta openai.ChatCompletionChunkChoiceDelta) string {
	field, ok := delta.JSON.ExtraFields["reasoning_content"]
	if !ok || !field.Valid() {
		return ""
	}
	var reasoning string
	if err := json.Unmarshal([]byte(field.Raw()), &reasoning); err != nil {
		return ""
	}
	return reasoning
}

func (openAISDKBackend) StreamChat(ctx context.Context, chatID int64, promptID int, uc chan contract.CompletionUpdate) error {
	model := getModel(chatID)
	client := getOpenAISDKClient(model)
	params := buildOpenAISDKParams(chatID, promptID, model)

	paramsJSON, err := json.MarshalIndent(params, "", "  ")
	if err != nil {
		logger.Error("failed to marshal openaisdk request", slog.String("func", "openAISDKBackend.StreamChat"), slog.Any("error", err))
		return err
	}
	logger.Debug("openaisdk chat request", slog.String("payload", string(paramsJSON)))

	stream := client.Chat.Completions.NewStreaming(ctx, params)
	defer stream.Close()

	var content, reasoning string
	var sentLast bool
	for stream.Next() {
		chunk := stream.Current()
		if len(chunk.Choices) == 0 {
			continue
		}

		choice := chunk.Choices[0]
		reasoningDelta := extractReasoningDelta(choice.Delta)
		content += choice.Delta.Content
		reasoning += reasoningDelta

		isLast := choice.FinishReason != ""
		if !isLast && choice.Delta.Content == "" && reasoningDelta == "" {
			continue
		}
		uc <- contract.CompletionUpdate{
			Content:   content,
			Reasoning: reasoning,
			IsLast:    isLast,
		}
		if isLast {
			sentLast = true
			return nil
		}
	}

	if err := stream.Err(); err != nil {
		logger.Error("openaisdk chat stream failed", slog.String("func", "openAISDKBackend.StreamChat"), slog.Any("error", err))
		if !sentLast {
			uc <- contract.CompletionUpdate{
				Content:   content,
				Reasoning: reasoning,
				IsLast:    true,
			}
		}
		return err
	}

	if !sentLast {
		uc <- contract.CompletionUpdate{
			Content:   content,
			Reasoning: reasoning,
			IsLast:    true,
		}
	}
	return nil
}
