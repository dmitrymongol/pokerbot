package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {

	botToken := os.Getenv("BOT_TOKEN")
	if botToken == "" {
		log.Panic("BOT_TOKEN environment variable is required")
	}
	
	bot, _ := tgbotapi.NewBotAPI(botToken)
	bot.Debug = true
	log.Printf("Bot activated: @%s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		log.Printf("[DEBUG] Raw message text: %s", update.Message.Text)

		if update.Message.Chat.IsGroup() || update.Message.Chat.IsSuperGroup() {
			log.Printf("Group message in '%s'", update.Message.Chat.Title)
			
			if isMessageForBot(bot, update.Message) {
				log.Println("Message is addressed to the bot")
				processMessage(bot, update.Message)
			}
		} else {
			processMessage(bot, update.Message)
		}
	}
}

func processMessage(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	// Очищаем текст от упоминания бота
	cleanText := removeBotMention(bot, msg.Text)
	cleanText = strings.ToLower(strings.TrimSpace(cleanText))
	
	log.Printf("[DEBUG] Cleaned text: '%s'", cleanText)

	if cleanText == "привет" {
		username := getUsername(msg.From)
		reply := fmt.Sprintf("Привет %s 👋", username)
		
		response := tgbotapi.NewMessage(msg.Chat.ID, reply)
		response.ReplyToMessageID = msg.MessageID
		
		if _, err := bot.Send(response); err != nil {
			log.Printf("[ERROR] Send failed: %v", err)
		} else {
			log.Println("[SUCCESS] Response sent")
		}
	}
}

func removeBotMention(bot *tgbotapi.BotAPI, text string) string {
	// Удаляем как @username, так и @username_bot
	mentionVariants := []string{
		"@" + bot.Self.UserName,
		"@" + strings.TrimSuffix(bot.Self.UserName, "_bot"),
	}

	for _, variant := range mentionVariants {
		text = strings.ReplaceAll(text, variant, "")
		text = strings.ReplaceAll(text, variant+" ", "") // С пробелом
	}

	return strings.TrimSpace(text)
}

func isMessageForBot(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) bool {
	// 1. Проверка прямого упоминания
	if strings.Contains(msg.Text, "@"+bot.Self.UserName) {
		return true
	}

	// 2. Ответ на сообщение бота
	if msg.ReplyToMessage != nil && msg.ReplyToMessage.From.ID == bot.Self.ID {
		return true
	}

	// 3. Команды (/start и т.д.)
	if msg.IsCommand() {
		return true
	}

	// 4. Личное сообщение
	if !msg.Chat.IsGroup() && !msg.Chat.IsSuperGroup() {
		return true
	}

	return false
}

func getUsername(user *tgbotapi.User) string {
	if user.UserName != "" {
		return "@" + user.UserName
	}
	return user.FirstName
}