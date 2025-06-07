package middleware

import (
	"context"
	"fmt"
	"github.com/digkill/posterAndGrabberBot/internal/botkit"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func AllAccess(next botkit.ViewFunc) botkit.ViewFunc {
	return func(ctx context.Context, bot *tgbotapi.BotAPI, update tgbotapi.Update) error {

		fmt.Println("🙀🙀🙀🙀🙀")
		return next(ctx, bot, update)
	}
}
