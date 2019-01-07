package repository

import (
	"crypto/md5" // nolint:gosec
	"encoding/hex"
	"fmt"
	"path"

	"github.schibsted.io/Yapo/yams-dav-sync/pkg/domain"
	"github.schibsted.io/Yapo/yams-dav-sync/pkg/interfaces"
	"github.schibsted.io/Yapo/yams-dav-sync/pkg/usecases"
)

// LocalImageRepo is a local storage representation
type LocalImageRepo struct {
	// path is the path to get objects to send to yams
	path string
	// fileSystem allows operations in local storage
	fileSystemView FileSystemView
}

// NewLocalImageRepo returns a fresh instance of LocalImageRepo
func NewLocalImageRepo(path string, fileSystemView FileSystemView) *LocalImageRepo {
	localImageRepo := LocalImageRepo{
		path:           path,
		fileSystemView: fileSystemView,
	}
	return &localImageRepo
}

// OpenFile opens a file from local storage
func (repo *LocalImageRepo) OpenFile(path string) (usecases.File, error) {
	return repo.fileSystemView.Open(path)
}

// GetLocalImage gets a single image from local repository
func (repo *LocalImageRepo) GetLocalImage(imagePath string) (domain.Image, error) {
	if len(imagePath) < 2 {
		return domain.Image{}, fmt.Errorf("ImagePath too short: %+v", imagePath)
	}
	filePath := path.Join(repo.path, imagePath[:2], imagePath)
	f, err := repo.OpenFile(filePath)
	if err != nil {
		return domain.Image{}, err
	}
	defer f.Close() // nolint:errcheck,gosec

	fileInfo, err := repo.fileSystemView.Info(filePath)
	if err != nil {
		return domain.Image{}, err
	}

	hash := md5.New() // nolint:gosec
	err = repo.fileSystemView.Copy(hash, f)
	if err != nil {
		return domain.Image{}, err
	}

	image := domain.Image{
		FilePath: filePath,
		Metadata: domain.ImageMetadata{
			ImageName: fileInfo.Name(),
			Size:      fileInfo.Size(),
			Checksum:  hex.EncodeToString(hash.Sum(nil)),
			ModTime:   fileInfo.ModTime(),
		},
	}
	return image, nil
}

// InitImageListScanner initialize scanner to read image list from file
func (repo *LocalImageRepo) InitImageListScanner(f usecases.File) interfaces.Scanner {
	return repo.fileSystemView.NewScanner(f)
}
