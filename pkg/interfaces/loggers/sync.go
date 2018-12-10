package loggers

import (
	"github.schibsted.io/Yapo/yams-dav-sync/pkg/domain"
	"github.schibsted.io/Yapo/yams-dav-sync/pkg/usecases"
)

type syncLogger struct {
	logger Logger
}

func (l *syncLogger) LogSentImage(img domain.Image) {
	l.logger.Info("< Sending object to Yams [Name: %s, Size: %+v, Time: %+v]",
		img.Metadata.ImageName,
		img.Metadata.Size,
		img.Metadata.ModTime,
	)
}

func (l *syncLogger) LogErrorGettingImages(err error) {
	l.logger.Error("> Error getting images from yams: %+v", err)
}

func (l *syncLogger) LogErrorSendingImage(img domain.Image, err error) {
	l.logger.Info("< Error sending image %+v to yams. Error: %+v", img.Metadata.ImageName, err)
}

func (l *syncLogger) LogErrorDeletingImage(img string, err error) {
	l.logger.Error("< Error deleting image %+v in Yams, error: %+v", img, err)
}

func (l *syncLogger) LogProcessImage(img domain.Image, sent, skipped, proccesed int) {
	l.logger.Info("Processing (%d/%d/%d): %+v", sent, skipped, proccesed, img.Metadata.ImageName)
}

func (l *syncLogger) LogUploadingImage(img domain.Image) {
	l.logger.Info("Uploading image: %+v", img.Metadata.ImageName)
}

func (l *syncLogger) ErrorDuplicatedImage(img domain.Image) {
	l.logger.Error("Error: Image duplicated: %+v", img.Metadata.ImageName)
}

func (l *syncLogger) ErrorDeletingImageInYams(imgID string, e error) {
	l.logger.Error("Error: Deleting image %+v in yams: %+v", imgID, e)
}
func (l *syncLogger) ErrorDeletingLastSyncInRepo(imgID string, e error) {
	l.logger.Error("Error: Deleting image status %+v in REDIS: %+v", imgID, e)
}

func (l *syncLogger) ImageSuccessfullyDelete(img domain.Image) {
	l.logger.Info("- Deleted: %+v ", img.Metadata.ImageName)
}

func (l *syncLogger) MarkingAsSynchronized(img domain.Image) {
	l.logger.Info("+ Synchronized: %+v ", img.Metadata.ImageName)
}

func (l *syncLogger) PassingOver(img domain.Image) {
	l.logger.Info("~ Skipped: %+v  ", img.Metadata.ImageName)
}

// MakeSyncLogger sets up a SyncLogger instrumented via the provided logger
func MakeSyncLogger(logger Logger) usecases.SyncLogger {
	return &syncLogger{
		logger: logger,
	}
}
