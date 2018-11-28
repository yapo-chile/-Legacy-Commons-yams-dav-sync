package usecases

import (
	"errors"

	"github.schibsted.io/Yapo/yams-dav-sync/pkg/domain"
)

// SyncInteractor executes operations for syncher between local storage and yams bucket
type SyncInteractor struct {
	YamsRepo        YamsRepository
	LocalRepo       LocalImageRepository
	ImageStatusRepo ImageStatusRepository
	Logger          SyncLogger
}

// SyncLogger logs synchronization events
type SyncLogger interface {
	LogProcessImage(img domain.Image, sent, skipped, proccessed int)
	LogUploadingImage(img domain.Image)
	ErrorDuplicatedImage(img domain.Image)
	ErrorDeletingImageInYams(imgID string, e error)
	ErrorDeletingImageStatusInRepo(imgID string, e error)
	ImageSuccessfullyDelete(img domain.Image)
	MarkingAsSynchronized(img domain.Image)
	PassingOver(img domain.Image)
	LogErrorGettingImages(err error)
	LogErrorSendingImage(img domain.Image, err error)
}

// LocalImageRepository allows local storage operations
type LocalImageRepository interface {
	GetImages() []domain.Image
}

var errImageNotFound = errors.New("Image Not Found")

// Run executes the synchronization of images between local storage and yams bucket
func (i *SyncInteractor) Run(limitPerExecution int, images []domain.Image) error {
	sentImages := 0
	processedImages := 0
	skippedImages := 0
	for _, image := range images {
		// Search the image status in database. Error check not being apply
		// because registered checksum won't match with actual checksum
		registeredHash, _ := i.ImageStatusRepo.GetImageStatus(image.Metadata.ImageName)
		i.Logger.LogProcessImage(image, sentImages, skippedImages, processedImages)
		//	fmt.Printf("\n Processing %+v", image)
		processedImages++

		// if the actual image md5 checksum does not match with the registered checksum
		if image.Metadata.Checksum != registeredHash {
			i.Logger.LogUploadingImage(image)
			// try to synchronize with yams
			if err := i.YamsRepo.PutImage(image); err != nil {
				switch err {
				case ErrYamsDuplicate:
					i.Logger.ErrorDuplicatedImage(image)
					externalHash, _ := i.YamsRepo.HeadImage(image.Metadata.ImageName)
					// Check if the error is only because the name or the name and content
					if externalHash != image.Metadata.Checksum {
						// if the image content is different force to update the image
						// TODO: force to update method.
						// Actually yams does not have a method to force an update
						// so the current solution is delete the image from yams
						// and delete the image status from redis and the next execution of
						// the script will upload the image
						if e := i.YamsRepo.DeleteImage(image.Metadata.ImageName, true); e != nil {
							i.Logger.ErrorDeletingImageInYams(image.Metadata.ImageName, e)
							// anyways try to delete from imageStatusRepos, next execution it will try
							// again
						}
						if e := i.ImageStatusRepo.DelImageStatus(image.Metadata.ImageName); e != nil {
							i.Logger.ErrorDeletingImageStatusInRepo(image.Metadata.ImageName, e)
						}
						i.Logger.ImageSuccessfullyDelete(image)

					} else {
						// the image already synchronized but not marked by redis
						i.Logger.MarkingAsSynchronized(image)
						i.ImageStatusRepo.SetImageStatus(image.Metadata.ImageName, image.Metadata.Checksum)
					}

				default:
					continue
					// with another kind of errors pass over the image
				}
				// continue with next image
			}
			// if the synchronization works then register in image status repository as sent
			i.ImageStatusRepo.SetImageStatus(image.Metadata.ImageName, image.Metadata.Checksum)
			sentImages++

		} else {
			// already marked in redis as synchronized
			i.Logger.PassingOver(image)
			skippedImages++
		}
		if sentImages == limitPerExecution {
			return nil
		}
	}
	// TODO: Consider case when image is in Yams' directory but not in local folder.
	return nil
}

// Send executes the synchronization of images between local storage and yams bucket
func (i *SyncInteractor) Send(image domain.Image) error {
	// Search the image status in database. Error check not being apply
	// because registered checksum won't match with actual checksum
	registeredHash, _ := i.ImageStatusRepo.GetImageStatus(image.Metadata.ImageName)
	// if the actual image md5 checksum does not match with the registered checksum
	if image.Metadata.Checksum != registeredHash {
		i.Logger.LogUploadingImage(image)
		// try to synchronize with yams
		if err := i.YamsRepo.PutImage(image); err != nil {
			switch err {
			case ErrYamsDuplicate:
				i.Logger.ErrorDuplicatedImage(image)
				externalHash, _ := i.YamsRepo.HeadImage(image.Metadata.ImageName)
				// Check if the error is only because the name or the name and content
				if externalHash != image.Metadata.Checksum {
					if e := i.YamsRepo.DeleteImage(image.Metadata.ImageName, true); e != nil {
						i.Logger.ErrorDeletingImageInYams(image.Metadata.ImageName, e)
					}
					if e := i.ImageStatusRepo.DelImageStatus(image.Metadata.ImageName); e != nil {
						i.Logger.ErrorDeletingImageStatusInRepo(image.Metadata.ImageName, e)
					}
					i.Logger.ImageSuccessfullyDelete(image)
				} else {
					// the image already synchronized but not marked by redis
					i.Logger.MarkingAsSynchronized(image)
					i.ImageStatusRepo.SetImageStatus(image.Metadata.ImageName, image.Metadata.Checksum)
				}

			default:
				// with another kind of errors pass over the image
			}
			// continue with next image
		}
		// if the synchronization works then register in image status repository as sent
		i.ImageStatusRepo.SetImageStatus(image.Metadata.ImageName, image.Metadata.Checksum)

	} else {
		// already marked in redis as synchronized
		i.Logger.PassingOver(image)
	}

	// TODO: Consider case when image is in Yams' directory but not in local folder.
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
			i.Logger.ErrorDeletingImageInYams(img.ID, err)
			return err
		}
		i.ImageStatusRepo.DelImageStatus(img.ID)
	}
	return nil
}

// Delete deletes an the images of yams bucket
func (i *SyncInteractor) Delete(imageName string) error {
	if err := i.YamsRepo.DeleteImage(imageName, true); err != nil {
		i.Logger.ErrorDeletingImageInYams(imageName, err)
		return err
	}
	return nil
}
