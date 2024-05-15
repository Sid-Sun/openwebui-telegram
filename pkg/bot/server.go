package bot

import (
	"encoding/json"
	"log/slog"
	"os"
	"os/signal"

	"github.com/sid-sun/openwebui-bot/cmd/config"
	"github.com/sid-sun/openwebui-bot/pkg/bot/router"
	"github.com/sid-sun/openwebui-bot/pkg/bot/store"
)

// StartBot starts the bot, inits all the requited submodules and routine for shutdown
func StartBot(cfg config.Config) {
	store.NewStore()
	loadStore()
	ch := router.New(cfg.Bot)

	slog.Info("[StartBot] Starting Bot")
	go dumpStore(ch)
	ch.Start()
}

// Dump store data to disk as JSON on interrupt
func dumpStore(ch *router.Bot) {
	shutDown := make(chan os.Signal, 1)
	signal.Notify(shutDown, os.Interrupt)
	<-shutDown
	slog.Info("[DumpStore] Dumping store data to disk")
	// Implement store dumping logic here
	x, err := json.MarshalIndent(store.ChatStore, "", "  ")
	if err != nil {
		slog.Error("[DumpStore] Error dumping store data to disk", slog.Any("error", err))
		return
	}
	err = os.WriteFile("./store/chat_store.json", x, 0644)
	if err != nil {
		slog.Error("[DumpStore] Error writing store data to file", slog.Any("error", err))
		return
	}
	slog.Info("[LoadStore] Dumped store data to disk")
	ch.Stop()
}

// load data from JSON file on startup
func loadStore() {
	slog.Info("[LoadStore] Loading store data from disk")
	data, err := os.ReadFile("./store/chat_store.json")
	if err != nil {
		slog.Error("[LoadStore] Error reading store data from file", slog.Any("error", err))
		return
	}
	err = json.Unmarshal(data, &store.ChatStore)
	if err != nil {
		slog.Error("[LoadStore] Error unmarshaling store data from file", slog.Any("error", err))
		return
	}
	slog.Info("[LoadStore] Loaded store data from disk")
}
