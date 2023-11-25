![maestro banner](banner.png)

Maestro is a terminal assistant powered by LLM models, it transforms your instructions into bash commands. Use it 100% offline, or use OpenAI's ChatGPT API.

![](maestro.svg)


## Features

- **Easy**: just type your instructions and press enter
- **Execute**: Use the `-e` flag to execute the command directly
  - You will be prompted to confirm the command before executing it
- **Context**: Maestro always has your system context: current directory, system, user, etc.
- **Multiple LLM models**
  - 100% offline usage with [Ollama](https://ollama.ai), with [more than 40 models](https://ollama.ai/library)
  - Use GPT4-Turbo or GPT3.5-Turbo

## Example



## Basic Installation

1. Get the latest binary from the [releases page](https://github.com/pluja/maestro/releases)
2. Run it: `./maestro -h`

> You'll have a much better experience if put the binary in a directory that's in your `$PATH` and rename it to `maestro`. For example: `sudo mv ./maestro /usr/local/bin/maestro`. Then you can run it with `maestro` from anywhere!

### Offline Usage with [Ollama](https://ollama.ai)

1. [Install ollama](https://ollama.ai/download), or [run it with docker](https://hub.docker.com/r/ollama/ollama)
2. Pull some models with `ollama pull <model-name>`
    - You must pull at least the default `codellama:7b-instruct` model: `ollama pull codellama:7b-instruct`
3. Start the server with `ollama serve` (default docker behavior)
4. Set the ollama server URL with `./maestro -set-ollama-url <ollama-url>`
   - Example: `./maestro -set-ollama-url http://localhost:8080`

### Usage with OpenAI's API

1. Get an API token from [OpenAI](https://platform.openai.com/)
2. Set the api token with `./maestro -set-openai-token <your-token>`

After setting your token, you can use the `-4` flag to use GPT4-Turbo or `-3` for GPT3.5-Turbo.