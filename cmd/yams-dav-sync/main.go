package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	_ "github.com/lib/pq"
	"github.com/mattes/migrate"
	mpgsql "github.com/mattes/migrate/database/postgres"
	_ "github.com/mattes/migrate/source/file"

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
		os.Exit(1)
	}

	opt := flag.String("command", "list", "command to execute syncher script")
	dumpFile := flag.String("dumpfile", "", "dump file with the list of images to upload")
	threadsStr := flag.String("threads", "5", "threads limit to be synchronized with yams")

	object := flag.String("object", "", "image name to be deleted in yams")
	flag.Parse()

	threads, e := strconv.Atoi(*threadsStr)
	if e != nil {
		logger.Error("Error: %+v. Threads set as %+v", e, threads)
	}
	// Setting up insfrastructure
	HTTPHandler := infrastructure.NewHTTPHandler()

	signer := infrastructure.NewJWTSigner(conf.YamsConf.PrivateKeyFile, logger)

	dbHandler, err := infrastructure.NewPgsqlHandler(conf.Database, logger)
	if err != nil {
		logger.Error("%s\n", err)
		os.Exit(2)
	}

	setUpMigrations(conf, dbHandler, logger)

	localImageRepo := repository.NewLocalImageRepo(
		conf.LocalStorageConf.Path,
		infrastructure.NewLocalFileSystemView(),
	)

	yamsRepo := repository.NewYamsRepository(
		signer,
		conf.YamsConf.MgmtURL,
		conf.YamsConf.AccessKeyID,
		conf.YamsConf.TenantID,
		conf.YamsConf.DomainID,
		conf.YamsConf.BucketID,
		localImageRepo,
		loggers.MakeYamsRepoLogger(logger),
		HTTPHandler,
		conf.YamsConf.TimeOut,
		conf.YamsConf.MaxConcurrentConns,
	)

	defaultLastSyncDate, err := time.Parse(conf.LastSync.DefaultLayout, conf.LastSync.DefaultDate)
	if err != nil {
		fmt.Printf("Wrong date layout %+v for date %+v",
			conf.LastSync.DefaultLayout,
			conf.LastSync.DefaultDate)
		os.Exit(3)
	}
	lastSyncRepo := repository.NewLastSyncRepo(dbHandler, defaultLastSyncDate)

	errorControlRepo := repository.NewErrorControlRepo(
		dbHandler,
		conf.ErrorControl.MaxResultsPerPage,
	)

	syncInteractor := usecases.SyncInteractor{
		YamsRepo:         yamsRepo,
		ImageRepo:        localImageRepo,
		LastSyncRepo:     lastSyncRepo,
		ErrorControlRepo: errorControlRepo,
	}
	CLIYams := interfaces.CLIYams{
		Interactor: syncInteractor,
		Logger:     loggers.MakeCLIYamsLogger(logger),
	}

	maxErrorQty := conf.ErrorControl.MaxRetriesPerError

	switch *opt {
	case "sync":
		if *dumpFile != "" && threads > 0 {
			if e := CLIYams.Sync(threads, maxErrorQty, *dumpFile); e != nil {
				logger.Error("Error with synchornization: %+v", e)
			}
		} else {
			logger.Error("make start command=sync threads=[number] dump-file=[path]")
		}
	case "list":
		if e := CLIYams.List(); e != nil {
			logger.Error("Error listing: %+v", e)
		}
	case "deleteAll":
		if threads > 0 {
			if e := CLIYams.DeleteAll(threads); e != nil {
				logger.Error("Error deleting: %+v ", e)
			}
		} else {
			logger.Error("make start command=deleteAll threads=[number]")
		}
	case "delete":
		if e := CLIYams.Delete(*object); e != nil {
			logger.Error("Error deleting: %+v", e)
		}
	default:
		logger.Error("Make start command=[commmand]\nCommand list:\n- sync \n- list\n- deleteAll\n")
	}
}

// Autoexecute database migrations
func setUpMigrations(conf infrastructure.Config, dbHandler *infrastructure.PgsqlHandler, logger loggers.Logger) {
	driver, err := mpgsql.WithInstance(dbHandler.Conn, &mpgsql.Config{})
	if err != nil {
		logger.Error("Error to instance migrations %v", err)
		return
	}
	mig, err := migrate.NewWithDatabaseInstance(
		"file://"+conf.Database.MgFolder,
		conf.Database.MgDriver,
		driver,
	)
	if err != nil {
		logger.Error("Consume migrations sources err %#v", err)
		return
	}
	version, _, e := mig.Version()
	if e != nil {
		logger.Error("Error getting current migration version: %#v", e)
	}
	logger.Info("Migrations Actual Version %d", version)
	err = mig.Up()
	if err != nil && err != migrate.ErrNoChange {
		logger.Info("Migration message: %v", err)
		return
	}
	version, _, e = mig.Version()
	if e != nil {
		logger.Error("Error getting current migration version: %#v", e)
	}
	logger.Info("Migrations upgraded to version %d", version)
}
