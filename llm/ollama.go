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
	Model     string `json:"model,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	Response  string `json:"response,omitempty"`
	Done      bool   `json:"done,omitempty"`
	Error     string `json:"error,omitempty"`
}

func (ol Ollama) Ask(prompt string, four bool) (Response, error) {
	var response Response

	type RequestBody struct {
		Model  string `json:"model"`
		Prompt string `json:"prompt"`
		Format string `json:"format"`
		System string `json:"system"`
		Stream bool   `json:"stream"`
	}

	data := RequestBody{
		Model:  ol.Model,
		Prompt: prompt,
		Format: "json",
		System: SystemPrompt,
		Stream: false,
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

	err = json.Unmarshal([]byte(ollamaResponse.Response), &response)
	if err != nil {
		return response, err
	}

	return response, nil
}
