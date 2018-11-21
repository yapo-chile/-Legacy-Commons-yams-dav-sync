package interfaces

import (
	"fmt"

	"github.schibsted.io/Yapo/yams-dav-sync/pkg/usecases"
)

// CLIHandler is a yams client that executes operation on yams repository
type CLIHandler struct {
	SyncUseCase usecases.SyncUseCase
}

// Sync synchornize images between local repository and yams repository
func (handler *CLIHandler) Sync() error {
	return handler.SyncUseCase.Run()
}

// List prints a list of available images in yams repository
func (handler *CLIHandler) List() error {
	list, err := handler.SyncUseCase.List()
	for i, img := range list {
		fmt.Printf("%v - %+v\n", i+1, img)
	}
	return err
}

// DeleteAll deletes all the objects in yams repository
func (handler *CLIHandler) DeleteAll() error {
	return handler.SyncUseCase.DeleteAll()
}
