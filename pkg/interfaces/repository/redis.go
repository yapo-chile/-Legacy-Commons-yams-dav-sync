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
	Get(key string) (interface{}, error)
	Del(key string) error
	Exists(key string) bool
}

// redisRepository wrapper struct for the RedisHandler
type redisRepository struct {
	handler RedisHandler
}

// NewRedisRepo constructor for a redis repository
func NewRedisRepo(redisHandler RedisHandler) RedisHandler {
	return &redisRepository{
		handler: redisHandler,
	}
}

func (repo *redisRepository) Set(key string, values interface{}, expiration time.Duration) error {
	return repo.handler.Set(key, values, expiration)
}

func (repo *redisRepository) Get(key string) (interface{}, error) {
	return repo.handler.Get(key)
}

func (repo *redisRepository) Del(key string) error {
	return repo.handler.Del(key)
}

func (repo *redisRepository) Exists(key string) bool {
	return repo.handler.Exists(key)
}
