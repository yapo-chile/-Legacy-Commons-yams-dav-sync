package usecases

import (
	"errors"

	"github.schibsted.io/Yapo/yams-dav-sync/pkg/domain"
)

// SyncInteractor executes operations for syncher between local storage and yams bucket
type SyncInteractor struct {
	YamsRepo      YamsRepository
	LocalRepo     LocalImageRepository
	LastSyncRepo  LastSyncRepository
	SyncErrorRepo ErrorControlRepository
	Logger        SyncLogger
}

// SyncLogger logs synchronization events
type SyncLogger interface {
	LogProcessImage(img domain.Image, sent, skipped, proccessed int)
	LogUploadingImage(img domain.Image)
	ErrorDuplicatedImage(img domain.Image)
	ErrorDeletingImageInYams(imgID string, e error)
	ErrorDeletingLastSyncInRepo(imgID string, e error)
	ImageSuccessfullyDelete(img domain.Image)
	MarkingAsSynchronized(img domain.Image)
	PassingOver(img domain.Image)
	LogErrorGettingImages(err error)
	LogErrorSendingImage(img domain.Image, err error)
}

// LocalImageRepository allows local storage operations
type LocalImageRepository interface {
	GetImage(imagePath string) (domain.Image, error)
}

var errImageNotFound = errors.New("Image Not Found")

// ValidateChecksum validates if a given imagen contains match with the checksum
// stored in LastSync repository
func (i *SyncInteractor) ValidateChecksum(image domain.Image) bool {
	registeredHash, _ := i.YamsRepo.HeadImage(image.Metadata.ImageName)
	return image.Metadata.Checksum == registeredHash
}

// Send executes the synchronization of images between local storage and yams bucket
func (i *SyncInteractor) Send(image domain.Image) error {
	i.Logger.LogUploadingImage(image)
	return i.YamsRepo.PutImage(image)
}

// List get a list of available images in yams bucket
func (i *SyncInteractor) List() ([]YamsObject, error) {
	return i.YamsRepo.GetImages()
}

// Delete deletes an the images of yams bucket
func (i *SyncInteractor) Delete(imageName string) error {
	if err := i.YamsRepo.DeleteImage(imageName, true); err != nil {
		i.Logger.ErrorDeletingImageInYams(imageName, err)
		return err
	}
	return nil
}
