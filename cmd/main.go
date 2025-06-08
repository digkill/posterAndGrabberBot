package main

import (
	"context"
	"errors"
	config "github.com/digkill/posterAndGrabberBot/internal"

	"github.com/digkill/posterAndGrabberBot/internal/fetcher"
	"github.com/digkill/posterAndGrabberBot/internal/helpers"

	poster "github.com/digkill/posterAndGrabberBot/internal/services/poster"
	"github.com/digkill/posterAndGrabberBot/internal/summary"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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

	var (
		newFetcher = fetcher.NewFetcher(
			config.Get().VKToken,
			config.Get().FetchInterval,
			config.Get().FilterKeywords,
		)

		newOpenAI = summary.NewOpenAI(
			config.Get().OpenAIKey,
			config.Get().OpenAIModel,
			config.Get().OpenAIPrompt,
		)

		newPoster = poster.NewPoster(
			config.Get().ImagesDirectory,
			config.Get().NotificationInterval,
			botAPI,
			config.Get().TelegramChannelID,
			newOpenAI,
		)
	)

	newsBot := helpers.NewBot(botAPI)

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	go func(ctx context.Context) {
		log.Printf("[INFO] newFetcher started")
		if err := newFetcher.Start(ctx); err != nil {
			if !errors.Is(err, context.Canceled) {
				log.Printf("[ERROR] failed to run newFetcher: %v", err)
				return
			}

			log.Printf("[INFO] newFetcher stopped")
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
		log.Printf("[ERROR] failed to run bot: %v", err)
	}
}
