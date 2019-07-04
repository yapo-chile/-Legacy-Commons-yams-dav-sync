package infrastructure

import (
	"bufio"
	"io"
	"os"

	"github.mpi-internal.com/Yapo/yams-dav-sync/pkg/interfaces"
	"github.mpi-internal.com/Yapo/yams-dav-sync/pkg/interfaces/loggers"
	"github.mpi-internal.com/Yapo/yams-dav-sync/pkg/interfaces/repository"
	"github.mpi-internal.com/Yapo/yams-dav-sync/pkg/usecases"
)

// LocalFileSystemView implements fileSystem using the local disk.
type LocalFileSystemView struct {
	logger loggers.Logger
}

// NewLocalFileSystemView will create a new instance of a custom fileSystem
func NewLocalFileSystemView(logger loggers.Logger) repository.FileSystemView {
	return &LocalFileSystemView{logger: logger}
}

// Open Opens a file from local storage
func (*LocalFileSystemView) Open(name string) (usecases.File, error) {
	return os.Open(name) // nolint: gosec
}

// NewScanner initializes a the localFileSystemView.Scanner to read from file
func (l *LocalFileSystemView) NewScanner(file usecases.File) interfaces.Scanner {
	scanner := bufio.NewScanner(file)
	return scanner
}

// Copy copies from src to dst until either EOF is reached on src or an error occurs.
func (l *LocalFileSystemView) Copy(dst io.Writer, src io.Reader) (err error) {
	_, err = io.Copy(dst, src)
	return
}

// Info returns FileInfo of a specific file
func (l *LocalFileSystemView) Info(name string) (repository.FileInfo, error) {
	return os.Stat(name)
}
