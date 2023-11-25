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

	if err := handleConfigSettings(cfg); err != nil {
		panic(err)
	}

	if cfg.query == "" {
		fmt.Println("Usage: maestro <query>")
		os.Exit(1)
	}

	if err := processQuery(cfg); err != nil {
		panic(err)
	}
}

func parseFlags() *config {
	maestroFlags := flag.NewFlagSet("maestro", flag.ExitOnError)
	cfg := &config{
		four:               maestroFlags.Bool("4", false, "Use OpenAI GPT-4"),
		three:              maestroFlags.Bool("3", false, "Use OpenAI GPT-3"),
		execFlag:           maestroFlags.Bool("e", false, "Run the command instead of printing it"),
		ollamaModel:        maestroFlags.String("m", "codellama:7b-instruct", "Model to use"),
		oaiToken:           maestroFlags.String("set-openai-token", "", "Set OpenAI API token"),
		ollamaUrl:          maestroFlags.String("set-ollama-url", "", "Set the ollama server URL"),
		ollamaDefaultModel: maestroFlags.String("set-ollama-model", "", "Set the default ollama model"),
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
		os.Exit(0)
	}

	if *cfg.ollamaUrl != "" {
		endpoint := sanitizeEndpoint(*cfg.ollamaUrl)
		if err := db.Badger.Set("ollama-url", endpoint); err != nil {
			return err
		}
		fmt.Println(utils.ColorGreen + "Ollama URL set." + utils.ColorReset)
		os.Exit(0)
	}

	if *cfg.ollamaDefaultModel != "" {
		if err := db.Badger.Set("ollama-model", *cfg.ollamaDefaultModel); err != nil {
			return err
		}
		fmt.Println(utils.ColorGreen + "Ollama default model set to " + *cfg.ollamaDefaultModel + utils.ColorReset)
		os.Exit(0)
	}
	return nil
}

func processQuery(cfg *config) error {
	context, err := utils.GetContext()
	if err != nil {
		return err
	}

	prompt := context + "\n\n" + cfg.query + "\n"
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
			fmt.Println("Command execution cancelled.")
			os.Exit(0)
		}
	}
}

func sanitizeEndpoint(url string) string {
	if !strings.HasPrefix(url, "http") && !strings.HasPrefix(url, "https") {
		url = fmt.Sprintf("http://%s", url)
	}
	url = strings.TrimSuffix(url, "/")
	url = strings.ReplaceAll(url, "/api/generate", "")
	return fmt.Sprintf("%s/api/generate", url)
}

type config struct {
	four               *bool
	three              *bool
	execFlag           *bool
	ollamaModel        *string
	oaiToken           *string
	ollamaUrl          *string
	ollamaDefaultModel *string
	query              string
}
