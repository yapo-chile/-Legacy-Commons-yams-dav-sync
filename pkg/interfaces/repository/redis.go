package repository

import (
	"time"
)

// RedisResult interface for a result obtained from executing a get command in redis
type RedisResult interface {
	Scan(dest interface{}) error
}

// RedisHandler interface for a redis handler
type RedisHandler interface {
	Set(key string, values interface{}, expiration time.Duration) error
	Get(key string) (RedisResult, error)
	Del(key string) error
}

// RedisRepo wrapper struct for the RedisHandler
type RedisRepository struct {
	Handler RedisHandler
}

// NewRedisFavoritesRepo constructor for a RedisFavoritesRepo
func NewRedisRepo(redisHandler RedisHandler) *RedisRepository {
	return &RedisRepository{
		Handler: redisHandler,
	}
}

func (repo *RedisRepository) Set(key string, values interface{}, expiration time.Duration) error {
	return repo.Handler.Set(key, values, expiration)
}

func (repo *RedisRepository) Get(key string) (RedisResult, error) {
	return repo.Handler.Get(key)
}

func (repo *RedisRepository) Del(key string) error {
	return repo.Handler.Del(key)
}
