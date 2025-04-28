package repository

import (
	"context"

	"github.com/dmitrymongol/pokerbot/internal/model"
)

// Объявляем интерфейсы в отдельном файле
type UserRepository interface {
	CreateOrUpdate(ctx context.Context, user *model.User) error
	GetByID(ctx context.Context, id int64) (*model.User, error)
}

type MessageRepository interface {
	Save(ctx context.Context, msg *model.Message) error
	GetLastMessages(ctx context.Context, userID int64, limit int) ([]*model.Message, error)
}