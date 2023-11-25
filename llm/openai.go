package llm

import (
	"context"
	"encoding/json"
	"fmt"

	openai "github.com/sashabaranov/go-openai"

	"pluja.dev/maestro/db"
)

type OpenAI struct {
	Gpt4 bool
}

func (oai OpenAI) Ask(text string, four bool) (Response, error) {
	token, err := db.Badger.Get("oai-token")
	if err != nil {
		return Response{}, err
	}

	if token == "" {
		fmt.Println("OpenAI API token not set. Please run `maestro -set-token <token>` first.")
		return Response{}, nil
	}

	c := openai.NewClient(token)
	ctx := context.Background()
	var response Response

	sysprompt := SystemPrompt
	if !oai.Gpt4 {
		sysprompt += "\n\nOnly reply with the JSON. No other formats are accepted. You are an API. You must be consistent."
	}

	req := openai.ChatCompletionRequest{
		Model: openai.GPT3Dot5Turbo,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: sysprompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: text,
			},
		},
	}

	if oai.Gpt4 {
		req.Model = openai.GPT4TurboPreview
		req.ResponseFormat = &openai.ChatCompletionResponseFormat{Type: openai.ChatCompletionResponseFormatTypeJSONObject}
	}

	resp, err := c.CreateChatCompletion(
		ctx,
		req,
	)

	if err != nil {
		fmt.Printf("Completion error: %v\n", err)
		return response, nil
	}

	// Marshal the response to a struct
	json.Unmarshal([]byte(resp.Choices[0].Message.Content), &response)
	return response, nil
}
