package interfaces

import (
	"sync"

	"github.schibsted.io/Yapo/yams-dav-sync/pkg/domain"
	"github.schibsted.io/Yapo/yams-dav-sync/pkg/usecases"
)

// CLIYams is a yams client that executes operation on yams repository
type CLIYams struct {
	Interactor usecases.SyncInteractor
	Logger     CLIYamsLogger
}

// CLIYamsLogger logs CLI yams events
type CLIYamsLogger interface {
	LogImage(int, usecases.YamsObject)
}

// getImages gets images from local repository. The images are validated by
// image status repository to be uploaded to yams repository
func (handler *CLIYams) getImages(limit int) []domain.Image {
	localFiles := handler.Interactor.LocalRepo.GetImages()
	images := []domain.Image{}
	imagesToProcess := 0
	for i := range localFiles {
		md5Checksume, _ := handler.Interactor.ImageStatusRepo.GetImageStatus(
			localFiles[i].Metadata.ImageName,
		)
		if md5Checksume != localFiles[i].Metadata.Checksum {
			images = append(images, localFiles[i])
			imagesToProcess++
		}
		if imagesToProcess >= limit {
			return images
		}
	}
	return images
}

// Sync synchronizes images between local repository and yams repository
// using go concurrency
func (handler *CLIYams) Sync(limit, threads int) error {
	if threads > limit {
		threads = limit
	}

	images := handler.getImages(limit)

	if limit > len(images) {
		limit = len(images)
	}

	jobs := make(chan domain.Image)
	var waitGroup sync.WaitGroup

	for w := 0; w < threads; w++ {
		go handler.sendWorker(w, jobs, &waitGroup)
	}

	for _, image := range images {
		jobs <- image
	}

	close(jobs)
	waitGroup.Wait()

	return nil
}

// sendWorker sends every image to yams repository
func (handler *CLIYams) sendWorker(id int, jobs <-chan domain.Image, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()
	for image := range jobs {
		validChecksum := handler.Interactor.ValidateChecksum(image)
		if !validChecksum {
			handler.Interactor.Send(image)
		}
	}
}

// List prints a list of available images in yams repository
func (handler *CLIYams) List() error {
	list, err := handler.Interactor.List()
	for i, img := range list {
		handler.Logger.LogImage(i+1, img)
	}
	return err
}

// Delete deletes an object in yams repository
func (handler *CLIYams) Delete(imageName string) error {
	return handler.Interactor.Delete(imageName)
}

// DeleteAll deletes every imagen in yams repository and redis using concurency
func (handler *CLIYams) DeleteAll(threads int) error {
	images, _ := handler.Interactor.YamsRepo.GetImages()

	jobs := make(chan string)
	var waitGroup sync.WaitGroup

	for w := 0; w < threads; w++ {
		go handler.deleteWorker(w, jobs, &waitGroup)
	}

	for _, image := range images {
		jobs <- image.ID
	}

	close(jobs)
	waitGroup.Wait()

	return nil
}

// deleteWorker deletes every image to yams repository
func (handler *CLIYams) deleteWorker(id int, jobs <-chan string, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()
	for j := range jobs {
		handler.Interactor.Delete(j)
		handler.Interactor.ImageStatusRepo.DelImageStatus(j)
	}
}
