package infrastructure

import (
	"os"

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
	return os.Open(name)
}

// Stat gets the stats of a file in local storage
func (LocalFileSystemView) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}
