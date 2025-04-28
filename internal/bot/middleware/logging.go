package middleware

import (
	"context"
	"time"

	"github.com/dmitrymongol/pokerbot/pkg/logger"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type HandlerFunc func(ctx context.Context, bot *tgbotapi.BotAPI, update *tgbotapi.Update)

func LoggingMiddleware(log logger.Logger) func(HandlerFunc) HandlerFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context, bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
			start := time.Now()
			
			// Проверяем наличие сообщения
			if update.Message == nil {
				log.Warn().Msg("Received update without message")
				return
			}

			// Создаем child logger с контекстом
			requestLog := log.With().
				Int64("message_id", int64(update.Message.MessageID)). // Явное преобразование в int64
				Str("username", update.Message.From.UserName).
				Logger()

			// Добавляем логирование в контекст
			ctx = requestLog.WithContext(ctx)

			defer func() {
				duration := time.Since(start)
				requestLog.Info().
					Str("text", update.Message.Text).
					Str("chat_type", update.Message.Chat.Type).
					Dur("duration", duration).
					Msg("Message processed")
			}()

			// Логирование входящего сообщения
			requestLog.Debug().
				Str("text", update.Message.Text).
				Str("chat_type", update.Message.Chat.Type).
				Msg("Received message")

			next(ctx, bot, update)
		}
	}
}