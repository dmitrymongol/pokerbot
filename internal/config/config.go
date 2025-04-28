package config

import (
	"log"

	"github.com/kelseyhightower/envconfig"
)

// type DeepSeekConfig struct {
//     APIKey  string `envconfig:"DEEPSEEK_API_KEY,required"`
//     BaseURL string `envconfig:"DEEPSEEK_API_URL" envDefault:"https://api.deepseek.com/v1"`
// }
type Config struct {
	Env  string `envconfig:"ENV" default:"development"`
	DB   DBConfig
	Telegram TelegramConfig
	Yandex YandexConfig
	AllowedChats string `envconfig:"ALLOWED_CHATS" required:"true"`
	AdminUsers string `envconfig:"ADMIN_USERS" required:"true"`
}

type DBConfig struct {
	DSN           string `envconfig:"DB_DSN" default:""`
	MigrationsPath string `envconfig:"DB_MIGRATIONS_PATH" default:"./migrations"`
}

type TelegramConfig struct {
	Token string `envconfig:"TELEGRAM_TOKEN" required:"true"`
	
}
type YandexConfig struct {
	YandexOauthToken string `envconfig:"YANDEX_OAUTH_TOKEN" required:"true"`
	YandexFolderID string `envconfig:"YANDEX_FOLDER_ID" required:"true"`
	
	
}

func Load() *Config {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		log.Fatal("Failed to load configuration: ", err)
	}
	return &cfg
}