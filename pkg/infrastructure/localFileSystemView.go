package infrastructure

import (
	"os"
	"time"

	"github.schibsted.io/Yapo/yams-dav-sync/pkg/interfaces/repository"
	"github.schibsted.io/Yapo/yams-dav-sync/pkg/usecases"
)

// LocalFileSystemView implements fileSystem using the local disk.
type LocalFileSystemView struct{}

// NewLocalFileSystemView will create a new instance of a custom fileSystem
func NewLocalFileSystemView() repository.FileSystemView {
	return &LocalFileSystemView{}
}

// Open opens a file from local storage
func (LocalFileSystemView) Open(name string) (usecases.File, error) {
	return os.Open(name) // nolint: gosec
}

// ModTime gets the file mod timein local storage
func (LocalFileSystemView) ModTime(name string) time.Time {
	fileInfo, err := os.Stat(name)
	if err != nil {
		return time.Time{}
	}
	return fileInfo.ModTime()
}

// Name gets the file name in local storage
func (LocalFileSystemView) Name(name string) string {
	fileInfo, err := os.Stat(name)
	if err != nil {
		return ""
	}
	return fileInfo.Name()
}

// Size gets the file size in local storage
func (LocalFileSystemView) Size(name string) int64 {
	fileInfo, err := os.Stat(name)
	if err != nil {
		return 0
	}
	return fileInfo.Size()
}
