package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type AIService struct {
	ApiKey string
	Model  string
	// SystemPrompt string // میتونی از متغیر استفاده کنی یا از فایل بفرستی
	SystemPrompt string
}

type ParsedSystemOutput struct {
	Action         string                 `json:"action"`
	Data           map[string]interface{} `json:"data"`
	Filters        map[string]interface{} `json:"filters,omitempty"`
	AssistantReply string                 `json:"assistant_reply,omitempty"`
}

func NewAIService(apiKey, model, systemPrompt string) *AIService {
	return &AIService{ApiKey: apiKey, Model: model, SystemPrompt: systemPrompt}
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIRequest struct {
	Model     string        `json:"model"`
	Messages  []chatMessage `json:"messages"`
	MaxTokens int           `json:"max_tokens,omitempty"`
}

type openAIResponse struct {
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func (s *AIService) ProcessMessage(userMessage string) (*ParsedSystemOutput, string, error) {
	// Build request
	req := openAIRequest{
		Model: s.Model,
		Messages: []chatMessage{
			{Role: "system", Content: s.SystemPrompt},
			{Role: "user", Content: userMessage},
		},
		MaxTokens: 800,
	}

	b, _ := json.Marshal(req)
	httpReq, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(b))
	if err != nil {
		return nil, "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+s.ApiKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	var oresp openAIResponse
	if err := json.Unmarshal(body, &oresp); err != nil {
		// return raw body in error for debugging
		return nil, string(body), fmt.Errorf("openai response parse error: %v", err)
	}
	if len(oresp.Choices) == 0 {
		return nil, string(body), fmt.Errorf("no choices returned")
	}

	fmt.Println(oresp)
	assistantText := oresp.Choices[0].Message.Content
	fmt.Println(assistantText)

	var result ParsedSystemOutput
	if err := json.Unmarshal([]byte(assistantText), &result); err != nil {
		return nil, assistantText, fmt.Errorf("invalid json: %v", err)
	}

	// // extract <system_output> block
	// re := regexp.MustCompile(`(?s)<system_output>\s*(\{.*?\})\s*</system_output>`)
	// m := re.FindStringSubmatch(assistantText)
	// if len(m) < 2 {
	// 	return nil, assistantText, fmt.Errorf("no system_output block found in assistant message")
	// }

	// // fmt.Println(assistantText, m)
	// var parsed ParsedSystemOutput
	// if err := json.Unmarshal([]byte(m[1]), &parsed); err != nil {
	// 	return nil, assistantText, fmt.Errorf("failed to parse system_output JSON: %v", err)
	// }

	return &result, assistantText, nil
}

// helper to load API key from env if needed
func NewAIServiceFromEnv(systemPrompt string) *AIService {
	// const key = "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"
	const key = ""
	const model = "gpt-4.1-mini"
	return NewAIService(key, model, systemPrompt)
}
