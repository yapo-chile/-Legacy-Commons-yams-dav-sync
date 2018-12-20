package usecases

import (
	"io"
	"time"
)

// LastSyncRepository allows LastSyncRepository's operations
type LastSyncRepository interface {
	GetLastSync() time.Time
	SetLastSync(value string) error
}

// ErrorControlRepository allows ErrorControlRepository's operations
type ErrorControlRepository interface {
	GetSyncErrors(nPage, maxErrorTolerance int) (result []string, err error)
	GetPagesQty(maxErrorTolerance int) int
	DelSyncError(imgPath string) error
	AddSyncError(imagePath string) (err error)
	SetErrorCounter(imagePath string, count int) error
}

// File allows File's operations in local storage
type File interface {
	io.Closer
	io.Reader
}
