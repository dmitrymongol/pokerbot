package bot

import (
	"context"

	"pokerbot/internal/bot/middleware"
	"pokerbot/internal/service"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Router struct {
	middlewares []func(middleware.HandlerFunc) middleware.HandlerFunc
	handler     middleware.HandlerFunc
}

type HandlerFunc = middleware.HandlerFunc

func NewRouter(log service.Logger) *Router {
	r := &Router{}
	
	// Инициализируем с дефолтными middleware
	r.Use(
		middleware.RecoveryMiddleware(log),
		middleware.LoggingMiddleware(log),
	)
	
	return r
}

func (r *Router) Use(middlewares ...func(HandlerFunc) HandlerFunc) {
	r.middlewares = append(r.middlewares, middlewares...)
}

func (r *Router) Handle(handler HandlerFunc) {
	// Собираем цепочку middleware
	finalHandler := handler
	for i := len(r.middlewares) - 1; i >= 0; i-- {
		finalHandler = r.middlewares[i](finalHandler)
	}
	r.handler = finalHandler
}

func (r *Router) HandleUpdate(ctx context.Context, bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	if r.handler != nil {
		r.handler(ctx, bot, update)
	}
}