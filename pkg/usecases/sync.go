package usecases

import (
	"github.schibsted.io/Yapo/yams-dav-sync/pkg/domain"
)

// SyncInteractor executes operations for syncher between local storage and yams bucket
type SyncInteractor struct {
	YamsRepo         YamsRepository
	LocalStorageRepo LocalStorageRepository
	LastSyncRepo     LastSyncRepository
	SyncErrorRepo    ErrorControlRepository
}

// LocalStorageRepository allows local storage operations
type LocalStorageRepository interface {
	GetImage(imagePath string) (domain.Image, error)
}

// ValidateChecksum returns true if a given image exists in yams repository, otherwise
// returns false
func (i *SyncInteractor) ValidateChecksum(image domain.Image) bool {
	registeredHash, _ := i.YamsRepo.HeadImage(image.Metadata.ImageName)
	return image.Metadata.Checksum == registeredHash
}

// Send sends images from local storage to yams bucket
func (i *SyncInteractor) Send(image domain.Image) error {
	return i.YamsRepo.PutImage(image)
}

// List gets list of available images in yams bucket
func (i *SyncInteractor) List() ([]YamsObject, error) {
	return i.YamsRepo.GetImages()
}

// Delete deletes image from yams bucket
func (i *SyncInteractor) Delete(imageName string) error {
	return i.YamsRepo.DeleteImage(imageName, domain.YAMSForceRemoval)
}
