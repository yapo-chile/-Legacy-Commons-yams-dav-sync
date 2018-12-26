package loggers

import (
	"fmt"

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
