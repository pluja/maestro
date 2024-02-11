package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	openai "github.com/sashabaranov/go-openai"
	"pluja.dev/maestro/db"
)

var SystemPrompt = `As an expert Shell command interpreter, your directives are:
- Ensure most optimal and direct solutions.
- Beware of the user's environment.
- Make sure commands are compatible with the OS.
- Keep comments short and concise.
- Ensure commands are executable as provided, requiring no alterations.
- Avoid text editors; use sed for file editing.
- Employ sudo as needed for administrative tasks.
- Use echo for file creation.

Adhere to this JSON response structure:
{
	"commands": [
		{
			"command": "Your bash command here",
			"comment": "Concise explanation or context"
		}
	]
}
\n\nOnly JSON replies allowed. No other formats are accepted. You are an API. Be consistent.
`

type Llm struct {
	Oai    bool
	Ollama Ollama
	Openai OpenAI
}

type Ollama struct {
	Endpoint string
	Model    string
}

type OpenAI struct {
	Gpt4 bool
}

type Response struct {
	Commands []Command `json:"commands"`
}

type Command struct {
	Command string `json:"command"`
	Comment string `json:"comment"`
}

func (llm Llm) Ask(text string) (Response, error) {
	client, model, err := llm.setupClientAndModel()
	if err != nil {
		return Response{}, err
	}

	ctx := context.Background()
	req := llm.prepareCompletionRequest(text, model) // Pass the model here

	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		log.Printf("Completion error: %v\n", err)
		return Response{}, err
	}

	var response Response
	if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), &response); err != nil {
		return Response{}, err
	}
	return response, nil
}

func (llm Llm) setupClientAndModel() (*openai.Client, string, error) {
	if llm.Oai {
		token, err := db.Badger.Get("oai-token")
		if err != nil {
			return nil, "", err
		}
		if token == "" {
			return nil, "", fmt.Errorf("OpenAI API token not set. Please run `maestro -set-token <token>` first")
		}
		client := openai.NewClient(token)
		model := openai.GPT3Dot5Turbo0125
		if llm.Openai.Gpt4 {
			model = openai.GPT4TurboPreview
		}
		return client, model, nil
	}
	openaiConfig := openai.DefaultConfig("ollama")
	openaiConfig.BaseURL = fmt.Sprintf("%s/v1", llm.Ollama.Endpoint)
	client := openai.NewClientWithConfig(openaiConfig)
	return client, llm.Ollama.Model, nil
}

func (llm Llm) prepareCompletionRequest(text string, model string) openai.ChatCompletionRequest { // Accept model as a parameter
	systemP := SystemPrompt

	return openai.ChatCompletionRequest{
		Model: model, // Use the model parameter
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: systemP},
			{Role: openai.ChatMessageRoleUser, Content: text},
		},
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONObject,
		},
	}
}
