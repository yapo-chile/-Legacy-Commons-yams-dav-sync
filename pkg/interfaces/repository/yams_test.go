package repository

import (
	"fmt"
	"io/ioutil"
	"path"
	"testing"
	"time"

	"github.schibsted.io/Yapo/yams-dav-sync/pkg/domain"
	infra "github.schibsted.io/Yapo/yams-dav-sync/pkg/infrastructure"
)

// TODO: FIX ALL OF THIS.
func TestPutImages(t *testing.T) {

	jwtSigner := infra.NewJWTSigner("../../config/private.key")

	yamsRepo := NewYamsRepository(jwtSigner, "mgmt-us-east-1-yams.schibsted.com", "17c82c157c50a0c4",
		"e5ce1008-0145-4b91-9670-390db782ed9c", "fa5881b0-3092-4c80-b37b-0ab08519951f", "71464924-d0a4-41c9-b9c9-09d3c8a9456b")
	yamsRepo.Debug = true

	testImage := domain.Image{
		FilePath: "../../5768222092.jpg",
		Metadata: domain.ImageMetadata{
			ImageName: "5768222092.jpg",
		},
	}

	err := yamsRepo.PutImage(testImage)
	if err != nil {
		t.Errorf("PutImage failed: %s", err.ErrorString)
	}

}

func TestDeleteImages(t *testing.T) {

	jwtSigner := infra.NewJWTSigner("../../config/private.key")

	yamsRepo := NewYamsRepository(jwtSigner, "mgmt-us-east-1-yams.schibsted.com", "17c82c157c50a0c4",
		"e5ce1008-0145-4b91-9670-390db782ed9c", "fa5881b0-3092-4c80-b37b-0ab08519951f", "71464924-d0a4-41c9-b9c9-09d3c8a9456b")
	yamsRepo.Debug = true

	err := yamsRepo.DeleteImage("5768222093.jpg", true)
	if err != nil {
		t.Errorf("PutImage failed: %s", err.ErrorString)
	}

}

func TestHeadImages(t *testing.T) {

	jwtSigner := infra.NewJWTSigner("../../config/private.key")

	yamsRepo := NewYamsRepository(jwtSigner, "mgmt-us-east-1-yams.schibsted.com", "17c82c157c50a0c4",
		"e5ce1008-0145-4b91-9670-390db782ed9c", "fa5881b0-3092-4c80-b37b-0ab08519951f", "71464924-d0a4-41c9-b9c9-09d3c8a9456b")
	yamsRepo.Debug = true

	err := yamsRepo.HeadImage("5768222094.jpg")
	if err != nil {
		t.Errorf("PutImage failed: %s", err.ErrorString)
	}

}

func TestMultiPutImage(t *testing.T) {
	jwtSigner := infra.NewJWTSigner("../../config/private.key")

	yamsRepo := NewYamsRepository(jwtSigner, "mgmt-us-east-1-yams.schibsted.com", "17c82c157c50a0c4",
		"e5ce1008-0145-4b91-9670-390db782ed9c", "fa5881b0-3092-4c80-b37b-0ab08519951f", "71464924-d0a4-41c9-b9c9-09d3c8a9456b")
	yamsRepo.Debug = false

	dirPath := "/Users/maurodelucca/YapoImages"
	fileInfo, _ := ioutil.ReadDir(dirPath)

	c := make(chan bool)

	logFunc := func(imgName, msg string) {
		fmt.Printf("[%d - %s] %s\n", time.Now().UnixNano(), imgName, msg)
	}

	putImageFunc := func(image domain.Image) {
		logFunc(image.Metadata.ImageName, "Starting Put")
		err := yamsRepo.PutImage(image)
		if err != nil {
			logFunc(image.Metadata.ImageName, "Error: "+err.ErrorString)
		}

		logFunc(image.Metadata.ImageName, "Starting Delete")
		err = yamsRepo.DeleteImage(image.Metadata.ImageName, true)
		if err != nil {
			logFunc(image.Metadata.ImageName, "Error: "+err.ErrorString)
		}

		logFunc(image.Metadata.ImageName, "Done")
		c <- true
	}

	count := 0
	for _, file := range fileInfo {
		if extRegex.MatchString(file.Name()) {
			testImage := domain.Image{
				FilePath: path.Join(dirPath, file.Name()),
				Metadata: domain.ImageMetadata{
					ImageName: file.Name(),
				},
			}

			go putImageFunc(testImage)

			count++
			if count > 100 {
				break
			}
		}
	}

	for i := 0; i < count; i++ {
		// wait all to finish
		<-c
	}
}
