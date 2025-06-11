package main

import (
	"context"
	"errors"
	"fmt"
	config "github.com/digkill/posterAndGrabberBot/internal"
	"github.com/digkill/posterAndGrabberBot/internal/fetcher"
	"github.com/digkill/posterAndGrabberBot/internal/helpers"
	"github.com/digkill/posterAndGrabberBot/internal/services/nutsdb"
	"time"

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

	newNutsDB := nutsdb.NewNutsDB()
	defer newNutsDB.Close()
	if err := newNutsDB.InitBuckets(); err != nil {
		log.Fatalf("InitBuckets error: %v", err)
	}
	if err := newNutsDB.TestCreateAndPush(); err != nil {
		log.Fatalf("TestCreateAndPush error: %v", err)
	}

	newFetcher := fetcher.NewFetcher(
		config.Get().VKToken,
		config.Get().FetchInterval,
		config.Get().FilterKeywords,
		newNutsDB,
	)

	newOpenAI := summary.NewOpenAI(
		config.Get().OpenAIKey,
		config.Get().OpenAIModel,
		config.Get().OpenAIPrompt,
	)

	newPoster := poster.NewPoster(
		config.Get().ImagesDirectory,
		config.Get().NotificationInterval,
		botAPI,
		config.Get().TelegramChannelID,
		newOpenAI,
	)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	go func(ctx context.Context) {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				links, err := newNutsDB.GetAllPendingVideoLinks()
				if err != nil {
					log.Println("NutsDB error:", err)
					continue
				}
				for _, url := range links {
					if err := helpers.DownloadVKVideo(url); err != nil {
						log.Println("yt-dlp error:", err)
					}
					err = newNutsDB.MarkVideoURLProcessed(url)
					if err != nil {
						log.Println("MarkVideoURLProcessed:", err)
					}

					err = newNutsDB.RemoveVideoLink(url)
					if err != nil {
						fmt.Printf("[ERROR] failed to remove link %s: %v", url, err)
					}

				}
			}
		}
	}(ctx)

	newsBot := helpers.NewBot(botAPI)

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

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
