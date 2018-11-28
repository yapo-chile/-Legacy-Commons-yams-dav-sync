package interfaces

import (
	"fmt"
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

// Sync synchronizes images between local repository and yams repository
func (handler *CLIYams) Sync(limit int) error {
	images := handler.Interactor.LocalRepo.GetImages()
	return handler.Interactor.Run(limit, images)
}

var wg sync.WaitGroup

func (handler *CLIYams) goSync(limit int, images []domain.Image) {
	runner := handler.Interactor
	err := runner.Run(limit, images)
	// TODO: better error control
	if err != nil {
		fmt.Printf("\n Error:  %+v", err)
	}
	fmt.Printf("\n Done\n")

	defer wg.Done()

}

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

// ConcurrentSync synchronizes images between local repository and yams repository
// using go concurrency
func (handler *CLIYams) ConcurrentSync(limit, threads int) error {
	images := handler.getImages(limit)
	if threads > limit {
		return fmt.Errorf("limit can't be lower than threads quantity")
	}
	if limit > len(images) {
		limit = len(images)
	}
	interval := limit / threads
	offset := limit - (interval * threads)
	min, max, increment := 0, interval, 0

	for i := 0; i < threads; i++ {
		wg.Add(1)
		go handler.goSync(interval+increment, images[min:max])
		fmt.Println(min, max)
		offset, increment = offsetDistribution(offset)
		min = max
		max = max + interval + increment

	}
	wg.Wait()

	return nil
}

func offsetDistribution(offset int) (newOffset int, increment int) {
	if offset > 0 {
		return offset - 1, 1
	}
	return offset, 0

}

// List prints a list of available images in yams repository
func (handler *CLIYams) List() error {
	list, err := handler.Interactor.List()
	for i, img := range list {
		handler.Logger.LogImage(i+1, img)
	}
	return err
}

// DeleteAll deletes all the objects in yams repository
func (handler *CLIYams) DeleteAll() error {
	return handler.Interactor.DeleteAll()
}

// Delete deletes an object in yams repository
func (handler *CLIYams) Delete(imageName string) error {
	return handler.Interactor.Delete(imageName)
}
