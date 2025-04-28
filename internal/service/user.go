package service

// import (
// 	"context"
// 	"github.com/dmitrymongol/pokerbot/internal/domain/model"
// 	"github.com/dmitrymongol/pokerbot/internal/repository"
// )

// type UserService struct {
// 	repo repository.UserRepository
// }

// func NewUserService(repo repository.UserRepository) *UserService {
// 	return &UserService{repo: repo}
// }

// func (s *UserService) ProcessUser(ctx context.Context, tgUser *model.User) (*model.User, error) {
// 	existingUser, err := s.repo.GetByTelegramID(ctx, tgUser.TelegramID)
// 	if err != nil {
// 		return nil, err
// 	}

// 	if existingUser == nil {
// 		return s.repo.CreateOrUpdate(ctx, tgUser)
// 	}

// 	// Обновление данных пользователя
// 	existingUser.Username = tgUser.Username
// 	existingUser.FirstName = tgUser.FirstName
// 	existingUser.LastName = tgUser.LastName

// 	return s.repo.CreateOrUpdate(ctx, existingUser)
// }