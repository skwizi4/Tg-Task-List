package ActiveUsers

import (
	"github.com/skwizi4/lib/Redis"
	"strconv"
)

func userKey(userID int64) string {
	return "user:" + strconv.FormatInt(userID, 10)
}
func InitRedisCache(password string, DB int, addr string) *RedisCache {
	return &RedisCache{
		Redis: Redis.New(password, DB, addr),
	}
}
