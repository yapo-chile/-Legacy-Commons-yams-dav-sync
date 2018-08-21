package usecases

import (
	"errors"
	"fmt"

	"github.schibsted.io/Yapo/yams-dav-sync/pkg/domain"
)

type SyncUseCase struct {
	YamsRepo     YamsRepository
	LocalRepo    LocalImageRepository
	YamsImageDir LocalYamsImageDirectory
}

type YamsRepository interface {
	PutImage(domain.Image) error
}

type LocalImageRepository interface {
	GetImages() []domain.Image
}

type LocalYamsImageDirectory interface {
	StoreImageMetadata(metadata domain.ImageMetadata) error
	GetImageMetadata(imageName string) (domain.ImageMetadata, error)
}

var errImageNotFound = errors.New("Image Not Found")

func (uc *SyncUseCase) Run() error {
	fmt.Println("Sync Executed")
	images := uc.LocalRepo.GetImages()
	for _, image := range images {
		yamsMetadata, err := uc.YamsImageDir.GetImageMetadata(image.Metadata.ImageName)
		if err != null {
			switch err {
			case errImageNotFound:
			default:

			}
		}
		uc.YamsRepo.PutImage(image)
	}

	//TODO: Consider case when image is in Yams' directory but not in local folder.
	return nil
}

func (uc *SyncUseCase) compareMetadata(domain.Image) {

}

func areEqual(a, b domain.ImageMetadata) bool {
	if a.Size == b.Size && a.ModTime == b.ModTime {
		return true
	} else {
		return false
	}
}
