package contract

type ChatCompletionPayloadMinimal struct {
	BasicModelTweaks
	ModelOptions
	Messages []ChatMessage `json:"messages"`
}

type BasicModelTweaks struct {
	ContextLength int     `json:"context_length"`
	MaxTokens     int     `json:"max_tokens"`
	Temperature   float64 `json:"temperature"`
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
	FrequencyPenalty float64 `json:"frequency_penalty"`
	PresencePenalty  float64 `json:"presence_penalty"`
	Temperature      float64 `json:"temperature"`
	RepeatPenalty    float64 `json:"repeat_penalty"`
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
