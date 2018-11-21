package main

import (
	"flag"
	"fmt"
	"os"

	"github.schibsted.io/Yapo/yams-dav-sync/pkg/infrastructure"
	"github.schibsted.io/Yapo/yams-dav-sync/pkg/interfaces"
	"github.schibsted.io/Yapo/yams-dav-sync/pkg/interfaces/loggers"
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

	opt := flag.String("command", "list", "command to execute syncher script")
	flag.Parse()

	// Setting up insfrastructure
	HTTPHandler := infrastructure.NewHTTPHandler()

	signer := infrastructure.NewJWTSigner(conf.YamsConf.PrivateKeyFile)

	redisHandler := infrastructure.NewRedisHandler(conf.Redis.Address, logger)
	yamsRepo := repository.NewYamsRepository(
		signer,
		conf.YamsConf.MgmtURL,
		conf.YamsConf.AccessKeyID,
		conf.YamsConf.TenantID,
		conf.YamsConf.DomainID,
		conf.YamsConf.BucketID,
		loggers.MakeYamsRepoLogger(logger),
		HTTPHandler,
	)

	imageStatusRepo := repository.NewImageStatusRepository(redisHandler, "", 0)
	localRepo := repository.NewLocalRepo(
		conf.LocalStorageConf.Path,
		logger,
	)

	syncInteractor := usecases.SyncInteractor{
		YamsRepo:        yamsRepo,
		LocalRepo:       localRepo,
		ImageStatusRepo: imageStatusRepo,
		Logger:          loggers.MakeSyncLogger(logger),
	}
	CLIYams := interfaces.CLIYams{
		Interactor: syncInteractor,
		Logger:     loggers.MakeCLIYamsLogger(logger),
	}

	switch *opt {
	case "sync":
		CLIYams.Sync()
	case "list":
		CLIYams.List()
	case "deleteAll":
		CLIYams.DeleteAll()
	default:
		fmt.Printf("Make start command=[commmand]\nCommand list:\n- sync \n- list\n- deleteAll\n")

	}
}
