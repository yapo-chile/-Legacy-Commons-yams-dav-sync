package repository

import (
	"time"

	"github.mpi-internal.com/Yapo/yams-dav-sync/pkg/interfaces"
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
	return repo.db.Insert(`
		INSERT INTO last_sync(last_sync_date)
		VALUES ($1)`,
		date.Format(repo.dateLayout),
	)
}

// Reset deletes the last synchronization date mark to run the process again from the last
// checkpoint
func (repo *lastSyncRepo) Reset() (err error) {
	result, err := repo.db.Query(`
		DELETE FROM last_sync
		WHERE last_sync_id
		IN (SELECT last_sync_id FROM last_sync
			ORDER BY last_sync_id DESC LIMIT 1)
		`)
	result.Close() // nolint
	return
}

// Get gets a list of synchronization marks order by newer to older
func (repo *lastSyncRepo) Get() (marks []string, err error) {
	result, err := repo.db.Query(`
		SELECT last_sync_date
		FROM last_sync
		ORDER BY last_sync_id DESC`)
	if err != nil {
		return []string{}, err
	}
	defer result.Close()
	for result.Next() {
		var temp string
		err := result.Scan(&temp)
		if err != nil {
			return []string{}, err
		}
		marks = append(marks, temp)
	}
	return
}
