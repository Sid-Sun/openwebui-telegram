# OpenWebUI-Telegram

🤖 A Telegram bot that integrates with OpenWeb UI's OpenAI compatible APIs to provide chat functionality.

## Configuration

The configuration is loaded a yaml file, en example is provided in `example.yaml` and should be self-explanatory, it has to be stored in `config.yaml` under either:

- current directory
- `data` directory
- `config` directory
- `data/config` directory

## Usage

Only private chat is tested as of now.

### Chatting

If you send messages without a reply, bot treats that as starting a new chat / thread. To continue a chat / thread, reply to the last (or previous) message you want to pick up the conversation from.

### Commands

The bot supports three commands:

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
