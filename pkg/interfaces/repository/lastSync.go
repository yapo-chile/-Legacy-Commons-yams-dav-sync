package repository

import (
	"fmt"
	"time"

	"github.schibsted.io/Yapo/yams-dav-sync/pkg/interfaces"
)

// lastSyncRepo repository to save current synchronization date mark
type lastSyncRepo struct {
	db          DbHandler
	defaultDate time.Time
	dateLayout  string
}

// NewLastSyncRepo makes a new LastSyncRepo instance
func NewLastSyncRepo(dbHandler DbHandler, dateLayout string, defaultLastSyncDate time.Time) interfaces.LastSync {
	return &lastSyncRepo{
		db:          dbHandler,
		defaultDate: defaultLastSyncDate,
		dateLayout:  dateLayout,
	}
}

// GetLastSynchronizationMark returns the last synchronization date mark
func (repo *lastSyncRepo) GetLastSynchronizationMark() (lastSyncDate time.Time) {
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

// SetLastSynchronizationMark saves a new synchronization date mark
func (repo *lastSyncRepo) SetLastSynchronizationMark(date time.Time) (err error) {
	row, err := repo.db.Query(fmt.Sprintf(`
		INSERT INTO last_sync(last_sync_date)
		VALUES ('%+v')`,
		date.Format(repo.dateLayout),
	))
	defer row.Close() // nolint
	return err
}
