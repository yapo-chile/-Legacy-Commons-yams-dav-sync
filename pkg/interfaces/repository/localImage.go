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

// OpenFile opens a file from local storage
func (repo *LocalImageRepo) OpenFile(path string) (usecases.File, error) {
	return repo.fileSystemView.Open(path)
}

// GetImage gets a single image from local repository
func (repo *LocalImageRepo) GetImage(imagePath string) (domain.Image, error) {
	if len(imagePath) < 2 {
		return domain.Image{}, fmt.Errorf("ImagePath too short: %+v", imagePath)
	}
	filePath := path.Join(repo.path, imagePath[:2], imagePath)
	f, err := repo.OpenFile(filePath)
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

// GetImageListElement gets tuple element from image List, element format must be
// [date][space][imagepath]
func (repo *LocalImageRepo) GetImageListElement() string {
	return repo.fileSystemView.Text()
}

// NextImageListElement returns true if there is more elements in Image List, otherwise returns false
func (repo *LocalImageRepo) NextImageListElement() bool {
	return repo.fileSystemView.Scan()
}

// ErrorScanningImageList returns error if the process of get element from image list failed
func (repo *LocalImageRepo) ErrorScanningImageList() error {
	return repo.fileSystemView.Err()
}

// InitImageListScanner initialize scanner to read image list from file
func (repo *LocalImageRepo) InitImageListScanner(f usecases.File) {
	repo.fileSystemView.NewScanner(f)
}
