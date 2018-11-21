package main

import (
	"fmt"
	"os"

	"github.schibsted.io/Yapo/yams-dav-sync/pkg/infrastructure"
	"github.schibsted.io/Yapo/yams-dav-sync/pkg/interfaces"
	"github.schibsted.io/Yapo/yams-dav-sync/pkg/interfaces/repository"
	"github.schibsted.io/Yapo/yams-dav-sync/pkg/usecases"
)

func main() {

	var conf infrastructure.Config
	infrastructure.LoadFromEnv(&conf)

	// Setting up logger
	logger, err := infrastructure.MakeYapoLogger(&conf.LoggerConf)
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}

	opt := os.Args[1]
	signer := infrastructure.NewJWTSigner(conf.YamsConf.PrivateKeyFile)

	yamsRepo := repository.NewYamsRepository(
		signer,
		conf.YamsConf.MgmtURL,
		conf.YamsConf.AccessKeyID,
		conf.YamsConf.TenantID,
		conf.YamsConf.DomainID,
		conf.YamsConf.BucketID,
		false,
	)

	localRepo := repository.NewLocalRepo(
		conf.LocalStorageConf.Path,
		logger,
	)

	sync := usecases.SyncUseCase{
		YamsRepo:  yamsRepo,
		LocalRepo: localRepo,
	}
	cliHandler := interfaces.CLIHandler{
		SyncUseCase: sync,
	}

	switch opt {
	case "sync":
		cliHandler.Sync()
	case "list":
		cliHandler.List()
	case "deleteAll":
		cliHandler.DeleteAll()
	default:
		fmt.Printf("Make start command=[commmand]\nCommand list:\n- sync \n- list\n- deleteAll")

	}
}
