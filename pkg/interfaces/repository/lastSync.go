package repository

import (
	"fmt"
	"time"

	"github.schibsted.io/Yapo/yams-dav-sync/pkg/usecases"
)

// lastSyncRepo repository to save current synchronization date mark
type lastSyncRepo struct {
	db          DbHandler
	defaultDate time.Time
}

// NewLastSyncRepo makes a new LastSyncRepo instance
func NewLastSyncRepo(dbHandler DbHandler, defaultLastSyncDate time.Time) usecases.LastSyncRepository {
	return &lastSyncRepo{
		db:          dbHandler,
		defaultDate: defaultLastSyncDate,
	}
}

// GetLastSync returns the last synchronization date mark
func (repo *lastSyncRepo) GetLastSync() (lastSyncDate time.Time) {
	result, err := repo.db.Query(`
		SELECT last_sync_date
		FROM last_sync 
		ORDER BY last_sync_id DESC 
		LIMIT 1`)
	defer result.Close() // nolint
	if err != nil {
		return repo.defaultDate
	}
	if result.Next() {
		err = result.Scan(&lastSyncDate)
		if err != nil {
			return repo.defaultDate
		}
	}

	return lastSyncDate
}

// GetLastSync saves a new synchronization date mark
func (repo *lastSyncRepo) SetLastSync(dateMark string) (err error) {
	if dateMark == "" {
		return fmt.Errorf("dateMark not valid")
	}

	row, err := repo.db.Query(fmt.Sprintf(`
		INSERT INTO last_sync(last_sync_date)
		VALUES ('%s')`,
		dateMark,
	))
	defer row.Close() // nolint
	if err != nil {
		err = fmt.Errorf("There was an error creating last synchronization mark: %+v", err)
		return
	}
	return
}
