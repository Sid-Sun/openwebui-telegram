package service

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"
	"sync"

	"github.com/ollama/ollama/api"
	"github.com/sid-sun/openwebui-telegram/cmd/config"
	"github.com/sid-sun/openwebui-telegram/pkg/bot/contract"
)

// ollamaBackend talks to a native Ollama daemon (POST /api/chat) via the
// official ollama/api client.
type ollamaBackend struct{}

var (
	ollamaClientMu       sync.Mutex
	ollamaClient         *api.Client
	ollamaClientEndpoint string
)

func getOllamaClient(model config.Model) (*api.Client, error) {
	backend := config.GetConfig().Backends[model.Backend]
	endpoint := backend.Endpoint
	ollamaClientMu.Lock()
	defer ollamaClientMu.Unlock()
	if ollamaClient != nil && ollamaClientEndpoint == endpoint {
		return ollamaClient, nil
	}
	base, err := url.Parse(endpoint)
	if err != nil {
		logger.Error("invalid ollama endpoint", slog.String("func", "ollamaBackend.StreamChat"), slog.Any("error", err))
		return nil, err
	}
	ollamaClient = api.NewClient(base, http.DefaultClient)
	ollamaClientEndpoint = endpoint
	return ollamaClient, nil
}

// mapThink converts the model's reasoning_effort setting to Ollama's think
// parameter. Empty means "server default" (nil), "none" disables thinking,
// and named levels are passed through verbatim.
func mapThink(reasoningEffort string) *api.ThinkValue {
	switch reasoningEffort {
	case "":
		return nil
	case "none":
		return &api.ThinkValue{Value: false}
	default:
		return &api.ThinkValue{Value: reasoningEffort}
	}
}

// mapOllamaOptions converts the model's tweaks to Ollama runtime options.
// context_length -> num_ctx and max_tokens -> num_predict; the penalties map
// 1:1. As with the OpenAI payload, the extended penalties are only sent for
// models with tweak_level advanced.
func mapOllamaOptions(model config.Model) map[string]any {
	t := model.Tweaks
	opts := make(map[string]any)
	if t.Temperature != 0 {
		opts["temperature"] = t.Temperature
	}
	if t.ContextLength != 0 {
		opts["num_ctx"] = t.ContextLength
	}
	if t.MaxTokens != 0 {
		opts["num_predict"] = t.MaxTokens
	}
	if !model.UseMinimalTweaks() {
		if t.RepeatPenalty != 0 {
			opts["repeat_penalty"] = t.RepeatPenalty
		}
		if t.PresencePenalty != 0 {
			opts["presence_penalty"] = t.PresencePenalty
		}
		if t.FrequencyPenalty != 0 {
			opts["frequency_penalty"] = t.FrequencyPenalty
		}
	}
	if len(opts) == 0 {
		return nil
	}
	return opts
}

func (ollamaBackend) StreamChat(ctx context.Context, chatID int64, promptID int, uc chan contract.CompletionUpdate) error {
	model := getModel(chatID)

	client, err := getOllamaClient(model)
	if err != nil {
		return err
	}

	history := generateMessages(chatID, promptID, []contract.ChatMessage{{
		Role:    "system",
		Content: getSystemPrompt(chatID),
	}})
	messages := make([]api.Message, len(history))
	for i, m := range history {
		messages[i] = api.Message{Role: m.Role, Content: m.Content}
	}

	stream := true
	req := &api.ChatRequest{
		Model:    model.Model,
		Messages: messages,
		Stream:   &stream,
		Think:    mapThink(model.ReasoningEffort),
		Options:  mapOllamaOptions(model),
	}

	reqJSON, err := json.Marshal(req)
	if err != nil {
		logger.Error("failed to marshal ollama request", slog.String("func", "ollamaBackend.StreamChat"), slog.Any("error", err))
		return err
	}
	logger.Debug("ollama chat request", slog.String("payload", string(reqJSON)))

	var content, thinking string
	var sentLast bool
	err = client.Chat(ctx, req, func(resp api.ChatResponse) error {
		content += resp.Message.Content
		thinking += resp.Message.Thinking

		isLast := resp.Done
		if !isLast && resp.Message.Content == "" && resp.Message.Thinking == "" {
			return nil
		}
		uc <- contract.CompletionUpdate{
			Content:   content,
			Reasoning: thinking,
			IsLast:    isLast,
		}
		if isLast {
			sentLast = true
		}
		return nil
	})
	if err != nil {
		logger.Error("ollama chat request failed", slog.String("func", "ollamaBackend.StreamChat"), slog.Any("error", err))
		return err
	}

	if !sentLast {
		// The stream ended without a terminal chunk - make sure the consumer
		// isn't left blocked waiting on IsLast forever. Mirrors the OpenAI
		// backend's fallback.
		uc <- contract.CompletionUpdate{
			Content:   content,
			Reasoning: thinking,
			IsLast:    true,
		}
	}
	return nil
}
