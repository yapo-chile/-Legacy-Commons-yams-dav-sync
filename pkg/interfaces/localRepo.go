package interfaces

import (
	"io/ioutil"
	"log"
	"path"

	"github.schibsted.io/Yapo/yams-dav-sync/pkg/domain"
)

type LocalRepo struct {
	Path   string
	Logger Logger
}

func (repo *LocalRepo) GetImages() []domain.Image {
	files, err := ioutil.ReadDir(".")
	if err != nil {
		log.Fatal(err)
	}

	var images []domain.Image
	for _, file := range files {
		if !file.IsDir() {
			image := domain.Image{
				FilePath: path.Join("./", file.Name()),
			}
			images = append(images, image)
		}
	}

	return images
}
