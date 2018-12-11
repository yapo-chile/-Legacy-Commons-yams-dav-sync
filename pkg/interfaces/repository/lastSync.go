package repository

import (
	"fmt"

	"github.schibsted.io/Yapo/yams-dav-sync/pkg/usecases"
)

// lastSyncRepo repository to save current synchronization date mark
type lastSyncRepo struct {
	db DbHandler
}

// NewLastSyncRepo makes a new LastSyncRepo instance
func NewLastSyncRepo(dbHandler DbHandler) usecases.LastSyncRepository {
	return &lastSyncRepo{
		db: dbHandler,
	}
}

// GetLastSync returns the last synchronization date mark
func (repo *lastSyncRepo) GetLastSync() (dateStr string, err error) {
	result, err := repo.db.Query(`
		SELECT last_sync_date
		FROM last_sync 
		ORDER BY last_sync_id DESC 
		LIMIT 1`)
	if result.Next() {
		result.Scan(&dateStr)
	}
	result.Close()
	return dateStr, err
}

// GetLastSync saves a new synchronization date mark
func (repo *lastSyncRepo) SetLastSync(dateMark string) (err error) {
	row, err := repo.db.Query(fmt.Sprintf(`
		INSERT INTO last_sync(last_sync_date)
		VALUES ('%s')`,
		dateMark,
	))

	if err != nil {
		err = fmt.Errorf("There was an erro creating last synchronization mark")
		return
	}

	row.Close()
	return
}
