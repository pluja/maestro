package llm

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"pluja.dev/maestro/db"
	"pluja.dev/maestro/utils"
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
	compatible, err := ol.CheckVersion()
	if err != nil {
		return Response{}, err
	}
	if !compatible {
		return Response{}, fmt.Errorf("ollama API version too old. Please update your Ollama instance to at least v0.1.14")
	}

	modelAvailable, err := ol.CheckModel(ol.Model)
	if err != nil {
		return Response{}, err
	}
	if !modelAvailable {
		return Response{}, fmt.Errorf("model %q not available. Please pull the model with 'ollama pull %s' and try again", ol.Model, ol.Model)
	}

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

	if strings.Contains(ol.Endpoint, "/api/chat") {
		ol.Endpoint = utils.SanitizeEndpoint(ol.Endpoint)
		db.Badger.Set("ollama-url", ol.Endpoint)
	}

	url := fmt.Sprintf("%s/api/chat", ol.Endpoint)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{
		Transport: tr,
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return response, err
	}

	req.Header.Set("Content-Type", "application/json")

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

func (ol Ollama) CheckVersion() (bool, error) {
	// APIVersionResponse represents the JSON response structure for the API version
	type APIVersionResponse struct {
		Version string `json:"version"`
	}

	url := strings.TrimSuffix(ol.Endpoint, "/")
	url = strings.ReplaceAll(url, "/api/chat", "")

	// Create a new TLS config, allowing insecure connections
	tlsConfig := &tls.Config{InsecureSkipVerify: true}

	// Set the TLS config on the client transport
	trans := &http.Transport{TLSClientConfig: tlsConfig}

	// Create a new client
	cli := &http.Client{Transport: trans}

	// Send a GET request to the API
	resp, err := cli.Get(fmt.Sprintf("%s/api/version", url))
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	// Decode the JSON response
	var apiResp APIVersionResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return false, err
	}

	// Compare the version
	return utils.CompareVersion(apiResp.Version, "0.1.14"), nil
}

func (ol Ollama) CheckModel(modelName string) (bool, error) {
	type Model struct {
		Name string `json:"name"`
	}
	type APIModelsResponse struct {
		Models []Model `json:"models"`
	}

	// Create a new TLS config, allowing insecure connections
	tlsConfig := &tls.Config{InsecureSkipVerify: true}

	// Set the TLS config on the client transport
	trans := &http.Transport{TLSClientConfig: tlsConfig}

	// Create a new client
	cli := &http.Client{Transport: trans}

	// Send a GET request to the API
	resp, err := cli.Get(fmt.Sprintf("%s/api/tags", ol.Endpoint))
	if err != nil {
		return false, err
	}

	// Decode the JSON response
	var apiResp APIModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return false, err
	}

	// Check if the model is available
	for _, model := range apiResp.Models {
		if model.Name == modelName {
			return true, nil
		}
	}

	return false, nil
}
