package loggers

import (
	"github.schibsted.io/Yapo/yams-dav-sync/pkg/interfaces/repository"
)

type yamsRepoLogger struct {
	logger Logger
}

func (l *yamsRepoLogger) LogRequestURI(url string) {
	l.logger.Debug("< Request URL for yams: %s", url)
}

func (l *yamsRepoLogger) LogResponse(body string, err error) {
	l.logger.Debug("> Yams body: %+v err: %+v", body, err)
}

func (l *yamsRepoLogger) LogStatus(statusCode int) {
	l.logger.Info("> Status: %+v", statusCode)
}

func (l *yamsRepoLogger) LogCannotDecodeErrorMessage(err error) {
	l.logger.Error("> Error: Can not decode Yams error message: %+v", err)
}

// MakeYamsRepoLogger sets up a SyncLogger instrumented via the provided logger
func MakeYamsRepoLogger(logger Logger) repository.YamsRepositoryLogger {
	return &yamsRepoLogger{
		logger: logger,
	}
}
