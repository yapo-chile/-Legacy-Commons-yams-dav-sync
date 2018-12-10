package repository

import (
	"fmt"

	"github.schibsted.io/Yapo/yams-dav-sync/pkg/usecases"
)

type errorControlRepo struct {
	db             DbHandler
	maxRetries     int
	resultsPerPage int
}

func NewErrorControlRepo(dbHandler DbHandler, maxRetries, resultsPerPage int) usecases.ErrorControlRepository {
	return &errorControlRepo{
		db:             dbHandler,
		maxRetries:     maxRetries,
		resultsPerPage: resultsPerPage,
	}
}

func (repo *errorControlRepo) GetErrorSync(nPage int) (result []string, err error) {
	rows, err := repo.db.Query(fmt.Sprintf(`
		SELECT 
			image_path	
		FROM 
			sync_error 
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

	imgPath := ""
	for rows.Next() {
		rows.Scan(&imgPath)
		fmt.Println(imgPath)
		result = append(result, imgPath)
	}

	defer rows.Close()
	return result, err
}

func (repo *errorControlRepo) GetPagesQty() (nPages int) {

	if repo.resultsPerPage < 1 {
		return 0
	}

	rows, _ := repo.db.Query(fmt.Sprintf(`
		SELECT 
		count(*) 
		FROM sync_error`,
	))

	if rows.Next() {
		rows.Scan(&nPages)
	}
	nPages = nPages / repo.resultsPerPage
	if nPages%repo.resultsPerPage > 0 && nPages > 0 {
		nPages++
	}
	defer rows.Close()
	return

}

func (repo *errorControlRepo) DelErrorSync(imgPath string) error {
	result, err := repo.db.Query(fmt.Sprintf(`
		DELETE  
		FROM sync_error
		where image_path = '%s'`,
		imgPath,
	))
	defer result.Close()
	return err
}

func (repo *errorControlRepo) SetErrorCounter(imagePath string, count int) (err error) {
	row, err := repo.db.Query(
		fmt.Sprintf(
			`INSERT INTO
					sync_error(
						image_path,
						error_counter
					)
				VALUES (
					'%s',
					 %d
				)
			ON CONFLICT ON CONSTRAINT sync_error_image_path_key
			DO UPDATE SET error_counter = %d
			`,
			imagePath,
			count,
			count,
		))

	if err != nil {
		err = fmt.Errorf("There was an error creating errors sync")
		return
	}

	defer row.Close()
	return

}

func (repo *errorControlRepo) SetErrorSync(imagePath string) (err error) {
	row, err := repo.db.Query(
		fmt.Sprintf(
			`INSERT INTO
					sync_error(
						image_path,
						error_counter
					)
				VALUES (
					'%s', 
					0
				)
			ON CONFLICT ON CONSTRAINT image_path_unique
			DO UPDATE SET error_counter = sync_error.error_counter + 1
			`,
			imagePath,
		))

	if err != nil {
		err = fmt.Errorf("There was an error creating errors sync")
		return
	}

	defer row.Close()
	return
}
