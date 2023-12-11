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

func main() {
	cfg := parseFlags()

	db.Init()
	defer db.Badger.Close()

	if err := handleConfigSettings(cfg); err != nil {
		fmt.Println(utils.ColorRed + "Error: " + err.Error() + utils.ColorReset)
	}

	if cfg.query == "" {
		fmt.Println("Usage: maestro <query>")
		db.Badger.Close()
		os.Exit(1)
	}

	if err := processQuery(cfg); err != nil {
		fmt.Println(utils.ColorRed + "Error: " + err.Error() + utils.ColorReset)
		db.Badger.Close()
		os.Exit(1)
	}
}

func parseFlags() *config {
	maestroFlags := flag.NewFlagSet("maestro", flag.ExitOnError)
	cfg := &config{
		four:                maestroFlags.Bool("4", false, "Use OpenAI GPT-4"),
		three:               maestroFlags.Bool("3", false, "Use OpenAI GPT-3"),
		dev:                 maestroFlags.Bool("dev", false, "Enable development mode"),
		execFlag:            maestroFlags.Bool("e", false, "Run the command instead of printing it"),
		enableFolderContext: maestroFlags.Bool("wc", false, "Enable the folder context (files and folders)"),
		ollamaModel:         maestroFlags.String("m", "codellama:7b-instruct", "Model to use"),
		oaiToken:            maestroFlags.String("set-openai-token", "", "Set OpenAI API token"),
		ollamaUrl:           maestroFlags.String("set-ollama-url", "", "Set the ollama server URL"),
		ollamaDefaultModel:  maestroFlags.String("set-ollama-model", "", "Set the default ollama model"),
	}

	if *cfg.ollamaModel != "codellama:7b-instruct" {
		fmt.Println(utils.ColorBlue + "Using model " + *cfg.ollamaModel + utils.ColorReset)
	}

	maestroFlags.Parse(os.Args[1:])
	cfg.query = strings.Join(maestroFlags.Args(), " ")
	return cfg
}

func handleConfigSettings(cfg *config) error {
	if *cfg.oaiToken != "" {
		if err := db.Badger.Set("oai-token", *cfg.oaiToken); err != nil {
			return err
		}
		fmt.Println("OpenAI API token set.")
		db.Badger.Close()
		os.Exit(0)
	}

	if *cfg.ollamaUrl != "" {
		endpoint := sanitizeEndpoint(*cfg.ollamaUrl)
		if err := db.Badger.Set("ollama-url", endpoint); err != nil {
			return err
		}
		fmt.Println(utils.ColorGreen + "Ollama URL set." + utils.ColorReset)
		db.Badger.Close()
		os.Exit(0)
	}

	if *cfg.ollamaDefaultModel != "" {
		if err := db.Badger.Set("ollama-model", *cfg.ollamaDefaultModel); err != nil {
			return err
		}
		fmt.Println(utils.ColorGreen + "Ollama default model set to " + *cfg.ollamaDefaultModel + utils.ColorReset)
		db.Badger.Close()
		os.Exit(0)
	}
	return nil
}

func processQuery(cfg *config) error {
	prompt := "**TASK: " + cfg.query + "?**\n"
	if cfg.enableFolderContext != nil && *cfg.enableFolderContext {
		context, err := utils.GetContext(*cfg.enableFolderContext)
		if err != nil {
			return err
		}
		prompt += "```CONTEXT: " + context + "```"
	}

	ai, err := selectAI(cfg)
	if err != nil {
		return err
	}

	response, err := ai.Ask(prompt, *cfg.four)
	if err != nil {
		return err
	}

	displayResponse(response, cfg)
	return nil
}

func selectAI(cfg *config) (llm.Llm, error) {
	if *cfg.four || *cfg.three {
		if *cfg.four && *cfg.three {
			fmt.Println(utils.ColorRed + "You can't use both -3 and -4 flags at the same time." + utils.ColorReset)
			os.Exit(1)
		}
		return llm.OpenAI{Gpt4: *cfg.four}, nil
	}

	endpoint, err := db.Badger.Get("ollama-url")
	if err != nil {
		return nil, fmt.Errorf(utils.ColorRed + "Ollama URL not set. Please run `maestro -set-ollama-url <url>` first." + utils.ColorReset)
	}

	model, err := db.Badger.Get("ollama-model")
	if err != nil {
		model = "codellama:7b-instruct"
	}
	if *cfg.ollamaModel != "" {
		model = *cfg.ollamaModel
	}

	return llm.Ollama{
		Model:    model,
		Endpoint: endpoint,
	}, nil
}

func displayResponse(response llm.Response, cfg *config) {
	for _, command := range response.Commands {
		fmt.Println(utils.ColorComment + "# " + command.Comment + utils.ColorReset)
		fmt.Println("$ " + utils.ColorGreen + command.Command + utils.ColorReset)
	}

	if *cfg.execFlag {
		executeCommands(response)
	}
}

func executeCommands(response llm.Response) {
	for _, command := range response.Commands {
		fmt.Println(utils.ColorYellow + "\n🔥 " + command.Command + utils.ColorReset)
		confirmation, _ := utils.GetUserConfirmation(fmt.Sprint(utils.ColorBlue + "🔥🏃 Run?" + utils.ColorReset))
		if confirmation {
			utils.RunCommand(command.Command)
		} else {
			fmt.Println(utils.ColorBlue + "[X] Command execution cancelled." + utils.ColorReset)
			db.Badger.Close()
			os.Exit(0)
		}
	}
}

func sanitizeEndpoint(url string) string {
	if !strings.HasPrefix(url, "http") && !strings.HasPrefix(url, "https") {
		url = fmt.Sprintf("http://%s", url)
	}
	url = strings.TrimSuffix(url, "/")
	url = strings.ReplaceAll(url, "/api/chat", "")
	return fmt.Sprintf("%s/api/chat", url)
}

type config struct {
	four                *bool
	three               *bool
	dev                 *bool
	execFlag            *bool
	ollamaModel         *string
	oaiToken            *string
	ollamaUrl           *string
	ollamaDefaultModel  *string
	query               string
	enableFolderContext *bool
}
