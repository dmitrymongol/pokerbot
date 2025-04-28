package repository

import (
	"context"
	"database/sql"
	"pokerbot/internal/model"
)

// Реализация заглушек

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgres(dsn string) (*sql.DB, error) {
	// Заглушка для подключения
	return &sql.DB{}, nil
}

func Migrate(db *sql.DB, path string) error {
	// Заглушка миграций
	return nil
}

// UserRepository implementation
type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) CreateOrUpdate(ctx context.Context, user *model.User) error {
	// Заглушка
	return nil
}

func (r *userRepository) GetByID(ctx context.Context, id int64) (*model.User, error) {
	// Заглушка
	return &model.User{}, nil
}

// MessageRepository implementation
type messageRepository struct {
	db *sql.DB
}

func NewMessageRepository(db *sql.DB) MessageRepository {
	return &messageRepository{db: db}
}

func (r *messageRepository) Save(ctx context.Context, msg *model.Message) error {
	// Заглушка
	return nil
}

func (r *messageRepository) GetLastMessages(ctx context.Context, userID int64, limit int) ([]*model.Message, error) {
	// Заглушка
	return []*model.Message{}, nil
}