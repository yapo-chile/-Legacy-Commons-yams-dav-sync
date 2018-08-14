package interfaces

import "github.schibsted.io/Yapo/yams-dav-sync/pkg/usecases"

type CLIHandler struct {
	SyncUseCase usecases.SyncUseCase
}

func (handler *CLIHandler) Sync() error {
	handler.SyncUseCase.Run()
	return nil
}
