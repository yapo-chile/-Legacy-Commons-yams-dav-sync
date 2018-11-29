package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.schibsted.io/Yapo/yams-dav-sync/pkg/infrastructure"
	"github.schibsted.io/Yapo/yams-dav-sync/pkg/interfaces"
	"github.schibsted.io/Yapo/yams-dav-sync/pkg/interfaces/loggers"
	"github.schibsted.io/Yapo/yams-dav-sync/pkg/interfaces/repository"
	"github.schibsted.io/Yapo/yams-dav-sync/pkg/usecases"
)

// elapsed estimated execution processing time since a defer elapsed is placed
func elapsed(process string) func() {
	start := time.Now()
	return func() {
		fmt.Printf("%s took %v\n", process, time.Since(start))
	}
}

func main() {
	defer elapsed("exec")()

	var conf infrastructure.Config
	infrastructure.LoadFromEnv(&conf)

	// Setting up logger
	logger, err := infrastructure.MakeYapoLogger(&conf.LoggerConf)
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}

	opt := flag.String("command", "list", "command to execute syncher script")
	limitStr := flag.String("limit", "100", "images quantity limit to be synchronized with yams")
	threadsStr := flag.String("threads", "5", "threads limit to be synchronized with yams")

	object := flag.String("object", "", "image name to be deleted in yams")
	flag.Parse()

	limit, _ := strconv.Atoi(*limitStr)
	threads, _ := strconv.Atoi(*threadsStr)
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

	imageStatusRepo := repository.NewImageStatusRepo(redisHandler, "", 0)
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
		if limit > 0 && threads > 0 {
			CLIYams.Sync(limit, threads)
		} else {
			fmt.Println("make start command=sync threads=[number] limit=[number]")
		}
	case "list":
		CLIYams.List()
	case "deleteAll":
		if threads > 0 {
			CLIYams.DeleteAll(threads)
		} else {
			fmt.Println("make start command=deleteAll threads=[number]")
		}
	case "delete":
		CLIYams.Delete(*object)
	default:
		fmt.Printf("Make start command=[commmand]\nCommand list:\n- sync \n- list\n- deleteAll\n")
	}
}
