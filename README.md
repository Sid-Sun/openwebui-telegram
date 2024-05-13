# OpenWebUI-Telegram

🤖 A Telegram bot that integrates with OpenWeb UI's OpenAI compatible APIs to provide chat functionality.

## Configuration

The configuration is loaded from environment variables, here are the environment variables you can set:

### OpenAI Endpoint Options

- `OPENAI_ENDPOINT`: The base OpenAI API compatible endpoint. example: `http://localhost:3000/ollama/v1/`. http & https both work.
- `OPENAI_API_KEY`: The `Bearer` token API Key, example: `sk-12345667890abcdefghijkl`

### Model Options

- `MODEL`: The model to use. Default is `llama3:instruct`.
- `MODEL_TWEAK_LEVEL`: The level of tweaking to apply to the model. Set to `advanced` to make Penalty tweaks take effect, else - none are provided to API. Default is `minimal`, i.e. no Penalty parameters.

### Model Tweaks

- `MAX_TOKENS`: The maximum number of tokens that the model can generate. Default is `1024`.
- `TEMPERATURE`: Controls the randomness of the model's output. Higher values make the output more random. Default is `0.8`.
- `REPEAT_PENALTY`: Penalty for repeating the same token. Default is `1.2`.
- `CONTEXT_LENGTH`: The maximum number of tokens in the context. Default is `8192`.
- `PRESENCE_PENALTY`: Penalty for using tokens that are not in the context. Default is `1.5`.
- `FREQUENCY_PENALTY`: Penalty for using tokens that are used frequently. Default is `1.0`.

### Bot Configuration

- `API_TOKEN`: The Telegram Bot API token.

## Usage

Only private chat is tested as of now.

### Chatting

If you send messages without a reply, bot treats that as starting a new chat / thread. To continue a chat / thread, reply to the last (or previous) message you want to pick up the conversation from.

### Commands

The bot supports two commands:

1. To set system prompt, use `/reset <system prompt>`.
   - The default system prompt is `You are a friendly assistant`.
   - Once you set a custom system prompt, it will remain set until you either change it or bot is restarted.
2. To regenerate a response, reply to the message you want to regenerate from and send `/resend`
   - Once the bot is restarted, the conversation history is lost and thread can't be continued.

## How to run

### Docker Compose

Copy the docker-compose.yml in this repo, create your env file in `dev.env` and run:

```bash
docker compose up
```

### Shell

In the fish shell, you can just do:

```fish
env (cat dev.env | xargs -L 1) make serve
```

bash:

```bash
env $(cat dev.env | xargs -L 1) make serve
```


## Contributing

Contributions are welcome. Please submit a pull request or create an issue if you have any improvements or suggestions.
