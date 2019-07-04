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

	"github.mpi-internal.com/Yapo/yams-dav-sync/pkg/domain"
	"github.mpi-internal.com/Yapo/yams-dav-sync/pkg/infrastructure"
	"github.mpi-internal.com/Yapo/yams-dav-sync/pkg/interfaces"
	"github.mpi-internal.com/Yapo/yams-dav-sync/pkg/interfaces/loggers"
	"github.mpi-internal.com/Yapo/yams-dav-sync/pkg/interfaces/repository"
)

// elapsed estimated execution processing time since a defer elapsed is placed
func elapsed(process string) func() {
	start := time.Now()
	return func() {
		fmt.Printf("%s took %v\n", process, time.Since(start))
	}
}

func main() { // nolint: gocyclo
	defer elapsed("exec")()

	shutdownSequence := infrastructure.NewShutdownSequence()

	shutdownSequence.Listen()

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
	threadsStr := flag.String("threads", "5", "threads limit to make sync with yams")
	limitStr := flag.String("limit", "0", "images qty. limit to upload to yams")
	totalStr := flag.String("total", "0", "images qty. total to upload to yams")

	object := flag.String("object", "", "image name to be deleted in yams")
	flag.Parse()

	threads, e := strconv.Atoi(*threadsStr)
	if e != nil {
		logger.Error("Error: %+v. Threads set as %+v", e, threads)
	}
	limit, e := strconv.Atoi(*limitStr)
	if e != nil {
		logger.Error("Error: %+v. Limit set as %+v", e, limit)
	}
	total, e := strconv.Atoi(*totalStr)
	if e != nil {
		logger.Error("Error: %+v. total set as %+v", e, total)
	}
	// Setting up insfrastructure

	dialer, err := infrastructure.NewProxyDialerHandler(
		conf.BandwidthProxyConf.ConnType,
		conf.BandwidthProxyConf.Host,
	)
	if err != nil {
		logger.Error("%s\n", err)
		os.Exit(2)
	}

	// Metrics exporter
	prometheus := infrastructure.NewPrometheusExporter(conf.MetricsConf.Port)

	// Set the first metric: Total of images to send using the syncher
	prometheus.SetGauge(domain.TotalImages, float64(total))

	circuitBreaker := infrastructure.NewCircuitBreaker(
		conf.CircuitBreakerConf.Name,
		conf.CircuitBreakerConf.ConsecutiveFailure,
		conf.CircuitBreakerConf.FailureRatio,
		conf.CircuitBreakerConf.Timeout,
		conf.CircuitBreakerConf.Interval,
		logger,
	)

	HTTPHandler := infrastructure.NewHTTPHandler(dialer, circuitBreaker, logger)

	signer := infrastructure.NewJWTSigner(conf.YamsConf.PrivateKeyFile, logger)

	dbHandler, err := infrastructure.NewPgsqlHandler(conf.Database, logger)
	if err != nil {
		logger.Error("%s\n", err)
		os.Exit(2)
	}

	shutdownSequence.Push(dbHandler)

	setUpMigrations(conf, dbHandler, logger)

	localImageRepo := repository.NewLocalImageRepo(
		conf.LocalStorageConf.Path,
		infrastructure.NewLocalFileSystemView(logger),
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
		conf.YamsConf.ErrorControlHeader,
		conf.YamsConf.ErrorControlValue,
		conf.YamsConf.MaxConcurrentConns,
	)

	defaultLastSyncDate, err := time.Parse(
		conf.LastSync.DefaultLayout,
		conf.LastSync.DefaultDate,
	)

	if err != nil {
		fmt.Printf("Wrong date layout %+v for date %+v",
			conf.LastSync.DefaultLayout,
			conf.LastSync.DefaultDate)
		os.Exit(3)
	}

	lastSyncRepo := repository.NewLastSyncRepo(
		dbHandler,
		conf.LocalStorageConf.DefaultFilesDateLayout,
		defaultLastSyncDate,
	)

	errorControlRepo := repository.NewErrorControlRepo(
		dbHandler,
		conf.ErrorControl.MaxResultsPerPage,
	)

	cliYams := interfaces.NewCLIYams(
		yamsRepo,
		errorControlRepo,
		lastSyncRepo,
		localImageRepo,
		loggers.MakeCLIYamsLogger(logger),
		defaultLastSyncDate,
		interfaces.NewStats(prometheus),
		conf.LocalStorageConf.DefaultFilesDateLayout,
	)

	shutdownSequence.Push(cliYams)

	maxErrorTolerance := conf.ErrorControl.MaxRetriesPerError
	go func() {
		switch *opt {
		case "sync":
			if *dumpFile != "" && threads > 0 {
				if e := cliYams.Sync(threads, limit, maxErrorTolerance, *dumpFile); e != nil {
					logger.Error("Error with synchornization: %+v", e)
				}
			} else {
				logger.Error("make start command=sync threads=[number] limit=[limit] dump-file=[path]")
			}

		case "list":
			if e := cliYams.List(limit); e != nil {
				logger.Error("Error listing: %+v", e)
			}

		case "deleteAll":
			if threads > 0 {
				if e := cliYams.DeleteAll(threads, limit); e != nil {
					logger.Error("Error deleting: %+v ", e)
				}
			} else {
				logger.Error("make start command=deleteAll threads=[number]")
			}

		case "delete":
			if e := cliYams.Delete(*object); e != nil {
				logger.Error("Error deleting: %+v", e)
			}

		case "reset":
			if e := cliYams.Reset(); e != nil {
				logger.Error("Error reseting: %+v", e)
			}

		case "marks":
			if e := cliYams.GetMarks(); e != nil {
				logger.Error("Error getting sync marks: %+v", e)
			}

		default:
			logger.Error("Make start command=[commmand]\nCommand list:\n- sync \n- list\n- deleteAll\n")
		}
		shutdownSequence.Done()
	}()

	shutdownSequence.Wait()
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
