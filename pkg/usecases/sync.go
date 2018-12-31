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

// NewSyncInteractor returns a new instance of SyncInteractor
func NewSyncInteractor(yamsRepo YamsRepository, imageRepo ImageRepository,
	lastSyncRepo LastSyncRepository, errorControlRepo ErrorControlRepository) *SyncInteractor {
	return &SyncInteractor{
		YamsRepo:         yamsRepo,
		ImageRepo:        imageRepo,
		LastSyncRepo:     lastSyncRepo,
		ErrorControlRepo: errorControlRepo,
	}
}

// ImageRepository allows local storage operations for images
type ImageRepository interface {
	GetImage(imagePath string) (domain.Image, error)
	OpenFile(Path string) (File, error)
	GetImageListElement() string
	NextImageListElement() bool
	ErrorScanningImageList() error
	InitImageListScanner(f File)
}

// ValidateChecksum returns true if a given image exists in yams repository, otherwise
// returns false
func (i *SyncInteractor) ValidateChecksum(image domain.Image) bool {
	registeredHash, _ := i.YamsRepo.HeadImage(image.Metadata.ImageName)
	return image.Metadata.Checksum == registeredHash
}

// Send sends images from local storage to yams bucket
func (i *SyncInteractor) Send(image domain.Image) *YamsRepositoryError {
	return i.YamsRepo.PutImage(image)
}

// List gets list of available images in yams bucket
func (i *SyncInteractor) List() ([]YamsObject, *YamsRepositoryError) {
	return i.YamsRepo.GetImages()
}

// RemoteDelete deletes image from yams bucket
func (i *SyncInteractor) RemoteDelete(imageName string) *YamsRepositoryError {
	return i.YamsRepo.DeleteImage(imageName, domain.YAMSForceRemoval)
}

// GetMaxConcurrency get maximum supported concurrency by yams
func (i *SyncInteractor) GetMaxConcurrency() int {
	return i.YamsRepo.GetMaxConcurrentConns()
}

// GetRemoteChecksum gets the checksum of image in YAMS
func (i *SyncInteractor) GetRemoteChecksum(imageName string) (string, *YamsRepositoryError) {
	return i.YamsRepo.HeadImage(imageName)
}

// GetErrorsPagesQty gets the number of pages for error pagination
func (i *SyncInteractor) GetErrorsPagesQty(maxErrorTolerance int) int {
	return i.ErrorControlRepo.GetPagesQty(maxErrorTolerance)
}

// GetPreviousErrors gets a list with previus errors, errors must have itsown counter
// over maxErrorTolerance
func (i *SyncInteractor) GetPreviousErrors(pagination, maxErrorTolerance int) ([]string, error) {
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

// GetLocalImage gets image from local storage parsed as domain.Image
func (i *SyncInteractor) GetLocalImage(imagePath string) (domain.Image, error) {
	return i.ImageRepo.GetImage(imagePath)
}

// OpenFile gets file from local storage returning readable File struct
func (i *SyncInteractor) OpenFile(imagePath string) (File, error) {
	return i.ImageRepo.OpenFile(imagePath)
}

// InitImageListScanner initialize scanner to read image list from file
func (i *SyncInteractor) InitImageListScanner(f File) {
	i.ImageRepo.InitImageListScanner(f)
}

// GetImageListElement gets tuple element from image List, element format must be
// [date][space][imagepath]
func (i *SyncInteractor) GetImageListElement() string {
	return i.ImageRepo.GetImageListElement()
}

// NextImageListElement returns true if there is more elements in Image List, otherwise returns false
func (i *SyncInteractor) NextImageListElement() bool {
	return i.ImageRepo.NextImageListElement()
}

// ErrorScanningImageList returns error if the process of get element from image list failed
func (i *SyncInteractor) ErrorScanningImageList() error {
	return i.ImageRepo.ErrorScanningImageList()
}

// GetLastSynchornizationMark gets the date of latest synchronizated image
func (i *SyncInteractor) GetLastSynchornizationMark() time.Time {
	return i.LastSyncRepo.GetLastSync()
}

// SetLastSynchornizationMark sets the date of latest synchronizated image
func (i *SyncInteractor) SetLastSynchornizationMark(imageDateStr string) error {
	return i.LastSyncRepo.SetLastSync(imageDateStr)
}
