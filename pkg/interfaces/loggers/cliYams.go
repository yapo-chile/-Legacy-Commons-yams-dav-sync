package loggers

import (
	"fmt"
	"time"

	"github.schibsted.io/Yapo/yams-dav-sync/pkg/interfaces"
	"github.schibsted.io/Yapo/yams-dav-sync/pkg/usecases"
)

type cliYamsLogger struct {
	logger Logger
}

func (l *cliYamsLogger) LogImage(position int, img usecases.YamsObject) {
	fmt.Printf("\n%v ) Name: %+v  MD5: %+v Size: %+v LasModified: %+v",
		position,
		img.ID,
		img.Md5,
		img.Size,
		img.LastModified,
	)
}

// MakeCLIYamsLogger sets up a cliYamsLogger instrumented via the provided logger
func MakeCLIYamsLogger(logger Logger) interfaces.CLIYamsLogger {
	return &cliYamsLogger{
		logger: logger,
	}
}

func (l *cliYamsLogger) LogErrorCleaningMarks(imgName string, err error) {
	l.logger.Error("Error cleaning error marks for %+v, error: %+v", imgName, err)
}

func (l *cliYamsLogger) LogErrorRemoteDelete(imgName string, err error) {
	l.logger.Error("Error deleting remote image %+v, error: %+v", imgName, err)
}

func (l *cliYamsLogger) LogErrorResetingErrorCounter(imgName string, err error) {
	l.logger.Error("Error reseting error counter for %+v, error: %+v", imgName, err)
}

func (l *cliYamsLogger) LogErrorIncreasingErrorCounter(imgName string, err error) {
	l.logger.Error("Error increasing error counter for %+v, error: %+v", imgName, err)
}

func (l *cliYamsLogger) LogErrorGettingRemoteChecksum(imgName string, err error) {
	l.logger.Error("Error getting checksum for %+v, error: %+v", imgName, err)
}

func (l *cliYamsLogger) LogErrorGettingImagesList(listPath string, err error) {
	l.logger.Error("Error getting images list in path %+v, error: %+v", listPath, err)
}

func (l *cliYamsLogger) LogErrorSettingSyncMark(mark time.Time, err error) {
	l.logger.Error("Error setting synchronization mark %+v error: %+v", mark, err)
}

func (l *cliYamsLogger) LogRetryPreviousFailedUploads() {
	l.logger.Info("Retrying to upload previous failed uploads...")
}

func (l *cliYamsLogger) LogReadingNewImages() {
	l.logger.Info("Reading new images from dump file...")
}

func (l *cliYamsLogger) LogUploadingNewImages() {
	l.logger.Info("Uploading new images to yams...")
}

func (l *cliYamsLogger) LogStats(timer int, stats *interfaces.Stats) {
	sent := <-stats.Sent
	errors := <-stats.Errors
	processed := <-stats.Processed
	duplicated := <-stats.Duplicated
	skipped := <-stats.Skipped
	notFound := <-stats.NotFound

	stats.Sent <- sent
	stats.Errors <- errors
	stats.Duplicated <- duplicated
	stats.Processed <- processed
	stats.Skipped <- skipped
	stats.NotFound <- notFound

	fmt.Printf("\r[ Timer: %ds ] ( \033[32mSent images: %d \033[0m "+
		"\033[31m Errors: %d \033[0m "+
		"\033[31m Duplicated: %d \033[0m "+
		"\033[33m Processed: %d \033[0m "+
		"\033[33m Skipped: %d \033[0m "+
		"\033[33m Not Found: %d \033[0m ) ", timer,
		sent, errors, duplicated, processed, skipped, notFound)
}
