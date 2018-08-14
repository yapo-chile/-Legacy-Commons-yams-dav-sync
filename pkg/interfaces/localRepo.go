package interfaces

import "github.schibsted.io/Yapo/yams-dav-sync/pkg/domain"

type LocalRepo struct {
	Path string
}

func (repo *LocalRepo) GetImages() []domain.Image {
	return nil
}
