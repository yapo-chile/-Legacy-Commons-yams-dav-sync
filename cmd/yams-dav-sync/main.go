package main

import (
	"fmt"
	"os"

	"github.schibsted.io/Yapo/yams-dav-sync/pkg/interfaces"
	"github.schibsted.io/Yapo/yams-dav-sync/pkg/usecases"
)

func main() {
	config, _ := interfaces.NewConfig()

	opt := os.Args[1]

	yamsRepo := interfaces.YamsRepo{
		MgmtURL:     config.MgmtURL,
		AccessKeyID: config.AccessKeyID,
		TenantID:    config.TenantID,
	}
	syncUC := usecases.SyncUseCase{
		YamsRepo: yamsRepo,
	}
	cliHandler := interfaces.CLIHandler{
		SyncUseCase: syncUC,
	}

	if opt == "sync" {
		cliHandler.Sync()
	} else {
		fmt.Printf("Options available are:\n\t* sync [folder]\n")
	}
}
