package usecases

import (
	"errors"
	"fmt"

	"github.schibsted.io/Yapo/yams-dav-sync/pkg/domain"
)

type SyncUseCase struct {
	YamsRepo  YamsRepository
	LocalRepo LocalImageRepository
}

type YamsRepository interface {
	PutImage(domain.Image) error
}

type LocalImageRepository interface {
	GetImages() []domain.Image
}

var errImageNotFound = errors.New("Image Not Found")

func (uc *SyncUseCase) Run() error {
	fmt.Println("Sync Executed")
	images := uc.LocalRepo.GetImages()
	for _, image := range images {
		uc.YamsRepo.PutImage(image)
	}

	//TODO: Consider case when image is in Yams' directory but not in local folder.
	return nil
}
