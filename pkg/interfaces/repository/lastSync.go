package repository

import (
	"fmt"

	"github.schibsted.io/Yapo/yams-dav-sync/pkg/usecases"
)

// lastSyncRepo repository to save current synchronization status of images.
// If a image is synchronized this repository will keep a key with the image name
// and the value will be the md5 checksum of the image. The image MD5 checksum
// musts match with the checksum kept by lastSync repository and Yams repository
type lastSyncRepo struct {
	db DbHandler
}

// NewLastSyncRepo makes a new LastSyncRepo instance
func NewLastSyncRepo(dbHandler DbHandler) usecases.LastSyncRepository {
	return &lastSyncRepo{
		db: dbHandler,
	}
}

// GetLastSync returns the status of synchronization process in redis
func (repo *lastSyncRepo) GetLastSync() (dateStr string, err error) {
	result, err := repo.db.Query(fmt.Sprintf(`
		SELECT 
		last_sync_date
		FROM last_sync 
		ORDER BY last_sync_id DESC 
		LIMIT 1`))
	if result.Next() {
		result.Scan(&dateStr)
	}
	defer result.Close()
	return dateStr, err
}

// SetLastSync saves the checksum of the image in redis repo
func (repo *lastSyncRepo) SetLastSync(value string) (err error) {
	row, err := repo.db.Query(
		fmt.Sprintf(
			`INSERT INTO
					last_sync(
						last_sync_date
					)
				VALUES (
					'%s'
				)`,
			value,
		))

	if err != nil {
		err = fmt.Errorf("There was an erro creating last synchronization mark")
		return
	}

	defer row.Close()
	return
}
