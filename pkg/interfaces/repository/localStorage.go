package repository

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"regexp"

	"github.schibsted.io/Yapo/yams-dav-sync/pkg/domain"
	"github.schibsted.io/Yapo/yams-dav-sync/pkg/usecases"
)

// LocalStorageRepo is a local storage representation
type LocalStorageRepo struct {
	// path is the path to get objects to send to yams
	path string
	// fileSystem allows operations in local storage
	fileSystem FileSystem
	// logger logs event messages
	logger interface{}
}

// NewLocalStorageRepo returns a fresh instance of LocalStorageRepo
func NewLocalStorageRepo(path string, fileSystem FileSystem, logger interface{}) *LocalStorageRepo {
	localStorageRepo := LocalStorageRepo{
		path:       path,
		fileSystem: fileSystem,
		logger:     logger,
	}
	return &localStorageRepo
}

// Open opens a file from local storage
func (repo *LocalStorageRepo) Open(path string) (usecases.File, error) {
	return repo.fileSystem.Open(path)
}

var extRegex = regexp.MustCompile("(?i).(png|bmp|jpg)$")

// GetImage gets a single image from local repository
func (repo *LocalStorageRepo) GetImage(imagePath string) (domain.Image, error) {
	if len(imagePath) < 2 {
		return domain.Image{}, fmt.Errorf("ImagePath too short: %+v", imagePath)
	}
	filePath := path.Join(repo.path, imagePath[:2], imagePath)
	f, err := repo.fileSystem.Open(filePath)
	if err != nil {
		return domain.Image{}, err
	}
	fileInfo, err := f.Stat()
	if err != nil {
		return domain.Image{}, err
	}
	hash := md5.New()
	io.Copy(hash, f)
	image := domain.Image{
		FilePath: filePath,
		Metadata: domain.ImageMetadata{
			ImageName: fileInfo.Name(),
			Size:      fileInfo.Size(),
			Checksum:  hex.EncodeToString(hash.Sum(nil)),
			ModTime:   fileInfo.ModTime(),
		},
	}
	f.Close()
	return image, nil
}

// GetImages returns all images inside the path defined, including inner directories.
func (repo *LocalStorageRepo) GetImages() []domain.Image {
	var imagePath string

	// Convert relative path to absolute path
	if !path.IsAbs(repo.path) {
		workingDir, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		imagePath = path.Join(workingDir, repo.path)
	} else {
		imagePath = repo.path
	}

	images, err := navigate(imagePath)
	if err != nil {
		log.Fatal(err)
	}
	return images
}

func navigate(root string) ([]domain.Image, error) {
	f, err := os.Open(root)
	if err != nil {
		return nil, err
	}

	fileInfo, err := f.Readdir(-1)
	f.Close()
	if err != nil {
		return nil, err
	}

	var images []domain.Image
	for _, file := range fileInfo {

		filePath := path.Join(root, file.Name())
		if !file.IsDir() {
			if extRegex.MatchString(file.Name()) {
				f, e := os.Open(filePath)
				if e != nil {
					continue
				}
				hash := md5.New()
				io.Copy(hash, f)
				image := domain.Image{
					FilePath: filePath,
					Metadata: domain.ImageMetadata{
						ImageName: file.Name(),
						Size:      file.Size(),
						Checksum:  hex.EncodeToString(hash.Sum(nil)),
					},
				}
				images = append(images, image)
				f.Close()
			}
		} else {
			innerImages, err := navigate(filePath)
			if err != nil {
				return nil, err
			}

			images = append(images, innerImages...)
		}
	}

	return images, nil
}
