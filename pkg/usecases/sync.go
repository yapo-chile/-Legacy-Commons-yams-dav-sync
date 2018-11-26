package usecases

import (
	"errors"
	"fmt"

	"github.schibsted.io/Yapo/yams-dav-sync/pkg/domain"
)

// SyncInteractor executes operations for syncher between local storage and yams bucket
type SyncInteractor struct {
	YamsRepo        YamsRepository
	LocalRepo       LocalImageRepository
	ImageStatusRepo ImageStatusRepository
	Logger          SyncLogger
}

type SyncLogger interface {
	LogSentImage(img domain.Image)
	LogErrorGettingImages(err error)
	LogErrorSendingImage(img domain.Image, err error)
	LogErrorDeletingImage(imgID string, err error)
}

// LocalImageRepository allows local storage operations
type LocalImageRepository interface {
	GetImages() []domain.Image
	GetFileCheckSum(path string) (hash string, err error)
}

var errImageNotFound = errors.New("Image Not Found")

// Run executes the synchronization of images between local storage and yams bucket
func (i *SyncInteractor) Run() error {
	sentImages := 0
	images := i.LocalRepo.GetImages()
	for _, image := range images {
		// Search the image status in database
		registeredHash, _ := i.ImageStatusRepo.GetImageStatus(image.Metadata.ImageName)

		actualHash, err := i.LocalRepo.GetFileCheckSum(image.FilePath)
		if err != nil {
			fmt.Printf("Error getting image from local %+v", err)

			// if error reading the file, continue next image
			break
		}
		fmt.Printf("\nProccessing: %+v", image.Metadata.ImageName)
		// if the actual image md5 hash does not match with the registered md5 hash
		if actualHash != registeredHash {
			fmt.Printf("\n... [Redis]: UPDATE IT!\n\n")
			// try to synchronize with yams
			i.Logger.LogSentImage(image)
			if err := i.YamsRepo.PutImage(image); err != nil {
				switch err {
				case ErrYamsDuplicate:
					fmt.Printf("\n...[YAMS]: Duplicated \n")
					externalHash, _ := i.YamsRepo.HeadImage(image.Metadata.ImageName)
					//fmt.Printf("externalHash: %+v == %+v actualHash", externalHash, actualHash)
					// Same name, different content
					if externalHash != actualHash {
						// replace image in bucket, because the name is the same but the content is different
						fmt.Printf("\n..Forced update\n\n")
						fmt.Printf("\n........deleting\n\n")

						e := i.YamsRepo.DeleteImage(image.Metadata.ImageName, true)
						if e != nil {
							fmt.Printf("delete error: %+v", e)
							break
						}
						e2 := i.ImageStatusRepo.DelImageStatus(image.Metadata.ImageName)
						if e2 != nil {
							fmt.Printf("delete error: %+v", e)
							break
						}
						fmt.Printf("\n[REDIS]............deleted\n\n")

					} else {
						fmt.Printf("\n[Yams]...I already have this (MD5 validated)\n\n")
						fmt.Printf("\n[Redis]...Marking as sent\n\n")
						i.ImageStatusRepo.SetImageStatus(image.Metadata.ImageName, actualHash)
					}

				default:
					// with another kind of errors pass over the image
				}
				// continue with next image
			} else {
				// if the synchronization works then register in image status repository as sent
				i.ImageStatusRepo.SetImageStatus(image.Metadata.ImageName, actualHash)
				sentImages++
			}
		} else {
			fmt.Printf("\n...[Redis] Image already synchronized (pass over - MD5 validated)\n\n")
		}

		// TODO: Make it smarter!
		limitPerExecution := 25
		if sentImages == limitPerExecution {
			return nil
		}
	}
	// Consider case when image is in Yams' directory but not in local folder.
	return nil
}

// List get a list of available images in yams bucket
func (i *SyncInteractor) List() ([]YamsObject, error) {
	return i.YamsRepo.GetImages()
}

// DeleteAll deletes all the images of yams bucket
func (i *SyncInteractor) DeleteAll() error {
	yamsResponse, err := i.YamsRepo.GetImages()
	if err != nil {
		i.Logger.LogErrorGettingImages(err)
		return err
	}
	for _, img := range yamsResponse {
		if err := i.YamsRepo.DeleteImage(img.ID, true); err != nil {
			i.Logger.LogErrorDeletingImage(img.ID, err)
			return err
		}
	}
	return nil
}
