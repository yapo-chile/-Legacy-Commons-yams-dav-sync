package repository

import (
	"time"

	"github.schibsted.io/Yapo/yams-dav-sync/pkg/usecases"
)

// imageStatusRepo repository to save current synchronization status of images.
// If a image is synchronized this repository will keep a key with the image name
// and the value will be the md5 checksum of the image. The image MD5 checksum
// musts match with the checksum kept by imageStatus repository and Yams repository
type imageStatusRepo struct {
	redisHandler   RedisHandler
	prefix         string
	expirationTime time.Duration
}

// NewImageStatusRepo makes a new ImageStatusRepo instance
func NewImageStatusRepo(handler RedisHandler, prefix string, expirationTime time.Duration) usecases.ImageStatusRepository {
	return &imageStatusRepo{
		redisHandler:   handler,
		prefix:         prefix,
		expirationTime: expirationTime,
	}
}

func (repo *imageStatusRepo) makeRedisKey(key string) string {
	return repo.prefix + key
}

// GetImageStatus returns the status of a image in redis
func (repo *imageStatusRepo) GetImageStatus(imageName string) (checksum string, err error) {
	res, err := repo.redisHandler.Get(repo.makeRedisKey(imageName))
	resp := res.(RedisResult)
	if err == nil {
		resp.Scan(&checksum)
	}
	return checksum, err
}

// DelImageStatus deletes the image status from redis repo
func (repo *imageStatusRepo) DelImageStatus(ImageName string) error {
	return repo.redisHandler.Del(repo.makeRedisKey(ImageName))
}

// SetImageStatus saves the checksum of the image in redis repo
func (repo *imageStatusRepo) SetImageStatus(imageName, checksum string) error {
	return repo.redisHandler.Set(repo.makeRedisKey(imageName), checksum, -1)
}
