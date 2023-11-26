![maestro banner](banner.png)

`maestro` converts natural language instructions into cli commands. It's designed for both offline use with Ollama and online integration with ChatGPT API.

![](maestro.svg)


## Key Features

- **Ease of Use**: Simply type your instructions and press enter.
- **Direct Execution**: Utilize the `-e` flag to directly execute commands with a confirmation prompt for safety.
- **Context Awareness**: Maestro understands your system's context, including the current directory, system, and user.
- **Support for Multiple LLM Models**: Choose from a variety of models for offline and online usage.
  - Offline: [Ollama](https://ollama.ai) with [over 40 models available](https://ollama.ai/library).
  - Online: GPT4-Turbo and GPT3.5-Turbo.
- **Lightweight**: Maestro is a single 9MB binary with no dependencies.

## Installation

1. Download the latest binary from the [releases page](https://github.com/pluja/maestro/releases).
2. Execute `./maestro -h` to start.

> Tip: Place the binary in a directory within your `$PATH` and rename it to `maestro` for global access, e.g., `sudo mv ./maestro /usr/local/bin/maestro`.

## Offline Usage with [Ollama](https://ollama.ai)

1. Install Ollama from [here](https://ollama.ai/download) (or use [ollama's docker image](https://hub.docker.com/r/ollama/ollama)).
2. Download models using `ollama pull <model-name>`. 
   - **IMPORTANT**: You must pull the default model: `ollama pull codellama:7b-instruct`
3. Start the Ollama server with `ollama serve`.
4. Configure Maestro to use Ollama with `./maestro -set-ollama-url <ollama-url>`, for example, `./maestro -set-ollama-url http://localhost:8080`.

## Online Usage with OpenAI's API

1. Obtain an API token from [OpenAI](https://platform.openai.com/).
2. Set the token using `./maestro -set-openai-token <your-token>`.
3. Choose between GPT4-Turbo with `-4` flag and GPT3.5-Turbo with `-3` flag.
    - Example: `./maestro -4 <prompt>`