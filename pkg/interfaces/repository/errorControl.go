package repository

import (
	"fmt"

	"github.schibsted.io/Yapo/yams-dav-sync/pkg/interfaces"
)

// errorControlRepo repository to store error marks in dav-yams synchronization process
type errorControlRepo struct {
	db             DbHandler
	resultsPerPage int
}

// NewErrorControlRepo creates a new instance of ErrorControl repository
func NewErrorControlRepo(dbHandler DbHandler, resultsPerPage int) interfaces.ErrorControl {
	return &errorControlRepo{
		db:             dbHandler,
		resultsPerPage: resultsPerPage,
	}
}

// GetSyncError gets all error marks in repository using pagination
func (repo *errorControlRepo) GetPreviousErrors(nPage, maxErrorTolerance int) (result []string, err error) {
	rows, err := repo.db.Query(fmt.Sprintf(`
		SELECT image_path	
		FROM sync_error 
		WHERE 
			error_counter <= %d
		ORDER BY
			sync_error_id 
		LIMIT 
			%d OFFSET %d*(%d-1)`,
		maxErrorTolerance,
		repo.resultsPerPage,
		repo.resultsPerPage,
		nPage,
	))
	defer rows.Close() // nolint

	if err != nil {
		return result, fmt.Errorf("Error getting synchronization marks: %+v", err)
	}

	var imgPath string
	for rows.Next() {
		rows.Scan(&imgPath) // nolint
		result = append(result, imgPath)
	}
	return result, nil
}

// GetErrorsPagesQty get the total pages number for pagination
func (repo *errorControlRepo) GetErrorsPagesQty(maxErrorTolerance int) (nPages int) {
	if repo.resultsPerPage < 1 {
		return 0
	}

	result, err := repo.db.Query(
		fmt.Sprintf(
			`SELECT count(*)
			FROM sync_error
			WHERE error_counter <= %d`,
			maxErrorTolerance,
		),
	)
	defer result.Close() // nolint
	if err != nil {
		return 0
	}
	rows := 0
	if result.Next() {
		err = result.Scan(&rows)
		if err != nil {
			return 0
		}
	}
	nPages = rows / repo.resultsPerPage
	if rows%repo.resultsPerPage > 0 && rows > 0 {
		nPages++
	}
	return
}

// CleanErrorMarks deletes the error mark for a specific image in repository
func (repo *errorControlRepo) CleanErrorMarks(imgPath string) error {
	result, err := repo.db.Query(fmt.Sprintf(`
		DELETE  
		FROM sync_error
		where image_path = '%s'`,
		imgPath,
	))
	defer result.Close() // nolint
	return err
}

// SetErrorCounter sets the error counter in repository for a specific image, if
// does not exist then create the error mark with a given counter
func (repo *errorControlRepo) SetErrorCounter(imagePath string, count int) (err error) {
	row, err := repo.db.Query(
		fmt.Sprintf(`
			INSERT INTO
				sync_error(image_path, error_counter)
			VALUES 
				('%s',%d)
			ON CONFLICT ON CONSTRAINT image_path_unique
				DO UPDATE SET error_counter = %d`,
			imagePath,
			count,
			count,
		))
	defer row.Close() // nolint

	if err != nil {
		err = fmt.Errorf("There was an error creating errors sync: %+v", err)
	}
	return
}

// IncreaseErrorCounter creates an error mark for a specific image, if exists then
// increases the error counter
func (repo *errorControlRepo) IncreaseErrorCounter(imagePath string) (err error) {
	row, err := repo.db.Query(
		fmt.Sprintf(`
			INSERT INTO
				sync_error(image_path, error_counter)
			VALUES
				('%s', 0)
			ON CONFLICT ON CONSTRAINT image_path_unique
				DO UPDATE SET 
				error_counter = sync_error.error_counter + 1`,
			imagePath,
		))
	defer row.Close() // nolint
	if err != nil {
		err = fmt.Errorf("There was an error creating errors sync: %+v", err)
	}
	return
}
