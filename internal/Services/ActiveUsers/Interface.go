package ActiveUsers

import (
	"main.go/internal/domain"
)

type Cache interface {
	Create(TelegramID int64, user domain.User) error
	Get(TelegramID int64) (domain.User, error)
	Delete(TelegramID int64) error
	Update(TelegramID int64, user domain.User) error
}
