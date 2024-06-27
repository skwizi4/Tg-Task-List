package ActiveUsers

import (
	"encoding/json"
	"main.go/internal/domain"
)

//CRUD methods

func (c *RedisCache) Create(TelegramID int64, user domain.User) error {
	userData, err := json.Marshal(user)
	if err != nil {
		return err
	}
	return c.Redis.Set(userKey(TelegramID), userData)
}

func (c *RedisCache) Get(TelegramID int64) (domain.User, error) {
	var user domain.User
	userData, err := c.Redis.GetBytes(userKey(TelegramID))
	if err != nil {
		return user, err
	}
	err = json.Unmarshal(userData, &user)
	return user, err
}

func (c *RedisCache) Delete(TelegramID int64) error {
	return c.Redis.Delete(userKey(TelegramID))
}

func (c *RedisCache) Update(TelegramID int64, user domain.User) error {
	if err := c.Redis.Delete(userKey(TelegramID)); err != nil {
		return err
	}
	userData, err := json.Marshal(user)
	if err != nil {
		return err
	}
	return c.Redis.Set(userKey(TelegramID), userData)
}
