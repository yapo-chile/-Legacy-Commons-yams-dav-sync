package repository

import (
	"time"

	"github.schibsted.io/Yapo/yams-dav-sync/pkg/usecases"
)

type imageStatusRepository struct {
	handler        RedisHandler
	prefix         string
	expirationTime time.Duration
}

func NewImageStatusRepository(handler RedisHandler, prefix string, expirationTime time.Duration) usecases.ImageStatusRepo {
	return &imageStatusRepository{
		handler:        handler,
		prefix:         prefix,
		expirationTime: expirationTime,
	}
}

func (repo *imageStatusRepository) makeRedisKey(listID string) string {
	return repo.prefix + listID
}

// GetAdCache returns the status of a cached ad
func (repo *imageStatusRepository) GetImageStatus(key string) (bool, error) {
	var adCachedStatus bool
	res, err := repo.handler.Get(repo.makeRedisKey(key))
	if err == nil {
		res.Scan(&adCachedStatus)
	}
	return adCachedStatus, err
}

// SetAdCache saves the status of the ad to redis
func (repo *imageStatusRepository) SetImageStatus(listID string, adCachedStatus bool) error {
	return repo.handler.Set(repo.makeRedisKey(listID), adCachedStatus, repo.expirationTime)
}
