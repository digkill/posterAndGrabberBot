package main

import (
	"context"
	"errors"
	config "github.com/digkill/posterAndGrabberBot/internal"
	"github.com/digkill/posterAndGrabberBot/internal/bot"
	"github.com/digkill/posterAndGrabberBot/internal/bot/middleware"
	"github.com/digkill/posterAndGrabberBot/internal/botkit"
	"github.com/digkill/posterAndGrabberBot/internal/notifier"
	poster "github.com/digkill/posterAndGrabberBot/internal/services/poster"
	"github.com/digkill/posterAndGrabberBot/internal/storage"
	"github.com/digkill/posterAndGrabberBot/internal/summary"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jmoiron/sqlx"
	_ "github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	botAPI, err := tgbotapi.NewBotAPI(config.Get().TelegramBotToken)
	if err != nil {
		log.Printf("[ERROR] failed to create botAPI: %v", err)
		return
	}

	db, err := sqlx.Connect("postgres", config.Get().DatabaseDSN)
	if err != nil {
		log.Printf("[ERROR] failed to connect to db: %v", err)
		return
	}
	defer func(db *sqlx.DB) {
		err := db.Close()
		if err != nil {
			log.Printf("[ERROR] failed to close db: %v", err)
		}
	}(db)

	var (
		articleStorage = storage.NewArticleStorage(db)
		sourceStorage  = storage.NewSourceStorage(db)
		/*fetcher        = fetcher.NewFetcher(
			articleStorage,
			sourceStorage,
			config.Get().FetchInterval,
			config.Get().FilterKeywords,
		)*/

		openAI = summary.NewOpenAI(
			config.Get().OpenAIKey,
			config.Get().OpenAIModel,
			config.Get().OpenAIPrompt,
		)

		newPoster = poster.NewPoster(
			config.Get().ImagesDirectory,
			config.Get().NotificationInterval,
			botAPI,
			config.Get().TelegramChannelID,
			openAI,
		)

		newNotifier = notifier.NewNotifier(
			articleStorage,
			openAI,
			botAPI,
			config.Get().NotificationInterval,
			2*config.Get().FetchInterval,
			config.Get().TelegramChannelID,
		)
	)

	newsBot := botkit.NewBot(botAPI)
	newsBot.RegisterCmdView(
		"addsource",
		middleware.AdminsOnly(
			config.Get().TelegramChannelID,
			bot.ViewCmdAddSource(sourceStorage),
		),
	)
	newsBot.RegisterCmdView(
		"setpriority",
		middleware.AdminsOnly(
			config.Get().TelegramChannelID,
			bot.ViewCmdSetPriority(sourceStorage),
		),
	)
	newsBot.RegisterCmdView(
		"getsource",
		middleware.AdminsOnly(
			config.Get().TelegramChannelID,
			bot.ViewCmdGetSource(sourceStorage),
		),
	)
	newsBot.RegisterCmdView(
		"listsources",
		middleware.AdminsOnly(
			config.Get().TelegramChannelID,
			bot.ViewCmdListSource(sourceStorage),
		),
	)
	newsBot.RegisterCmdView(
		"deletesource",
		middleware.AdminsOnly(
			config.Get().TelegramChannelID,
			bot.ViewCmdDeleteSource(sourceStorage),
		),
	)
	newsBot.RegisterCmdView(
		"start",
		middleware.AllAccess(
			bot.ViewCmdStart(sourceStorage),
		),
	)

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	/*go func(ctx context.Context) {
		if err := fetcher.Start(ctx); err != nil {
			if !errors.Is(err, context.Canceled) {
				log.Printf("[ERROR] failed to run fetcher: %v", err)
				return
			}

			log.Printf("[INFO] fetcher stopped")
		}
	}(ctx)*/

	go func(ctx context.Context) {
		if err := newNotifier.Start(ctx); err != nil {
			if !errors.Is(err, context.Canceled) {
				log.Printf("[ERROR] failed to run newNotifier: %v", err)
				return
			}

			log.Printf("[INFO] newNotifier stopped")
		}
	}(ctx)

	go func(ctx context.Context) {
		if err := newPoster.Start(ctx); err != nil {
			if !errors.Is(err, context.Canceled) {
				log.Printf("[ERROR] failed to run newPoster: %v", err)
				return
			}

			log.Printf("[INFO] newPoster stopped")
		}
	}(ctx)

	go func(ctx context.Context) {
		if err := http.ListenAndServe("0.0.0.0:8881", mux); err != nil {
			if !errors.Is(err, context.Canceled) {
				log.Printf("[ERROR] failed to run http server: %v", err)
				return
			}

			log.Printf("[INFO] http server stopped")
		}
	}(ctx)

	if err := newsBot.Run(ctx); err != nil {
		log.Printf("[ERROR] failed to run botkit: %v", err)
	}
}
