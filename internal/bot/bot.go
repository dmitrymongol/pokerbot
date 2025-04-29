package bot

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"pokerbot/internal/api"
	"pokerbot/internal/repository"
	"pokerbot/internal/service"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
    api      *tgbotapi.BotAPI
    logger   *service.Logger // Используем указатель
    userRepo repository.UserRepository
    msgRepo  repository.MessageRepository
	// deepSeek  *api.DeepSeekClient
	yandexGPT  *api.YandexGPTClient
	allowedChats map[int64]struct{}
    adminUserIDs map[int64]struct{}
}
func New(
	token string,
	log *service.Logger, // Принимаем указатель
	userRepo repository.UserRepository,
	msgRepo repository.MessageRepository,
	yandexToken string,
    yandexFolderID string,
	allowedChatsStr string,
	adminUserIDsStr string, // Строка с ID админов через запятую
) *Bot {
	tgbotapi, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create bot")
	}
	return &Bot{
		api:      tgbotapi,
		logger:   log,
		userRepo: userRepo,
		msgRepo:  msgRepo,
		yandexGPT: api.NewYandexGPTClient(yandexToken, yandexFolderID, log),
        allowedChats: parseIDs(allowedChatsStr, log, "allowed chat"),
        adminUserIDs: parseIDs(adminUserIDsStr, log, "admin user"),
	}
}

// Общая функция парсинга ID
func parseIDs(input string, log *service.Logger, idType string) map[int64]struct{} {
    ids := make(map[int64]struct{})
    if input == "" {
        return ids
    }

    for _, s := range strings.Split(input, ",") {
        s = strings.TrimSpace(s)
        id, err := strconv.ParseInt(s, 10, 64)
        if err != nil {
            log.Error().Str("type", idType).Str("value", s).Msg("Invalid ID")
            continue
        }
        ids[id] = struct{}{}
    }

    return ids
}

// Проверка прав доступа
func (b *Bot) isAdmin(userID int64) bool {
    _, ok := b.adminUserIDs[userID]
    return ok
}


func (b *Bot) Start() error {
	b.logger.Info().Msg("Starting bot...")

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			b.handleMessage(context.Background(), update.Message)
		}
	}

	return nil
}

func (b *Bot) handleMessage(ctx context.Context, msg *tgbotapi.Message) {
	// Проверяем адресовано ли сообщение боту
    if msg.Chat.IsPrivate() {
        if !b.isAdmin(msg.From.ID) {
            b.logger.Warn().
                Int64("user_id", msg.From.ID).
                Str("username", msg.From.UserName).
                Msg("Unauthorized private message attempt")
			reply := tgbotapi.NewMessage(msg.From.ID, msg.From.UserName + "Unauthorized toi send private messages")	
			b.api.Send(reply)		
            return
        }
		b.processMessage(msg)
        return
    }

    // Для групповых чатов проверяем вайтлист
    if _, ok := b.allowedChats[msg.Chat.ID]; !ok {
        b.logger.Debug().
            Int64("chat_id", msg.Chat.ID).
            Str("chat_title", msg.Chat.Title).
            Msg("Message from non-whitelisted group")
		reply := tgbotapi.NewMessage(msg.Chat.ID, "Chat non-whitelisted")	
		b.api.Send(reply)	
        return
    }

	b.processMessage(msg)

}

func (b *Bot)  processMessage(msg *tgbotapi.Message) {
	text := b.cleanMessageText(msg.Text, b.api.Self.UserName)

	if isPokerHandHistory(text) {
        result := b.analyzeHandHistory(text)
        reply := tgbotapi.NewMessage(msg.Chat.ID, result)
        reply.ParseMode = "Markdown"
        b.api.Send(reply)
        return
    }
}

// Проверяем, адресовано ли сообщение боту
func (b *Bot) isMessageForBot(msg *tgbotapi.Message) bool {
	// Личные сообщения всегда обрабатываем
	if !msg.Chat.IsGroup() && !msg.Chat.IsSuperGroup() {
		return true
	}

	// Проверяем упоминание бота
	if strings.Contains(msg.Text, "@"+b.api.Self.UserName) {
		return true
	}

	// Проверяем ответ на сообщение бота
	if msg.ReplyToMessage != nil && msg.ReplyToMessage.From.ID == b.api.Self.ID {
		return true
	}

	return false
}

// Очищаем текст от упоминаний бота
func (b *Bot) cleanMessageText(text, botUsername string) string {
	variants := []string{
		"@" + botUsername,
		"@ " + botUsername, // Для случаев с опечаткой
		"/start",
		"/help",
	}

	for _, variant := range variants {
		text = strings.ReplaceAll(text, variant, "")
		text = strings.ReplaceAll(text, variant+" ", "")
	}

	return strings.TrimSpace(text)
}

func isPokerHandHistory(text string) bool {
    patterns := []string{
        `(?i)Hand #[\dA-Z]+`,       // Номер руки с буквами и цифрами
        `\d+/[\d,]+\)`,             // Блайнды в уровне (например: Level12(400/800)
        `\*\*\* HOLE CARDS \*\*\*`, 
        `Dealt to .+\[.+\]`,        // Раздача карт игроку
    }
    
    matched := 0
    for _, pattern := range patterns {
        re := regexp.MustCompile(pattern)
        if re.MatchString(text) {
            matched++
        }
    }
    
    return matched >= 3
}

func (b *Bot) analyzeHandHistory(text string) string {
    history, err := service.ParseTextHandHistory(text)
    if err != nil {
        return "❌ Ошибка парсинга: " + err.Error()
    }

    // validationErrors := service.ValidateMysteryRoyale(history)
	validationErrors := []error{}
    result := formatAnalysisResult(history, validationErrors)

    if len(validationErrors) == 0 {
        advice, err := b.getGTOAdvice(text)
        if err != nil {
            b.logger.Error().Err(err).Msg("GPT API error")
            result += "\n\n⚠️ Не удалось получить совет"
        } else {
            result += formatGPTAdvice(advice)
        }
    }
    
    return result
}

func (b *Bot) getGTOAdvice(handHistory string) (string, error) {
    return b.yandexGPT.GetPokerAdvice(handHistory)
}

func formatGPTAdvice(advice string) string {
    // Упрощаем форматирование
    return fmt.Sprintf("\n\n🎓 **GTO Совет:**\n%s", 
        strings.ReplaceAll(advice, "\n", "\n• "))
}

// Обновляем функцию formatAnalysisResult в файле bot/bot.go
func formatAnalysisResult(hh *service.HandHistory, errors []error) string {
    builder := strings.Builder{}
    
    // Заголовок и базовая информация
    builder.WriteString(fmt.Sprintf("🃏 *Анализ раздачи #%s*\n", hh.HandID))
    if hh.TournamentID != "" {
        builder.WriteString(fmt.Sprintf("Турнир: `%s`\n", hh.TournamentID))
    }
    builder.WriteString(fmt.Sprintf("Блайнды: %s/%s\n", 
        formatNumber(hh.SmallBlind),
        formatNumber(hh.BigBlind)))
    builder.WriteString(fmt.Sprintf("Анте: %s\n", formatNumber(hh.Ante)))

    // Вывод Mystery-элементов
    if len(hh.MysteryElements) > 0 {
        builder.WriteString(fmt.Sprintf("Mystery-элементы: [%s]\n", 
            strings.Join(hh.MysteryElements, ", ")))
    }

    // Валидация
    if len(errors) == 0 {
        builder.WriteString("\n✅ *Раздача валидна*")
    } else {
        builder.WriteString("\n❌ *Нарушения:*\n")
        for _, err := range errors {
            builder.WriteString(fmt.Sprintf("• %s\n", err.Error()))
        }
    }
    
    builder.WriteString(fmt.Sprintf("\n_Проверено: %s_", 
        time.Now().Format("2006-01-02 15:04")))
    
    return builder.String()
}

func formatNumber(n int) string {
    if n == 0 {
        return "N/A"
    }
    
    s := strconv.Itoa(n)
    var result []rune
    counter := 0

    for i := len(s) - 1; i >= 0; i-- {
        counter++
        result = append([]rune{rune(s[i])}, result...)
        if counter%3 == 0 && i != 0 {
            result = append([]rune{','}, result...)
        }
    }
    
    return string(result)
}