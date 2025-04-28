package service

import (
	"context"
	"os"
	"sync"

	"github.com/rs/zerolog"
)

// Определяем кастомный тип для ключа контекста
type contextKey struct{}

var (
	once      sync.Once
	globalLog *Logger
	logKey    = &contextKey{}
)

type Logger struct {
	zerolog.Logger
}

// New инициализирует логгер с учетом окружения
func NewLogger(env string) *Logger {
	once.Do(func() {
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
		log := zerolog.New(os.Stdout).
			With().
			Timestamp().
			Logger()

		if env == "development" {
			log = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
		}

		globalLog = &Logger{log}
	})

	return globalLog
}


// WithContext добавляет логгер в контекст
func (l *Logger) WithContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, logKey, l)
}

// FromContext безопасно извлекает логгер из контекста
func FromContext(ctx context.Context) *Logger {
	if log, ok := ctx.Value(logKey).(*Logger); ok {
		return log
	}
	// Возвращаем глобальный логгер, если не найден в контексте
	return globalLog
}

// Helpers для уровней логирования
func (l *Logger) Debug() *zerolog.Event {
	return l.Logger.Debug()
}

func (l *Logger) Info() *zerolog.Event {
	return l.Logger.Info()
}

func (l *Logger) Warn() *zerolog.Event {
	return l.Logger.Warn()
}

func (l *Logger) Error() *zerolog.Event {
	return l.Logger.Error()
}

func (l *Logger) Fatal() *zerolog.Event {
	return l.Logger.Fatal()
}

// Пример использования:
/*
func main() {
    logger := logger.New("development")
    ctx := logger.WithContext(context.Background())
    
    log := logger.FromContext(ctx)
    log.Info().Msg("Test message")
}
*/