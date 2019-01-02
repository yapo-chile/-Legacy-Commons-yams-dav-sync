package usecases

import (
	"io"
	"time"
)

// LastSyncRepository allows LastSyncRepository's operations
type LastSyncRepository interface {
	GetLastSynchronizationMark() time.Time
	SetLastSync(value string) error
}

// File allows File's operations in local storage
type File interface {
	io.Closer
	io.Reader
}
