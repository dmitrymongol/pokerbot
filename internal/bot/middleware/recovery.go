package middleware

import (
	"context"
	"runtime/debug"

	"github.com/dmitrymongol/pokerbot/pkg/logger"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func RecoveryMiddleware(log logger.Logger) func(HandlerFunc) HandlerFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context, bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
			defer func() {
				if r := recover(); r != nil {
					log.Error().
						Interface("recover", r).
						Bytes("stack", debug.Stack()).
						Msg("Panic recovered")
				}
			}()

			next(ctx, bot, update)
		}
	}
}