package loggers

import (
	"github.schibsted.io/Yapo/yams-dav-sync/pkg/interfaces"
	"github.schibsted.io/Yapo/yams-dav-sync/pkg/usecases"
)

type cliYamsLogger struct {
	logger Logger
}

func (l *cliYamsLogger) LogImage(position int, img usecases.YamsObject) {
	l.logger.Info("%v ) Name: %+v  MD5: %+v Size: %+v LasModified: %+v",
		position,
		img.ID,
		img.Md5,
		img.Size,
		img.LastModified)
}

// MakeCLIYamsLogger sets up a cliYamsLogger instrumented via the provided logger
func MakeCLIYamsLogger(logger Logger) interfaces.CLIYamsLogger {
	return &cliYamsLogger{
		logger: logger,
	}
}
