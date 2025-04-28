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
	deepSeek  *api.DeepSeekClient 
}
func New(
	token string,
	log *service.Logger, // Принимаем указатель
	userRepo repository.UserRepository,
	msgRepo repository.MessageRepository,
) *Bot {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create bot")
	}

	return &Bot{
		api:      api,
		logger:   log,
		userRepo: userRepo,
		msgRepo:  msgRepo,
	}
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
	if !b.isMessageForBot(msg) {
		return
	}

	// Очищаем текст от упоминания бота
	text := b.cleanMessageText(msg.Text, b.api.Self.UserName)

	if isPokerHandHistory(text) {
        result := b.analyzeHandHistory(text)
        reply := tgbotapi.NewMessage(msg.Chat.ID, result)
        reply.ParseMode = "Markdown"
        b.api.Send(reply)
        return
    }

	if strings.ToLower(strings.TrimSpace(text)) == "привет" {
		reply := tgbotapi.NewMessage(msg.Chat.ID, "Привет, "+msg.From.FirstName)
		reply.ReplyToMessageID = msg.MessageID
		b.api.Send(reply)
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
        `(?i)Hand #\d+`,
        `Blinds: [\d,]+/[\d,]+`,
        `\*\*\* HOLE CARDS \*\*\*`,
        `Dealt to .+\[.+\]`,
    }
    
    matched := 0
    for _, pattern := range patterns {
        re := regexp.MustCompile(pattern)
        if re.MatchString(text) {
            matched++
        }
    }
    
    return matched >= 3 // Минимум 3 совпадения
}

func (b *Bot) analyzeHandHistory(text string) string {
    // Парсим историю раздачи
    history, err := service.ParseTextHandHistory(text)
    if err != nil {
        return "❌ Ошибка парсинга: " + err.Error()
    }

    if err := history.ParseBlinds(text); err != nil {
        return "❌ Ошибка блайндов: " + err.Error()
    }
    
    // if err := history.ParseBlindIncrease(text); err != nil {
    //     return "⚠️ Предупреждение: " + err.Error()
    // }

    // Валидация для Mystery Battle Royale
    validationErrors := service.ValidateMysteryRoyale(history)

    result := formatAnalysisResult(history, validationErrors)
    
    if len(validationErrors) == 0 && strings.Contains(text, "Hero ?") {
        advice, err := b.deepSeek.GetPokerAdvice(text)
        if err == nil {
            result += "\n\n🎓 **Совет DeepSeek:**\n" + advice
        } else {
            b.logger.Error().Err(err).Msg("DeepSeek API error")
            result += "\n\n⚠️ Не удалось получить совет (сервис недоступен)"
        }
    }
    
    return result
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
        builder.WriteString("\n✅ *Валидна для Mystery Battle Royale*")
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