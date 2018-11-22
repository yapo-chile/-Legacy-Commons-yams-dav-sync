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

// MakeSyncLogger sets up a SyncLogger instrumented via the provided logger
func MakeSyncLogger(logger Logger) usecases.SyncLogger {
	return &syncLogger{
		logger: logger,
	}
}
