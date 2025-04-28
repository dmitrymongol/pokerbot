package config

import (
	"github.com/caarlos0/env/v9"
	"github.com/joho/godotenv"
)

type DeepSeekConfig struct {
    APIKey  string `env:"DEEPSEEK_API_KEY,required"`
    BaseURL string `env:"DEEPSEEK_API_URL" envDefault:"https://api.deepseek.com/v1"`
}
type Config struct {
	Env  string `env:"ENV" envDefault:"development"`
	DB   DBConfig
	Telegram TelegramConfig
}

type DBConfig struct {
	DSN           string `env:"DB_DSN" envDefault:""`
	MigrationsPath string `env:"DB_MIGRATIONS_PATH" envDefault:"./migrations"`
}

type TelegramConfig struct {
	Token string `env:"TELEGRAM_TOKEN,required"`
}

func Load() (*Config, error) {
	// Загружаем .env файл
	if err := godotenv.Load(); err != nil {
		return nil, err
	}

	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}