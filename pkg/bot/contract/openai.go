package contract

type ChatCompletionPayloadMinimal struct {
	BasicModelTweaks
	ModelOptions
	Messages []ChatMessage `json:"messages"`
}

type BasicModelTweaks struct {
	ContextLength int     `json:"context_length"`
	MaxTokens     int     `json:"max_tokens"`
	Temperature   float32 `json:"temperature"`
}

type ChatCompletionPayload struct {
	ModelOptions
	ModelTweaks
	Messages []ChatMessage `json:"messages"`
}

type ModelOptions struct {
	Model  string `json:"model"`
	Stream bool   `json:"stream"`
}

type ModelTweaks struct {
	ContextLength    int     `json:"context_length"`
	MaxTokens        int     `json:"max_tokens"`
	FrequencyPenalty float32 `json:"frequency_penalty"`
	PresencePenalty  float32 `json:"presence_penalty"`
	Temperature      float32 `json:"temperature"`
	RepeatPenalty    float32 `json:"repeat_penalty"`
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatCompletionResponse struct {
	Choices []ChatCompletionChoice `json:"choices"`
}

type ChatCompletionChoice struct {
	Delta struct {
		Content string `json:"content"`
	} `json:"delta"`
	FinishReason string `json:"finish_reason"`
}
