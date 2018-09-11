package interfaces

import (
	"testing"

	"github.schibsted.io/Yapo/yams-dav-sync/pkg/domain"
	infra "github.schibsted.io/Yapo/yams-dav-sync/pkg/infrastructure"
)

func TestPutImages(t *testing.T) {

	jwtSigner := infra.NewJWTSigner("../../config/private.key")

	yamsRepo := YamsRepo{
		JWTSigner:   jwtSigner,
		MgmtURL:     "mgmt-us-east-1-yams.schibsted.com",
		AccessKeyID: "17c82c157c50a0c4",
		TenantID:    "e5ce1008-0145-4b91-9670-390db782ed9c",
		DomainID:    "fa5881b0-3092-4c80-b37b-0ab08519951f",
	}

	testImage := domain.Image{
		FilePath: "../../5768222092.jpg",
		Metadata: domain.ImageMetadata{
			ImageName: "5768222092.jpg",
		},
	}

	yamsRepo.PutImage("465a8c49-cdb8-4fe4-8ce2-204860780391", testImage)

}
