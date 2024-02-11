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

type Config struct {
	Four                bool
	Three               bool
	Dev                 bool
	ExecFlag            bool
	OllamaModel         string
	OAIToken            string
	OllamaURL           string
	OllamaDefaultModel  string
	Query               string
	EnableFolderContext bool
}

func main() {
	cfg := parseFlags()

	db.Init()
	defer db.Badger.Close() // Simplified database closing logic

	if err := handleConfigSettings(&cfg); err != nil {
		exitWithError(err)
	}

	if cfg.Query == "" {
		fmt.Println("Usage: maestro [flags] <query>")
		os.Exit(1)
	}

	if err := processQuery(&cfg); err != nil {
		exitWithError(err)
	}
}

func parseFlags() Config {
	var cfg Config
	maestroFlags := flag.NewFlagSet("maestro", flag.ExitOnError)
	maestroFlags.BoolVar(&cfg.Four, "4", false, "Use OpenAI GPT-4")
	maestroFlags.BoolVar(&cfg.Three, "3", false, "Use OpenAI GPT-3")
	maestroFlags.BoolVar(&cfg.Dev, "dev", false, "Enable development mode")
	maestroFlags.BoolVar(&cfg.ExecFlag, "e", false, "Run the command instead of printing it")
	maestroFlags.BoolVar(&cfg.EnableFolderContext, "ctx", false, "Enable the folder context (files and folders list)")
	maestroFlags.StringVar(&cfg.OllamaModel, "m", "dolphin-mistral:latest", "Model to use")
	maestroFlags.StringVar(&cfg.OAIToken, "set-openai-token", "", "Set OpenAI API token")
	maestroFlags.StringVar(&cfg.OllamaURL, "set-ollama-url", "", "Set the ollama server URL")
	maestroFlags.StringVar(&cfg.OllamaDefaultModel, "set-ollama-model", "", "Set the default ollama model")
	maestroFlags.Parse(os.Args[1:])
	cfg.Query = strings.Join(maestroFlags.Args(), " ")
	return cfg
}

func handleConfigSettings(cfg *Config) error {
	if cfg.OAIToken != "" {
		return setAndExit("oai-token", cfg.OAIToken, "OpenAI API token set.")
	}

	if cfg.OllamaURL != "" {
		endpoint := utils.SanitizeEndpoint(cfg.OllamaURL)
		return setAndExit("ollama-url", endpoint, "Ollama URL set.")
	}

	if cfg.OllamaDefaultModel != "" {
		return setAndExit("ollama-model", cfg.OllamaDefaultModel, "Ollama default model set to "+cfg.OllamaDefaultModel)
	}
	return nil
}

func setAndExit(key, value, message string) error {
	if err := db.Badger.Set(key, value); err != nil {
		return err
	}
	fmt.Println(utils.ColorGreen + message + utils.ColorReset)
	os.Exit(0)
	return nil // This return is never reached but required by the compiler
}

func processQuery(cfg *Config) error {
	prompt := "**TASK: " + cfg.Query + "?**\n"
	context, err := utils.GetContext(cfg.EnableFolderContext)
	if err != nil {
		return err
	}
	prompt += "```CONTEXT: " + context + "```"

	ai, err := prepareAI(cfg)
	if err != nil {
		return err
	}

	response, err := ai.Ask(prompt)
	if err != nil {
		return err
	}

	displayResponse(response, cfg)
	return nil
}

func prepareAI(cfg *Config) (llm.Llm, error) {
	var ai llm.Llm

	ai.Oai = cfg.Four || cfg.Three
	ai.Openai.Gpt4 = cfg.Four
	ollamaEndpoint, err := db.Badger.Get("ollama-url")
	if err != nil {
		return ai, fmt.Errorf("ollama URL not set. Please run `maestro -set-ollama-url <url>` first")
	}

	ollamaModel, err := db.Badger.Get("ollama-model")
	if err != nil || cfg.OllamaModel != "dolphin-mistral:latest" {
		ollamaModel = cfg.OllamaModel
	}
	ai.Ollama.Endpoint = ollamaEndpoint
	ai.Ollama.Model = ollamaModel

	return ai, nil
}

func displayResponse(response llm.Response, cfg *Config) {
	for _, command := range response.Commands {
		fmt.Println(utils.ColorComment + "# " + command.Comment + utils.ColorReset)
		fmt.Println("$ " + utils.ColorGreen + command.Command + utils.ColorReset)
	}

	if cfg.ExecFlag {
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
			os.Exit(0)
		}
	}
}

func exitWithError(err error) {
	fmt.Printf("%sError: %s%s\n", utils.ColorRed, err.Error(), utils.ColorReset)
	os.Exit(1)
}
