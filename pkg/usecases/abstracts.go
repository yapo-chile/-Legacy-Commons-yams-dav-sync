package usecases

import "time"

// LastSyncRepository allows LastSyncRepository's operations
type LastSyncRepository interface {
	GetLastSync() time.Time
	SetLastSync(value string) error
}

// ErrorControlRepository allows ErrorControlRepository's operations
type ErrorControlRepository interface {
	GetErrorSync(nPage int) (result []string, err error)
	GetPagesQty() int
	DelErrorSync(imgPath string) error
	AddErrorSync(imagePath string) (err error)
	SetErrorCounter(imagePath string, count int) error
}

// ImageStatusRepository allows ImageStatusRepository's operations
type ImageStatusRepository interface {
	GetImageStatus(key string) (string, error)
	SetImageStatus(key, value string) error
	DelImageStatus(key string) error
}
