package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"pluja.dev/maestro/db"
	"pluja.dev/maestro/llm"
	"pluja.dev/maestro/utils"
)

var maestroFlags = flag.NewFlagSet("maestro", flag.ExitOnError)

var (
	four      = maestroFlags.Bool("4", false, "Use OpenAI GPT-4")
	three     = maestroFlags.Bool("3", false, "Use OpenAI GPT-3")
	execFlag  = maestroFlags.Bool("e", false, "Run the command instead of printing it")
	model     = maestroFlags.String("m", "codellama:7b-instruct", "Model to use")
	oaiToken  = maestroFlags.String("set-openai-token", "", "Set OpenAI API token")
	ollamaUrl = maestroFlags.String("set-ollama-url", "", "Set the ollama server URL")
)

func init() {
	maestroFlags.Parse(os.Args[1:])
	db.Init()
}

func main() {
	if *oaiToken != "" {
		db.Badger.Set("oai-token", *oaiToken)
		fmt.Println("OpenAI API token set.")
		os.Exit(0)
	}

	if *ollamaUrl != "" {
		endpoint := *ollamaUrl
		if !strings.HasPrefix(endpoint, "http") || !strings.HasPrefix(endpoint, "https") {
			endpoint = fmt.Sprintf("http://%s", endpoint)
		}
		endpoint = strings.TrimSuffix(endpoint, "/")
		endpoint = strings.ReplaceAll(endpoint, "/api/generate", "")
		endpoint = fmt.Sprintf("%s/api/generate", endpoint)
		db.Badger.Set("ollama-url", endpoint)
		fmt.Println(utils.ColorGreen + "Ollama URL set." + utils.ColorReset)
		os.Exit(0)
	}

	// Use flag.Args() to get non-flag arguments
	query := strings.Join(maestroFlags.Args(), " ")
	if query == "" {
		fmt.Println("Usage: maestro <query>")
		os.Exit(1)
	}

	context, err := utils.GetContext()
	if err != nil {
		panic(err)
	}

	prompt := context + "\n\n" + query + "\n"

	var ai llm.Llm

	if *four || *three {
		if *four && *three {
			fmt.Println(utils.ColorRed + "You can't use both -3 and -4 flags at the same time." + utils.ColorReset)
			os.Exit(1)
		}
		ai = llm.OpenAI{
			Gpt4: *four,
		}
	} else {
		endpoint, err := db.Badger.Get("ollama-url")
		if err != nil {
			fmt.Println(utils.ColorRed + "Ollama URL not set. Please run `maestro -set-ollama-url <url>` first." + utils.ColorReset)
			os.Exit(1)
		}
		ai = llm.Ollama{
			Model:    *model,
			Endpoint: endpoint,
		}
	}

	response, err := ai.Ask(prompt, *four)
	if err != nil {
		panic(err)
	}

	for _, command := range response.Commands {
		fmt.Println(utils.ColorComment + "# " + command.Comment + utils.ColorReset)
		fmt.Println("$ " + utils.ColorGreen + command.Command + utils.ColorReset)
	}

	if *execFlag {
		for _, command := range response.Commands {
			fmt.Println(utils.ColorYellow + "\n🔥 " + command.Command + utils.ColorReset)
			confirmation, _ := utils.GetUserConfirmation(fmt.Sprint(utils.ColorBlue + "🔥🏃 Run?" + utils.ColorReset))
			if confirmation {
				utils.RunCommand(command.Command)
			} else {
				fmt.Println("Command execution cancelled.")
				os.Exit(0)
			}
		}
	}
}
