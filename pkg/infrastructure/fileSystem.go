package infrastructure

import (
	"os"

	"github.schibsted.io/Yapo/yams-dav-sync/pkg/interfaces/repository"
	"github.schibsted.io/Yapo/yams-dav-sync/pkg/usecases"
)

// FileSystem implements fileSystem using the local disk.
type FileSystem struct{}

// NewFileSystem will create a new instance of a custom fileSystem
func NewFileSystem() repository.FileSystem {
	return &FileSystem{}
}

// Open opens a file from local storage
func (FileSystem) Open(name string) (usecases.File, error) {
	return os.Open(name)
}

// Stat gets the stats of a file in local storage
func (FileSystem) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}
