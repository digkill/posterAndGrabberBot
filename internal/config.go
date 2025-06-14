package config

import (
	"fmt"
	"github.com/cristalhq/aconfig"
	"github.com/cristalhq/aconfig/aconfighcl"
	"log"
	"os"
	"sync"
	"time"
)

type Config struct {
	TelegramBotToken     string        `hcl:"telegram_bot_token" env:"TELEGRAM_BOT_TOKEN" required:"true"`
	TelegramChannelID    int64         `hcl:"telegram_channel_id" env:"TELEGRAM_CHANNEL_ID" required:"true"`
	DatabaseDSN          string        `hcl:"database_dsn" env:"DATABASE_DSN" default:"postgres://postgres:postgres@localhost:5432/news_feed_bot?sslmode=disable"`
	FetchInterval        time.Duration `hcl:"fetch_interval" env:"FETCH_INTERVAL" default:"10m"`
	NotificationInterval time.Duration `hcl:"notification_interval" env:"NOTIFICATION_INTERVAL" default:"1m"`
	FilterKeywords       []string      `hcl:"filter_keywords" env:"FILTER_KEYWORDS"`
	OpenAIKey            string        `hcl:"openai_key" env:"OPENAI_KEY"`
	OpenAIPrompt         string        `hcl:"openai_prompt" env:"OPENAI_PROMPT"`
	OpenAIModel          string        `hcl:"openai_model" env:"OPENAI_MODEL" default:"gpt-4o"`
	ImagesDirectory      string        `hcl:"images_directory" env:"IMAGES_DIRECTORY" default:"./internal/storage/images"`
	VKToken              string        `hcl:"vk_token" env:"VK_TOKEN" default:""`
}

var (
	cfg  Config
	once sync.Once
)

func Get() Config {
	once.Do(func() {
		loader := aconfig.LoaderFor(&cfg, aconfig.Config{
			//	EnvPrefix: "NFB",
			Files: []string{"./config.local.hcl", "./config.hcl", "$HOME/.config/posterAndGrabberBot/config.hcl"},
			FileDecoders: map[string]aconfig.FileDecoder{
				".hcl": aconfighcl.New(),
			},
		})

		if err := loader.Load(); err != nil {
			log.Printf("[ERROR] failed to load config: %v", err)
		}
	})

	// using the function
	mydir, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(mydir)

	return cfg
}
