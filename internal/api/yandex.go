package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"pokerbot/internal/service"
)

type YandexGPTClient struct {
    oauthToken   string
    folderID     string
	modelURI     string
    apiURL       string
    model        string
    temperature  float64
    maxTokens    int
    
    tokenMutex   sync.RWMutex
    iamToken     string
    expiresAt    time.Time
	logger       *service.Logger
	httpClient   *http.Client	
}

func NewYandexGPTClient(oauthToken, folderID string, logger *service.Logger) *YandexGPTClient {
    client := &YandexGPTClient{
        oauthToken:  oauthToken,
        folderID:    folderID,
		modelURI:    fmt.Sprintf("gpt://%s/yandexgpt", folderID),
        apiURL:      "https://llm.api.cloud.yandex.net/foundationModels/v1/completion",
        model:       "general",
        temperature: 0.3,
        maxTokens:   2000,
		logger: logger,
		httpClient:  &http.Client{Timeout: 30 * time.Second},
    }

    go client.tokenRefreshDaemon()
    return client
}

func (c *YandexGPTClient) tokenRefreshDaemon() {
    for {
        if err := c.refreshIAMTokenWithRetry(); err != nil {
            c.logger.Error().Err(err).Msg("background token refresh failed")
        }
        time.Sleep(55 * time.Minute)
    }
}
func (c *YandexGPTClient) refreshIAMTokenWithRetry() error {
    const maxRetries = 5
    backoff := []time.Duration{1, 2, 4, 8, 16} // секунды
    
    var lastErr error
    for attempt := 0; attempt < maxRetries; attempt++ {
        err := c.refreshIAMToken()
        if err == nil {
            return nil
        }
        
        lastErr = err
        if attempt < maxRetries-1 {
            time.Sleep(backoff[attempt] * time.Second)
        }
    }
    
    return fmt.Errorf("token refresh failed after %d attempts: %w", maxRetries, lastErr)
}

func (c *YandexGPTClient) refreshIAMToken() error {
    reqBody := struct {
        YandexPassportOauthToken string `json:"yandexPassportOauthToken"`
    }{
        YandexPassportOauthToken: c.oauthToken,
    }

    body, err := json.Marshal(reqBody)
    if err != nil {
        return fmt.Errorf("marshal error: %w", err)
    }

    req, err := http.NewRequest(
        "POST", 
        "https://iam.api.cloud.yandex.net/iam/v1/tokens",
        bytes.NewReader(body),
    )
    if err != nil {
        return fmt.Errorf("request creation failed: %w", err)
    }
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{Timeout: 10 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        return fmt.Errorf("HTTP request failed: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
    }

    var result struct {
        IamToken  string    `json:"iamToken"`
        ExpiresAt time.Time `json:"expiresAt"`
    }

    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return fmt.Errorf("JSON decode error: %w", err)
    }

    c.tokenMutex.Lock()
    defer c.tokenMutex.Unlock()
    c.iamToken = result.IamToken
    c.expiresAt = result.ExpiresAt

    return nil
}

func (c *YandexGPTClient) ensureValidToken() error {
    c.tokenMutex.RLock()
    tokenValid := time.Until(c.expiresAt) > 5*time.Minute
    c.tokenMutex.RUnlock()

    if tokenValid {
        return nil
    }

    return c.refreshIAMTokenWithRetry()
}

func (c *YandexGPTClient) GetPokerAdvice(handHistory string) (string, error) {
	if err := c.ensureValidToken(); err != nil {
        return "", fmt.Errorf("token validation failed: %w", err)
    }

    requestBody := map[string]interface{}{
        "modelUri": c.modelURI,
        "completionOptions": map[string]interface{}{
            "stream": false,
            "temperature": c.temperature,
            "maxTokens": strconv.Itoa(c.maxTokens),
        },
        "messages": []map[string]string{
            {
                "role": "system",
                "text": "Ты профессиональный покерный тренер GTO. Анализируй раздачи по следующим аспектам: " +
                    "1. Префлоп: Рейнджи открытия, 3-беты\n" +
                    "2. Постфлоп: Линии ставок, баланс валью/блеф\n" +
                    "3. Размеры ставок\n" + 
                    "4. Пот оддсы и эквити\n" +
                    "5. Возможные ошибки игроков\n" +
					"Совет должен быть структурированным, кратким и использовать покерную терминологию",
            },
            {
                "role": "user",
                "text": handHistory,
            },
        },
    }

    body, err := json.Marshal(requestBody)
    if err != nil {
        return "", fmt.Errorf("request marshal error: %w", err)
    }

    req, err := http.NewRequest("POST", c.apiURL, bytes.NewReader(body))
    if err != nil {
        return "", fmt.Errorf("request creation failed: %w", err)
    }

    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+c.iamToken)
    req.Header.Set("x-folder-id", c.folderID)

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return "", fmt.Errorf("API request failed: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return "", fmt.Errorf("API error: status %d, body: %s", resp.StatusCode, string(body))
    }

    var response struct {
        Result struct {
            Alternatives []struct {
                Message struct {
                    Text string `json:"text"`
                } `json:"message"`
            } `json:"alternatives"`
        } `json:"result"`
    }

    if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
        return "", fmt.Errorf("response decode error: %w", err)
    }

    if len(response.Result.Alternatives) == 0 {
        return "", errors.New("no alternatives in response")
    }

    return response.Result.Alternatives[0].Message.Text, nil
}