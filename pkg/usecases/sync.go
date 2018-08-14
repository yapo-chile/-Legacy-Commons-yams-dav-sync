package usecases

import (
	"fmt"

	"github.schibsted.io/Yapo/yams-dav-sync/pkg/domain"
)

type SyncUseCase struct {
	YamsRepo  YamsRepository
	LocalRepo LocalImageRepository
}

type YamsRepository interface {
	GetDomains() string
	PutImage(domain.Image) error
}

type LocalImageRepository interface {
	GetImages() []domain.Image
}

func (uc *SyncUseCase) Run() error {
	fmt.Println("Sync Executed")
	images := uc.LocalRepo.GetImages()
	for _, image := range images {
		uc.YamsRepo.PutImage(image)
	}
	return nil
}
