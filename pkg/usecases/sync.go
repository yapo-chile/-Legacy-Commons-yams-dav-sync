package usecases

import (
	"errors"
	"fmt"

	"github.schibsted.io/Yapo/yams-dav-sync/pkg/domain"
)

// SyncUseCase executes operations for syncher between local storage and yams bucket
type SyncUseCase struct {
	YamsRepo  YamsRepository
	LocalRepo LocalImageRepository
}

// LocalImageRepository allows local storage operations
type LocalImageRepository interface {
	GetImages() []domain.Image
}

var errImageNotFound = errors.New("Image Not Found")

// Run executes the synchronization of images between local storage and yams bucket
func (uc *SyncUseCase) Run() error {
	fmt.Println("Sync Executed")
	count := 0
	images := uc.LocalRepo.GetImages()
	for _, image := range images {
		fmt.Printf("\n - SENDING %+v  .\n", image)
		err := uc.YamsRepo.PutImage(image)
		if err == nil {
			count++
		}
		if count == 1 {
			return nil
		}
	}
	// TODO: Make it smarter !
	// Consider case when image is in Yams' directory but not in local folder.
	return nil
}

// List get a list of available images in yams bucket
func (uc *SyncUseCase) List() ([]YamsObject, error) {
	return uc.YamsRepo.GetImages()
}

// DeleteAll deletes all the images of yams bucket
func (uc *SyncUseCase) DeleteAll() error {
	images, err := uc.YamsRepo.GetImages()
	if err != nil {
		return err
	}
	for _, img := range images {
		err := uc.YamsRepo.DeleteImage(img.ID, true)
		if err != nil {
			return err
		}
	}
	return nil
}
