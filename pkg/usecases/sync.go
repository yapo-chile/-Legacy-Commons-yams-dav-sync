package usecases

import (
	"errors"

	"github.schibsted.io/Yapo/yams-dav-sync/pkg/domain"
)

// SyncInteractor executes operations for syncher between local storage and yams bucket
type SyncInteractor struct {
	YamsRepo        YamsRepository
	LocalRepo       LocalImageRepository
	ImageStatusRepo ImageStatusRepository
	Logger          SyncLogger
}

// SyncLogger logs synchronization events
type SyncLogger interface {
	LogProcessImage(img domain.Image, sent, skipped, proccessed int)
	LogUploadingImage(img domain.Image)
	ErrorDuplicatedImage(img domain.Image)
	ErrorDeletingImageInYams(imgID string, e error)
	ErrorDeletingImageStatusInRepo(imgID string, e error)
	ImageSuccessfullyDelete(img domain.Image)
	MarkingAsSynchronized(img domain.Image)
	PassingOver(img domain.Image)
	LogErrorGettingImages(err error)
	LogErrorSendingImage(img domain.Image, err error)
}

// LocalImageRepository allows local storage operations
type LocalImageRepository interface {
	GetImages() []domain.Image
}

var errImageNotFound = errors.New("Image Not Found")

// ValidateChecksum validates if a given imagen contains match with the checksum
// stored in ImageStatus repository
func (i *SyncInteractor) ValidateChecksum(image domain.Image) bool {
	registeredHash, _ := i.ImageStatusRepo.GetImageStatus(image.Metadata.ImageName)
	return image.Metadata.Checksum == registeredHash
}

// Send executes the synchronization of images between local storage and yams bucket
func (i *SyncInteractor) Send(image domain.Image) error {
	i.Logger.LogUploadingImage(image)
	if err := i.YamsRepo.PutImage(image); err != nil {
		switch err {
		case ErrYamsDuplicate:
			i.Logger.ErrorDuplicatedImage(image)
			externalHash, _ := i.YamsRepo.HeadImage(image.Metadata.ImageName)
			// Check if the error is only because the name or content
			if externalHash != image.Metadata.Checksum {
				if e := i.YamsRepo.DeleteImage(image.Metadata.ImageName, true); e != nil {
					i.Logger.ErrorDeletingImageInYams(image.Metadata.ImageName, e)
				}
				if e := i.ImageStatusRepo.DelImageStatus(image.Metadata.ImageName); e != nil {
					i.Logger.ErrorDeletingImageStatusInRepo(image.Metadata.ImageName, e)
				}
				i.Logger.ImageSuccessfullyDelete(image)
			} else {
				// the image was already synchronized but not marked by redis
				i.Logger.MarkingAsSynchronized(image)
				i.ImageStatusRepo.SetImageStatus(image.Metadata.ImageName, image.Metadata.Checksum)
			}
		default:
			// with another kind of errors pass over the image
		}
	}
	i.ImageStatusRepo.SetImageStatus(image.Metadata.ImageName, image.Metadata.Checksum)

	// TODO: Consider case when image is in Yams' directory but not in local folder.
	return nil
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
