package repository

import (
	"crypto/md5" // nolint:gosec
	"encoding/hex"
	"fmt"
	"io"
	"path"

	"github.schibsted.io/Yapo/yams-dav-sync/pkg/domain"
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

// Open opens a file from local storage
func (repo *LocalImageRepo) Open(path string) (usecases.File, error) {
	return repo.fileSystemView.Open(path)
}

// GetImage gets a single image from local repository
func (repo *LocalImageRepo) GetImage(imagePath string) (domain.Image, error) {
	if len(imagePath) < 2 {
		return domain.Image{}, fmt.Errorf("ImagePath too short: %+v", imagePath)
	}
	filePath := path.Join(repo.path, imagePath[:2], imagePath)
	f, err := repo.Open(filePath)
	if err != nil {
		return domain.Image{}, err
	}
	defer f.Close() // nolint:errcheck,gosec

	hash := md5.New() // nolint:gosec
	_, err = io.Copy(hash, f)
	if err != nil {
		return domain.Image{}, err
	}
	image := domain.Image{
		FilePath: filePath,
		Metadata: domain.ImageMetadata{
			ImageName: repo.fileSystemView.Name(filePath),
			Size:      repo.fileSystemView.Size(filePath),
			Checksum:  hex.EncodeToString(hash.Sum(nil)),
			ModTime:   repo.fileSystemView.ModTime(filePath),
		},
	}
	return image, nil
}
