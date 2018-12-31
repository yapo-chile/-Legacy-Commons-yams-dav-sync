package infrastructure

import (
	"bufio"
	"io"
	"os"
	"time"

	"github.schibsted.io/Yapo/yams-dav-sync/pkg/interfaces"
	"github.schibsted.io/Yapo/yams-dav-sync/pkg/interfaces/loggers"
	"github.schibsted.io/Yapo/yams-dav-sync/pkg/interfaces/repository"
	"github.schibsted.io/Yapo/yams-dav-sync/pkg/usecases"
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

// ModTime gets a file modTime in local storage
func (l *LocalFileSystemView) ModTime(name string) time.Time {
	fileInfo, err := os.Stat(name)
	if err != nil {
		l.logger.Error("Error getting file info: %+v", err)
		return time.Time{}
	}
	return fileInfo.ModTime()
}

// Name gets the file name in local storage
func (l *LocalFileSystemView) Name(name string) string {
	fileInfo, err := os.Stat(name)
	if err != nil {
		l.logger.Error("Error getting file info: %+v", err)
		return ""
	}
	return fileInfo.Name()
}

// Size gets the file size in local storage
func (l *LocalFileSystemView) Size(name string) int64 {
	fileInfo, err := os.Stat(name)
	if err != nil {
		l.logger.Error("Error getting file info: %+v", err)
		return 0
	}
	return fileInfo.Size()
}

// NewScanner initializes a the localFileSystemView.Scanner to read from file
func (l *LocalFileSystemView) NewScanner(file usecases.File) interfaces.Scanner {
	scanner := bufio.NewScanner(file)
	return scanner
}

// Copy copies from src to dst until either EOF is reached on src or an error occurs.
func (l *LocalFileSystemView) Copy(dst io.Writer, src io.Reader) (int64, error) {
	return io.Copy(dst, src)
}
