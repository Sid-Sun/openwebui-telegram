package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/sid-sun/openwebui-telegram/cmd/config"
	"github.com/sid-sun/openwebui-telegram/pkg/bot/contract"
)

// Backend streams a chat completion for the prompt identified by
// (chatID, promptID), emitting cumulative CompletionUpdates on uc. The final
// update must have IsLast set so consumers can stop ranging.
type Backend interface {
	StreamChat(ctx context.Context, chatID int64, promptID int, uc chan contract.CompletionUpdate) error
}

func newBackend(model config.Model) (Backend, error) {
	backendType := config.GetConfig().Backends[model.Backend].Type
	switch backendType {
	case "openai":
		return openAIBackend{}, nil
	case "openaisdk":
		return openAISDKBackend{}, nil
	case "ollama":
		return ollamaBackend{}, nil
	default:
		return nil, fmt.Errorf("unknown backend %q for model %q", model.Backend, model.Name)
	}
}

// GetChatResponseStream dispatches to the backend configured for the chat's
// active model. The signature is unchanged so existing callers (e.g. the
// completion handler) work with either backend.
func GetChatResponseStream(chatID int64, promptID int, uc chan contract.CompletionUpdate) error {
	model := getModel(chatID)
	backend, err := newBackend(model)
	if err != nil {
		logger.Error("unknown backend", slog.String("func", "GetChatResponseStream"), slog.Any("error", err))
		return err
	}
	return backend.StreamChat(context.Background(), chatID, promptID, uc)
}
