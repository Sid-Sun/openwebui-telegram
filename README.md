# OpenWebUI-Telegram

🤖 A Telegram bot that integrates with OpenWeb UI's OpenAI compatible APIs to provide chat functionality.

## Configuration

The configuration is loaded from a YAML file (an example is provided in `example.yaml` and should be self-explanatory). By default, it is read from `config.yaml` under one of:

- current directory
- `data` directory
- `config` directory
- `data/config` directory

### Storage Engine (Remote Configuration & Hot-Reload)

You can optionally back the configuration with a remote [Storage Engine](https://github.com/fitant/storage-engine-go) object, allowing you to hot-reload configuration changes at runtime without restarting the bot.

Note that a valid local `config.yaml` is still loaded first on startup before the Storage Engine config is fetched. To load configuration from Storage Engine on startup, set the following environment variables:

- `SE_CONFIG_ENABLED`: Set to `TRUE` (case-insensitive) to load the configuration from Storage Engine at startup.
- `SE_URL`: The Storage Engine server URL.
- `SE_OBJ_ID`: The Storage Engine object ID containing your YAML configuration.
- `SE_OBJ_PASS`: The password for the Storage Engine object.

> **Note:** Hot-reloading updates backends and models in-place, but the Telegram bot `api_token` itself is not hot-reloaded — the bot will continue running with the token it started with.

## Usage

Only private chat is tested as of now.

### Chatting

Just keep sending messages — the bot automatically continues the conversation from the most recent message, so you don't need to reply to keep a thread going.

To branch off from an earlier point instead, reply to the specific message you want to pick the conversation up from. Explicit replies always take precedence over the automatic continuation, which is also what `/resend` relies on.

To start a completely fresh thread, send `/new`; your next message begins a new conversation. Previous threads are kept and can still be resumed by replying to one of their messages.

### Commands

The bot supports the following commands:

1. To set system prompt, use `/reset <system prompt>`.
   - The default system prompt is `You are a friendly assistant`.
   - Once you set a custom system prompt, it will remain set until you either change it or bot is restarted.
   - This takes effect immediately after you set it, even in earlier conversations.
2. To regenerate a response, reply to the message you want to regenerate from and send `/resend`.
   - Once the bot is restarted, the conversation history is lost and thread can't be continued.
3. To change the model being used, use `/models`.
   - It will present you with the available model, friendly names and basic config.
   - Select the model using the inline keyboard.
   - This takes effect immediately after you set it, even in earlier conversations.
4. To hot-reload the configuration from Storage Engine, use `/reload`.
   - Fetches the latest YAML configuration from the configured Storage Engine object (`SE_URL`, `SE_OBJ_ID`, `SE_OBJ_PASS`) and updates backends and models in-place.
   - The bot `api_token` does not hot-reload.

## How to run

### Docker Compose

Copy the docker-compose.yml in this repo, create your config file in `data/config/config.yaml` and run:

```bash
docker compose up
```

### Shell

In the shell, you can just do:

```bash
make serve
```

## Contributing

Contributions are welcome. Please submit a pull request or create an issue if you have any improvements or suggestions.
