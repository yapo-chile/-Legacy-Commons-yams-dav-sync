package interfaces

import (
	"testing"

	"github.schibsted.io/Yapo/yams-dav-sync/pkg/domain"
	infra "github.schibsted.io/Yapo/yams-dav-sync/pkg/infrastructure"
)

func TestPutImages(t *testing.T) {

	jwtSigner := infra.NewJWTSigner("../../config/private.key")

	yamsRepo := NewYamsRepository(jwtSigner, "mgmt-us-east-1-yams.schibsted.com", "17c82c157c50a0c4",
		"e5ce1008-0145-4b91-9670-390db782ed9c", "fa5881b0-3092-4c80-b37b-0ab08519951f")
	yamsRepo.Debug = true

	testImage := domain.Image{
		FilePath: "../../5768222092.jpg",
		Metadata: domain.ImageMetadata{
			ImageName: "5768222092.jpg",
		},
	}

	yamsRepo.PutImage("71464924-d0a4-41c9-b9c9-09d3c8a9456b", testImage)

}

func TestDeleteImages(t *testing.T) {

	jwtSigner := infra.NewJWTSigner("../../config/private.key")

	yamsRepo := NewYamsRepository(jwtSigner, "mgmt-us-east-1-yams.schibsted.com", "17c82c157c50a0c4",
		"e5ce1008-0145-4b91-9670-390db782ed9c", "fa5881b0-3092-4c80-b37b-0ab08519951f")
	yamsRepo.Debug = true

	yamsRepo.DeleteImage("71464924-d0a4-41c9-b9c9-09d3c8a9456b", "5768222092.jpg", false)

}
