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
		os.Exit(2)
	}

	opt := flag.String("command", "list", "command to execute syncher script")
	dumpFile := flag.String("dumpfile", "", "dump file with the list of images to upload")
	threadsStr := flag.String("threads", "5", "threads limit to be synchronized with yams")

	object := flag.String("object", "", "image name to be deleted in yams")
	flag.Parse()

	threads, _ := strconv.Atoi(*threadsStr)
	// Setting up insfrastructure
	HTTPHandler := infrastructure.NewHTTPHandler()

	signer := infrastructure.NewJWTSigner(conf.YamsConf.PrivateKeyFile)

	dbHandler, err := infrastructure.NewPgsqlHandler(conf.Database, logger)
	if err != nil {
		logger.Error("%s\n", err)
		os.Exit(2)
	}

	setUpMigrations(conf, dbHandler, logger)

	yamsRepo := repository.NewYamsRepository(
		signer,
		conf.YamsConf.MgmtURL,
		conf.YamsConf.AccessKeyID,
		conf.YamsConf.TenantID,
		conf.YamsConf.DomainID,
		conf.YamsConf.BucketID,
		loggers.MakeYamsRepoLogger(logger),
		HTTPHandler,
		conf.YamsConf.TimeOut,
		conf.YamsConf.MaxConcurrentConns,
	)

	defaultLastSyncDate, _ := time.Parse("02-01-2006", conf.LastSync.DefaultDate)
	lastSyncRepo := repository.NewLastSyncRepo(dbHandler, defaultLastSyncDate)

	syncErrorRepo := repository.NewErrorControlRepo(
		dbHandler,
		conf.ErrorControl.MaxRetriesPerError,
		conf.ErrorControl.MaxResultsPerPage,
	)

	localRepo := repository.NewLocalRepo(
		conf.LocalStorageConf.Path,
		logger,
	)
	syncInteractor := usecases.SyncInteractor{
		YamsRepo:      yamsRepo,
		LocalRepo:     localRepo,
		LastSyncRepo:  lastSyncRepo,
		SyncErrorRepo: syncErrorRepo,
		Logger:        loggers.MakeSyncLogger(logger),
	}
	CLIYams := interfaces.CLIYams{
		Interactor: syncInteractor,
		Logger:     loggers.MakeCLIYamsLogger(logger),
	}

	maxErrorQty := conf.ErrorControl.MaxRetriesPerError

	switch *opt {
	case "sync":
		if *dumpFile != "" && threads > 0 {
			CLIYams.Sync(threads, maxErrorQty, *dumpFile)
		} else {
			fmt.Println("make start command=sync threads=[number] dump-file=[path]")
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
	version, _, _ := mig.Version()
	logger.Info("Migrations Actual Version %d", version)
	err = mig.Up()
	if err != nil && err != migrate.ErrNoChange {
		logger.Info("Migration message: %v", err)
		return
	}
	version, _, _ = mig.Version()
	logger.Info("Migrations upgraded to version %d", version)
}
