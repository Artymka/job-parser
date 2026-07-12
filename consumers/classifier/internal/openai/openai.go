// package openai

// import (
// 	"context"

// 	"github.com/Azure/azure-sdk-for-go/sdk/ai/azopenai"
// 	"github.com/artymka/jobparser-consumer-classifier/internal/config"
// )

// type Client struct {
// 	client *azopenai.Client
// 	model  string
// }

// func NewClient(config *config.Config) (*Client, error) {
// 	// Создаем credential с API ключом
// 	credential := azopenai.NewKeyCredential(config.AIKey)

// 	// Создаем клиент с endpoint
// 	client, err := azopenai.NewClientWithKeyCredential(
// 		config.AIEndpoint,
// 		credential,
// 		&azopenai.ClientOptions{
// 			APIVersion: "2024-12-01-preview",
// 		},
// 	)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &Client{
// 		client: client,
// 		model:  config.AIModel,
// 	}, nil
// }

// func (c *Client) Request(prompt string) (string, error) {
// 	resp, err := c.client.GetChatCompletions(
// 		context.Background(),
// 		azopenai.ChatCompletionsOptions{
// 			Messages: []azopenai.ChatMessage{
// 				{
// 					Role:    azopenai.ChatRoleUser.Ptr(),
// 					Content: &prompt,
// 				},
// 			},
// 			Model: &c.model,
// 		},
// 		nil,
// 	)
// 	if err != nil {
// 		return "", err
// 	}

// 	return *resp.Choices[0].Message.Content, nil
// }

// =======================================================================================================

// package openai

// import (
// 	"context"

// 	"github.com/artymka/jobparser-consumer-classifier/internal/config"
// 	"github.com/openai/openai-go/v3"
// 	"github.com/openai/openai-go/v3/azure"
// )

// type Client struct {
// 	client     *openai.Client
// 	aiKey      string
// 	aiModel    string
// 	aiEndpoint string
// }

// func NewClient(config *config.Config) *Client {
// 	client := openai.NewClient(
// 		azure.WithEndpoint(config.AIEndpoint, "2024-12-01-preview"),
// 		azure.WithAPIKey(config.AIKey),
// 	)

// 	res := Client{
// 		client:     &client,
// 		aiKey:      config.AIKey,
// 		aiModel:    config.AIModel,
// 		aiEndpoint: config.AIEndpoint,
// 	}
// 	return &res
// }

// func (c *Client) Request(prompt string) (string, error) {
// 	resp, err := c.client.Chat.Completions.New(context.Background(), openai.ChatCompletionNewParams{
// 		Messages: []openai.ChatCompletionMessageParamUnion{
// 			openai.UserMessage(prompt),
// 		},
// 		Model: openai.ChatModel(c.aiModel),
// 	})

// 	if err != nil {
// 		return "", err
// 	}

// 	return resp.Choices[0].Message.Content, nil
// }

package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/artymka/jobparser-consumer-classifier/internal/config"
)

type Client struct {
	endpoint string
	apiKey   string
	model    string
	httpCli  *http.Client
}

func NewClient(config *config.Config) *Client {
	return &Client{
		endpoint: config.AIEndpoint,
		apiKey:   config.AIKey,
		model:    config.AIModel,
		httpCli: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Messages []chatMessage `json:"messages"`
	Model    string        `json:"model"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func (c *Client) Request(prompt string) (string, error) {
	// Формируем тело запроса
	reqBody := chatRequest{
		Messages: []chatMessage{
			{Role: "system", Content: SystemPrompt},
			{Role: "user", Content: prompt},
		},
		Model: c.model,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	// Формируем URL
	url := fmt.Sprintf("%s/chat/completions?api-version=2024-12-01-preview", c.endpoint)

	// Создаем запрос
	req, err := http.NewRequestWithContext(context.Background(), "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	// Заголовки как в Python
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", c.apiKey)
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	// Отправляем
	resp, err := c.httpCli.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Проверяем статус
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("status %d: %s", resp.StatusCode, string(body))
	}

	// Парсим ответ
	var chatResp chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return "", err
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return chatResp.Choices[0].Message.Content, nil
}
