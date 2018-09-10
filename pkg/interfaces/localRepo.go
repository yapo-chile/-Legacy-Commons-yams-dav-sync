package interfaces

import (
	"log"
	"os"
	"path"
	"regexp"

	"github.schibsted.io/Yapo/yams-dav-sync/pkg/domain"
)

type LocalRepo struct {
	Path   string
	Logger Logger
}

func NewLocalRepo(path string, logger Logger) LocalRepo {
	localRepo := LocalRepo{
		Path:   path,
		Logger: logger,
	}
	return localRepo
}

// TODO: Make supported images configurable
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
