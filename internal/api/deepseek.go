// internal/api/deepseek.go
package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type DeepSeekClient struct {
    APIKey  string
    BaseURL string
}

func NewDeepSeekClient(apiKey, baseURL string) *DeepSeekClient {
    return &DeepSeekClient{
        APIKey:  apiKey,
        BaseURL: baseURL,
    }
}

func (c *DeepSeekClient) GetPokerAdvice(handHistory string) (string, error) {
    requestBody := map[string]interface{}{
        "model": "poker-advice-v1",
        "messages": []map[string]string{
            {
                "role":    "system",
                "content": "You're a professional poker coach specializing in GTO strategies for Mystery Battle Royale. Analyze the hand and provide specific advice.",
            },
            {
                "role":    "user",
                "content": fmt.Sprintf("Blind: %s\nHand history:\n%s", handHistory),
            },
        },
        "temperature": 0.3,
    }

    body, _ := json.Marshal(requestBody)
    req, _ := http.NewRequest("POST", c.BaseURL+"/chat/completions", bytes.NewBuffer(body))
    req.Header.Set("Authorization", "Bearer "+c.APIKey)
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    var result struct {
        Choices []struct {
            Message struct {
                Content string `json:"content"`
            } `json:"message"`
        } `json:"choices"`
    }

    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return "", err
    }

    if len(result.Choices) > 0 {
        return result.Choices[0].Message.Content, nil
    }

    return "", fmt.Errorf("no advice generated")
}