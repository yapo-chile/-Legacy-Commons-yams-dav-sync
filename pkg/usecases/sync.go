package usecases

import (
	"errors"

	"github.schibsted.io/Yapo/yams-dav-sync/pkg/domain"
)

// SyncInteractor executes operations for syncher between local storage and yams bucket
type SyncInteractor struct {
	YamsRepo  YamsRepository
	LocalRepo LocalImageRepository
	Logger    SyncLogger
}

type SyncLogger interface {
	LogSentImage(img domain.Image)
	LogErrorGettingImages(err error)
	LogErrorSendingImage(img domain.Image, err error)
	LogErrorDeletingImage(imgID string, err error)
}

// LocalImageRepository allows local storage operations
type LocalImageRepository interface {
	GetImages() []domain.Image
}

var errImageNotFound = errors.New("Image Not Found")

// Run executes the synchronization of images between local storage and yams bucket
func (i *SyncInteractor) Run() error {
	count := 0
	images := i.LocalRepo.GetImages()
	for _, image := range images {
		i.Logger.LogSentImage(image)
		err := i.YamsRepo.PutImage(image)
		if err == nil {
			count++
		}
		// TODO: Make it smarter !
		if count == 1 {
			return nil
		}
	}
	// Consider case when image is in Yams' directory but not in local folder.
	return nil
}

// List get a list of available images in yams bucket
func (i *SyncInteractor) List() ([]YamsObject, error) {
	return i.YamsRepo.GetImages()
}

// DeleteAll deletes all the images of yams bucket
func (i *SyncInteractor) DeleteAll() error {
	yamsResponse, err := i.YamsRepo.GetImages()
	if err != nil {
		i.Logger.LogErrorGettingImages(err)
		return err
	}
	for _, img := range yamsResponse {
		err := i.YamsRepo.DeleteImage(img.ID, true)
		if err != nil {
			i.Logger.LogErrorDeletingImage(img.ID, err)
			return err
		}
	}
	return nil
}
