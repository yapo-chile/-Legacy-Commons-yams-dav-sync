package interfaces

import (
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
	return handler.Interactor.Run(limit)
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
