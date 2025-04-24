package models

type OpenAIModelResponse struct {
	ID      string `json:"id"`
	Created int64  `json:"created"`
	Object  string `json:"object"`
}

type OpenAIChatCompletionsRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

// 定义结构体
type MessagePart struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type Message struct {
	Role    string        `json:"role"`
	Content string        `json:"content"`
	Parts   []MessagePart `json:"parts,omitempty"`
}

func (oai *Message) ToSciraMessage() Message {
	return Message{
		Role:    oai.Role,
		Content: oai.Content,
		Parts: []MessagePart{
			{
				Type: "text",
				Text: oai.Content,
			},
		},
	}
}

type SciraChatCompletionsRequest struct {
	ID            string    `json:"id"`
	Messages      []Message `json:"messages"`
	SelectedModel string    `json:"selectedModel"`
	McpServers    []any     `json:"mcpServers"`
	ChatId        string    `json:"chatId"`
	UserId        string    `json:"userId"`
}

func (oai *OpenAIChatCompletionsRequest) ToSciraChatCompletionsRequest(model string, chatId string, userId string) *SciraChatCompletionsRequest {
	sciraMessages := make([]Message, len(oai.Messages))
	for i, message := range oai.Messages {
		sciraMessages[i] = message.ToSciraMessage()
	}

	return &SciraChatCompletionsRequest{
		ID:            chatId,
		SelectedModel: model,
		McpServers:    []any{},
		ChatId:        chatId,
		UserId:        userId,
		Messages:      sciraMessages,
	}

}

type OpenAIChatCompletionsStreamResponse struct {
	ID                string   `json:"id"`
	Object            string   `json:"object"`
	Provider          string   `json:"provider"`
	Model             string   `json:"model"`
	SystemFingerprint string   `json:"system_fingerprint"`
	Created           int64    `json:"created"`
	Choices           []Choice `json:"choices"`
	Usage             Usage    `json:"usage"`
}

type Choice struct {
	Index               int    `json:"index"`
	Delta               Delta  `json:"delta"`
	FinishReason        string `json:"finish_reason"`
	NaturalFinishReason string `json:"natural_finish_reason"`
	Logprobs            any    `json:"logprobs"`
}

type Delta struct {
	Role             string `json:"role"`
	Content          string `json:"content"`
	ReasoningContent string `json:"reasoning_content"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

func NewOaiStreamResponse(id string, time int64, model string, choices []Choice) *OpenAIChatCompletionsStreamResponse {
	return &OpenAIChatCompletionsStreamResponse{
		ID:       id,
		Object:   "chat.completion.chunk",
		Provider: "scira",
		Model:    model,
		Created:  time,
		Choices:  choices,
	}
}

func NewChoice(content string, reasoningContent string, finishReason string) []Choice {
	return []Choice{
		{
			Index:        0,
			Delta:        Delta{Role: "assistant", Content: content, ReasoningContent: reasoningContent},
			FinishReason: finishReason,
		},
	}
}
