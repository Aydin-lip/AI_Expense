package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"example/AI/internal/models"
)

type AIService struct {
	ApiKey string
	Model  string
	// SystemPrompt string // میتونی از متغیر استفاده کنی یا از فایل بفرستی
	SystemPrompt string
}

type ParsedSystemOutput struct {
	Action         string                 `json:"action"`
	RequestContext map[string]interface{} `json:"request_context,omitempty"`
	Data           map[string]interface{} `json:"data"`
	Filters        map[string]interface{} `json:"filters"`
	Analysis       map[string]interface{} `json:"analysis"`
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

func (s *AIService) ProcessMessage(userMessage string, userID int, username, role string) (*ParsedSystemOutput, string, error) {
	// Build request
	req := openAIRequest{
		Model: s.Model,
		Messages: []chatMessage{
			{Role: "system", Content: s.SystemPrompt},
			{Role: "user", Content: fmt.Sprintf("(userID: %s, username: %s, role: %s) \n\n %s", userID, username, role, userMessage)},
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
	const key = "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"
	const model = "gpt-4.1-mini"
	return NewAIService(key, model, systemPrompt)
}

// اضافه کن داخل فایل services/ai_service.go یا همون جایی که AIService تعریف شده

// GenerateNaturalAnalysis: دریافت یک payload (نتایج محاسبات SQL) و تولید یک reply طبیعی توسط مدل
// برمی‌گرداند: (naturalText, rawAssistantText, err)
func (s *AIService) GenerateNaturalAnalysis(payload map[string]interface{}) (string, string, error) {
	// 1. marshal payload to pretty json for prompt
	payloadBytes, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal payload: %w", err)
	}
	payloadStr := string(payloadBytes)

	// 2. build prompt: system + user
	// system: لحن و قواعد (دوستانه، فارسی، بدون ایموجی)، user: داده ها و خواسته ی تحلیل
	systemMsg := `You are an AI assistant that summarizes numeric analytics into a short friendly Persian message.
Tone: friendly and conversational in Persian. No emojis.
Output: return ONLY a short human-readable Persian paragraph (1-4 sentences) that explains the provided analytics clearly and gives one actionable suggestion if appropriate.`

	userMsg := fmt.Sprintf(`Here is analytics data (JSON). Produce a short, friendly Persian summary (1-4 sentences) for the end user. Do NOT output JSON or metadata—only plain Persian text.

DATA:
%s

Rules:
- Mention main numbers (total, top category) if present.
- Use rounding for big numbers (e.g., 2.3M, 1.2k) if appropriate.
- Keep it short and practical.
`, payloadStr)

	// 3. build request to OpenAI Chat Completions (raw HTTP)
	reqBody := map[string]interface{}{
		"model": s.Model, // e.g., "gpt-4.1-mini"
		"messages": []map[string]string{
			{"role": "system", "content": systemMsg},
			{"role": "user", "content": userMsg},
		},
		"max_tokens":  300,
		"temperature": 0.2,
	}

	reqBytes, _ := json.Marshal(reqBody)
	httpReq, err := http.NewRequestWithContext(context.Background(), "POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(reqBytes))
	if err != nil {
		return "", "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+s.ApiKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	respBytes, _ := ioutil.ReadAll(resp.Body)
	// Try parse minimal response structure
	var o struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error interface{} `json:"error,omitempty"`
	}
	if err := json.Unmarshal(respBytes, &o); err != nil {
		// return raw for debugging
		return "", string(respBytes), fmt.Errorf("failed to parse openai response: %w", err)
	}
	if len(o.Choices) == 0 {
		return "", string(respBytes), fmt.Errorf("no choices returned")
	}

	assistantText := o.Choices[0].Message.Content
	// assistantText is the natural text we want
	return assistantText, string(respBytes), nil
}

// SumAmount already داشتیم؛ بیارش
func (s *PurchaseService) CountPurchases(filter models.PurchaseFilter) (int64, error) {
	db := s.DB.Model(&models.Purchase{})
	// apply same filters...
	var cnt int64
	if err := db.Count(&cnt).Error; err != nil {
		return 0, err
	}
	return cnt, nil
}
