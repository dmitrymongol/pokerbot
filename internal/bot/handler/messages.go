package handler

// import (
// 	"context"
// 	"strings"

// 	"github.com/dmitrymongol/pokerbot/internal/model"
// 	"github.com/dmitrymongol/pokerbot/internal/service"
// 	"github.com/dmitrymongol/pokerbot/pkg/logger"

// 	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
// )

// type MessageHandler struct {
// 	logger      logger.Logger
// 	userService *service.UserService
// }

// func NewMessageHandler(
// 	logger logger.Logger,
// 	userService *service.UserService,
// ) *MessageHandler {
// 	return &MessageHandler{
// 		logger:      logger,
// 		userService: userService,
// 	}
// }

// func (h *MessageHandler) HandleMessage(ctx context.Context, bot *tgbotapi.BotAPI, update *tgbotapi.Update) {

// 	log := logger.FromContext(ctx)

// 	msg := update.Message
// 	if msg == nil {
// 		log.Warn().Msg("Empty message received")
// 		return
// 	}

// 	// Пример использования контекстного логгера
// 	log.Info().
// 		Int64("user_id", msg.From.ID).
// 		Msg("Processing message")

// 	// Обработка пользователя
// 	user := model.User{
// 		TelegramID:   msg.From.ID,
// 		Username:     msg.From.UserName,
// 		FirstName:    msg.From.FirstName,
// 		LastName:     msg.From.LastName,
// 		LanguageCode: msg.From.LanguageCode,
// 	}

// 	if _, err := h.userService.ProcessUser(ctx, &user); err != nil {
// 		h.logger.Error().Err(err).Msg("Failed to process user")
// 	}

// 	// Обработка сообщения
// 	text := strings.ToLower(strings.TrimSpace(msg.Text))
// 	if text == "привет" {
// 		reply := tgbotapi.NewMessage(msg.Chat.ID, "Привет, "+msg.From.FirstName)
// 		reply.ReplyToMessageID = msg.MessageID
// 		bot.Send(reply)
// 	}
// }