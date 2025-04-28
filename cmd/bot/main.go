package main

import (
	"log"

	"pokerbot/internal/bot"
	"pokerbot/internal/config"
	"pokerbot/internal/repository"
	"pokerbot/pkg/logger"
)

func main() {
	// Инициализация конфига
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Config error: %v", err)
	}

	// Инициализация логгера
	logger := logger.New(cfg.Env)

	// Подключение к БД
	db, err := repository.NewPostgres(cfg.DB.DSN)
	if err != nil {
		logger.Fatal().Err(err).Msg("DB connection failed")
	}
	defer db.Close()

	// Миграции
	if err := repository.Migrate(db, cfg.DB.MigrationsPath); err != nil {
		logger.Fatal().Err(err).Msg("Migration failed")
	}

	// Создание репозиториев
	userRepo := repository.NewUserRepository(db)
	messageRepo := repository.NewMessageRepository(db)

	// Запуск бота
	bot := bot.New(
		cfg.Telegram.Token,
		logger,
		userRepo,
		messageRepo,
	)

	if err := bot.Start(); err != nil {
		logger.Fatal().Err(err).Msg("Bot stopped")
	}
}