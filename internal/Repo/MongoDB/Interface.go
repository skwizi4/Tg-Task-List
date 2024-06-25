package MongoDB

import "main.go/internal/domain"

type Mongo interface {
	Create(User domain.User) error
	Delete(id int64) error
	Update(id int64, User domain.User) error
	Get(id int64) (*domain.User, error)
}
