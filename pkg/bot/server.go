package bot

import (
	"log/slog"

	"github.com/sid-sun/openwebui-bot/cmd/config"
	"github.com/sid-sun/openwebui-bot/pkg/bot/router"
	"github.com/sid-sun/openwebui-bot/pkg/bot/store"
)

// StartBot starts the bot, inits all the requited submodules and routine for shutdown
func StartBot(cfg config.Config) {
	store.NewStore()
	ch := router.New(cfg.Bot).NewUpdateChan()

	slog.Info("[StartBot] Starting Bot")
	ch.ListenAndServe()
}
