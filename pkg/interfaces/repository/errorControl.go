package repository

import (
	"fmt"

	"github.schibsted.io/Yapo/yams-dav-sync/pkg/usecases"
)

// errorControlRepo repository to store error marks in dav-yams synchronization process
type errorControlRepo struct {
	db             DbHandler
	maxRetries     int
	resultsPerPage int
}

// NewErrorControlRepo creates a new instance of ErrorControl repository
func NewErrorControlRepo(dbHandler DbHandler, maxRetries, resultsPerPage int) usecases.ErrorControlRepository {
	return &errorControlRepo{
		db:             dbHandler,
		maxRetries:     maxRetries,
		resultsPerPage: resultsPerPage,
	}
}

// GetErrorSync gets all error marks in repository using pagination
func (repo *errorControlRepo) GetErrorSync(nPage int) (result []string, err error) {
	rows, err := repo.db.Query(fmt.Sprintf(`
		SELECT image_path	
		FROM sync_error 
		WHERE 
			error_counter <= %d
		ORDER BY
			sync_error_id 
		LIMIT 
			%d OFFSET %d*(%d-1)`,
		repo.maxRetries,
		repo.resultsPerPage,
		repo.resultsPerPage,
		nPage,
	))

	if err != nil {
		return result, fmt.Errorf("Error getting synchronization marks: %+v", err)
	}

	var imgPath string
	for rows.Next() {
		rows.Scan(&imgPath)
		result = append(result, imgPath)
	}
	rows.Close()
	return result, nil
}

// GetPagesQty get the total pages number for pagination
func (repo *errorControlRepo) GetPagesQty() (nPages int) {
	if repo.resultsPerPage < 1 {
		return 0
	}

	rows, err := repo.db.Query(`
		SELECT 
		count(*) 
		FROM sync_error
		`)
	if err != nil {
		return 0
	}

	if rows.Next() {
		err = rows.Scan(&nPages)
		if err != nil {
			return 0
		}
	}
	nPages = nPages / repo.resultsPerPage
	if nPages%repo.resultsPerPage > 0 && nPages > 0 {
		nPages++
	}
	rows.Close()
	return
}

// DelErrorSync deletes the error mark for a specific image in repository
func (repo *errorControlRepo) DelErrorSync(imgPath string) error {
	result, err := repo.db.Query(fmt.Sprintf(`
		DELETE  
		FROM sync_error
		where image_path = '%s'`,
		imgPath,
	))
	if err != nil {
		return err
	}
	result.Close()
	return nil
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

	if err != nil {
		return fmt.Errorf("There was an error creating errors sync: %+v", err)
	}
	row.Close()
	return
}

// AddErrorSync creates a error mark for specific image, if exists then
// increases the error counter
func (repo *errorControlRepo) AddErrorSync(imagePath string) (err error) {
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

	if err != nil {
		return fmt.Errorf("There was an error creating errors sync: %+v", err)
	}
	row.Close()
	return
}
