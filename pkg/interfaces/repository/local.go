package repository

import (
	"log"
	"os"
	"path"
	"regexp"

	"github.schibsted.io/Yapo/yams-dav-sync/pkg/interfaces/logger"

	"github.schibsted.io/Yapo/yams-dav-sync/pkg/domain"
)

// LocalRepo is a local storage representation
type LocalRepo struct {
	// Path is the path to get objects to send to yams
	Path string
	// Logger logs event messages
	Logger logger.Logger
}

// NewLocalRepo returns a fresh instance of LocalRepo
func NewLocalRepo(path string, logger logger.Logger) *LocalRepo {
	localRepo := LocalRepo{
		Path:   path,
		Logger: logger,
	}
	return &localRepo
}

var extRegex = regexp.MustCompile("(?i).(png|bmp|jpg)$")

// GetImages returns all images inside the path defined, including inner directories.
func (repo *LocalRepo) GetImages() []domain.Image {
	var imagePath string

	// Convert relative path to absolute path
	if !path.IsAbs(repo.Path) {
		workingDir, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}

		imagePath = path.Join(workingDir, repo.Path)
	} else {
		imagePath = repo.Path
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
				image := domain.Image{
					FilePath: filePath,
					Metadata: domain.ImageMetadata{
						ImageName: file.Name(),
						Size:      file.Size(),
					},
				}
				images = append(images, image)
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
