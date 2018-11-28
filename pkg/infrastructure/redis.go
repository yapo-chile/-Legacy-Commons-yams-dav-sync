package infrastructure

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis"
	"github.schibsted.io/Yapo/yams-dav-sync/pkg/interfaces/loggers"
	"github.schibsted.io/Yapo/yams-dav-sync/pkg/interfaces/repository"
)

// RedisHandler handler for the request made to Redis
type RedisHandler struct {
	Client *redis.Client
	Logger loggers.Logger
}

// RedisResult result representing a result from a get
type RedisResult struct {
	Result *redis.StringCmd
	Logger loggers.Logger
}

// NewRedisHandler constructor for RedisHandler
func NewRedisHandler(address string, logger loggers.Logger) repository.RedisHandler {
	client := redis.NewClient(&redis.Options{
		Addr: address,
	})

	return &RedisHandler{
		Client: client,
		Logger: logger,
	}
}

// Get gets the result of a GET command with the given key
func (r RedisHandler) Get(key string) (interface{}, error) {
	redisResult := new(RedisResult)
	redisResult.Result = redis.NewStringCmd("get", key)
	redisResult.Logger = r.Logger
	result := r.Client.Get(key)
	err := result.Err()
	if err != nil {
		if err == redis.Nil {
			return redisResult, fmt.Errorf("KEY_NOT_FOUND: %s", key)
		}
		return redisResult, err
	}
	redisResult.Result = result
	return redisResult, err
}

// Set sets a value in redis with the given key
func (r RedisHandler) Set(key string, value interface{}, expiration time.Duration) error {
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	stringValue := string(b)
	status := r.Client.Set(key, stringValue, expiration)
	err = status.Err()
	return err
}

// Exists determines if given key exists in redis
func (r RedisHandler) Exists(key string) bool {
	status := r.Client.Exists(key)
	return status.Val() > 0
}

// Del deletes the given key from the database in redis
func (r RedisHandler) Del(key string) error {
	result := r.Client.Del(key)
	return result.Err()
}

// Scan stores the result row of a query execution into the given interfaces
func (r RedisResult) Scan(dest interface{}) error {
	stringResult, _ := r.Result.Result()
	if stringResult == "" {
		return fmt.Errorf("Empty string, nothing to parse")
	}
	err := json.Unmarshal([]byte(stringResult), dest)
	if err != nil {
		if e, ok := err.(*json.SyntaxError); ok {
			return fmt.Errorf("syntax error at byte offset %d", e.Offset)
		}
	}
	return err
}
