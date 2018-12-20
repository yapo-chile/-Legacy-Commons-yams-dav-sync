package usecases

import (
	"time"

	"github.schibsted.io/Yapo/yams-dav-sync/pkg/domain"
)

// SyncInteractor executes operations for syncher between local storage and yams bucket
type SyncInteractor struct {
	YamsRepo         YamsRepository
	ImageRepo        ImageRepository
	LastSyncRepo     LastSyncRepository
	ErrorControlRepo ErrorControlRepository
}

// ImageRepository allows local storage operations for images
type ImageRepository interface {
	GetImage(imagePath string) (domain.Image, error)
	Open(Path string) (File, error)
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

// RemoteDelete deletes image from yams bucket
func (i *SyncInteractor) RemoteDelete(imageName string) error {
	return i.YamsRepo.DeleteImage(imageName, domain.YAMSForceRemoval)
}

// GetMaxConcurrency get maximum supported concurrency by yams
func (i *SyncInteractor) GetMaxConcurrency() int {
	return i.YamsRepo.GetMaxConcurrentConns()
}

// GetRemoteChecksum gets the checksum of image in YAMS
func (i *SyncInteractor) GetRemoteChecksum(imageName string) (string, error) {
	return i.YamsRepo.HeadImage(imageName)
}

// GetErrorsPagesQty gets the number of pages for error pagination
func (i *SyncInteractor) GetErrorsPagesQty(maxErrorTolerance int) int {
	return i.ErrorControlRepo.GetPagesQty(maxErrorTolerance)
}

// GetPreviusErrors gets a list with previus errors, errors must have itsown counter
// over maxErrorTolerance
func (i *SyncInteractor) GetPreviusErrors(pagination, maxErrorTolerance int) ([]string, error) {
	return i.ErrorControlRepo.GetSyncErrors(pagination, maxErrorTolerance)
}

// CleanErrorMarks cleans every error mark associated with the image
func (i *SyncInteractor) CleanErrorMarks(imgName string) error {
	return i.ErrorControlRepo.DelSyncError(imgName)
}

// ResetErrorCounter sets the error counter to 0 for an specific image
func (i *SyncInteractor) ResetErrorCounter(imageName string) error {
	return i.ErrorControlRepo.SetErrorCounter(imageName, 0)
}

// IncreaseErrorCounter increase the error counter in one, if the image does not
// have error mark, the mark will be created
func (i *SyncInteractor) IncreaseErrorCounter(imageName string) error {
	return i.ErrorControlRepo.AddSyncError(imageName)
}

// GetLocalImage gets image form local storage
func (i *SyncInteractor) GetLocalImage(imagePath string) (domain.Image, error) {
	return i.ImageRepo.GetImage(imagePath)
}

// GetLastSynchornizationMark gets the date of latest synchronizated image
func (i *SyncInteractor) GetLastSynchornizationMark() time.Time {
	return i.LastSyncRepo.GetLastSync()
}

// SetLastSynchornizationMark sets the date of latest synchronizated image
func (i *SyncInteractor) SetLastSynchornizationMark(imageDateStr string) error {
	return i.LastSyncRepo.SetLastSync(imageDateStr)
}
