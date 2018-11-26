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
	err := handler.Interactor.Run(limit, images)
	if err != nil {
		fmt.Printf("\n Error:  %+v", err)
	}

	fmt.Printf("\n Done")

	defer wg.Done()

}

// ConcurrentSync synchronizes images between local repository and yams repository
// using go concurrency
func (handler *CLIYams) ConcurrentSync(limit, threads int) error {
	images := handler.Interactor.LocalRepo.GetImages()
	for i := 0; i < threads; i++ {
		wg.Add(1)
		// TODO: Improved
		go handler.goSync(limit/threads, images[(i*threads):(i*threads+threads)])
	}
	wg.Wait()

	return nil
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
