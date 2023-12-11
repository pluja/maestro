package llm

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
)

type Ollama struct {
	Endpoint string
	Model    string
}

type OllamaResponse struct {
	Model     string  `json:"model,omitempty"`
	CreatedAt string  `json:"created_at,omitempty"`
	Message   Message `json:"message,omitempty"`
	Done      bool    `json:"done,omitempty"`
	Error     string  `json:"error,omitempty"`
}
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func (ol Ollama) Ask(prompt string, four bool) (Response, error) {
	var response Response

	type RequestBody struct {
		Model    string    `json:"model"`
		Messages []Message `json:"messages"`
		Format   string    `json:"format"`
		Stream   bool      `json:"stream"`
	}

	var messages []Message
	messages = append(messages, Message{
		Role:    "system",
		Content: SystemPrompt,
	})

	messages = append(messages, Message{
		Role:    "user",
		Content: prompt,
	})

	data := RequestBody{
		Model:    ol.Model,
		Messages: messages,
		Format:   "json",
		Stream:   false,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return response, err
	}

	req, err := http.NewRequest("POST", ol.Endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return response, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return response, err
	}
	defer resp.Body.Close()

	var ollamaResponse OllamaResponse
	json.NewDecoder(resp.Body).Decode(&ollamaResponse)
	if ollamaResponse.Error != "" {
		return response, fmt.Errorf(ollamaResponse.Error)
	}

	err = json.Unmarshal([]byte(ollamaResponse.Message.Content), &response)
	if err != nil {
		return response, err
	}

	return response, nil
}
